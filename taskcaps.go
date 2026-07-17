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

package caps

import (
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"

	"github.com/thediveo/caps/v2/errno"
)

// TaskCapabilities represents the effective, permitted and inheritable
// capabilities sets (of a task).
//
// The zero value is a valid task capabilities set, devoid of any effective,
// permitted, and inheritable capabilities.
//
// These three capabilities sets of a task can be retrieved en block using
// [TaskCaps], modified, and then set with [SetTaskCaps] (again, en block). This
// design is imposed by the Linux [capget(2)] and [capset(2)] syscalls.
//
// For instance, to drop all effective capabilities of the current task, while
// keeping the current permitted and inheritable capabilities:
//
//	caps.OfCurrentTaskOrZero().Effective().Clear().ApplyToCurrentTask()
//
// In case the dropped effective capabilities should be restored at a later
// point:
//
//	orig, _ := caps.OfCurrentTask()
//	_ = orig.Effective().Clear().ApplyToCurrentTask()
//	// ...
//	_ = orig.ApplyToCurrentTask()
//
// [capget(2)]: https://www.man7.org/linux/man-pages/man2/capget.2.html
// [capset(2)]: https://www.man7.org/linux/man-pages/man2/capset.2.html
type TaskCapabilities struct {
	eff, perm, inh CapabilitiesSet
}

// OfTask returns the effective, permitted and inheritable capabilities triple
// for the specified task. If the triple sets cannot be queried from the Linux
// kernel, then an error is returned instead.
func OfTask(tid int) (taskcaps TaskCapabilities, err error) {
	return ofTask(tid)
}

// OfTaskOrZero returns the effective, permitted and inheritable capabilities
// triple for the specified task. It returns a zero value in case retrieving the
// capabilities fails.
func OfTaskOrZero(tid int) TaskCapabilities {
	taskcaps, _ := ofTask(tid)
	return taskcaps
}

// OfCurrentTask returns the effective, permitted and inheritable capabilities
// triple for the current task. If the triple sets cannot be queried from the
// Linux kernel, then an error is returned instead.
func OfCurrentTask() (taskcaps TaskCapabilities, err error) {
	return ofTask(0)
}

// OfCurrentTaskOrZero returns the effective, permitted and inheritable
// capabilities triple for the current task. It returns a zero value in case
// retrieving the capabilities fails.
func OfCurrentTaskOrZero() TaskCapabilities {
	taskcaps, _ := OfCurrentTask()
	return taskcaps
}

// Clone returns a fresh (deep) copy of this task capabilities that is wholy
// independent of the passed task capabilities.
func (t TaskCapabilities) Clone() TaskCapabilities {
	return TaskCapabilities{
		eff:  t.eff.Clone(),
		perm: t.perm.Clone(),
		inh:  t.inh.Clone(),
	}
}

// IsEmpty returns true if the effective, permitted, and inheritable
// capabilities are all zero.
func (t TaskCapabilities) IsEmpty() bool {
	return t.eff.IsEmpty() && t.perm.IsEmpty() && t.inh.IsEmpty()
}

// String returns a compact textual representation of the effective, permitted,
// and inheritable capabilities.
func (t TaskCapabilities) String() string {
	var s strings.Builder
	s.WriteRune('{')
	s.WriteString("effective:")
	s.WriteString(t.eff.String())
	s.WriteString(";permitted:")
	s.WriteString(t.perm.String())
	s.WriteString(";inheritable:")
	s.WriteString(t.inh.String())
	s.WriteRune('}')
	return s.String()
}

// ApplyToTask applies the effective, permitted, and inheritable capabilities to
// the specified task. The zero tid identifies the calling task (which then must
// have been locked to the calling go routine using [runtime.LockOSThread]).
//
// Please note that this never spreads these task capabilities to any other task
// in the same process. If needed, this must be done explicitly.
func (t TaskCapabilities) ApplyToTask(tid int) (TaskCapabilities, error) {
	err := setForTask(tid, t)
	return t, err
}

