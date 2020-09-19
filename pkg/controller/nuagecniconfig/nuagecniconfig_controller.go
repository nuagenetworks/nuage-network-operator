package nuagecniconfig

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	operv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	"github.com/nuagenetworks/nuage-network-operator/pkg/certs"
	"github.com/nuagenetworks/nuage-network-operator/pkg/names"
	"github.com/nuagenetworks/nuage-network-operator/pkg/network/cni"
	"github.com/nuagenetworks/nuage-network-operator/pkg/network/monitor"
	"github.com/nuagenetworks/nuage-network-operator/pkg/network/vrs"
	"github.com/nuagenetworks/nuage-network-operator/pkg/render"
	"github.com/openshift/api/network"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
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

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new NuageCNIConfig Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	r, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, r)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		log.Errorf("creating new discovery client failed")
		return &ReconcileNuageCNIConfig{}, err
	}

	clientset, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		log.Errorf("creating new direct client failed")
		return &ReconcileNuageCNIConfig{}, err
	}

	r := &ReconcileNuageCNIConfig{
		client:    mgr.GetClient(),
		scheme:    mgr.GetScheme(),
		dclient:   dc,
		clientset: clientset,
	}

	r.orchestrator, err = r.getOrchestratorType()
	if err != nil {
		log.Errorf("orchestrator type could not be set %v", err)
		return &ReconcileNuageCNIConfig{}, err
	}

	r.serviceAccountToken, err = r.getServiceAccountToken()
	if err != nil {
		log.Errorf("creating service account token failed %v", err)
		return &ReconcileNuageCNIConfig{}, err
	}

	r.apiServerURL = mgr.GetConfig().Host

	return r, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("nuagecniconfig-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource NuageCNIConfig
	err = c.Watch(&source.Kind{Type: &operv1.NuageCNIConfig{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner NuageCNIConfig
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operv1.NuageCNIConfig{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileNuageCNIConfig implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNuageCNIConfig{}

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

// ReconcileNuageCNIConfig reconciles a NuageCNIConfig object
type ReconcileNuageCNIConfig struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client                     client.Client
	dclient                    discovery.DiscoveryInterface
	scheme                     *runtime.Scheme
	orchestrator               OrchestratorType
	clusterNetworkCIDR         string
	apiServerURL               string
	serviceAccountToken        []byte
	clusterNetworkSubnetLength uint32
	clientset                  kubernetes.Interface
	ClusterServiceNetworkCIDR  string
}

// Reconcile reads that state of the cluster for a NuageCNIConfig object and makes changes based on the state read
// and what is in the NuageCNIConfig.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNuageCNIConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.SetLevel(log.DebugLevel)
	log.Infof("Reconciling NuageCNIConfig")

	// Fetch the Nuage custom resource instance
	instance := &operv1.NuageCNIConfig{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	if err := r.parse(instance); err != nil {
		log.Errorf("failed to parse crd config %v", err)
		return reconcile.Result{}, nil
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

	certificates := &operv1.TLSCertificates{}
	if err := r.GetConfigFromServer(certConfig, certificates); err == nil && certificates.CA == nil {
		log.Infof("No previous certificates found. creating certs first time")

		certificates, err = certs.GenerateCertificates(&operv1.CertGenConfig{})
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
	renderData := render.MakeRenderData(&operv1.RenderConfig{
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
		err = r.client.Update(context.TODO(), instance)
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
			log.Errorf("failed creating object, name %s in namespace %s type %s %v", obj.GetName(), obj.GetNamespace(), obj.GroupVersionKind(), err)
			log.Errorf("object is %v", obj)
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
	return reconcile.Result{}, nil
}

func (r *ReconcileNuageCNIConfig) deleteNuageResourceByName(objs []*unstructured.Unstructured, objName string) error {
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

func (r *ReconcileNuageCNIConfig) checkMonitVSDAddressChange(instance *operv1.NuageCNIConfig) (bool, error) {
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

func (r *ReconcileNuageCNIConfig) confirmPodsDeletion(CRDName string) error {
	log.Infof("Waiting for pods related to %s be deleted", CRDName)
	for {
		CRDPodsDeleted := true
		podList, err := r.clientset.CoreV1().Pods(names.Namespace).List(metav1.ListOptions{})
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

func (r *ReconcileNuageCNIConfig) addFinalizer(nuageOperator *operv1.NuageCNIConfig) error {
	if len(nuageOperator.GetFinalizers()) < 1 && nuageOperator.GetDeletionTimestamp() == nil {
		log.Infof("Adding Finalizer for the Nuage")
		nuageOperator.SetFinalizers([]string{nuageFinalizer})
		// Update CustomResource
		err := r.client.Update(context.TODO(), nuageOperator)
		if err != nil {
			log.Errorf("Failed to update NuageOperator with finalizer, %v", err)
			return err
		}
	}
	return nil
}

func (r *ReconcileNuageCNIConfig) getOrchestratorType() (OrchestratorType, error) {

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

func (r *ReconcileNuageCNIConfig) parse(instance *operv1.NuageCNIConfig) error {
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

func (r *ReconcileNuageCNIConfig) setPodNetworkConfig(p *operv1.PodNetworkConfigDefinition) {
	r.clusterNetworkCIDR = p.ClusterNetworkCIDR
	r.clusterNetworkSubnetLength = p.SubnetLength
	r.ClusterServiceNetworkCIDR = p.ClusterServiceNetworkCIDR
}

func (r *ReconcileNuageCNIConfig) getServiceAccountToken() ([]byte, error) {
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
