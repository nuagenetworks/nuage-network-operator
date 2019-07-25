package vrs

import (
	"testing"

	operv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	. "github.com/onsi/gomega"
)

func TestParse(t *testing.T) {
	g := NewGomegaWithT(t)

	c := &operv1.CNIConfigDefinition{}

	err := Parse(c)
	g.Expect(err).To(BeNil())
	g.Expect(c.MTU).To(Equal(1450))

	c = &operv1.CNIConfigDefinition{}
	c.NuageSiteID = -1
	err = Parse(c)
	g.Expect(err).To(BeNil())

	c = &operv1.CNIConfigDefinition{}
	c.NuageSiteID = 10
	err = Parse(c)
	g.Expect(err).Should(HaveOccurred())
	g.Expect(err.Error()).Should(ContainSubstring("non negative values"))

	c = &operv1.CNIConfigDefinition{}
	c.MTU = 1500
	err = Parse(c)
	g.Expect(err).Should(HaveOccurred())
	g.Expect(err.Error()).Should(ContainSubstring("mtu exceeds"))

	c = &operv1.CNIConfigDefinition{}
	c.MTU = 1450
	err = Parse(c)
	g.Expect(err).ShouldNot(HaveOccurred())
}
