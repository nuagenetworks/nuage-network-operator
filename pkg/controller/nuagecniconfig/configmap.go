// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package nuagecniconfig

import (
	"context"
	"encoding/json"

	"github.com/nuagenetworks/nuage-network-operator/pkg/names"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var certConfig = types.NamespacedName{
	Namespace: names.Namespace,
	Name:      names.NuageCertConfig,
}

var monitConfig = types.NamespacedName{
	Namespace: names.Namespace,
	Name:      names.NuageMonitorConfig,
}

var releaseConfig = types.NamespacedName{
	Namespace: names.Namespace,
	Name:      names.NuageReleaseConfig,
}

//CreateConfigMap creates a config map on api server
func (r *ReconcileNuageCNIConfig) CreateConfigMap(nsn types.NamespacedName, data string) error {
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nsn.Name,
			Namespace: nsn.Namespace,
		},
		Data: map[string]string{"applied": data},
	}

	err := r.ApplyObject(nsn, cm)
	if err != nil {
		return err
	}

	return nil
}

//GetConfigMap get a config map from api server
func (r *ReconcileNuageCNIConfig) GetConfigMap(nsn types.NamespacedName) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), nsn, cm)
	if err != nil {
		return nil, err
	}

	return cm, nil
}

//SaveConfigToServer stores the applied release config in api server
func (r *ReconcileNuageCNIConfig) SaveConfigToServer(nsn types.NamespacedName, c interface{}) error {
	app, err := json.Marshal(c)
	if err != nil {
		return err
	}

	err = r.CreateConfigMap(nsn, string(app))
	if err != nil {
		return err
	}

	return nil
}

//GetConfigFromServer fetches the stored config from server
func (r *ReconcileNuageCNIConfig) GetConfigFromServer(nsn types.NamespacedName, c interface{}) error {
	cm, err := r.GetConfigMap(nsn)
	if err != nil && apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(cm.Data["applied"]), c)
	if err != nil {
		return err
	}
	return nil
}
