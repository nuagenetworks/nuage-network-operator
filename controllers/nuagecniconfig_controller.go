// Copyright 2020 Nokia
// Licensed under the Apache License 2.0.
// SPDX-License-Identifier: Apache-2.0

/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/nuagenetworks/nuage-network-operator/controllers/certs"
	"github.com/nuagenetworks/nuage-network-operator/controllers/names"
	"github.com/nuagenetworks/nuage-network-operator/controllers/network/cni"
	"github.com/nuagenetworks/nuage-network-operator/controllers/network/monitor"
	"github.com/nuagenetworks/nuage-network-operator/controllers/network/vrs"
	"github.com/nuagenetworks/nuage-network-operator/controllers/render"
	"github.com/openshift/api/network"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"

	operatorv1alpha1 "github.com/nuagenetworks/nuage-network-operator/api/v1alpha1"
)

var (
	//ManifestPath is the path to templates directory
	ManifestPath   = "./bindata"
	monitDaemonset = types.NamespacedName{
		Namespace: names.Namespace,
		Name:      names.NuageMonitor,
	}
)

const nuageFinalizer = "finalizer.operator.nuage.io"

// OrchestratorType is for orchestrator type(k8s or ose)
type OrchestratorType string

const (
	//OrchestratorKubernetes if platform is Kubernetes
	OrchestratorKubernetes OrchestratorType = "k8s"
	//OrchestratorOpenShift if platform is OpenShift
	OrchestratorOpenShift OrchestratorType = "ose"
	//OrchestratorNone if platform could not be determined
	OrchestratorNone OrchestratorType = "none"
)

