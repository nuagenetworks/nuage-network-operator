package nuagecniconfig

import (
	"context"
	"testing"

	operv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	"github.com/nuagenetworks/nuage-network-operator/pkg/names"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	osv1 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func init() {
	s := scheme.Scheme
	err := osv1.Install(s)
	if err != nil {
		testlog.Error(err, "Failed to install openshift")
	}
	s.AddKnownTypes(configv1.SchemeGroupVersion, &configv1.Network{})
}

func TestClusterConfigUpdateStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	r := &ReconcileNuageCNIConfig{
		client:       fake.NewFakeClient(),
		orchestrator: OrchestratorOpenShift,
	}

	c := &configv1.Network{
		TypeMeta:   metav1.TypeMeta{APIVersion: configv1.GroupVersion.String(), Kind: "Network"},
		ObjectMeta: metav1.ObjectMeta{Name: "cluster"},
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

	d := &operv1.ClusterNetworkConfigDefinition{
		ClusterNetworkCIDR:         "a",
		ClusterNetworkSubnetLength: 8,
		ServiceNetworkCIDR:         "c",
	}

	err = r.UpdateClusterNetworkStatus(d)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestClusterConfigGet(t *testing.T) {
	g := NewGomegaWithT(t)

	r := &ReconcileNuageCNIConfig{
		client: fake.NewFakeClient(),
	}

	cnf, err := r.GetClusterNetworkInfo()
	g.Expect(err).ToNot(BeNil())
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

	cnf, err = r.GetOSEClusterNetworkInfo()
	g.Expect(err).To(BeNil())
	g.Expect(cnf).ToNot(BeNil())
	g.Expect(cnf.ClusterNetworkCIDR).To(Equal("70.70.0.0/16"))
	g.Expect(cnf.ServiceNetworkCIDR).To(Equal("192.168.0.0/16"))
	g.Expect(cnf.ClusterNetworkSubnetLength).To(Equal(uint32(24)))

}

func TestClusterConfigValidateOSE(t *testing.T) {
	g := NewGomegaWithT(t)

	type testvec struct {
		in  configv1.NetworkSpec
		out error
	}

	vec := []testvec{
		{
			in: configv1.NetworkSpec{
				ClusterNetwork: []configv1.ClusterNetworkEntry{
					{CIDR: "70.70.0.0/16", HostPrefix: 24},
				},
				ServiceNetwork: []string{"192.168.0.0/16"},
				NetworkType:    names.NuageSDN,
			},
			out: nil,
		},
		{
			in: configv1.NetworkSpec{
				ClusterNetwork: []configv1.ClusterNetworkEntry{
					{CIDR: "70.70.0.0/16", HostPrefix: 24},
				},
				ServiceNetwork: []string{"192.168.0.0/16"},
			},
			out: errors.Errorf("is not supported"),
		},
		{
			in: configv1.NetworkSpec{
				ClusterNetwork: []configv1.ClusterNetworkEntry{
					{CIDR: "70.70.0.0/16", HostPrefix: 24},
					{CIDR: "70.70.0.0/16", HostPrefix: 24},
				},
				ServiceNetwork: []string{"192.168.0.0/16"},
			},
			out: errors.Errorf("must have only one entry"),
		},
		{
			in: configv1.NetworkSpec{
				ClusterNetwork: []configv1.ClusterNetworkEntry{
					{CIDR: "70.70.0.0/16", HostPrefix: 24},
				},
				ServiceNetwork: []string{"192.168.0.0/16", "10.10.0.0/16"},
			},
			out: errors.Errorf("must have only one entry"),
		},
		{
			in: configv1.NetworkSpec{
				ClusterNetwork: []configv1.ClusterNetworkEntry{
					{CIDR: "70.70.0.0/16", HostPrefix: 24},
				},
				ServiceNetwork: []string{"192.168.0.0/116"},
			},
			out: errors.Errorf("could not parse spec.serviceNetwork"),
		},
		{
			in: configv1.NetworkSpec{
				ClusterNetwork: []configv1.ClusterNetworkEntry{
					{CIDR: "70.70.0.0/116", HostPrefix: 24},
				},
				ServiceNetwork: []string{"192.168.0.0/16"},
			},
			out: errors.Errorf("could not parse spec.clusterNetwork"),
		},
		{
			in: configv1.NetworkSpec{
				ClusterNetwork: []configv1.ClusterNetworkEntry{
					{CIDR: "70.70.0.0/16", HostPrefix: 15},
				},
				ServiceNetwork: []string{"192.168.0.0/16"},
			},
			out: errors.Errorf("is larger than its cidr"),
		},
		{
			in: configv1.NetworkSpec{
				ClusterNetwork: []configv1.ClusterNetworkEntry{
					{CIDR: "70.70.0.0/16", HostPrefix: 31},
				},
				ServiceNetwork: []string{"192.168.0.0/16"},
			},
			out: errors.Errorf("is too small"),
		},
	}

	for _, tt := range vec {
		err := ValidateOSEClusterConfig(tt.in)
		if tt.out == nil {
			g.Expect(err).To(BeNil())
		} else {
			g.Expect(err.Error()).To(ContainSubstring(tt.out.Error()))
		}
	}

}

func TestClusterConfigValidateK8S(t *testing.T) {
	g := NewGomegaWithT(t)

	type testvec struct {
		in  *operv1.ClusterNetworkConfigDefinition
		out error
	}
	vec := []testvec{
		{
			in:  &operv1.ClusterNetworkConfigDefinition{},
			out: errors.Errorf("invalid service network cidr found "),
		},
		{
			in: &operv1.ClusterNetworkConfigDefinition{
				ServiceNetworkCIDR:         "192.168.0.0/16",
				ClusterNetworkCIDR:         "70.70.0.0/16",
				ClusterNetworkSubnetLength: 20,
			},
			out: nil,
		},
		{
			in: &operv1.ClusterNetworkConfigDefinition{
				ServiceNetworkCIDR:         "192.168.0.0/16",
				ClusterNetworkCIDR:         "192.168.0/18",
				ClusterNetworkSubnetLength: 20,
			},
			out: errors.Errorf("invalid pod network cidr found 192.168.0/18"),
		},
		{
			in: &operv1.ClusterNetworkConfigDefinition{
				ServiceNetworkCIDR:         "192.168.0.0/16",
				ClusterNetworkCIDR:         "192.168.0.0/18",
				ClusterNetworkSubnetLength: 20,
			},
			out: errors.Errorf("CIDRs 192.168.0.0/16 and 192.168.0.0/18 overlap"),
		},
		{
			in: &operv1.ClusterNetworkConfigDefinition{
				ServiceNetworkCIDR:         "192.168.0.0/16",
				ClusterNetworkCIDR:         "70.70.0.0/18",
				ClusterNetworkSubnetLength: 16,
			},
			out: errors.Errorf("subnet length 16 is larger than its cidr 70.70.0.0/18"),
		},
		{
			in: &operv1.ClusterNetworkConfigDefinition{
				ServiceNetworkCIDR:         "192.168.0.0/16",
				ClusterNetworkCIDR:         "70.70.0.0/18",
				ClusterNetworkSubnetLength: 31,
			},
			out: errors.Errorf("subnet length 31 is too small, must be a /30 or larger"),
		},
	}

	for _, tt := range vec {
		err := ValidateK8SClusterConfig(tt.in)
		if tt.out == nil {
			g.Expect(err).To(BeNil())
		} else {
			g.Expect(err.Error()).To(Equal(tt.out.Error()))
		}
	}

}
