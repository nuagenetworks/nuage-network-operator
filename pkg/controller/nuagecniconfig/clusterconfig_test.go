package nuagecniconfig

import (
	"context"
	"fmt"
	"testing"

	"github.com/nuagenetworks/nuage-network-operator/pkg/names"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	osv1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestGetK8SClusterNetworkInfo(t *testing.T) {
	g := NewGomegaWithT(t)
	f := &fakeRestClient{
		client: fake.NewFakeClient(),
	}

	r := &ReconcileNuageCNIConfig{
		client: f,
	}

	//no test kube apiserver pods are created. should error
	c, err := r.GetK8SClusterNetworkInfo()
	g.Expect(err).To(HaveOccurred())

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-kube-apiserver",
			Namespace: "kube-system",
			Labels:    map[string]string{"component": "kube-apiserver"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Image:   "test-container-image",
				Name:    "test-container-name",
				Command: []string{"--service-cluster-ip-range=\"192.168.0.0/16\""},
			},
			},
		},
	}

	err = r.client.Create(context.TODO(), pod)
	g.Expect(err).ToNot(HaveOccurred())
	// should not error as there is a kube-apiserver pod
	c, err = r.GetK8SClusterNetworkInfo()
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(c.ServiceNetworkCIDR).To(Equal("192.168.0.0/16"))

	for _, arg := range []string{"--service-cluster-ip-range", "not-used-for-anything"} {
		pod.Spec.Containers[0].Command = []string{arg}
		err = r.client.Update(context.TODO(), pod)
		g.Expect(err).ToNot(HaveOccurred())
		c, err = r.GetK8SClusterNetworkInfo()
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(c.ServiceNetworkCIDR).To(Equal(DefaultServiceNetworkCIDR))
	}

	f.ListFunc = func(ctx context.Context, opts *client.ListOptions, obj runtime.Object) error {
		return fmt.Errorf(apiServerError)
	}
	//if api server returns an error, test should catch it
	_, err = r.GetK8SClusterNetworkInfo()
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(Equal(apiServerError))
}

func TestGetClusterConfig(t *testing.T) {
	g := NewGomegaWithT(t)
	s := scheme.Scheme
	osv1.Install(s)
	s.AddKnownTypes(configv1.SchemeGroupVersion, &configv1.Network{})

	r := &ReconcileNuageCNIConfig{
		client: fake.NewFakeClient(),
	}

	cnf, err := r.GetClusterNetworkInfo(reconcile.Request{})
	g.Expect(err).To(BeNil())
	g.Expect(cnf).To(BeNil())

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

	cnf, err = r.GetClusterNetworkInfo(reconcile.Request{types.NamespacedName{Name: "network"}})
	g.Expect(err).To(BeNil())
	g.Expect(cnf).ToNot(BeNil())
	g.Expect(cnf.ClusterNetworkCIDR).To(Equal("70.70.0.0/16"))
	g.Expect(cnf.ServiceNetworkCIDR).To(Equal("192.168.0.0/16"))
	g.Expect(cnf.ClusterNetworkSubnetLength).To(Equal(uint32(24)))

}

func TestValidateClusterConfig(t *testing.T) {
	g := NewGomegaWithT(t)

	//	r := &ReconcileNuageCNIConfig{
	//		client: fake.NewFakeClient(),
	//	}

	spec := configv1.NetworkSpec{
		ClusterNetwork: []configv1.ClusterNetworkEntry{
			{CIDR: "70.70.0.0/16", HostPrefix: 24},
		},
		ServiceNetwork: []string{"192.168.0.0/16"},
		NetworkType:    names.NuageSDN,
	}

	err := ValidateClusterConfig(spec)
	g.Expect(err).ToNot(HaveOccurred())

	spec1 := configv1.NetworkSpec{
		ClusterNetwork: []configv1.ClusterNetworkEntry{
			{CIDR: "70.70.0.0/16", HostPrefix: 24},
		},
		ServiceNetwork: []string{"192.168.0.0/16"},
	}

	err = ValidateClusterConfig(spec1)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("is not supported"))

	spec2 := configv1.NetworkSpec{
		ClusterNetwork: []configv1.ClusterNetworkEntry{
			{CIDR: "70.70.0.0/16", HostPrefix: 24},
			{CIDR: "70.70.0.0/16", HostPrefix: 24},
		},
		ServiceNetwork: []string{"192.168.0.0/16"},
	}
	err = ValidateClusterConfig(spec2)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("must have only one entry"))

	spec3 := configv1.NetworkSpec{
		ClusterNetwork: []configv1.ClusterNetworkEntry{
			{CIDR: "70.70.0.0/16", HostPrefix: 24},
		},
		ServiceNetwork: []string{"192.168.0.0/16", "10.10.0.0/16"},
	}
	err = ValidateClusterConfig(spec3)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("must have only one entry"))

	spec4 := configv1.NetworkSpec{
		ClusterNetwork: []configv1.ClusterNetworkEntry{
			{CIDR: "70.70.0.0/16", HostPrefix: 24},
		},
		ServiceNetwork: []string{"192.168.0.0/116"},
	}
	err = ValidateClusterConfig(spec4)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("could not parse spec.serviceNetwork"))

	spec5 := configv1.NetworkSpec{
		ClusterNetwork: []configv1.ClusterNetworkEntry{
			{CIDR: "70.70.0.0/116", HostPrefix: 24},
		},
		ServiceNetwork: []string{"192.168.0.0/16"},
	}
	err = ValidateClusterConfig(spec5)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("could not parse spec.clusterNetwork"))

	spec6 := configv1.NetworkSpec{
		ClusterNetwork: []configv1.ClusterNetworkEntry{
			{CIDR: "70.70.0.0/16", HostPrefix: 15},
		},
		ServiceNetwork: []string{"192.168.0.0/16"},
	}
	err = ValidateClusterConfig(spec6)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("is larger than its cidr"))

	spec7 := configv1.NetworkSpec{
		ClusterNetwork: []configv1.ClusterNetworkEntry{
			{CIDR: "70.70.0.0/16", HostPrefix: 31},
		},
		ServiceNetwork: []string{"192.168.0.0/16"},
	}
	err = ValidateClusterConfig(spec7)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("is too small"))
}
