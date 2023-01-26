package errno

import (
	"syscall"
)

// Do the interface allocations only once for common
// Errno values.
var (
	errEAGAIN   error = syscall.EAGAIN
	errEBADF    error = syscall.EBADF
	errEINVAL   error = syscall.EINVAL
	errENOENT   error = syscall.ENOENT
	errENOTSOCK error = syscall.ENOTSOCK
)

// Error turns a syscall.Errno in an ordinary error-type value -- this mimics
// the behavior of golang.org/x/sys/unix for returning boxed [syscall.EAGAIN],
// [syscall.EINVAL] and [syscall.ENOENT] instead of their unix package
// counterparts (The Source tells us that this prevents allocations at runtime).
// This function returns nil if the error number is zero.
func Error(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case syscall.EBADF: // Brian says: Be Different!
		return errEBADF
	case syscall.ENOTSOCK:
		return errENOTSOCK
	case syscall.EAGAIN:
		return errEAGAIN
	case syscall.EINVAL:
		return errEINVAL
	case syscall.ENOENT:
		return errENOENT
	}
	return e
}
