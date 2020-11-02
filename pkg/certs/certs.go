// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	operv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var certlog = logf.Log.WithName("certs")

// GenerateCertificates generate certs for client server authentication
func GenerateCertificates(config *operv1.CertGenConfig) (*operv1.TLSCertificates, error) {
	fillDefaults(config)

	priv, err := GeneratePrivateKey(config)
	if err != nil {
		certlog.Error(err, "private key generation failed")
		return nil, err
	}

	template, err := GenerateCertificateTemplate(config)
	if err != nil {
		certlog.Error(err, "Generating certificate template failed")
		return nil, err
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, publicKey(priv), priv)
	if err != nil {
		certlog.Error(err, "Failed to create certificate")
		return nil, err
	}

	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if cert == nil {
		err := fmt.Errorf("failed to encode cert to memory")
		certlog.Error(err, "cert creation failed")
		return nil, err
	}

	pemBlock, err := pemBlockForKey(priv)
	if err != nil {
		certlog.Error(err, "converting private key to pem block failed")
		return nil, err
	}

	key := pem.EncodeToMemory(pemBlock)
	if key == nil {
		err := fmt.Errorf("failed to encode key to memory")
		certlog.Error(err, "Key creation failed")
		return nil, err
	}

	certificate := string(cert)
	privatekey := string(key)

	return &operv1.TLSCertificates{
		CA:          &certificate,
		Certificate: &certificate,
		PrivateKey:  &privatekey,
	}, nil

}

// GenerateCertificateTemplate generates certificate  template
func GenerateCertificateTemplate(config *operv1.CertGenConfig) (*x509.Certificate, error) {
	var err error
	var notBefore time.Time
	var notAfter time.Time

	fillDefaults(config)

	if len(*config.ValidFrom) == 0 {
		notBefore = time.Now()
	} else {
		notBefore, err = time.Parse("Jan 2 15:04:05 2006", *config.ValidFrom)
		if err != nil {
			certlog.Error(err, "Failed to parse creation date")
			return nil, err
		}
	}

	if config.ValidFor == time.Duration(0) {
		fmt.Printf("duration not set\n")
		notAfter = notBefore.Add(365 * 24 * time.Hour)
	} else {
		fmt.Printf("duration set\n")
		notAfter = notBefore.Add(config.ValidFor)
	}
	fmt.Printf("%s\n", notAfter.String())

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		certlog.Error(err, "failed to generate serial number")
		return nil, err
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Nuage Networks"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	return template, nil
}

// GeneratePrivateKey generates a private key
func GeneratePrivateKey(config *operv1.CertGenConfig) (interface{}, error) {

	var priv interface{}
	var err error

	fillDefaults(config)

	switch *config.ECDSACurve {
	case "":
		fallthrough
	case "rsa":
		priv, err = rsa.GenerateKey(rand.Reader, config.RSABits)
	case "P224":
		priv, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		priv, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		priv, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		return nil, fmt.Errorf("Unrecognized elliptic curve %q", *config.ECDSACurve)
	}
	if err != nil {
		certlog.Error(err, "failed to generate private key")
		return nil, err
	}
	return priv, nil
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) (*pem.Block, error) {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(k),
		}, nil
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			certlog.Error(err, "Unable to marshal ECDSA private key")
			return nil, err
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}, nil
	default:
		return nil, fmt.Errorf("unknown key type passed")
	}
}

func fillDefaults(config *operv1.CertGenConfig) {
	curve := "rsa"
	validFrom := ""

	if config.RSABits == 0 {
		config.RSABits = 2048
	}
	if config.ECDSACurve == nil {
		config.ECDSACurve = &curve
	}
	if config.ValidFrom == nil {
		config.ValidFrom = &validFrom
	}
}
