// Copyright 2020 Nokia
// Licensed under the Apache License 2.0.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"testing"

	operv1 "github.com/nuagenetworks/nuage-network-operator/api/v1alpha1"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestAll(t *testing.T) {
	g := NewGomegaWithT(t)

	r := &NuageCNIConfigReconciler{
		Client: fake.NewFakeClient(),
	}

	rc := &operv1.ReleaseConfigDefinition{
		CNITag: "abc",
	}

	err := r.SaveConfigToServer(releaseConfig, rc)
	g.Expect(err).ToNot(HaveOccurred())

	cm := &corev1.ConfigMap{}
	err = r.Client.Get(context.TODO(), releaseConfig, cm)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cm.Data["applied"]).To(ContainSubstring("abc"))

	rc.CNITag = "def"
	err = r.SaveConfigToServer(releaseConfig, rc)
	g.Expect(err).ToNot(HaveOccurred())

	cm = &corev1.ConfigMap{}
	err = r.Client.Get(context.TODO(), releaseConfig, cm)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cm.Data["applied"]).To(ContainSubstring("def"))

	rc = &operv1.ReleaseConfigDefinition{}
	err = r.GetConfigFromServer(releaseConfig, rc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(rc.CNITag).To(Equal("def"))
}

func TestGet(t *testing.T) {
	g := NewGomegaWithT(t)

	r := &NuageCNIConfigReconciler{
		Client: fake.NewFakeClient(),
	}

	rc := &operv1.ReleaseConfigDefinition{}
	err := r.GetConfigFromServer(releaseConfig, rc)
	g.Expect(err).To(BeNil())
	g.Expect(len(rc.VRSTag)).To(Equal(0))

}
