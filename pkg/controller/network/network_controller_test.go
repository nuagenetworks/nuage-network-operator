package network

import (
	"context"
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	operv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	"github.com/nuagenetworks/nuage-network-operator/pkg/names"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	osv1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var nu *operv1.Network

func createNetworkConfig(g *GomegaWithT, r *ReconcileNetwork) {
	scheme1 := scheme.Scheme
	osv1.Install(scheme1)
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

	err := r.client.Create(context.TODO(), c)
	g.Expect(err).ToNot(HaveOccurred())
}

func createNetworkOperatorConfig(g *GomegaWithT, r *ReconcileNetwork) {
	scheme2 := scheme.Scheme
	osv1.Install(scheme2)
	scheme2.AddKnownTypes(operv1.SchemeGroupVersion, &operv1.Network{})

	nu = &operv1.Network{
		TypeMeta:   metav1.TypeMeta{APIVersion: operv1.SchemeGroupVersion.String(), Kind: "Network"},
		ObjectMeta: metav1.ObjectMeta{Name: "nuage-network"},
		Spec: operv1.NetworkSpec{
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
			},
		},
	}
	err := r.client.Create(context.TODO(), nu)
	g.Expect(err).ToNot(HaveOccurred())
}

func updateCNITag(g *GomegaWithT, r *ReconcileNetwork) {
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

func TestReconcile(t *testing.T) {
	g := NewGomegaWithT(t)

	r := &ReconcileNetwork{
		client: fake.NewFakeClient(),
	}

	// create Network.config.openshift.io
	createNetworkConfig(g, r)

	// create Network.operator.nuage.io
	createNetworkOperatorConfig(g, r)

	// set up env vars simulating pod env
	setupEnvVars(g)

	ManifestPath = "../../../bindata"

	res, err := r.Reconcile(reconcile.Request{types.NamespacedName{Name: "nuage-network"}})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(res).ToNot(BeNil())

	monitorDS := appsv1.DaemonSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "nuage-monitor", Namespace: "kube-system"}, &monitorDS)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(monitorDS.GetName()).To(Equal("nuage-monitor"))

	monitorConfig := corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "nuage-monitor-config-data", Namespace: "kube-system"}, &monitorConfig)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(monitorConfig.GetName()).To(Equal("nuage-monitor-config-data"))

	cniDS := appsv1.DaemonSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "nuage-cni", Namespace: "kube-system"}, &cniDS)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cniDS.GetName()).To(Equal("nuage-cni"))

	cniConfig := corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "nuage-cni-config-data", Namespace: "kube-system"}, &cniConfig)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cniConfig.GetName()).To(Equal("nuage-cni-config-data"))

	vrsDS := appsv1.DaemonSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "nuage-vrs", Namespace: "kube-system"}, &vrsDS)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(vrsDS.GetName()).To(Equal("nuage-vrs"))

	releaseConfig := corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), nsn, &releaseConfig)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(releaseConfig.GetName()).To(Equal(names.ConfigName))

	//This reconcile should be a no op
	res, err = r.Reconcile(reconcile.Request{types.NamespacedName{Name: "nuage-network"}})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(res).ToNot(BeNil())

	//if we updated cni tag. so the reconile would redeploy the daemonsets
	updateCNITag(g, r)
	res, err = r.Reconcile(reconcile.Request{types.NamespacedName{Name: "nuage-network"}})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(res).ToNot(BeNil())

}
