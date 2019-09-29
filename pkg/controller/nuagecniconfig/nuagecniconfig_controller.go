package nuagecniconfig

import (
	"context"
	"fmt"
	"os"
	"reflect"

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
	ManifestPath = "./bindata"
)

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

	r.serviceAccountToken, err = r.createServiceAccountToken()
	if err != nil {
		log.Errorf("creating service account token failed %v", err)
		return &ReconcileNuageCNIConfig{}, err
	}

	r.apiServerURL, err = buildAPIServerURL()
	if err != nil {
		log.Errorf("failed to get api server url %v", err)
		return &ReconcileNuageCNIConfig{}, err
	}

	return r, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("network-controller", mgr, controller.Options{Reconciler: r})
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
	clientset                  *kubernetes.Clientset
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

	clusterInfo := &operv1.ClusterNetworkConfigDefinition{}
	clusterInfo, err = r.GetClusterNetworkInfo()
	if err != nil {
		log.Errorf("failed to get cluster network config %v", err)
		return reconcile.Result{}, err
	}

	if clusterInfo == nil {
		log.Infof("could not populate network config")
		return reconcile.Result{}, nil
	}

	//TODO: Get the previous config
	rc := &operv1.ReleaseConfigDefinition{}
	if err := r.GetConfigFromServer(releaseConfig, rc); err == nil && len(rc.VRSTag) == 0 {
		log.Infof("no previous config found. creating objects first time")
	} else {
		if err != nil {
			log.Errorf("getting release config failed %v", err)
			return reconcile.Result{}, err
		}

		if reflect.DeepEqual(rc, &instance.Spec.ReleaseConfig) {
			log.Warnf("no config differences found. skipping reconcile")
			return reconcile.Result{}, nil
		}
	}

	certificates := &operv1.TLSCertificates{}
	if err := r.GetConfigFromServer(certConfig, certificates); err == nil && certificates.CA == nil {
		log.Infof("no previous certificates found. creating certs first time")

		certificates, err = certs.GenerateCertificates(&operv1.CertGenConfig{})
		if err != nil {
			log.Errorf("failed to generate certs %v", err)
			return reconcile.Result{}, err
		}
	} else if err != nil {
		log.Errorf("getting previous certificates failed %v", err)
		return reconcile.Result{}, err
	}

	//Render the templates and get the objects
	renderData := render.MakeRenderData(&operv1.RenderConfig{
		instance.Spec,
		r.apiServerURL,
		string(r.serviceAccountToken),
		certificates,
		clusterInfo,
	})

	var objs []*unstructured.Unstructured
	if objs, err = render.RenderDir(ManifestPath, &renderData); err != nil {
		//TODO: update operator status
		log.Errorf("failed to render templates %v", err)
		return reconcile.Result{}, err
	}

	//Create or update the objects against API server
	for _, obj := range objs {
		if err := r.ApplyObject(types.NamespacedName{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
		}, obj); err != nil {
			log.Errorf("failed creating object, name %s namespace %s type %s", obj.GetName(), obj.GetNamespace(), obj.GroupVersionKind())
		}
	}

	if err := r.SaveConfigToServer(releaseConfig, &instance.Spec.ReleaseConfig); err != nil {
		log.Errorf("saving the release config failed %v", err)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func buildAPIServerURL() (string, error) {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		return "", fmt.Errorf("neither kubernetes service host nor service port can be empty")
	}

	return "https://" + host + ":" + port, nil
}

func (r *ReconcileNuageCNIConfig) getOrchestratorType() (OrchestratorType, error) {

	if r.dclient == nil {
		return OrchestratorNone, fmt.Errorf("discovery client not initialized. platform cannot be determined")
	}

	apis, err := r.dclient.ServerGroups()
	if err != nil {
		return OrchestratorNone, fmt.Errorf("couldn't fetch api groups from api server")
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
		// TODO: update the operator status to the same and dont requeue
		log.Errorf("failed to parse monitor config %v", err)
		return err
	}

	if err := cni.Parse(&instance.Spec.CNIConfig); err != nil {
		//invalid config passed.
		//TODO: update the operator status to the same and dont requeue
		log.Errorf("failed to parse cni config %v", err)
		return err
	}

	if err := vrs.Parse(&instance.Spec.VRSConfig); err != nil {
		//invalid config passed.
		//TODO: update the operator status to the same and dont requeue
		log.Errorf("failed to parse vrs config %v", err)
		return err
	}

	return nil
}

func (r *ReconcileNuageCNIConfig) setPodNetworkConfig(p *operv1.PodNetworkConfigDefinition) {
	r.clusterNetworkCIDR = p.ClusterNetworkCIDR
	r.clusterNetworkSubnetLength = p.SubnetLength
}

func (r *ReconcileNuageCNIConfig) createServiceAccountToken() ([]byte, error) {
	//	err := r.CreateServiceAccount(names.ServiceAccountName, names.Namespace)
	//	if err != nil {
	//		log.Errorf("failed to create service account %v", err)
	//		return []byte{}, err
	//	}

	secret, err := r.GetSecret(names.ServiceAccountName, names.Namespace)
	if err != nil {
		log.Errorf("failed to get secret for sa %s in ns %s", names.ServiceAccountName, names.Namespace)
		return []byte{}, err
	}

	token, err := r.ExtractSecretToken(secret)
	if err != nil {
		log.Errorf("token extraction failed %v", err)
		return []byte{}, err
	}

	return token, nil
}
