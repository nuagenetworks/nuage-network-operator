// Copyright 2020 Nokia
// Licensed under the Apache License 2.0.
// SPDX-License-Identifier: Apache-2.0

package vrs

import (
	"testing"

	operv1 "github.com/nuagenetworks/nuage-network-operator/api/v1alpha1"
	. "github.com/onsi/gomega"
)

func TestParse(t *testing.T) {
	g := NewGomegaWithT(t)

	c := &operv1.VRSConfigDefinition{}
	err := Parse(c)
	g.Expect(err).Should(HaveOccurred())
	g.Expect(err.Error()).Should(ContainSubstring("validating vrs config"))

	c = &operv1.VRSConfigDefinition{
		Controllers:    []string{"1.1.1.1"},
		UnderlayUplink: "eth0",
	}
	err = Parse(c)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(c.Platform).Should(Equal(VRSPlatform))

	c = &operv1.VRSConfigDefinition{
		Controllers:    []string{"1.1.1.1", "2.2.2.2"},
		UnderlayUplink: "eth0",
	}
	err = Parse(c)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(c.Platform).Should(Equal(VRSPlatform))

	c = &operv1.VRSConfigDefinition{
		Controllers:    []string{"1.1.1.1000"},
		UnderlayUplink: "eth0",
	}
	err = Parse(c)
	g.Expect(err).Should(HaveOccurred())
	g.Expect(err.Error()).Should(ContainSubstring("controller ip is not valid"))

	c = &operv1.VRSConfigDefinition{
		UnderlayUplink: "eth0",
	}
	err = Parse(c)
	g.Expect(err).Should(HaveOccurred())
	g.Expect(err.Error()).Should(ContainSubstring("atleast one controller is expected"))

	c = &operv1.VRSConfigDefinition{
		Controllers: []string{"1.1.1.1"},
	}
	err = Parse(c)
	g.Expect(err).Should(HaveOccurred())
	g.Expect(err.Error()).Should(ContainSubstring("underlay uplink cannot be empty"))
}
