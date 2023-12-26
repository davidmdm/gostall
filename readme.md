# gostall: Go Binary Installer

## Overview

gostall is a lightweight Go build tool that empowers you to take control of your binary names when installing go binaries. Whether you're working with local or remote Go projects, gostall simplifies the build and installation process.

## Why?

Sometimes we need to install binaries but do not want them to be installed under their module name or folder name.
For local packages we can use the `go build -o PATH_TO_BIN` however this does not work when working with remote packages. There's no equivalent `go install -o` flag either.

Gostall makes installing Go packages under any name easy with a single unified command.

## Usage

_path_: The path to the Go project. It can be a local or remote (GitHub) repository.
_name_: Your preferred name for the binary output.

```bash
gostall [path] [name]
```

## Examples

### Local Path

```bash
gostall ./myproject mybinary
```

### Remote Path

```bash
gostall github.com/user/repo@latest mybinary
```

### Local Output

By default _gostall_ builds your binaries under GOBIN with the name provided for them.
If however the name argument is a multi segment filepath it will build it to that location instead.

```bash
# Install to under GOBIN
gostall github.com/user/repo@latest example

# Install/Build it to the current working directory as `./example`
gostall github.com/user/repo@latest ./example
```

## Installation

```bash
go install github.com/davidmdm/gostall@latest

# if you wish to install it under a different you can use gostall to do so!
gostall github.com/davidmdm/gostall@latest what-i-want
```

### Requirements

Go 1.16 or later

## Configuration

GOBIN: The location where the binary will be installed.

## License

This project is licensed under the MIT License.
