package render

import (
	"testing"

	operv1 "github.com/nuagenetworks/nuage-network-operator/pkg/apis/operator/v1alpha1"
	. "github.com/onsi/gomega"
)

var c = &operv1.RenderConfig{}

// TestRenderSimple tests rendering a single object with no templates
func TestRenderSimple(t *testing.T) {
	g := NewGomegaWithT(t)

	d := MakeRenderData(c)

	o1, err := RenderTemplate("testdata/simple.yaml", &d)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(o1).To(HaveLen(1))
	expected := `
{
	"apiVersion": "v1",
	"kind": "Pod",
	"metadata": {
		"name": "busybox1",
		"namespace": "ns"
	},
	"spec": {
		"containers": [
			{
  				"image": "busybox"
			}
		]
	}
}
`
	g.Expect(o1[0].MarshalJSON()).To(MatchJSON(expected))

	// test that json parses the same
	o2, err := RenderTemplate("testdata/simple.json", &d)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(o2).To(Equal(o1))
}

func TestRenderMultiple(t *testing.T) {
	g := NewGomegaWithT(t)

	p := "testdata/multiple.yaml"
	d := MakeRenderData(c)

	o, err := RenderTemplate(p, &d)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(o).To(HaveLen(3))

	g.Expect(o[0].GetObjectKind().GroupVersionKind().String()).To(Equal("/v1, Kind=Pod"))
	g.Expect(o[1].GetObjectKind().GroupVersionKind().String()).To(Equal("rbac.authorization.k8s.io/v1, Kind=ClusterRoleBinding"))
	g.Expect(o[2].GetObjectKind().GroupVersionKind().String()).To(Equal("/v1, Kind=ConfigMap"))
}

func TestTemplate(t *testing.T) {
	g := NewGomegaWithT(t)

	p := "testdata/template.yaml"

	// Test that missing variables are detected
	d := MakeRenderData(c)
	_, err := RenderTemplate(p, &d)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(HaveSuffix(`function "fname" not defined`))

	// Set expected function (but not variable)
	d.Funcs["fname"] = func(s string) string { return "test-" + s }
	_, err = RenderTemplate(p, &d)
	g.Expect(err).ToNot(HaveOccurred())

	// now we can render
	c.K8SAPIServerURL = "myns"
	o, err := RenderTemplate(p, &d)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(o[0].GetName()).To(Equal("test-podname"))
	g.Expect(o[0].GetNamespace()).To(Equal("myns"))

	var q int64
	q = 1
	g.Expect(o[0].Object["good"]).To(Equal(q))
	q = 0
	g.Expect(o[0].Object["bad"]).To(Equal(q))
}

func TestRenderDir(t *testing.T) {
	g := NewGomegaWithT(t)

	d := MakeRenderData(c)
	d.Funcs["fname"] = func(s string) string { return s }
	c.K8SAPIServerURL = "myns"
	c.VRSConfig.Controllers = []string{"b1", "b2"}

	o, err := RenderDir("testdata", &d)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(o).To(HaveLen(10))
}

func TestRenderDirOrder(t *testing.T) {
	g := NewGomegaWithT(t)

	d := MakeRenderData(c)
	d.Funcs["fname"] = func(s string) string { return s }
	c.K8SAPIServerURL = "myns"
	c.VRSConfig.Controllers = []string{"b1", "b2"}

	o, err := RenderDir("testdata", &d)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(o).To(HaveLen(10))
	g.Expect(o[0].GetObjectKind().GroupVersionKind().String()).To(Equal("/v1, Kind=ServiceAccount"))
	g.Expect(o[0].GetName()).To(Equal("nuage-cni"))
	g.Expect(o[1].GetObjectKind().GroupVersionKind().String()).To(Equal("rbac.authorization.k8s.io/v1, Kind=ClusterRole"))
	g.Expect(o[1].GetName()).To(Equal("nuage-cni"))
	g.Expect(o[2].GetObjectKind().GroupVersionKind().String()).To(Equal("rbac.authorization.k8s.io/v1, Kind=ClusterRoleBinding"))
	g.Expect(o[2].GetName()).To(Equal("nuage-cni"))

}

func TestRenderFile(t *testing.T) {
	g := NewGomegaWithT(t)

	p := "testdata/array.yaml"
	d := MakeRenderData(c)
	d.Funcs["fname"] = func(s string) string { return s }
	c.K8SAPIServerURL = "myns"
	c.VRSConfig.Controllers = []string{"b1", "b2"}
	o, err := RenderTemplate(p, &d)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(o[0].GetName()).To(Equal("test-podname"))
	g.Expect(o[0].GetNamespace()).To(Equal("myns"))

	c.VRSConfig.Controllers = []string{"b1"}
	o, err = RenderTemplate(p, &d)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(o[0].GetName()).To(Equal("test-podname"))
	g.Expect(o[0].GetNamespace()).To(Equal("myns"))
}
