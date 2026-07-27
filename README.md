<p align="center">
  <img src="logo.svg" alt="justx" width="440">
</p>

<div align="center">

[![Go Version](https://img.shields.io/github/go-mod/go-version/amberpixels/just-x)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-yellow.svg)](LICENSE)

</div>

---

**The [`just`](https://github.com/casey/just) companion.** 👾 `justx` is a small Go binary that wraps `just` and does two things:

1. **Expressive recipe names** - type `:`, `!`, and `?` in recipe names, which `just` doesn't allow ([#2669](https://github.com/casey/just/issues/2669), [#2587](https://github.com/casey/just/issues/2587)). justx translates them to plain names before calling real `just`.
2. **Scaffolding** - `j @init` detects your stack and writes a justfile from composable modules (`fmt`, `lint`, `fix`, `test`, `build`, `ci`).

```bash
j app:test           # runs → just app--test
j build!             # runs → just build-x
j ready?             # runs → just ready-q
j test               # unchanged
j @init              # detect project + scaffold a justfile
```

## Install

Requires [Go](https://go.dev/dl/). The installer builds the binary with `go install` and adds the `j` / `just` aliases to your shell.

```bash
git clone https://github.com/amberpixels/just-x.git
cd just-x && ./install.sh
```

Restart your shell (or `source` your rc file), then try `j @init`.

<details>
<summary>Manual install</summary>

```bash
go install github.com/amberpixels/just-x/cmd/justx@latest

# then add to ~/.zshrc (zsh) - noglob lets you type `?` unquoted:
alias j='noglob justx'
alias just='noglob justx'

# or ~/.bashrc (bash):
alias j='justx'
alias just='justx'
```

</details>

## Expressive Recipe Names

| You type | Runs | Mapping |
|---|---|---|
| `j app:build` | `just app--build` | `:` → `--` |
| `j dev!` | `just dev-x` | `!` → `-x` |
| `j ready?` | `just ready-q` | `?` → `-q` |

Name your recipes in the mapped form; type the expressive form in the terminal:

```justfile
lint--go:           # ← call with: j lint:go
    golangci-lint run

build-x:            # ← call with: j build!  (force rebuild, no cache)
    go build -a ./...

ready-q:            # ← call with: j ready?  (check if ready)
    ./check.sh
```

`j --list` (and `-l` / `--summary`) reverse-translates the output back to the expressive form, keeping comment alignment intact.

> **Why the alias?** A binary can't intercept the bare word `just`, and zsh expands `?` as a glob before any binary runs. The `noglob` alias solves both. `:` and `!` need no special handling; in bash, only `?` must be quoted (`j 'ready?'`).

## Scaffolding with `@init`

`j @init` detects your project (Go via `go.mod`, Node via `package.json`, shell via scripts in `./`, `bin/` or `scripts/`), shows a checkbox form of modules pre-ticked from detection, and writes a justfile. Use `--yes` to skip the form and accept the pre-ticked set.

Go projects that pin [standardgo](https://github.com/amberpixels/standardgo) as a tool dependency get `go tool standardgo` recipes for `fmt` / `lint` / `fix`; everything else falls back to `go fmt` and `golangci-lint`.

Shell projects get [shellcheck](https://www.shellcheck.net) for correctness and [shfmt](https://github.com/mvdan/sh) for formatting. Scripts are discovered at run time by `shfmt -f`, which matches on extension *and* on shebang, so an extensionless `bin/deploy` is covered without being listed anywhere. Shell is detected last, so a Go or Node repo that merely ships a few helper scripts keeps its real stack. There is no `test` module: no shell test runner is conventional enough to assume.

Each module is written inside provenance fences so a future `j @upgrade` can re-sync it without touching your edits:

```justfile
# >>> justx:lint (managed) — `j @upgrade` re-syncs; remove these fences to take over
# lint Go code - reports findings, changes nothing
lint:
    golangci-lint run
# <<< justx:lint
```

`@init` refuses to overwrite an existing justfile (merging is `@upgrade`, coming in a later release).

## Meta-Commands

Meta-commands live under the `@` sigil and are handled by justx, never passed to `just`:

```
j @init [--yes]   detect project + scaffold a justfile
j @help           show help
j @version        print version
```

Planned: `@add`, `@upgrade`, `@doctor`, `@templates`.

## Configuration

Override the character mappings via environment variables (set before the aliases in your rc file):

```bash
export JUST_X_BANG="-x"       # ! replacement (default: -x)
export JUST_X_QUESTION="-q"   # ? replacement (default: -q)
export JUST_X_COLON="--"      # : replacement (default: --)
```

## Uninstall

```bash
./uninstall.sh
```

## Requirements

- [Go 1.26+](https://go.dev/dl/)
- [just](https://github.com/casey/just)
- zsh or bash

## Feedback

justx is a solo, opinionated project - but if you stumbled upon it and have ideas, questions, or bug reports, an [issue](https://github.com/amberpixels/just-x/issues) is always welcome :)

## License

[MIT](LICENSE) © [amberpixels](https://amberpixels.io)
