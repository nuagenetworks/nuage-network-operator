package nuagecniconfig

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/nuagenetworks/nuage-network-operator/pkg/names"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	osv1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
