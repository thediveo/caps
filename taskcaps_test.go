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
	"os"
	"runtime"
	"syscall"

	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("task capabilities", func() {

	It("returns an error when asking capabilities of a non-existing task", func() {
		Expect(OfTask(-1)).Error().To(MatchError(syscall.EINVAL))
	})

	It("returns an error when trying to set the capabilities of a non-existing task", func() {
		Expect(SetForTask(-1, TaskCapabilities{})).Error().To(MatchError(syscall.EPERM))
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

			before := Successful(OfThisTask())
			By(fmt.Sprintf("original task capabilities: %#v", before))

			By("dopping all capabilities before trying to create a raw 'sucket'")
			powerless := before.Clone()
			powerless.Effective = NewCapabilitiesSet()
			Expect(SetForThisTask(powerless)).To(Succeed())
			_, err := unix.Socket(unix.AF_INET, unix.SOCK_RAW, 254) // returns -1 as fd
			Expect(err).To(HaveOccurred())

			By("regaining CAP_NET_RAW before creating a raw socket")
			powerless.Effective.Add(CAP_NET_RAW)
			Expect(SetForThisTask(powerless)).To(Succeed())
			unix.Close(Successful(unix.Socket(unix.AF_INET, unix.SOCK_RAW, 254)))
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

			orig := Successful(OfThisTask())
			before := Successful(SetEffectiveCaps(CAP_NET_RAW))
			Expect(orig.Effective).To(Equal(before.Effective))
			current := Successful(OfThisTask())
			Expect(current.Effective.Has(CAP_NET_RAW)).To(BeTrue())
			before = Successful(AddEffectiveCaps(CAP_SYS_ADMIN))
			Expect(before.Effective).To(Equal(current.Effective))
			Expect(Successful(OfThisTask()).Effective.Has(CAP_SYS_ADMIN)).To(BeTrue())
		}()
		Eventually(done).Should(BeClosed())
	})

})
