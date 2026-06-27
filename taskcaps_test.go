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

package caps_test

import (
	"fmt"
	"os"
	"runtime"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/thediveo/caps/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("task capabilities", func() {

	When("retrieving task capabilities", func() {

		It("gets the current task's capabilities", func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}
			Expect(Successful(caps.OfTask(0)).IsEmpty()).NotTo(BeTrue())
			Expect(Successful(caps.OfCurrentTask()).IsEmpty()).NotTo(BeTrue())
			Expect(caps.OfCurrentTaskOrZero().IsEmpty()).NotTo(BeTrue())
		})

		It("returns an error for a non-existing task", func() {
			Expect(caps.OfTask(-1)).Error().To(MatchError(syscall.EINVAL))
		})

		It("returns zero capabilities for non-existing task", func() {
			Expect(caps.OfTaskOrZero(-1).IsEmpty()).To(BeTrue())
		})

	})

	It("self tests", func() {
		tc := caps.TaskCapabilities{}.Effective().Add(caps.CAP_SYS_ADMIN).TaskCapabilities
		Expect(tc.Eff()).To(ShareBackingArrayWith(tc.Eff()))
	})

	It("indicates if no capabilities at all", func() {
		tc := caps.TaskCapabilities{}
		Expect(tc.IsEmpty()).To(BeTrue())
		Expect(tc.Permitted().Add(caps.CAP_SYS_ADMIN).TaskCapabilities.IsEmpty()).To(BeFalse())
	})

	It("prints the task's capabilities", func() {
		Expect(caps.TaskCapabilities{}.String()).To(Equal("{effective:;permitted:;inheritable:}"))
		tc := caps.TaskCapabilities{}.
			Effective().Add(caps.CAP_SYS_ADMIN).
			Permitted().Add(caps.CAP_AUDIT_READ).
			Inheritable().Add(caps.CAP_BPF)
		Expect(tc.String()).To(Equal("{effective:CAP_SYS_ADMIN;permitted:CAP_AUDIT_READ;inheritable:CAP_BPF}"))
	})

	When("mutating, it returns independent new values", func() {

		It("returns a new clone", func() {
			before := caps.TaskCapabilities{}
			after := before.Clone()
			Expect(after.Eff()).NotTo(ShareBackingArrayWith(before.Eff()))
			Expect(after.Perm()).NotTo(ShareBackingArrayWith(before.Perm()))
			Expect(after.Inh()).NotTo(ShareBackingArrayWith(before.Inh()))
		})

		It("adds a capability and returns a new value", func() {
			before := caps.TaskCapabilities{}
			after := before.Effective().Add(caps.CAP_SYS_ADMIN).TaskCapabilities
			Expect(after.Eff()).NotTo(ShareBackingArrayWith(before.Eff()))
			Expect(after.Perm()).NotTo(ShareBackingArrayWith(before.Perm()))
			Expect(after.Inh()).NotTo(ShareBackingArrayWith(before.Inh()))

			after = before.Permitted().Add(caps.CAP_SYS_ADMIN).TaskCapabilities
			Expect(after.Eff()).NotTo(ShareBackingArrayWith(before.Eff()))
			Expect(after.Perm()).NotTo(ShareBackingArrayWith(before.Perm()))
			Expect(after.Inh()).NotTo(ShareBackingArrayWith(before.Inh()))

			after = before.Inheritable().Add(caps.CAP_SYS_ADMIN).TaskCapabilities
			Expect(after.Eff()).NotTo(ShareBackingArrayWith(before.Eff()))
			Expect(after.Perm()).NotTo(ShareBackingArrayWith(before.Perm()))
			Expect(after.Inh()).NotTo(ShareBackingArrayWith(before.Inh()))
		})

		It("drops a capability and returns a new value", func() {
			before := caps.TaskCapabilities{}.Effective().All().TaskCapabilities
			after := before.Effective().Drop(caps.CAP_SYS_ADMIN).TaskCapabilities
			Expect(after.Eff()).NotTo(ShareBackingArrayWith(before.Eff()))

			before = before.Permitted().All().TaskCapabilities
			after = before.Permitted().Drop(caps.CAP_SYS_ADMIN).TaskCapabilities
			Expect(after.Perm()).NotTo(ShareBackingArrayWith(before.Perm()))

			before = before.Inheritable().All().TaskCapabilities
			after = before.Inheritable().Drop(caps.CAP_SYS_ADMIN).TaskCapabilities
			Expect(after.Inh()).NotTo(ShareBackingArrayWith(before.Inh()))
		})

		It("clears all capabilities and returns a new value", func() {
			before := caps.TaskCapabilities{}.
				Effective().All().
				Permitted().All().
				Inheritable().All().
				TaskCapabilities
			after := before.Effective().Clear().TaskCapabilities
			Expect(after.Eff().IsEmpty()).To(BeTrue())
			Expect(after.Eff()).NotTo(ShareBackingArrayWith(before.Eff()))

			after = before.Permitted().Clear().TaskCapabilities
			Expect(after.Perm()).NotTo(ShareBackingArrayWith(before.Perm()))

			after = before.Inheritable().Clear().TaskCapabilities
			Expect(after.Inh()).NotTo(ShareBackingArrayWith(before.Inh()))
		})

		It("clears existing capabilities, then sets the specified ones, and returns a new value", func() {
			before := caps.TaskCapabilities{}.Effective().Add(caps.CAP_SYS_ADMIN)
			after := before.Effective().Replace(caps.CAP_NET_ADMIN).TaskCapabilities
			Expect(after.Eff()).NotTo(ShareBackingArrayWith(before.Eff()))

			after = before.Permitted().Drop(caps.CAP_SYS_ADMIN).TaskCapabilities
			Expect(after.Perm()).NotTo(ShareBackingArrayWith(before.Perm()))

			after = before.Inheritable().Drop(caps.CAP_SYS_ADMIN).TaskCapabilities
			Expect(after.Inh()).NotTo(ShareBackingArrayWith(before.Inh()))
		})

		It("adds multiple capabilities", func() {
			tc := caps.TaskCapabilities{}.
				Effective().Add(caps.CAP_SYS_ADMIN).
				Add(caps.CAP_NET_ADMIN, caps.CAP_SYS_BOOT).
				TaskCapabilities
			Expect(tc.Effective().Has(caps.CAP_SYS_ADMIN)).To(BeTrue())
			Expect(tc.Effective().Has(caps.CAP_NET_ADMIN)).To(BeTrue())
			Expect(tc.Effective().Has(caps.CAP_SYS_BOOT)).To(BeTrue())
		})

		It("drops multiple capabilities", func() {
			tc := caps.TaskCapabilities{}.
				Effective().All().Drop(caps.CAP_SYS_ADMIN).
				Drop(caps.CAP_NET_ADMIN, caps.CAP_SYS_BOOT).
				TaskCapabilities
			Expect(tc.Effective().IsEmpty()).To(BeFalse())
			Expect(tc.Effective().Has(caps.CAP_SYS_ADMIN)).To(BeFalse())
			Expect(tc.Effective().Has(caps.CAP_NET_ADMIN)).To(BeFalse())
			Expect(tc.Effective().Has(caps.CAP_SYS_BOOT)).To(BeFalse())
		})

		It("replaces multiple capabilities", func() {
			tc := caps.TaskCapabilities{}.
				Effective().Add(caps.CAP_SYS_ADMIN).
				Replace(caps.CAP_NET_ADMIN, caps.CAP_SYS_BOOT).
				TaskCapabilities
			Expect(tc.Effective().Has(caps.CAP_SYS_ADMIN)).To(BeFalse())
			Expect(tc.Effective().Has(caps.CAP_NET_ADMIN)).To(BeTrue())
			Expect(tc.Effective().Has(caps.CAP_SYS_BOOT)).To(BeTrue())
		})

		It("copies effective capabilities", func() {
			before := caps.TaskCapabilities{}.Effective().Add(caps.CAP_SYS_ADMIN).TaskCapabilities
			after := before.Permitted().SameAsEffective().TaskCapabilities
			Expect(after.Permitted().Has(caps.CAP_SYS_ADMIN)).To(BeTrue())
			Expect(after.Perm()).NotTo(ShareBackingArrayWith(after.Eff()))
		})

		It("copies permitted capabilities", func() {
			before := caps.TaskCapabilities{}.Permitted().Add(caps.CAP_SYS_ADMIN).TaskCapabilities
			after := before.Inheritable().SameAsPermitted().TaskCapabilities
			Expect(after.Inheritable().Has(caps.CAP_SYS_ADMIN)).To(BeTrue())
			Expect(after.Inh()).NotTo(ShareBackingArrayWith(after.Perm()))
		})

		It("copies inheritable capabilities", func() {
			before := caps.TaskCapabilities{}.Inheritable().Add(caps.CAP_SYS_ADMIN).TaskCapabilities
			after := before.Effective().SameAsInheritable().TaskCapabilities
			Expect(after.Effective().Has(caps.CAP_SYS_ADMIN)).To(BeTrue())
			Expect(after.Eff()).NotTo(ShareBackingArrayWith(after.Inh()))
		})

		It("returns the spotlighted capabilities", func() {
			tc := caps.TaskCapabilities{}.
				Effective().Add(caps.CAP_SYS_ADMIN).
				TaskCapabilities
			eff := tc.Effective().Capabilities()
			Expect(tc.Eff()).To(ShareBackingArrayWith(eff))
		})

	})

	When("setting task capabilities", func() {

		It("returns an error for a non-existing task", func() {
			Expect(caps.TaskCapabilities{}.ApplyToTask(-1)).Error().To(MatchError(syscall.EPERM))
		})

		It("drops and reinstates capabilities", func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}
			done := make(chan struct{})
			go func() {
				defer GinkgoRecover()
				defer close(done)
				runtime.LockOSThread()

				before := Successful(caps.OfCurrentTask())
				By(fmt.Sprintf("original task capabilities: %#v", before))

				By("dopping all capabilities before trying to create a raw 'sucket'")
				powerless := Successful(before.Clone().Effective().Clear().ApplyToCurrentTask())
				Expect(powerless.Effective().IsEmpty()).To(BeTrue())
				_, err := unix.Socket(unix.AF_INET, unix.SOCK_RAW, 254) // returns -1 as fd
				Expect(err).To(HaveOccurred())

				By("regaining CAP_NET_RAW before creating a raw socket")
				Expect(powerless.Effective().Add(caps.CAP_NET_RAW).ApplyToCurrentTask()).Error().To(Succeed())
				Expect(unix.Close(Successful(unix.Socket(unix.AF_INET, unix.SOCK_RAW, 254)))).To(Succeed())
			}()
			Eventually(done).Should(BeClosed())
		})

		It("sets the effective capabilities", func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}
			done := make(chan struct{})
			go func() {
				defer GinkgoRecover()
				defer close(done)
				runtime.LockOSThread()

				orig := Successful(caps.OfCurrentTask())
				Expect(orig.Effective().Has(caps.CAP_NET_RAW)).To(BeTrue())
				Expect(orig.Effective().Replace(caps.CAP_NET_RAW).ApplyToCurrentTask()).Error().NotTo(HaveOccurred())
				current := Successful(caps.OfCurrentTask())
				Expect(current.Effective().Has(caps.CAP_NET_RAW)).To(BeTrue())
				Expect(current.Effective().Has(caps.CAP_NET_ADMIN)).To(BeFalse())

				Expect(current.Effective().Add(caps.CAP_NET_ADMIN).ApplyToCurrentTask()).Error().NotTo(HaveOccurred())
				current = Successful(caps.OfCurrentTask())
				Expect(current.Effective().Has(caps.CAP_NET_ADMIN)).To(BeTrue())
			}()
			Eventually(done).Should(BeClosed())
		})

	})

})