// ApplyToCurrentTask applies the effective, permitted, and inheritable
// capabilities to calling task, that is, the locked(!) task of the calling go
// routine. Important: ApplyToCurrentTask never calls [runtime.LockOSThread] by
// itself.
func (t TaskCapabilities) ApplyToCurrentTask() (TaskCapabilities, error) {
	return t.ApplyToTask(0)
}

// Effective causes the following chained method calls to work on the effective
// capabilities of this task.
func (t TaskCapabilities) Effective() EffectiveCaps {
	return EffectiveCaps{
		SpotlightedCapabilities: SpotlightedCapabilities{
			TaskCapabilities: t,
			selected:         onEffectiveCaps,
		},
	}
}

// Permitted causes the following chained method calls to work on the permitted
// capabilities of this task.
func (t TaskCapabilities) Permitted() PermittedCaps {
	return PermittedCaps{
		SpotlightedCapabilities: SpotlightedCapabilities{
			TaskCapabilities: t,
			selected:         onPermittedCaps,
		},
	}
}

// Inheritable causes the following chained method calls to work on the inheritable
// capabilities of this task.
func (t TaskCapabilities) Inheritable() InheritableCaps {
	return InheritableCaps{
		SpotlightedCapabilities: SpotlightedCapabilities{
			TaskCapabilities: t,
			selected:         onInheritableCaps,
		},
	}
}

// spotlightedCapabilitiesSet indicates which one out of the three capabilities
// sets of effective, permitted, inheritable we're going to deal with.
type spotlightedCapabilitiesSet int

const (
	onEffectiveCaps spotlightedCapabilitiesSet = iota + 1
	onPermittedCaps
	onInheritableCaps
)

// SpotlightedCapabilities spotlights one set of capabilities out of the triple
// sets of a task, keeping the triple sets.
type SpotlightedCapabilities struct {
	TaskCapabilities
	selected spotlightedCapabilitiesSet
}

// Add returns a new task capabilities triple set where in the currently
// spotlighted set the specified capabilities are set. Capabilities are
// identified by their numbers, such as [CAP_SYS_ADMIN], et cetera.
func (t SpotlightedCapabilities) Add(capno int, morecapnos ...int) SpotlightedCapabilities {
	// nota bene: we're already a *copy* of t
	cs := t.spotlight()
	*cs = cs.Add(capno, morecapnos...) // CapabilitiesSets are immutable.
	return t
}

// All returns a new task capabilities triple set where in the currently
// spotlighted set all capabilities are set. “All” means all capabilities
// reported by the currently running kernel.
func (t SpotlightedCapabilities) All() SpotlightedCapabilities {
	// nota bene: we're already a *copy* of t
	cs := t.spotlight()
	*cs = cs.All() // CapabilitiesSets are immutable.
	return t
}

// Clear returns a new task capabilities triple set where the currently
// spotlighted set is devoid of all its capabilities.
func (t SpotlightedCapabilities) Clear() SpotlightedCapabilities {
	// nota bene: we're already a *copy* of t
	*t.spotlight() = NewCapabilitiesSet()
	return t
}

// Drop returns a new task capabilities triple set where in the currently
// spotlighted set the specified capabilities have been removed. Capabilities
// are identified by their numbers, such as [CAP_SYS_ADMIN], et cetera.
func (t SpotlightedCapabilities) Drop(capno int, morecapnos ...int) SpotlightedCapabilities {
	// nota bene: we're already a *copy* of t
	cs := t.spotlight()
	*cs = cs.Drop(capno, morecapnos...) // CapabilitiesSets are immutable.
	return t
}

// Replace clears all currently spotlighted capabilities, setting only the
// specified capabilities. Capabilities are identified by their numbers, such as
// [CAP_SYS_ADMIN], et cetera.
func (t SpotlightedCapabilities) Replace(capno int, morecapnos ...int) SpotlightedCapabilities {
	// nota bene: we're already a *copy* of t
	*t.spotlight() = NewCapabilitiesSet().Add(capno, morecapnos...) // CapabilitiesSets are immutable.
	return t
}

