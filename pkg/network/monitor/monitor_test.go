package monitor

import (
	"testing"

	operv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	. "github.com/onsi/gomega"
)

func TestParse(t *testing.T) {
	g := NewGomegaWithT(t)

	c := &operv1.MonitorConfigDefinition{}
	err := Parse(c)
	g.Expect(err).To(HaveOccurred())

	c = &operv1.MonitorConfigDefinition{
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
	err = Parse(c)
	g.Expect(err).ToNot(HaveOccurred())

	c = &operv1.MonitorConfigDefinition{
		VSDAddress: "127.0.0.1",
		VSDPort:    8443,
		VSDMetadata: operv1.Metadata{
			Enterprise: "test",
			Domain:     "test",
			User:       "test",
			UserCert:   "test",
			UserKey:    "test",
		},
	}
	err = Parse(c)
	g.Expect(err).ToNot(HaveOccurred())

	c = &operv1.MonitorConfigDefinition{
		VSDAddress: "127.0.0.1000",
		VSDPort:    8443,
		VSDMetadata: operv1.Metadata{
			Enterprise: "test",
			Domain:     "test",
			User:       "test",
			UserCert:   "test",
			UserKey:    "test",
		},
	}
	err = Parse(c)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("validating monitor config failed"))
	g.Expect(err.Error()).To(ContainSubstring("invalid vsd ip address"))

	c = &operv1.MonitorConfigDefinition{
		VSDAddress: "127.0.0.1",
		VSDPort:    -1,
		VSDMetadata: operv1.Metadata{
			Enterprise: "test",
			Domain:     "test",
			User:       "test",
			UserCert:   "test",
			UserKey:    "test",
		},
	}
	err = Parse(c)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("validating monitor config failed"))
	g.Expect(err.Error()).To(ContainSubstring("invalid vsd port address"))

	c = &operv1.MonitorConfigDefinition{
		VSDAddress: "127.0.0.1",
		VSDPort:    8443,
		VSDMetadata: operv1.Metadata{
			Enterprise: "test",
			Domain:     "test",
			User:       "test",
			UserCert:   "test",
		},
	}
	err = Parse(c)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("validating monitor config failed"))
	g.Expect(err.Error()).To(ContainSubstring("vsd metadata validation failed"))
	g.Expect(err.Error()).To(ContainSubstring("user key cannot be empty"))
}
