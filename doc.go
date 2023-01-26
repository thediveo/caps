/*
Package caps provides a minimalist interface to getting and setting the
capabilities of Linux tasks (threads). It is a pure Go implementation that does
not need any linking with the [C libcap]. However, it isn't any drop-in
replacement for the [libcap.git] Go module (if that even is possible).

The focus of this module is on dropping and regaining effective capabilities, as
well as dropping permitted capabilities. That is, a more Go-like API to
the [capget(2)] and capset(2) Linux syscalls.

# Dropping and Regaining Effective Capabilities

To drop the calling task's effective capabilities only, without dropping the
permitted capabilities:

	// Make sure to lock this Go routine to its current OS-level task (thread).
	runtime.LockOSThread()

	origcaps := caps.OfThisTask()
	dropped := origcaps.Clone()
	dropped.Effective.Clear()
	caps.SetForThisTask(dropped)

To regain only a specific effective capability:

	dropped.Effective.Add(caps.CAP_SYS_ADMIN)
	caps.SetForThisTask(dropped)

And finally to regain all originally effective capabilities:

	caps.SetForThisTask(origcaps)

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
