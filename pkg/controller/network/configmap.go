package network

import (
	"bytes"
	"context"
	"encoding/json"

	operv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	"github.com/nuagenetworks/nuage-network-operator/pkg/names"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var nsn = types.NamespacedName{
	Namespace: names.Namespace,
	Name:      names.ConfigName,
}

// GetReleaseConfig fetches the previous applied release config
func (r *ReconcileNetwork) GetReleaseConfig() (*operv1.ReleaseConfigDefinition, error) {
	cm := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), nsn, cm)
	if err != nil && apierrors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	c := &operv1.ReleaseConfigDefinition{}
	err = json.Unmarshal([]byte(cm.Data["applied"]), c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

//SetReleaseConfig stores the applied release config in api server
func (r *ReconcileNetwork) SetReleaseConfig(c *operv1.ReleaseConfigDefinition) error {
	app, err := json.Marshal(c)
	if err != nil {
		return err
	}
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: names.Namespace,
			Name:      names.ConfigName,
		},
		Data: map[string]string{
			"applied": string(app),
		},
	}

	err = r.client.Get(context.TODO(), nsn, cm)
	if err != nil && apierrors.IsNotFound(err) {
		err = r.client.Create(context.TODO(), cm)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}

	cm.Data["applied"] = string(app)
	err = r.client.Update(context.TODO(), cm)
	if err != nil {
		return err
	}

	return nil
}

//IsDiffConfig return 0 if both the configs are same
func (r *ReconcileNetwork) IsDiffConfig(prev, curr *operv1.ReleaseConfigDefinition) (int, error) {
	var s1, s2 []byte
	var err error

	s1, err = json.Marshal(prev)
	if err != nil {
		return -1, err
	}
	s2, err = json.Marshal(curr)
	if err != nil {
		return -1, err
	}

	return bytes.Compare(s1, s2), err
}
