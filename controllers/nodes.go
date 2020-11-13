// Copyright 2020 Nokia
// Licensed under the Apache License 2.0.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"encoding/json"

	"github.com/nuagenetworks/nuage-network-operator/controllers/names"
	log "github.com/sirupsen/logrus"
	jsonpatch "gopkg.in/evanphx/json-patch.v4"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

//ListNodes fetches the list of nodes from api server that matches listOptions
func (r *NuageCNIConfigReconciler) ListNodes(listOptions metav1.ListOptions) ([]corev1.Node, error) {

	nodes, err := r.clientset.CoreV1().Nodes().List(context.TODO(), listOptions)
	if err != nil {
		return []corev1.Node{}, err
	}

	return nodes.Items, nil
}

//ListMasterNodes fetches the list of master nodes
func (r *NuageCNIConfigReconciler) ListMasterNodes() ([]corev1.Node, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: "node-role.kubernetes.io/master",
	}

	return r.ListNodes(listOptions)
}

//LabelMasterNodes labels master nodes with nodeSelector if not already present
func (r *NuageCNIConfigReconciler) LabelMasterNodes() error {
	masters, err := r.ListMasterNodes()
	if err != nil {
		return err
	}

	for _, m := range masters {
		if _, ok := m.Labels[names.MasterNodeSelector]; !ok {
			oldData, _ := json.Marshal(m)
			m.Labels[names.MasterNodeSelector] = ""
			newData, _ := json.Marshal(m)

			patch, err := jsonpatch.CreateMergePatch(oldData, newData)
			if err != nil {
				log.Errorf("creating patch failed: %v", err)
				continue
			}

			_, err = r.clientset.CoreV1().Nodes().Patch(context.TODO(), m.Name, types.MergePatchType, patch, metav1.PatchOptions{})
			if err != nil {
				log.Errorf("failed to add node selector label to %s: %v", m.Name, err)
			}
		}
	}

	return nil
}
