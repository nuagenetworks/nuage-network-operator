// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package nuagecniconfig

import (
	"context"

	"github.com/nuagenetworks/nuage-network-operator/pkg/names"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

// DeleteResource deletes object and pods if the resource exists
func (r *ReconcileNuageCNIConfig) UpdateDaemonsetpods(nsn types.NamespacedName) error {
	daemonset := &appsv1.DaemonSet{}
	err := r.client.Get(context.TODO(), nsn, daemonset)
	if err != nil {
		return err
	}
	err = r.client.Update(context.TODO(), daemonset)
	if err != nil {
		return err
	}
	labelMap, err := metav1.LabelSelectorAsMap(daemonset.Spec.Selector)
	if err != nil {
		return err
	}

	options := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelMap).String(),
	}
	if pods, err := r.clientset.CoreV1().Pods(names.Namespace).List(options); err != nil {
		log.Errorf("List Pods of Daemonset[%s] error:%v", daemonset.GetName(), err)
		return err
	} else {
		for _, v := range pods.Items {
			log.Infof("Terminating pod %s on node %s to update daemonset %s", v.GetName(), v.Spec.NodeName, daemonset.GetName())
			po := &metav1.DeleteOptions{}
			if err := r.clientset.CoreV1().Pods(v.Namespace).Delete(v.Name, po); err != nil {
				log.Errorf("Failed to delete pod '%s in namespace %s': %v", v.Name, v.Namespace, err)
				return err
			}
		}
	}
	return nil
}
