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

package errno

import (
	"syscall"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("boxed errno error values", func() {

	DescribeTable("returns boxed error values for common errno values",
		func(e syscall.Errno, err error) {
			Expect(Error(e)).To(BeIdenticalTo(err))
		},
		Entry("EBADF", syscall.EBADF, errEBADF),
		Entry("ENOTSOCK", syscall.ENOTSOCK, errENOTSOCK),
		Entry("EAGAIN", syscall.EAGAIN, errEAGAIN),
		Entry("EINVAL", syscall.EINVAL, errEINVAL),
		Entry("ENOENT", syscall.ENOENT, errENOENT),
	)

	It("returns nil for errno 0", func() {
		Expect(Error(0)).To(BeNil())
	})

	It("returns the errno as error", func() {
		Expect(Error(42000)).To(MatchError("errno 42000"))
	})

})
