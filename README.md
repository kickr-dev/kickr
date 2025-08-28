# kickr <!-- omit in toc -->

<p align="center">
  <img alt="GitHub Release" src="https://img.shields.io/github/v/release/kickr-dev/kickr?include_prereleases&sort=semver&style=for-the-badge">
  <img alt="GitHub Issues" src="https://img.shields.io/github/issues-raw/kickr-dev/kickr?style=for-the-badge">
  <img alt="GitHub License" src="https://img.shields.io/github/license/kickr-dev/kickr?style=for-the-badge">
  <img alt="GitHub Actions" src="https://img.shields.io/github/actions/workflow/status/kickr-dev/kickr/integration.yml?style=for-the-badge">
  <img alt="Coverage" src="https://img.shields.io/codecov/c/github/kickr-dev/kickr?style=for-the-badge">
  <img alt="Go Version" src="https://img.shields.io/github/go-mod/go-version/kickr-dev/kickr?style=for-the-badge">
  <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/kickr-dev/kickr?style=for-the-badge">
  <img alt="OpenSSF Scorecard" src="https://img.shields.io/ossf-scorecard/github.com/kickr-dev/kickr?label=OpenSSF+Scorecard&style=for-the-badge">
</p>

---

- [How to use ?](#how-to-use-)
  - [Go](#go)
  - [Linux](#linux)
- [Commands](#commands)
  - [Init](#init)
  - [Generate](#generate)
- [Kickr file](#kickr-file)

## How to use ?

### Go

```sh
go install github.com/kickr-dev/kickr/cmd/kickr@latest
```

### Linux

```sh
OS="linux" # change it depending on your case
ARCH="amd64" # change it depending on your case
INSTALL_DIR="$HOME/.local/bin" # change it depending on your case

new_version=$(curl -fsSL "https://api.github.com/repos/kickr-dev/kickr/releases/latest" | jq -r '.tag_name')
url="https://github.com/kickr-dev/kickr/releases/download/$new_version/kickr_${OS}_${ARCH}.tar.gz"
curl -fsSL "$url" | (mkdir -p "/tmp/kickr/$new_version" && cd "/tmp/kickr/$new_version" && tar -xz)
cp "/tmp/kickr/$new_version/kickr" "$INSTALL_DIR/kickr"
```

## Commands

```
Kickr initializes or generates kickr projects. Kickr projects are only defined by a .kickr file
and multiple files automatically generated to avoid multiple hours to setup Continuous Integration, coverage, security analyses, helm chart, etc.

Kickr generation can be done with 'kickr' command or 'kickr generate' command.
Additional generation command are available to generate only subparts of kickr layout (like 'kickr chart').

Usage:
  kickr [flags]
  kickr [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  generate    Generate project layout
  help        Help about any command
  init        Initialize kickr project
  version     Show current kickr version

Flags:
  -d, --dir string          set directory where generation will be made (default is current directory)
  -h, --help                help for kickr
      --log-format string   set logging format (either "text" or "json") (default "text")
      --log-level string    set logging level (default "info")

Use "kickr [command] --help" for more information about a command.
```

### Init

```
Initialize new kickr project

Usage:
  kickr init [flags]

Flags:
  -h, --help   help for init

Global Flags:
  -d, --dir string          set directory where generation will be made (default is current directory)
      --log-format string   set logging format (either "text" or "json") (default "text")
      --log-level string    set logging level (default "info")
```

### Generate

```
Generate project layout

Usage:
  kickr generate [flags]

Flags:
  -h, --help   help for generate

Global Flags:
  -d, --dir string          set directory where generation will be made (default is current directory)
      --log-format string   set logging format (either "text" or "json") (default "text")
      --log-level string    set logging level (default "info")
```

## Kickr file

TBD
