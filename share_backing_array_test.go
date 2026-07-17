// Copyright 2026 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package caps_test

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/gcustom"
	"github.com/onsi/gomega/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// ShareBackingArrayWith matches if actual is a slice that has the same backing
// array as the expected slice. If one or both slices are nil,
// ShareBackingArrayWith never matches. ShareBackingArrayWith will fail with an
// error in case the actual slice is not of the same type as the expected slice.
//
// Note: the slice headers can be different and we don't care, but we care if
// the backing arrays are the same, as this means the different slices (slice
// headers) share (some of) the underlying elements.
func ShareBackingArrayWith[S ~[]E, E any](expected S) types.GomegaMatcher {
	return gcustom.MakeMatcher(func(actual S) (bool, error) {
		return shareBackingArrayWith(actual, expected)
	}).WithTemplate(
		"Expected:\n" +
			"    {{if .Actual}}{{.Actual.String}}{{else}}<nil>{{end}}\n" +
			"{{.To}} have the same capabilities set backing array as:\n" +
			"    {{if .Data}}{{.Data.String}}{{else}}<nil>{{end}}",
	).WithTemplateData(expected)
}

func shareBackingArrayWith[S ~[]E, E any](actual, expected S) (bool, error) {
	// nil slices do not have a backing array and thus cannot share it.
	if actual == nil || expected == nil {
		return false, nil
	}

	expectedT := reflect.TypeOf(expected)
	actualT := reflect.TypeOf(actual)
	if actualT != expectedT {
		return false, fmt.Errorf("actual and expected must be the same slice type, expected: %T, got: %T",
			expected, actual)
	}

	expectedV := reflect.ValueOf(expected)
	actualV := reflect.ValueOf(actual)
	if actualV.Cap() == 0 || expectedV.Cap() == 0 {
		return false, nil
	}

	elemsize := actualT.Elem().Size()

	actualStart := actualV.Pointer()
	actualEnd := actualStart + uintptr(actualV.Cap())*elemsize

	expectedStart := expectedV.Pointer()
	expectedEnd := expectedStart + uintptr(expectedV.Cap())*elemsize

	return actualStart < expectedEnd && expectedStart < actualEnd, nil
}

var _ = Describe("slice backing array matchers", func() {

	It("rejects slices of different types", func() {
		m := ShareBackingArrayWith([]int{})
		Expect(m.Match([]string{})).Error().To(
			MatchError(ContainSubstring("expected actual of type <[]int>.")))
	})

	It("returns false if one or both slices are without backing array", func() {
		zeroBacking := make([]int, 0)
		m := ShareBackingArrayWith(zeroBacking)
		Expect(m.Match([]int{42})).To(BeFalse())
		Expect(m.Match(zeroBacking)).To(BeFalse())

		m = ShareBackingArrayWith([]int{666})
		Expect(m.Match(zeroBacking)).To(BeFalse())
	})

	It("return false in case of nil slices", func() {
		m := ShareBackingArrayWith([]int(nil))
		Expect(m.Match([]int{})).To(BeFalse())

		m = ShareBackingArrayWith([]string{})
		Expect(m.Match(nil)).To(BeFalse())
	})

	It("returns true if the backing arrays are the same", func() {
		s := []int{1, 42, 666}
		m := ShareBackingArrayWith(s)
		Expect(m.Match(s)).To(BeTrue())
		Expect(m.Match(s[1:2])).To(BeTrue())
	})

})