// NuageCNIConfigReconciler reconciles a NuageCNIConfig object
type NuageCNIConfigReconciler struct {
	dclient                    discovery.DiscoveryInterface
	orchestrator               OrchestratorType
	clusterNetworkCIDR         string
	apiServerURL               string
	serviceAccountToken        []byte
	clusterNetworkSubnetLength uint32
	clientset                  kubernetes.Interface
	ClusterServiceNetworkCIDR  string
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=operator.nuage.io,resources=nuagecniconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.nuage.io,resources=nuagecniconfigs/status,verbs=get;update;patch

func (r *NuageCNIConfigReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("nuagecniconfig", req.NamespacedName)

	// your logic here

	log.SetLevel(log.DebugLevel)
	log.Infof("Reconciling NuageCNIConfig")

	// Fetch the Nuage custom resource instance
	instance := &operatorv1alpha1.NuageCNIConfig{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	if err := r.parse(instance); err != nil {
		log.Errorf("failed to parse crd config %v", err)
		return reconcile.Result{}, err
	}

	if r.orchestrator == OrchestratorKubernetes {
		r.setPodNetworkConfig(&instance.Spec.PodNetworkConfig)
	}

	clusterInfo, err := r.GetClusterNetworkInfo()
	if err != nil {
		log.Errorf("failed to get cluster network config %v", err)
		return reconcile.Result{}, err
	}

	if clusterInfo == nil {
		log.Infof("could not populate network config")
		return reconcile.Result{}, nil
	}

	certificates := &operatorv1alpha1.TLSCertificates{}
	if err := r.GetConfigFromServer(certConfig, certificates); err == nil && certificates.CA == nil {
		log.Infof("No previous certificates found. creating certs first time")

		certificates, err = certs.GenerateCertificates(&operatorv1alpha1.CertGenConfig{})
		if err != nil {
			log.Errorf("failed to generate certs %v", err)
			return reconcile.Result{}, err
		}

		if err := r.SaveConfigToServer(certConfig, certificates); err != nil {
			log.Errorf("saving the release config failed %v", err)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		log.Errorf("getting previous certificates failed %v", err)
		return reconcile.Result{}, err
	}

	//Render the templates and get the objects
	renderData := render.MakeRenderData(&operatorv1alpha1.RenderConfig{
		NuageCNIConfigSpec:   instance.Spec,
		K8SAPIServerURL:      r.apiServerURL,
		ServiceAccountToken:  string(r.serviceAccountToken),
		Certificates:         certificates,
		ClusterNetworkConfig: clusterInfo,
	})

	var objs []*unstructured.Unstructured
	if objs, err = render.RenderDir(ManifestPath, &renderData); err != nil {
		//TODO: update operator status
		log.Errorf("Failed to render templates %v", err)
		return reconcile.Result{}, err
	}

	if instance.GetDeletionTimestamp() != nil {
		// Run finalization logic for nuageFinalizer. If the
		// finalization logic fails, don't remove the finalizer so
		// that we can retry during the next reconciliation.
		nuage_crd_names := []string{"nuage-infra", "nuage-monitor", "nuage-cni", "nuage-vrs"}
		for _, nuage_crd_name := range nuage_crd_names {
			err = r.deleteNuageResourceByName(objs, nuage_crd_name)
			if err != nil {
				return reconcile.Result{}, err
			} else {
				log.Infof("Deleted %s CRD objects", nuage_crd_name)
			}
			if nuage_crd_name == "nuage-infra" {
				err = r.confirmPodsDeletion(nuage_crd_name)
				if err != nil {
					return reconcile.Result{}, err
				}
			}
		}

		// Remove nuageFinalizer.
		instance.SetFinalizers(nil)

		// Update CR
		err = r.Client.Update(context.TODO(), instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	monitVSDAddressChange, err := r.checkMonitVSDAddressChange(instance)
	if err != nil {
		return reconcile.Result{}, nil
	}

	//Create or update the objects against API server
	for _, obj := range objs {
		if err := r.ApplyObject(types.NamespacedName{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
		}, obj); err != nil {
			log.Errorf("Appying object, name %s in namespace %s type %s %v", obj.GetName(), obj.GetNamespace(), obj.GroupVersionKind(), err)
		} else {
			log.Infof("Processed config for object %s in namespace %s type %s", obj.GetName(), obj.GetNamespace(), obj.GroupVersionKind())
		}
	}

	if err := r.LabelMasterNodes(); err != nil {
		log.Errorf("labeling master node with selector failed %v", err)
	}

	if err := r.SaveConfigToServer(releaseConfig, &instance.Spec.ReleaseConfig); err != nil {
		log.Errorf("Saving the release config failed %v", err)
		return reconcile.Result{}, err
	}

	//update cluster network status for openshift
	if err := r.UpdateClusterNetworkStatus(clusterInfo); err != nil {
		log.Errorf("updating cluster network status failed %v", err)
		return reconcile.Result{}, err
	}

	if monitVSDAddressChange {
		if err = r.UpdateDaemonsetpods(monitDaemonset); err != nil {
			log.Errorf("Updating daemonset pods failed %v", err)
			return reconcile.Result{}, err
		}
	}

	// Add finalizer for this CR
	if err := r.addFinalizer(instance); err != nil {
		return reconcile.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *NuageCNIConfigReconciler) deleteNuageResourceByName(objs []*unstructured.Unstructured, objName string) error {
	//delete nuage infra objects and pods against API server
	for _, obj := range objs {
		if obj.GetName() == objName {
			if err := r.DeleteResource(types.NamespacedName{
				Name:      obj.GetName(),
				Namespace: obj.GetNamespace(),
			}, obj); err != nil {
				log.Errorf("Failed deleting resource, name %s namespace %s type %s %v", obj.GetName(), obj.GetNamespace(), obj.GroupVersionKind(), err)
				log.Errorf("Object is %v", obj)
				return err
			}
			log.Infof("Deleted pod resources successfully for %s", objName)
		}
	}
	return nil
}

func (r *NuageCNIConfigReconciler) checkMonitVSDAddressChange(instance *operatorv1alpha1.NuageCNIConfig) (bool, error) {
	monitConfigMap, err := r.GetConfigMap(monitConfig)
	if err == nil {
		for _, monitConfigData := range monitConfigMap.Data {
			for _, lineData := range strings.Split(monitConfigData, "\n") {
				re, err := regexp.Compile(`vsdApiUrl`)
				match := re.FindStringIndex(lineData)
				if match != nil {
					currentAddress := strings.Split(lineData, "//")[len(strings.Split(lineData, "//"))-1]
					newAddress := fmt.Sprintf("%s:%d", instance.Spec.MonitorConfig.VSDAddress, instance.Spec.MonitorConfig.VSDPort)
					if currentAddress != newAddress {
						log.Infof("Current VSDAddress %s to be updated to %s", currentAddress, newAddress)
						return true, nil
					}
				} else if err != nil {
					log.Errorf("Error finding VSDURL in configMap %v", err)
				}
			}
		}
	} else if apierrors.IsNotFound(err) {
		log.Infof("No previous monitor configMap found, a new configmap will be created")
	} else {
		log.Errorf("Error getting monit configMap %v", err)
		return false, err
	}
	return false, nil
}

func (r *NuageCNIConfigReconciler) confirmPodsDeletion(CRDName string) error {
	log.Infof("Waiting for pods related to %s be deleted", CRDName)
	for {
		CRDPodsDeleted := true
		podList, err := r.clientset.CoreV1().Pods(names.Namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Errorf("Cannot retrieve pods from namespace %s: %s", names.Namespace, err)
			return err
		}
		for _, pod := range podList.Items {
			if pod.ObjectMeta.GenerateName == CRDName+"-" {
				CRDPodsDeleted = false
				break
			}
		}
		if CRDPodsDeleted {
			break
		} else {
			continue
		}
	}
	return nil
}

func (r *NuageCNIConfigReconciler) addFinalizer(nuageOperator *operatorv1alpha1.NuageCNIConfig) error {
	if len(nuageOperator.GetFinalizers()) < 1 && nuageOperator.GetDeletionTimestamp() == nil {
		log.Infof("Adding Finalizer for the Nuage")
		nuageOperator.SetFinalizers([]string{nuageFinalizer})
		// Update CustomResource
		err := r.Client.Update(context.TODO(), nuageOperator)
		if err != nil {
			log.Errorf("Failed to update NuageOperator with finalizer, %v", err)
			return err
		}
	}
	return nil
}

func (r *NuageCNIConfigReconciler) getOrchestratorType() (OrchestratorType, error) {

	if r.dclient == nil {
		return OrchestratorNone, fmt.Errorf("Discovery client not initialized. platform cannot be determined")
	}

	apis, err := r.dclient.ServerGroups()
	if err != nil {
		return OrchestratorNone, fmt.Errorf("Couldn't fetch api groups from api server")
	}

	for _, group := range apis.Groups {
		if group.Name == network.GroupName {
			return OrchestratorOpenShift, nil
		}
	}
	return OrchestratorKubernetes, nil
}

func (r *NuageCNIConfigReconciler) parse(instance *operatorv1alpha1.NuageCNIConfig) error {
	if err := monitor.Parse(&instance.Spec.MonitorConfig); err != nil {
		//invalid config passed.
		// TODO: update the operator status to the same and don't requeue
		log.Errorf("Failed to parse monitor config %v", err)
		return err
	}

	if err := cni.Parse(&instance.Spec.CNIConfig); err != nil {
		//invalid config passed.
		//TODO: update the operator status to the same and don't requeue
		log.Errorf("Failed to parse cni config %v", err)
		return err
	}

	if err := vrs.Parse(&instance.Spec.VRSConfig); err != nil {
		//invalid config passed.
		//TODO: update the operator status to the same and don't requeue
		log.Errorf("Failed to parse vrs config %v", err)
		return err
	}
	return nil
}

func (r *NuageCNIConfigReconciler) setPodNetworkConfig(p *operatorv1alpha1.PodNetworkConfigDefinition) {
	r.clusterNetworkCIDR = p.ClusterNetworkCIDR
	r.clusterNetworkSubnetLength = p.SubnetLength
	r.ClusterServiceNetworkCIDR = p.ClusterServiceNetworkCIDR
}

func (r *NuageCNIConfigReconciler) getServiceAccountToken() ([]byte, error) {
	secret, err := r.GetSecret(names.ServiceAccountName, names.Namespace)
	if err != nil {
		log.Errorf("Failed to get secret for sa %s in ns %s", names.ServiceAccountName, names.Namespace)
		return []byte{}, err
	}

	token, err := r.ExtractSecretToken(secret)
	if err != nil {
		log.Errorf("Token extraction failed %v", err)
		return []byte{}, err
	}

	return token, nil
}

func (r *NuageCNIConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	dc, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		log.Errorf("creating new discovery client failed")
		return err
	}
	r.dclient = dc

	r.clientset, err = kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		log.Errorf("creating new direct client failed")
		return err
	}

	r.Client = mgr.GetClient()

	r.orchestrator, err = r.getOrchestratorType()
	if err != nil {
		log.Errorf("orchestrator type could not be set %v", err)
		return err
	}

	r.serviceAccountToken, err = r.getServiceAccountToken()
	if err != nil {
		log.Errorf("creating service account token failed %v", err)
		return err
	}

	r.apiServerURL = mgr.GetConfig().Host

	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.NuageCNIConfig{}).
		Complete(r)
}
