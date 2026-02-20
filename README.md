# ⚡ just-x

[![Shell: bash | zsh](https://img.shields.io/badge/shell-bash%20%7C%20zsh-green.svg)](#requirements)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

<p align="center"><i>Use <code>!</code>, <code>?</code>, and <code>:</code> in <a href="https://github.com/casey/just">just</a> recipe names.</i></p>

[`just`](https://github.com/casey/just) keeps recipe names clean and simple by design — no `!`, `?`, or `:` allowed ([#2587](https://github.com/casey/just/issues/2587), [#2669](https://github.com/casey/just/issues/2669)).
`just-x` is a thin shell wrapper that adds them back via convention-based character substitution.

```bash
just lint:frontend   # runs → just lint--frontend
just lint:go         # runs → just lint--go
just build!          # runs → just build-x
just ready?          # runs → just ready-q
just test            # unchanged
```

## 🚀 Install

```bash
curl -fsSL https://raw.githubusercontent.com/amberpixels/just-x/main/install.sh | bash
```

Restart your shell. Done — `just` now supports extended recipe names.

<details>
<summary>Other install methods</summary>

**From source:**

```bash
git clone https://github.com/amberpixels/just-x.git
cd just-x && ./install.sh
```

**Manual:**

```bash
cp just-x ~/.local/bin/
echo 'eval "$(just-x init)"' >> ~/.zshrc
```

</details>

## 🔀 How it works

| You type | Runs | Mapping |
|---|---|---|
| `just app:build` | `just app--build` | `:` → `--` |
| `just dev!` | `just dev-x` | `!` → `-x` |
| `just ready?` | `just ready-q` | `?` → `-q` |

Name your justfile recipes using the mapped form, type the expressive form in the terminal:

```justfile
# justfile

lint--go:           # ← call with: just lint:go
    golangci-lint run

lint--frontend:     # ← call with: just lint:frontend
    eslint src/

build:              # regular build
    cargo build

build-x:            # ← call with: just build!  (force rebuild, skip cache)
    cargo build --release --force

deploy:             # deploy with confirmation prompt
    ./deploy.sh --confirm

deploy-x:           # ← call with: just deploy!  (skip confirmation, YOLO)
    ./deploy.sh --yes

ready-q:            # ← call with: just ready?  (check if ready to deploy)
    ./check-deps.sh && ./check-env.sh
```

## ⚙️ Configuration

Override mappings via environment variables (set before the `eval` line in your rc file):

```bash
export JUST_X_BANG="-x"       # ! replacement (default: -x)
export JUST_X_QUESTION="-q"   # ? replacement (default: -q)
export JUST_X_COLON="--"      # : replacement (default: --)
```

By default, `just-x init` creates a function named `just`. Use a custom name:

```bash
eval "$(just-x init j)"    # use j instead
```

## 🗑️ Uninstall

```bash
curl -fsSL https://raw.githubusercontent.com/amberpixels/just-x/main/uninstall.sh | bash
```

## 📋 Requirements

- [just](https://github.com/casey/just)
- Bash or Zsh

## 📄 License

[MIT](LICENSE)
