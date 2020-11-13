// Copyright 2020 Nokia
// Licensed under the Apache License 2.0.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

////CreateServiceAccount creates a service account
//func (r *ReconcileNuageCNIConfig) CreateServiceAccount(name, namespace string) error {
//	sa := &corev1.ServiceAccount{
//		TypeMeta: metav1.TypeMeta{
//			APIVersion: "v1",
//			Kind:       "ServiceAccount",
//		},
//		ObjectMeta: metav1.ObjectMeta{
//			Namespace: namespace,
//			Name:      name,
//		},
//	}
//
//	err := r.ApplyObject(types.NamespacedName{
//		Namespace: namespace,
//		Name:      name,
//	}, sa)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

////GetServiceAccount fetch the service account
//func (r *ReconcileNuageCNIConfig) GetServiceAccount(name, namespace string) (*corev1.ServiceAccount, error) {
//	sa, err := r.Clientset.CoreV1().ServiceAccounts(namespace).Get(name, metav1.GetOptions{})
//	if err != nil {
//		return nil, err
//	}
//
//	log.Errorf("%v", sa)
//	return sa, nil
//}

//GetSecret fetches secret from api server
func (r *NuageCNIConfigReconciler) GetSecret(saname, namespace string) (*corev1.Secret, error) {

	secList, err := r.clientset.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, sec := range secList.Items {
		if sec.ObjectMeta.Annotations["kubernetes.io/service-account.name"] == saname {
			return &sec, nil
		}
	}

	return nil, fmt.Errorf("could not find the secret")
}

//ExtractSecretToken extract the token from the secret
func (r *NuageCNIConfigReconciler) ExtractSecretToken(s *corev1.Secret) ([]byte, error) {
	token, ok := s.Data[corev1.ServiceAccountTokenKey]
	if !ok {
		return []byte{}, fmt.Errorf("could find key %s in secret %v", corev1.ServiceAccountTokenKey, s)
	}

	return token, nil
}

////ExtractSecrets extracts secret name from service account yaml
//func (r *ReconcileNuageCNIConfig) ExtractSecrets(sa *corev1.ServiceAccount) []corev1.ObjectReference {
//	return sa.Secrets
//}
