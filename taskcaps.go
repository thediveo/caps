// Copyright 2023 Harald Albrecht.
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

//go:build linux

package caps

import (
	"os"
	"strconv"
	"strings"
	"unsafe"

	"github.com/thediveo/caps/errno"
	"golang.org/x/sys/unix"
)

// TaskCapabilities represents the effective, permitted and inheritable
// capabilities sets.
//
// The three capabilities sets (effective, permitted, inherited) of a task can
// be retrieved using [TaskCaps] and all three sets set with [SetTaskCaps].
//
// Often, only the effective capabilities of a task are to be changed, this can
// be done by first obtaining suitable task capabilities via [AddEffectiveCaps]
// and [SetEffectiveCaps], and with the result obtained then calling
// [SetTaskCaps].
type TaskCapabilities struct {
	Effective   CapabilitiesSet
	Permitted   CapabilitiesSet
	Inheritable CapabilitiesSet
}

// Clone returns an independent clone of the task capabilities. Modifications to
// the source task capabilities won't change the cloned task capabilities.
func (t TaskCapabilities) Clone() TaskCapabilities {
	return TaskCapabilities{
		Effective:   t.Effective.Clone(),
		Permitted:   t.Permitted.Clone(),
		Inheritable: t.Inheritable.Clone(),
	}
}

// AddEffectiveCaps retrieves the current task's capabilities sets, adds the
// specified effective capabilities and sets them as the new current task's
// capabilities. AddEffectiveCaps returns the previous capabilities sets when
// successful. For more complex use cases where also the permitted capabilities
// are to be changed, please use [TaskCaps] first, then change the permitted
// capabilities in the task capabilities acquired, and finally call
// [SetTaskCaps].
func AddEffectiveCaps(capno int, morecapsno ...int) (capsbefore TaskCapabilities, err error) {
	capsbefore, err = OfThisTask()
	if err != nil {
		return
	}
	newcaps := capsbefore.Clone()
	newcaps.Effective.Add(capno, morecapsno...)
	return capsbefore, SetForThisTask(newcaps)
}

// SetEffectiveCaps retrieves the current task's capabilities sets, then sets
// only the specified effective capabilities and then sets them as the new
// current task's capabilities. SetEffectiveCaps returns the previous
// capabilities sets when successful.
func SetEffectiveCaps(capno int, morecapsno ...int) (capsbefore TaskCapabilities, err error) {
	capsbefore, err = OfThisTask()
	if err != nil {
		return
	}
	newcaps := capsbefore.Clone()
	newcaps.Effective = NewCapabilitiesSet()
	newcaps.Effective.Add(capno, morecapsno...)
	return capsbefore, SetForThisTask(newcaps)
}

const capDataElements = LINUX_CAPABILITY_U32S_3

// KernelCapabilityVersion returns the version of the capabilities user-space
// data structure that the Linux kernel we're just running on "natively" uses.
// In case the version could not properly be detected, 0 is returned instead.
func KernelCapabilityVersion() uint32 { return linuxCapabilityVersion }

var linuxCapabilityVersion uint32

// As can be glanced from (when you know it's there)
// https://elixir.bootlin.com/linux/v6.1/source/kernel/capability.c#L100, the
// Linux kernel returns the version it natively supports of the capabilities
// user-space data structure when trying to get capabilities using a
// non-existing version; the best bet is 0, as this is a version that was never
// used, nor will ever be used.
func init() {
	var capHeader = unix.CapUserHeader{Version: 0} // never was, won't ever be.

	_, _, _ = unix.RawSyscall(
		unix.SYS_CAPGET,
		uintptr(unsafe.Pointer(&capHeader)),
		0,
		0)
	linuxCapabilityVersion = capHeader.Version // now "should have been" changed by the kernel.
}

// LastCapability returns the number of the highest capability supported by the
// kernel we're now running on. This value might differ from
// [MaxCapabilityNumber] that is known to this package.
func LastCapability() int { return lastCapability }

var lastCapability int

func init() {
	contents, _ := os.ReadFile("/proc/sys/kernel/cap_last_cap")
	lastCapability, _ = strconv.Atoi(strings.TrimSuffix(string(contents), "\n"))
	if lastCapability == 0 {
		lastCapability = MaxCapabilityNumber
	}
}

// OfThisTask returns the effective, permitted and inheritable capability sets
// for the current task. If the sets cannot be queried from the Linux kernel,
// then an error is returned instead with a zero set of capabilities.
func OfThisTask() (taskcaps TaskCapabilities, err error) {
	return OfTask(0)
}

// OfTask returns the effective, permitted and inheritable capability sets for
// the specified task. If the sets cannot be queried from the Linux kernel, then
// an error is returned instead with a zero set of capabilities.
func OfTask(tid int) (taskcaps TaskCapabilities, err error) {
	var capHeader = unix.CapUserHeader{
		Version: unix.LINUX_CAPABILITY_VERSION_3,
		Pid:     int32(tid),
	}
	var capData [capDataElements]unix.CapUserData

	_, _, e := unix.RawSyscall(
		unix.SYS_CAPGET,
		uintptr(unsafe.Pointer(&capHeader)),
		uintptr(unsafe.Pointer(&capData[0])),
		0)
	if e != 0 {
		return TaskCapabilities{}, errno.Error(e)
	}

	caps := CapabilitiesSet(make([]uint32, capDataElements))
	for idx := 0; idx < capDataElements; idx++ {
		caps[idx] = capData[idx].Effective
	}
	taskcaps.Effective = caps

	caps = CapabilitiesSet(make([]uint32, capDataElements))
	for idx := 0; idx < capDataElements; idx++ {
		caps[idx] = capData[idx].Permitted
	}
	taskcaps.Permitted = caps

	caps = CapabilitiesSet(make([]uint32, capDataElements))
	for idx := 0; idx < capDataElements; idx++ {
		caps[idx] = capData[idx].Inheritable
	}
	taskcaps.Inheritable = caps

	return
}

// SetForThisTask sets the capability sets (effective, permitted and
// inheritable) for the current task.
func SetForThisTask(taskcaps TaskCapabilities) error {
	return SetForTask(0, taskcaps)
}

// SetForTask sets the capability sets (effective, permitted and inheritable)
// for the specified task.
func SetForTask(tid int, taskcaps TaskCapabilities) error {
	var capHeader = unix.CapUserHeader{
		Version: unix.LINUX_CAPABILITY_VERSION_3,
		Pid:     int32(tid),
	}
	var capData [capDataElements]unix.CapUserData

	for idx := 0; idx < capDataElements; idx++ {
		if idx < len(taskcaps.Effective) {
			capData[idx].Effective = taskcaps.Effective[idx]
		}
		if idx < len(taskcaps.Permitted) {
			capData[idx].Permitted = taskcaps.Permitted[idx]
		}
		if idx < len(taskcaps.Inheritable) {
			capData[idx].Inheritable = taskcaps.Inheritable[idx]
		}
	}

	_, _, e := unix.RawSyscall(
		unix.SYS_CAPSET,
		uintptr(unsafe.Pointer(&capHeader)),
		uintptr(unsafe.Pointer(&capData[0])),
		0)
	if e != 0 {
		return errno.Error(e)
	}
	return nil
}
