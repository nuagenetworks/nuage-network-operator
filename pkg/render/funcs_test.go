// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package render

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestAddEscapeChar(t *testing.T) {
	g := NewGomegaWithT(t)
	orig := "0.0.0.0/0"
	exp := `0.0.0.0\\/0`
	got := addEscapeChar(orig)
	g.Expect(exp).To(Equal(got))

	orig = "nothing"
	exp = "nothing"
	got = addEscapeChar(orig)
	g.Expect(exp).To(Equal(got))
}
