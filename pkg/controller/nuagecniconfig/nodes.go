// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package nuagecniconfig

import (
	"encoding/json"

	"github.com/nuagenetworks/nuage-network-operator/pkg/names"
	log "github.com/sirupsen/logrus"
	jsonpatch "gopkg.in/evanphx/json-patch.v4"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

//ListNodes fetches the list of nodes from api server that matches listOptions
func (r *ReconcileNuageCNIConfig) ListNodes(listOptions metav1.ListOptions) ([]corev1.Node, error) {

	nodes, err := r.clientset.CoreV1().Nodes().List(listOptions)
	if err != nil {
		return []corev1.Node{}, err
	}

	return nodes.Items, nil
}

//ListMasterNodes fetches the list of master nodes
func (r *ReconcileNuageCNIConfig) ListMasterNodes() ([]corev1.Node, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: "node-role.kubernetes.io/master",
	}

	return r.ListNodes(listOptions)
}

//LabelMasterNodes labels master nodes with nodeSelector if not already present
func (r *ReconcileNuageCNIConfig) LabelMasterNodes() error {
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

			_, err = r.clientset.CoreV1().Nodes().Patch(m.Name, types.MergePatchType, patch)
			if err != nil {
				log.Errorf("failed to add node selector label to %s: %v", m.Name, err)
			}
		}
	}

	return nil
}
