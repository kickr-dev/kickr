# kickr <!-- omit in toc -->

<div align="center">
  <img alt="GitLab Release" src="https://img.shields.io/gitlab/v/release/kickr-dev%2Fkickr?gitlab_url=https%3A%2F%2Fgitlab.com&include_prereleases&sort=semver&style=for-the-badge">
  <img alt="GitLab Issues" src="https://img.shields.io/gitlab/issues/open/kickr-dev%2Fkickr?gitlab_url=https%3A%2F%2Fgitlab.com&style=for-the-badge">
  <img alt="GitLab License" src="https://img.shields.io/gitlab/license/kickr-dev%2Fkickr?gitlab_url=https%3A%2F%2Fgitlab.com&style=for-the-badge">
  <img alt="GitLab CICD" src="https://img.shields.io/gitlab/pipeline-status/kickr-dev%2Fkickr?gitlab_url=https%3A%2F%2Fgitlab.com&branch=main&style=for-the-badge">
  <img alt="Go Version" src="https://img.shields.io/gitlab/go-mod/go-version/kickr-dev/kickr?style=for-the-badge">
  <img alt="Go Report Card" src="https://goreportcard.com/badge/gitlab.com/kickr-dev/kickr?style=for-the-badge">
</div>

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

Aliases:
  generate, g

Flags:
  -f, --force   force generation of all files initially created by kickr (README.md, SECURITY.md, etc.) even if the initial generated notice has been removed
  -h, --help    help for generate

Global Flags:
  -d, --dir string          set directory where generation will be made (default is current directory)
      --log-format string   set logging format (either "text" or "json") (default "text")
      --log-level string    set logging level (default "info")
```

## Kickr file

TBD
