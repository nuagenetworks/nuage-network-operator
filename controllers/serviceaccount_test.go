// Copyright 2020 Nokia
// Licensed under the Apache License 2.0.
// SPDX-License-Identifier: Apache-2.0

package controllers

//import . "github.com/onsi/gomega"

//func TestServiceAccount(t *testing.T) {
//	expErr := fmt.Errorf("error")
//	g := NewGomegaWithT(t)
//	f := &fakeRestClient{
//		Client: fake.NewFakeClient(),
//	}
//	r := &ReconcileNuageCNIConfig{
//		client: f,
//	}
//	fun := func(a runtime.Object) {
//		err := r.Client.Get(context.TODO(), types.NamespacedName{Namespace: names.Namespace, Name: names.ServiceAccountName}, a)
//		g.Expect(err).To(HaveOccurred())
//		g.Expect(err.Error()).To(Equal(expErr.Error()))
//	}
//
//	sa, err := r.GetServiceAccount(names.ServiceAccountName, names.Namespace)
//	g.Expect(err).To(HaveOccurred())
//
//	err = r.CreateServiceAccount(names.ServiceAccountName, names.Namespace)
//	g.Expect(err).ToNot(HaveOccurred())
//
//	sa, err = r.GetServiceAccount(names.ServiceAccountName, names.Namespace)
//	g.Expect(err).ToNot(HaveOccurred())
//	g.Expect(sa.ObjectMeta.Name).To(Equal(names.ServiceAccountName))
//
//	f.GetFunc = func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
//		return expErr
//	}
//
//	fun(&corev1.ServiceAccount{})
//	fun(&corev1.Secret{})
//
//}

//func TestSecret(t *testing.T) {
//	g := NewGomegaWithT(t)
//	f := &fakeRestClient{
//		Client: fake.NewFakeClient(),
//	}
//	r := &ReconcileNuageCNIConfig{
//		client: f,
//	}
//
//	sec := &corev1.Secret{
//		TypeMeta: metav1.TypeMeta{
//			APIVersion: "v1",
//			Kind:       "Secret",
//		},
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      names.ServiceAccountName,
//			Namespace: names.Namespace,
//		},
//	}
//
//	err := r.Client.Create(context.TODO(), sec)
//	g.Expect(err).ToNot(HaveOccurred())
//
//	sec, err = r.GetSecret(names.ServiceAccountName, names.Namespace)
//	g.Expect(err).ToNot(HaveOccurred())
//	g.Expect(names.ServiceAccountName).To(Equal(sec.ObjectMeta.Name))
//
//	_, err = r.ExtractSecretToken(sec)
//	g.Expect(err).To(HaveOccurred())
//}
