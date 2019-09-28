package nuagecniconfig

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

//CreateServiceAccount creates a service account
func (r *ReconcileNuageCNIConfig) CreateServiceAccount(name, namespace string) error {
	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}

	err := r.ApplyObject(types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, sa)
	if err != nil {
		return err
	}

	return nil
}

//GetServiceAccount fetch the service account
func (r *ReconcileNuageCNIConfig) GetServiceAccount(name, namespace string) (*corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{}
	san := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}

	err := r.client.Get(context.TODO(), san, sa)
	if err != nil {
		return nil, err
	}

	return sa, nil
}

//GetSecret fetches secret from api server
func (r *ReconcileNuageCNIConfig) GetSecret(name, namespace string) (*corev1.Secret, error) {
	s := &corev1.Secret{}
	ns := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}

	err := r.client.Get(context.TODO(), ns, s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

//ExtractSecretToken extract the token from the secret
func (r *ReconcileNuageCNIConfig) ExtractSecretToken(s *corev1.Secret) ([]byte, error) {
	token, ok := s.Data[corev1.ServiceAccountTokenKey]
	if !ok {
		return []byte{}, fmt.Errorf("could find key %s in secret %v", corev1.ServiceAccountTokenKey, s)
	}

	return token, nil
}

//ExtractSecrets extracts secret name from service account yaml
func (r *ReconcileNuageCNIConfig) ExtractSecrets(sa *corev1.ServiceAccount) []corev1.ObjectReference {
	return sa.Secrets
}
