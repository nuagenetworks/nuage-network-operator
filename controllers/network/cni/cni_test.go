// Copyright 2020 Nokia
// Licensed under the Apache License 2.0.
// SPDX-License-Identifier: Apache-2.0

package cni

import (
	"testing"

	operv1 "github.com/nuagenetworks/nuage-network-operator/api/v1alpha1"
	. "github.com/onsi/gomega"
)

func TestParse(t *testing.T) {
	g := NewGomegaWithT(t)

	c := &operv1.CNIConfigDefinition{
		LoadBalancerURL: "https://127.0.0.1:9443",
	}

	err := Parse(c)
	g.Expect(err).To(BeNil())
	g.Expect(c.MTU).To(Equal(1450))

	c.NuageSiteID = -1
	err = Parse(c)
	g.Expect(err).To(BeNil())

	c.NuageSiteID = 10
	err = Parse(c)
	g.Expect(err).Should(HaveOccurred())
	g.Expect(err.Error()).Should(ContainSubstring("non negative values"))

	c.MTU = 1500
	c.NuageSiteID = -1
	err = Parse(c)
	g.Expect(err).Should(HaveOccurred())
	g.Expect(err.Error()).Should(ContainSubstring("mtu exceeds"))

	c.MTU = 1450
	err = Parse(c)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(c.ServiceAccountName).To(Equal(DefaultResourceName))

	c.LoadBalancerURL = ""
	err = Parse(c)
	g.Expect(err).Should(HaveOccurred())
	g.Expect(err.Error()).Should(ContainSubstring("load balancer url cannot be empty"))
}
