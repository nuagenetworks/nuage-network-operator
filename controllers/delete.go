// Copyright 2020 Nokia
// Licensed under the Apache License 2.0.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeleteResource deletes object and pods if the resource exists
func (r *NuageCNIConfigReconciler) DeleteResource(nsn types.NamespacedName, obj runtime.Object) error {

	tmp := obj.DeepCopyObject()

	err := r.Client.Get(context.TODO(), nsn, tmp)
	if err != nil && strings.Contains(err.Error(), "not found") {
		return nil
	}
	err = r.Client.Delete(context.TODO(), obj)
	if err != nil && strings.Contains(err.Error(), "not found") {
		return nil
	}
	pod := &v1.Pod{}
	opts := []client.DeleteAllOfOption{
		client.InNamespace(nsn.Namespace),
		client.MatchingLabels{"k8s-app": nsn.Name},
	}
	err = r.Client.DeleteAllOf(context.TODO(), pod, opts...)
	if err != nil && strings.Contains(err.Error(), "not found") {
		return nil
	}
	return err
}
