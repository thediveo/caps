// Copyright 2026 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package caps

import (
	"os"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

// capDataElements is the number of “unsigned 32bit integer slots” used in the
// 3rd version of the Linux capget and capset syscalls.
const capDataElements = LINUX_CAPABILITY_U32S_3

// KernelCapabilityVersion returns the version of the capabilities user-space
// data structure that the Linux kernel we're just running on "natively" uses.
// In case the version could not properly be detected, 0 is returned instead.
func KernelCapabilityVersion() uint32 { return linuxCapabilityVersion }

// As can be glanced from (when you already know that it's there)
// https://elixir.bootlin.com/linux/v6.1/source/kernel/capability.c#L100, the
// Linux kernel returns the version it natively supports of the capabilities
// user-space data structure when trying to get capabilities using a
// non-existing version; the best bet is 0, as this is a version that was never
// used, nor will ever be used.
var linuxCapabilityVersion uint32 = func() uint32 {
	var capHeader = unix.CapUserHeader{Version: 0} // never was, won't ever be.
	_, _, _ = unix.RawSyscall(
		unix.SYS_CAPGET,
		uintptr(unsafe.Pointer(&capHeader)),
		0,
		0)
	return capHeader.Version // now "should have been" changed by the kernel.
}()

// LastCapability returns the number of the highest capability supported by the
// kernel we're now running on. This value might differ from
// [MaxCapabilityNumber] that is known to this package.
func LastCapability() int { return lastCapability }

var lastCapability int = func() int {
	contents, _ := os.ReadFile("/proc/sys/kernel/cap_last_cap")
	lastcap, _ := strconv.Atoi(strings.TrimSuffix(string(contents), "\n"))
	if lastcap == 0 {
		return MaxCapabilityNumber
	}
	return lastcap
}()
