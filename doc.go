/*
Package caps provides a minimalist interface to getting and setting the
capabilities of Linux tasks (threads). It is a pure Go implementation that does
not need any linking with the [C libcap]. On purpose, this package isn't any
drop-in replacement for the [libcap.git] Go module (if that even is possible).

The focus of this module is on dropping and regaining effective capabilities, as
well as dropping permitted capabilities. That is, a more Go-like fluent API to
the [capget(2)] and capset(2) Linux syscalls.

# Migrating v0 → v2: Capabilities are Immutable Values

caps v2 treats both individual capabilities sets (in form of [CapabilitiesSet])
as well as task capabilities triplets (that is, the [TaskCapabilities] type) as
immutable. This in turn naturally leads to a fluent chaining API: instead of
returning the modified original CapabilitiesSet or TaskCapabilities, chain
methods always return a new immutable object for the new state, keeping the old
object immutable.

Unless in the hottest of hot paths, we consider immutability to be significantly
more important than “ludicrous speed”, as immutability decreases the chance of
hard-to-track unintended shared state modifications at a slightly increased GC
cost. Please see below for usage examples.

In case of [TaskCapabilities] the three chain methods
[TaskCapabilities.Effective], [TaskCapabilities.Permitted] and
[TaskCapabilities.Inheritable] tell the following modification methods (such as
[SpotlightedCapabilities.Add], [SpotlightedCapabilities.Drop], et cetera) which
capabilities of the triple set to modify. Please note that it's not possible to
modify the same capability in two or more sets simultaneously, which hardly
should be a real use case under any circumstances.

# Dropping and Regaining Effective Capabilities

To drop the calling task's effective capabilities only, without dropping the
permitted capabilities:

	// Make sure to lock this Go routine to its current OS-level task (thread).
	runtime.LockOSThread()
	origcaps, err := caps.OfThisTask()
	dropped, err := origcaps.Effective().Clear().ApplyToThisTask()

To regain only a specific effective capability (the first returned value is the
applied capabilities triple set itself):

	_, err = dropped.Effective().Add(caps.CAP_SYS_ADMIN).ApplyToThisTask()

And finally to regain all originally effective capabilities (again, the first
returned value it the applied capabilities triple set itself):

	_, err = origcaps.ApplyToThisTask()

# Notes

This package assumes at least a kernel version 2.65 or later and does not
support older kernels.

The Linux kernel actually [returns the version of the capabilities] user-space
structure it uses “natively” in the capabilities header version field if this
version field is set to an invalid or unsupported version (such as 0 which was
never be used and won't ever). In this case, EINVAL is returned.

[returns the version of the capabilities]: https://elixir.bootlin.com/linux/v6.1/source/kernel/capability.c#L100
[C libcap]: https://git.kernel.org/pub/scm/libs/libcap/libcap.git/
[libcap.git]: https://pkg.go.dev/git.kernel.org/pub/scm/libs/libcap/libcap.git
[capget(2)]: https://man7.org/linux/man-pages/man2/capget.2.html
*/
package caps
