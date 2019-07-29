package network

import (
	"context"

	operatorv1alpha1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	"github.com/nuagenetworks/nuage-network-operator/pkg/certs"
	"github.com/nuagenetworks/nuage-network-operator/pkg/network/cni"
	"github.com/nuagenetworks/nuage-network-operator/pkg/network/monitor"
	"github.com/nuagenetworks/nuage-network-operator/pkg/network/vrs"
	"github.com/nuagenetworks/nuage-network-operator/pkg/render"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	log          = logf.Log.WithName("controller_network")
	ManifestPath = "./bindata"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Network Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNetwork{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("network-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Network
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.Network{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Network
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.Network{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileNetwork implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNetwork{}

// ReconcileNetwork reconciles a Network object
type ReconcileNetwork struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Network object and makes changes based on the state read
// and what is in the Network.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNetwork) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Network")

	clusterInfo, err := r.GetClusterNetworkInfo(request)
	if err != nil {
		reqLogger.Error(err, "failed to get cluster network config")
		return reconcile.Result{}, err
	}

	if clusterInfo == nil {
		reqLogger.Info("could not find network config. object must have been deleted")
		return reconcile.Result{}, nil
	}

	// Fetch the Nuage custom resource instance
	instance := &operatorv1alpha1.Network{}
	err = r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if err := monitor.Parse(&instance.Spec.MonitorConfig); err != nil {
		//invalid config passed.
		// TODO: update the operator status to the same and dont requeue
		return reconcile.Result{}, nil
	}

	if err := cni.Parse(&instance.Spec.CNIConfig); err != nil {
		//invalid config passed.
		//TODO: update the operator status to the same and dont requeue
		return reconcile.Result{}, nil
	}

	if err := vrs.Parse(&instance.Spec.VRSConfig); err != nil {
		//invalid config passed.
		//TODO: update the operator status to the same and dont requeue
		return reconcile.Result{}, nil
	}

	certificates := &operatorv1alpha1.TLSCertificates{}
	certificates, err = certs.GenerateCertificates(&operatorv1alpha1.CertGenConfig{})
	if err != nil {
		return reconcile.Result{}, err
	}

	renderData := render.MakeRenderData(&operatorv1alpha1.RenderConfig{
		instance.Spec,
		"https://0.0.0.0:9443",
		certificates,
		clusterInfo,
	})

	var objs []*unstructured.Unstructured
	if objs, err = render.RenderDir(ManifestPath, &renderData); err != nil {
		//TODO: update operator status
		return reconcile.Result{}, err
	}

	for _, obj := range objs {
		if err := r.client.Create(context.TODO(), obj); err != nil {
			log.Error(err, "error creating the object %v", err)
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}