// SameAsEffective copies the effective capabilities to the current spotlighted
// capabilities.
func (t SpotlightedCapabilities) SameAsEffective() SpotlightedCapabilities {
	// nota bene: we're already a *copy* of t
	*t.spotlight() = t.eff.Clone()
	return t
}

// SameAsPermitted copies the permitted capabilities to the current spotlighted
// capabilities.
func (t SpotlightedCapabilities) SameAsPermitted() SpotlightedCapabilities {
	// nota bene: we're already a *copy* of t
	*t.spotlight() = t.perm.Clone()
	return t
}

// SameAsInheritable copies the inheritable capabilities to the current
// spotlighted capabilities.
func (t SpotlightedCapabilities) SameAsInheritable() SpotlightedCapabilities {
	// nota bene: we're already a *copy* of t
	*t.spotlight() = t.inh.Clone()
	return t
}

// IsEmpty returns true if all spotlighted capabilities are unset.
func (t SpotlightedCapabilities) IsEmpty() bool { return t.spotlight().IsEmpty() }

// Has returns true if the specified capability is set in the spotlighted
// capabilities.
func (t SpotlightedCapabilities) Has(capno int) bool { return t.spotlight().Has(capno) }

// Capabilities returns the currently spotlighted set of capabilities out of the
// effective, permitted, or inheritable capabilities.
func (t SpotlightedCapabilities) Capabilities() CapabilitiesSet { return *t.spotlight() }

// spotlight returns a pointer to the currently spotlighted CapabilitiesSet.
//
// Note that we don't store the pointer (reference) permanently in a
// SpotlightedCapabilities object because when it gets copined/cloned we would
// need to ensure with each copy step that the pointer gets correctly updated in
// the copy. So we rather stick to the "late binding" pattern to avoid subtle
// bugs after forgetting to update the reference after copy.
func (t *SpotlightedCapabilities) spotlight() *CapabilitiesSet {
	switch t.selected {
	case onEffectiveCaps:
		return &t.eff
	case onPermittedCaps:
		return &t.perm
	case onInheritableCaps:
		return &t.inh
	default:
		panic("internal error: no spotlight")
	}
}

// ofTask returns the effective, permitted and inheritable capability sets for
// the specified task. If the sets cannot be queried from the Linux kernel, then
// an error is returned instead with a zero set of capabilities.
func ofTask(tid int) (taskcaps TaskCapabilities, err error) {
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
	for idx := range capDataElements {
		caps[idx] = capData[idx].Effective
	}
	taskcaps.eff = caps

	caps = CapabilitiesSet(make([]uint32, capDataElements))
	for idx := range capDataElements {
		caps[idx] = capData[idx].Permitted
	}
	taskcaps.perm = caps

	caps = CapabilitiesSet(make([]uint32, capDataElements))
	for idx := range capDataElements {
		caps[idx] = capData[idx].Inheritable
	}
	taskcaps.inh = caps

	return
}

// setForTask sets the capability sets (effective, permitted and inheritable)
// for the specified task.
func setForTask(tid int, taskcaps TaskCapabilities) error {
	var capHeader = unix.CapUserHeader{
		Version: unix.LINUX_CAPABILITY_VERSION_3,
		Pid:     int32(tid),
	}
	var capData [capDataElements]unix.CapUserData

	for idx := range capDataElements {
		if idx < len(taskcaps.eff) {
			capData[idx].Effective = taskcaps.eff[idx]
		}
		if idx < len(taskcaps.perm) {
			capData[idx].Permitted = taskcaps.perm[idx]
		}
		if idx < len(taskcaps.inh) {
			capData[idx].Inheritable = taskcaps.inh[idx]
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

// EffectiveCaps allows the effective capabilities of a task to be manipulated
// in following chained method calls.
type EffectiveCaps = struct{ SpotlightedCapabilities }

// PermittedCaps allows the permitted capabilities of a task to be manipulated
// in following chained method calls.
type PermittedCaps = struct{ SpotlightedCapabilities }

// InheritableCaps fallows the inheritable capabilities of a task to be
// manipulated in following chained method calls.
type InheritableCaps = struct{ SpotlightedCapabilities }
