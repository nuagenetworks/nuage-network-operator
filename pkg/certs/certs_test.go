package certs

import (
	"testing"
	"time"

	operv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	. "github.com/onsi/gomega"
)

func TestGeneratePrivateKey(t *testing.T) {
	g := NewGomegaWithT(t)

	config := &operv1.CertGenConfig{
		RSABits: 2048,
	}
	for _, curve := range []string{"", "rsa", "P224", "P256", "P384", "P521"} {
		config.ECDSACurve = &curve
		_, err := GeneratePrivateKey(config)
		g.Expect(err).NotTo(HaveOccurred())
	}
	for _, curve := range []string{"abc", "ra", "224"} {
		config.ECDSACurve = &curve
		priv, err := GeneratePrivateKey(config)
		g.Expect(err).To(HaveOccurred())
		g.Expect(priv).To(BeNil())
	}

	curve := ""
	config.RSABits = 0
	config.ECDSACurve = &curve
	priv, err := GeneratePrivateKey(config)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(priv).ToNot(BeNil())
}

func TestGenerateCertificateTemplate(t *testing.T) {
	g := NewGomegaWithT(t)

	validFrom := ""
	config := &operv1.CertGenConfig{
		ValidFrom: &validFrom,
		ValidFor:  time.Hour,
	}

	cert, err := GenerateCertificateTemplate(config)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cert.SerialNumber).NotTo(BeZero())
	g.Expect(cert.NotBefore).Should(BeTemporally("~", time.Now()))
	g.Expect(cert.NotAfter).Should(BeTemporally("~", time.Now().Add(time.Hour)))
	g.Expect(cert.IsCA).Should(BeTrue())
}

func TestGenerateCertificates(t *testing.T) {
	g := NewGomegaWithT(t)

	validFrom := ""
	curve := "rsa"
	config := &operv1.CertGenConfig{
		ValidFrom:  &validFrom,
		ValidFor:   time.Hour,
		ECDSACurve: &curve,
		RSABits:    2048,
	}

	c, err := GenerateCertificates(config)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(len(*c.CA)).ShouldNot(BeZero())
	g.Expect(len(*c.Certificate)).ShouldNot(BeZero())
	g.Expect(len(*c.PrivateKey)).ShouldNot(BeZero())

	config2 := &operv1.CertGenConfig{}
	c, err = GenerateCertificates(config2)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(len(*c.CA)).ShouldNot(BeZero())
	g.Expect(len(*c.Certificate)).ShouldNot(BeZero())
	g.Expect(len(*c.PrivateKey)).ShouldNot(BeZero())
}
