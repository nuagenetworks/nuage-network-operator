// Copyright 2020 Nokia
// Licensed under the Apache License 2.0.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"github.com/nuagenetworks/nuage-network-operator/controllers/names"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

// DeleteResource deletes object and pods if the resource exists
func (r *NuageCNIConfigReconciler) UpdateDaemonsetpods(nsn types.NamespacedName) error {
	daemonset := &appsv1.DaemonSet{}
	err := r.Client.Get(context.TODO(), nsn, daemonset)
	if err != nil {
		return err
	}
	err = r.Client.Update(context.TODO(), daemonset)
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
	if pods, err := r.clientset.CoreV1().Pods(names.Namespace).List(context.TODO(), options); err != nil {
		log.Errorf("List Pods of Daemonset[%s] error:%v", daemonset.GetName(), err)
		return err
	} else {
		for _, v := range pods.Items {
			log.Infof("Terminating pod %s on node %s to update daemonset %s", v.GetName(), v.Spec.NodeName, daemonset.GetName())
			if err := r.clientset.CoreV1().Pods(v.Namespace).Delete(context.TODO(), v.Name, metav1.DeleteOptions{}); err != nil {
				log.Errorf("Failed to delete pod '%s in namespace %s': %v", v.Name, v.Namespace, err)
				return err
			}
		}
	}
	return nil
}
