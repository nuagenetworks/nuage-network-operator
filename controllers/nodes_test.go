// Copyright 2020 Nokia
// Licensed under the Apache License 2.0.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var g *GomegaWithT
var r *NuageCNIConfigReconciler
var exp []*corev1.Node

func initData(t *testing.T) {
	g = NewGomegaWithT(t)
	exp = []*corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node1",
				Labels: map[string]string{
					"node-role.kubernetes.io/master": "",
				},
			},
		},
		{ObjectMeta: metav1.ObjectMeta{Name: "node2"}},
	}

	clientset := fake.NewSimpleClientset(exp[0], exp[1])

	r = &NuageCNIConfigReconciler{
		clientset: clientset,
	}

}

func TestNodesList(t *testing.T) {
	initData(t)

	obs, err := r.ListNodes(metav1.ListOptions{})
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(len(obs)).To(Equal(len(exp)))
}

func TestNodesListMasters(t *testing.T) {
	initData(t)

	obs, err := r.ListMasterNodes()
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(len(obs)).To(Equal(1))
}

func TestNodesLabelMasters(t *testing.T) {
	initData(t)

	err := r.LabelMasterNodes()
	g.Expect(err).ToNot(HaveOccurred())
}
