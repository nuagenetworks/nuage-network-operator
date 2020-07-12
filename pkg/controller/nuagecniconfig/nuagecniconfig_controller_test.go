package nuagecniconfig

import (
	"context"
	"fmt"
	"os"
	"testing"

	operv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	"github.com/nuagenetworks/nuage-network-operator/pkg/names"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/api/network"
	osv1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	fakeRest "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var testlog = logf.Log.WithName("cluster_config_test")

var nu *operv1.NuageCNIConfig

//Thanks to https://github.com/jaegertracing/jaeger-operator/
type fakeDiscoveryClient struct {
	discovery.DiscoveryInterface
	ServerGroupsFunc func() (apiGroupList *metav1.APIGroupList, err error)
}

func (d *fakeDiscoveryClient) ServerGroups() (apiGroupList *metav1.APIGroupList, err error) {
	if d.ServerGroupsFunc == nil {
		return &metav1.APIGroupList{}, nil
	}
	return d.ServerGroupsFunc()
}

func createNetworkConfig(g *GomegaWithT, r *ReconcileNuageCNIConfig) {
	scheme1 := scheme.Scheme
	err := osv1.Install(scheme1)
	if err != nil {
		testlog.Error(err, "Failed to install scheme")
	}
	scheme1.AddKnownTypes(configv1.SchemeGroupVersion, &configv1.Network{})
	c := &configv1.Network{
		TypeMeta:   metav1.TypeMeta{APIVersion: configv1.GroupVersion.String(), Kind: "Network"},
		ObjectMeta: metav1.ObjectMeta{Name: "network"},
		Spec: configv1.NetworkSpec{
			ClusterNetwork: []configv1.ClusterNetworkEntry{
				{CIDR: "70.70.0.0/16", HostPrefix: 24},
			},
			ServiceNetwork: []string{"192.168.0.0/16"},
			NetworkType:    names.NuageSDN,
		},
	}

	err = r.client.Create(context.TODO(), c)
	g.Expect(err).ToNot(HaveOccurred())
}

func createNetworkOperatorConfig(g *GomegaWithT, r *ReconcileNuageCNIConfig) {
	scheme2 := scheme.Scheme
	err := osv1.Install(scheme2)
	if err != nil {
		testlog.Error(err, "Failed to install scheme")
	}
	scheme2.AddKnownTypes(operv1.SchemeGroupVersion, &operv1.NuageCNIConfig{})

	nu = &operv1.NuageCNIConfig{
		TypeMeta:   metav1.TypeMeta{APIVersion: operv1.SchemeGroupVersion.String(), Kind: "NuageCNIConfig"},
		ObjectMeta: metav1.ObjectMeta{Name: "nuage-network"},
		Spec: operv1.NuageCNIConfigSpec{
			VRSConfig: operv1.VRSConfigDefinition{
				Controllers:    []string{"10.10.0.0", "10.10.0.1"},
				UnderlayUplink: "eth0",
			},
			CNIConfig: operv1.CNIConfigDefinition{
				LoadBalancerURL: "https://127.0.0.1:9443",
			},
			MonitorConfig: operv1.MonitorConfigDefinition{
				VSDAddress: "10.10.0.2",
				VSDPort:    8443,
				VSDMetadata: operv1.Metadata{
					Enterprise: "ent",
					Domain:     "dom",
					User:       "user",
					UserCert:   "cert",
					UserKey:    "key",
				},
			},
			ReleaseConfig: operv1.ReleaseConfigDefinition{
				Registry: operv1.RegistryConfig{
					URL:      "https://registry.mv.nuagenetworks.net/",
					Username: "username",
					Password: "password",
				},
				VRSTag:     "0.0.0",
				CNITag:     "0.0.0",
				MonitorTag: "0.0.0",
				InfraTag:   "0.0.0",
			},
		},
	}
	err = r.client.Create(context.TODO(), nu)
	g.Expect(err).ToNot(HaveOccurred())
}

func updateCNITag(g *GomegaWithT, r *ReconcileNuageCNIConfig) {
	nu.Spec.ReleaseConfig.CNITag = "0.0.1"
	err := r.client.Update(context.TODO(), nu)
	g.Expect(err).ToNot(HaveOccurred())
}

func setupEnvVars(g *GomegaWithT) {
	err := os.Setenv("KUBERNETES_SERVICE_HOST", "192.168.0.1")
	g.Expect(err).ToNot(HaveOccurred())
	err = os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	g.Expect(err).ToNot(HaveOccurred())
}

func TestGetOrchestratorType(t *testing.T) {
	g := NewGomegaWithT(t)
	dcl := &fakeDiscoveryClient{}

	r := &ReconcileNuageCNIConfig{
		client: fake.NewFakeClient(),
	}

	o, err := r.getOrchestratorType()
	g.Expect(err).To(HaveOccurred())
	g.Expect(o).To(Equal(OrchestratorNone))

	r.dclient = dcl
	o, err = r.getOrchestratorType()
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(o).To(Equal(OrchestratorKubernetes))

	dcl.ServerGroupsFunc = func() (*metav1.APIGroupList, error) {
		return nil, fmt.Errorf("error")
	}
	o, err = r.getOrchestratorType()
	g.Expect(err).To(HaveOccurred())
	g.Expect(o).To(Equal(OrchestratorNone))

	dcl.ServerGroupsFunc = func() (*metav1.APIGroupList, error) {
		return &metav1.APIGroupList{
			Groups: []metav1.APIGroup{
				{
					Name: network.GroupName,
				},
			},
		}, nil
	}
	o, err = r.getOrchestratorType()
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(o).To(Equal(OrchestratorOpenShift))

}

func TestReconcile(t *testing.T) {
	g := NewGomegaWithT(t)

	r := &ReconcileNuageCNIConfig{
		client:    fake.NewFakeClient(),
		dclient:   &fakeDiscoveryClient{},
		clientset: fakeRest.NewSimpleClientset(),
	}

	// create Network.config.openshift.io
	createNetworkConfig(g, r)

	// create NuageCNIConfig.operator.nuage.io
	createNetworkOperatorConfig(g, r)

	// set up env vars simulating pod env
	setupEnvVars(g)

	ManifestPath = "../../../bindata"

	res, err := r.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: "nuage-network"}})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(res).ToNot(BeNil())

	monitorDS := appsv1.DaemonSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "nuage-monitor", Namespace: names.Namespace}, &monitorDS)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(monitorDS.GetName()).To(Equal("nuage-monitor"))

	monitorConfig := corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "nuage-monitor-config-data", Namespace: names.Namespace}, &monitorConfig)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(monitorConfig.GetName()).To(Equal("nuage-monitor-config-data"))

	cniDS := appsv1.DaemonSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "nuage-cni", Namespace: names.Namespace}, &cniDS)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cniDS.GetName()).To(Equal("nuage-cni"))

	cniConfig := corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "nuage-cni-config-data", Namespace: names.Namespace}, &cniConfig)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cniConfig.GetName()).To(Equal("nuage-cni-config-data"))

	vrsDS := appsv1.DaemonSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "nuage-vrs", Namespace: names.Namespace}, &vrsDS)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(vrsDS.GetName()).To(Equal("nuage-vrs"))

	rc := corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), releaseConfig, &rc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(rc.GetName()).To(Equal(names.NuageReleaseConfig))

	//This reconcile should be a no op
	res, err = r.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: "nuage-network"}})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(res).ToNot(BeNil())

	//if we updated cni tag. so the reconile would redeploy the daemonsets
	updateCNITag(g, r)
	res, err = r.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: "nuage-network"}})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(res).ToNot(BeNil())

}
