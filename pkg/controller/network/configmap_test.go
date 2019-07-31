package network

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

	r := &ReconcileNetwork{
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

	r := &ReconcileNetwork{
		client: fake.NewFakeClient(),
	}

	rc, err := r.GetReleaseConfig()
	g.Expect(err).To(BeNil())
	g.Expect(rc).To(BeNil())

}
