// Copyright 2020 Nokia
// Licensed under the Apache License 2.0.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

//ApplyObject creates if it does not exist or updates it if it exists
func (r *NuageCNIConfigReconciler) ApplyObject(nsn types.NamespacedName, obj runtime.Object) error {

	tmp := obj.DeepCopyObject()

	err := r.Client.Get(context.TODO(), nsn, tmp)
	if err != nil && strings.Contains(err.Error(), "not found") {
		err = r.Client.Create(context.TODO(), obj)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}

	err = r.Client.Update(context.TODO(), obj)
	if err != nil {
		return err
	}
	return nil
}
