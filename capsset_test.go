// Copyright 2023 Harald Albrecht.
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

package caps

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("capabilities sets", func() {

	It("has no capability gaps", func() {
		for capno := 0; capno <= MaxCapabilityNumber; capno++ {
			Expect(CapabilityNameByNumber).To(HaveKey(capno))
		}
	})

	DescribeTable("anonymous capabilities",
		func(name string, isAnonymous bool) {
			Expect(isAnonymousCapability(name)).To(Equal(isAnonymous))
		},
		Entry(nil, "8BALL", false),
		Entry(nil, "CAP_FOO_BAR", false),
		Entry(nil, "CAP_8BALL", false),
		Entry(nil, "CAP_666", true),
	)

	DescribeTable("sorting capabilities",
		func(a, b string, order int) {
			Expect(cmpCapName(a, b)).To(Equal(order))
		},
		Entry(nil, "CAP_FOO_BAR", "CAP_ZOO", -1),
		Entry(nil, "CAP_FOO_BAR", "CAP_BAR", 1),
		Entry(nil, "CAP_FOO_BAR", "CAP_42", -1),
		Entry(nil, "CAP_42", "CAP_FOO", 1),
		Entry(nil, "CAP_42", "CAP_88", -1),
		Entry(nil, "CAP_42", "CAP_42", 0),
		Entry(nil, "CAP_100", "CAP_99", -1), // sic!
	)

	It("sets all capabilities", func() {
		max := LastCapability()
		Expect(max).NotTo(BeZero())

		caps := AllCapabilities()
		Expect(caps).To(HaveLen(max/32 + 1))
		Expect(caps[max/32]).To(Equal((^uint32(0)) >> (31 - max%32)))
	})

	It("adds and drops capabilities", func() {
		caps := NewCapabilitiesSet()
		caps.Add(CAP_SYS_ADMIN, CAP_SYS_CHROOT, CAP_BPF)
		Expect(caps).To(Equal(CapabilitiesSet([]uint32{0x00240000, 0x00000080})))
		caps.Drop(CAP_SYS_ADMIN)
		Expect(caps).To(Equal(CapabilitiesSet([]uint32{0x00040000, 0x00000080})))
		caps.Drop(CAP_SYS_CHROOT)
		Expect(caps).To(Equal(CapabilitiesSet([]uint32{0x00000000, 0x00000080})))
		caps.Drop(CAP_SYS_CHROOT)
		Expect(caps).To(Equal(CapabilitiesSet([]uint32{0x00000000, 0x00000080})))
	})

	It("drops dropped caps without enlarging the set", func() {
		caps := NewCapabilitiesSet()
		caps.Drop(CAP_SYS_ADMIN)
		Expect(caps).To(HaveLen(0))
	})

	It("clears all capabilities", func() {
		caps := NewCapabilitiesSet()
		caps.Add(CAP_SYS_ADMIN)
		caps.Clear()
		Expect(caps.Has(CAP_SYS_ADMIN)).To(BeFalse())
	})

	It("tests capabilities", func() {
		caps := NewCapabilitiesSet()
		caps.Add(CAP_SYS_ADMIN, CAP_SYS_CHROOT)
		Expect(caps.Has(CAP_SYS_ADMIN)).To(BeTrue())
		Expect(caps.Has(CAP_BPF)).To(BeFalse())
	})

	It("panics for negative capability number", func() {
		caps := NewCapabilitiesSet()
		Expect(func() {
			caps.Add(-1)
		}).To(Panic())
	})

	It("clones a set", func() {
		caps := NewCapabilitiesSet()
		caps.Add(CAP_SYS_ADMIN, CAP_SYS_CHROOT)
		capsclone := caps.Clone()
		Expect(capsclone).To(Equal(caps))
		caps.Drop(CAP_SYS_ADMIN)
		Expect(capsclone).NotTo(Equal(caps))
	})

	It("returns capability names set ordered by capability number", func() {
		caps := NewCapabilitiesSet()
		caps.Add(CAP_SYS_ADMIN, CAP_SYS_CHROOT, MaxCapabilityNumber+1)
		Expect(caps.Names()).To(ConsistOf([]string{
			"CAP_SYS_ADMIN", "CAP_SYS_CHROOT", fmt.Sprintf("CAP_%d", MaxCapabilityNumber+1),
		}))
	})

	It("returns a lexicographically sorted list of capability names", func() {
		caps := NewCapabilitiesSet()
		caps.Add(CAP_NET_ADMIN, CAP_SYS_ADMIN, CAP_SYS_CHROOT)
		Expect(caps.String()).To(Equal("CAP_NET_ADMIN, CAP_SYS_ADMIN, CAP_SYS_CHROOT"))

		caps.Add(MaxCapabilityNumber + 1)
		Expect(caps.SortedNames()).To(ConsistOf(
			"CAP_NET_ADMIN",
			"CAP_SYS_ADMIN",
			"CAP_SYS_CHROOT",
			fmt.Sprintf("CAP_%d", MaxCapabilityNumber+1)))
	})

	It("returns correct hexadecimal representation", func() {
		Expect(CapabilitiesSet{}.Hex()).To(
			Equal(strings.Repeat("00000000", capDataElements)))
		caps := CapabilitiesSet{}
		caps.Add(CAP_SYS_ADMIN)
		Expect(caps.Hex()).To(HaveSuffix("00200000"))
	})

	It("parses the hexadecimal capability set representation", func() {
		caps := Successful(CapabilitiesFromHex("00"))
		Expect(caps).To(Equal(CapabilitiesSet{0x0}))

		caps = Successful(CapabilitiesFromHex("80002001"))
		Expect(caps).To(Equal(CapabilitiesSet{0x80002001}))

		caps = Successful(CapabilitiesFromHex("1180002001"))
		Expect(caps).To(Equal(CapabilitiesSet{0x80002001, 0x11}))
	})

	It("returns errors for invalid hexadecimal capability set representations", func() {
		Expect(CapabilitiesFromHex("0")).Error().To(HaveOccurred())
		Expect(CapabilitiesFromHex("abcdefg")).Error().To(HaveOccurred())
	})

})
