package nuagecniconfig

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	apiServerError string = "api server error"
)

//Wrapper over runtime client.
type fakeRestClient struct {
	client     client.Client
	GetFunc    func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	ListFunc   func(ctx context.Context, opts *client.ListOptions, list runtime.Object) error
	CreateFunc func(ctx context.Context, obj runtime.Object) error
	DeleteFunc func(ctx context.Context, obj runtime.Object, opts ...client.DeleteOptionFunc) error
	UpdateFunc func(ctx context.Context, obj runtime.Object) error
	StatusFunc func() client.StatusWriter
}

func (f *fakeRestClient) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	if f.GetFunc != nil {
		return f.GetFunc(ctx, key, obj)
	}

	return f.client.Get(ctx, key, obj)
}

func (f *fakeRestClient) List(ctx context.Context, opts *client.ListOptions, list runtime.Object) error {
	if f.ListFunc != nil {
		return f.ListFunc(ctx, opts, list)
	}

	return f.client.List(ctx, opts, list)
}

func (f *fakeRestClient) Create(ctx context.Context, obj runtime.Object) error {
	if f.CreateFunc != nil {
		return f.CreateFunc(ctx, obj)
	}

	return f.client.Create(ctx, obj)
}

func (f *fakeRestClient) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOptionFunc) error {
	if f.DeleteFunc != nil {
		return f.DeleteFunc(ctx, obj, opts...)
	}

	return f.client.Delete(ctx, obj, opts...)
}

func (f *fakeRestClient) Update(ctx context.Context, obj runtime.Object) error {
	if f.UpdateFunc != nil {
		return f.UpdateFunc(ctx, obj)
	}

	return f.client.Update(ctx, obj)
}

func (f *fakeRestClient) Status() client.StatusWriter {
	if f.StatusFunc != nil {
		return f.StatusFunc()
	}

	return f.client.Status()
}
