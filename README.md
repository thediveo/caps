# caps

[![PkgGoDev](https://img.shields.io/badge/-reference-blue?logo=go&logoColor=white&labelColor=505050)](https://pkg.go.dev/github.com/thediveo/caps)
[![License](https://img.shields.io/github/license/thediveo/caps)](https://img.shields.io/github/license/thediveo/caps)
![Build and Test](https://github.com/thediveo/caps/actions/workflows/buildandtest.yaml/badge.svg?branch=master)
![Coverage](https://img.shields.io/badge/Coverage-96.6%25-brightgreen)
[![Go Report Card](https://goreportcard.com/badge/github.com/thediveo/caps)](https://goreportcard.com/report/github.com/thediveo/caps)

A pure-Go minimalist package for getting and setting the capabilities of Linux
tasks (threads). No need for linking with `libcap`.

For devcontainer instructions, please see the [section "DevContainer"
below](#devcontainer).

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

## DevContainer

> [!CAUTION]
>
> Do **not** use VSCode's "~~Dev Containers: Clone Repository in Container
> Volume~~" command, as it is utterly broken by design, ignoring
> `.devcontainer/devcontainer.json`.

1. `git clone https://github.com/thediveo/enumflag`
2. in VSCode: Ctrl+Shift+P, "Dev Containers: Open Workspace in Container..."
3. select `enumflag.code-workspace` and off you go...

## Supported Go Versions

`native` supports versions of Go that are noted by the [Go release
policy](https://golang.org/doc/devel/release.html#policy), that is, major
versions _N_ and _N_-1 (where _N_ is the current major version).

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md).

## Copyright and License

`caps` is Copyright 2023, 2025 Harald Albrecht, and licensed under the Apache
License, Version 2.0.
