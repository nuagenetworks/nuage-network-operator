// Copyright 2020 Nokia
// Licensed under the Apache License 2.0.
// SPDX-License-Identifier: Apache-2.0

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
