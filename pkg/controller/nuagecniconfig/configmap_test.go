package nuagecniconfig

import (
	"context"
	"testing"

	operv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestAll(t *testing.T) {
	g := NewGomegaWithT(t)

	r := &ReconcileNuageCNIConfig{
		client: fake.NewFakeClient(),
	}

	rc := &operv1.ReleaseConfigDefinition{
		CNITag: "abc",
	}

	err := r.SetReleaseConfig(rc)
	g.Expect(err).ToNot(HaveOccurred())

	cm := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), nsn, cm)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cm.Data["applied"]).To(ContainSubstring("abc"))

	rc.CNITag = "def"
	err = r.SetReleaseConfig(rc)
	g.Expect(err).ToNot(HaveOccurred())

	cm = &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), nsn, cm)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cm.Data["applied"]).To(ContainSubstring("def"))

	rc, err = r.GetReleaseConfig()
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(rc).ToNot(BeNil())
	g.Expect(rc.CNITag).To(Equal("def"))
}

func TestGet(t *testing.T) {
	g := NewGomegaWithT(t)

	r := &ReconcileNuageCNIConfig{
		client: fake.NewFakeClient(),
	}

	rc, err := r.GetReleaseConfig()
	g.Expect(err).To(BeNil())
	g.Expect(rc).To(BeNil())

}

func TestIsDiffConfig(t *testing.T) {
	g := NewGomegaWithT(t)

	r := &ReconcileNuageCNIConfig{
		client: fake.NewFakeClient(),
	}

	prev := &operv1.ReleaseConfigDefinition{}
	curr := &operv1.ReleaseConfigDefinition{}

	p, err := r.IsDiffConfig(prev, curr)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(p).To(Equal(0))

	prev.CNITag = "0.0.0"
	curr.CNITag = "0.0.1"
	p, err = r.IsDiffConfig(prev, curr)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(p).To(Equal(-1))

	prev.CNITag = "0.0.1"
	curr.CNITag = "0.0.0"
	p, err = r.IsDiffConfig(prev, curr)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(p).To(Equal(1))
}
