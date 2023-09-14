# caps

[![PkgGoDev](https://img.shields.io/badge/-reference-blue?logo=go&logoColor=white&labelColor=505050)](https://pkg.go.dev/github.com/thediveo/caps)
[![License](https://img.shields.io/github/license/thediveo/caps)](https://img.shields.io/github/license/thediveo/caps)
![Build and Test](https://github.com/thediveo/caps/workflows/build%20and%20test/badge.svg?branch=master)
![Coverage](https://img.shields.io/badge/Coverage-96.6%25-brightgreen)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/caps)](https://goreportcard.com/report/github.com/thediveo/caps)

A pure-Go minimalist package for getting and setting the capabilities of Linux
tasks (threads). No need for linking with `libcap`.

## Example: Dropping and Regaining Effective Capabilities

To drop the calling task's effective capabilities only, without dropping the
permitted capabilities:

```go
// Make sure to lock this Go routine to its current OS-level task (thread).
runtime.LockOSThread()

origcaps := caps.OfThisTask()
dropped := origcaps.Clone()
dropped.Effective.Clear()
caps.SetForThisTask(dropped)
```

To regain only a specific effective capability:

```go
dropped.Effective.Add(caps.CAP_SYS_ADMIN)
caps.SetForThisTask(dropped)
```

And finally to regain all originally effective capabilities:

```go
caps.SetForThisTask(origcaps)
```

## Go Version Support

`caps` supports versions of Go that are noted by the Go release policy, that is,
major versions _N_ and _N_-1 (where _N_ is the current major version).

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md).

## Copyright and License

`caps` is Copyright 2023 Harald Albrecht, and licensed under the Apache License,
Version 2.0.
