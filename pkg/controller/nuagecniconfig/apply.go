package nuagecniconfig

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

//ApplyObject creates if it does not exist or updates it if it exists
func (r *ReconcileNuageCNIConfig) ApplyObject(nsn types.NamespacedName, obj runtime.Object) error {

	tmp := obj.DeepCopyObject()

	err := r.client.Get(context.TODO(), nsn, tmp)
	if err != nil && strings.Contains(err.Error(), "not found") {
		err = r.client.Create(context.TODO(), obj)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}

	err = r.client.Update(context.TODO(), obj)
	if err != nil {
		return err
	}
	return nil
}
