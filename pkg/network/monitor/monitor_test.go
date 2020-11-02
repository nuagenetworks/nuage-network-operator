// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package monitor

import (
	"testing"

	operv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	. "github.com/onsi/gomega"
)

var c = &operv1.MonitorConfigDefinition{
	VSDAddress: "127.0.0.1",
	VSDPort:    8443,
	VSDMetadata: operv1.Metadata{
		Enterprise: "test",
		Domain:     "test",
		User:       "test",
		UserCert:   "test",
		UserKey:    "test",
	},
	VSDFlags: operv1.Flags{
		UnderlayEnabled:  true,
		StatsEnabled:     true,
		AutoScaleSubnets: true,
	},
}

func TestParse(t *testing.T) {
	g := NewGomegaWithT(t)

	b := &operv1.MonitorConfigDefinition{}
	err := Parse(b)
	g.Expect(err).To(HaveOccurred())

	err = Parse(c)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(c.RestServerAddress).To(Equal(DefaultRestServerAddress))
	g.Expect(c.RestServerPort).To(Equal(DefaultRestServerPort))
	g.Expect(c.ServiceAccountName).To(Equal(DefaultResourceName))

	c.RestServerAddress = "acv"
	c.RestServerPort = 1000
	err = Parse(c)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(c.RestServerPort).To(Equal(1000))
	g.Expect(c.RestServerAddress).To(Equal("acv"))

	c.VSDAddress = ""
	err = Parse(c)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("validating monitor config failed"))
	g.Expect(err.Error()).To(ContainSubstring("invalid vsd ip address"))

	c.VSDAddress = "127.0.0.1"
	c.VSDPort = -1
	err = Parse(c)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("invalid vsd port address"))

	c.VSDPort = 100
	c.VSDMetadata.UserKey = ""
	err = Parse(c)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("validating monitor config failed"))
	g.Expect(err.Error()).To(ContainSubstring("vsd metadata validation failed"))
	g.Expect(err.Error()).To(ContainSubstring("user key cannot be empty"))

	c.VSDMetadata.UserKey = "abc"
	c.RestServerPort = -1
	err = Parse(c)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("invalid rest server port"))

	c.RestServerPort = 100
	c.ServiceAccountName = "test-name"
	err = Parse(c)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(c.ServiceAccountName).To(Equal("test-name"))
}
