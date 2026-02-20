# just-x

Extended recipe names for [just](https://github.com/casey/just).

`just` is a fantastic command runner — simple, fast, and rock-solid. It intentionally doesn't allow `:`, `!`, `?` in recipe names ([#2587](https://github.com/casey/just/issues/2587), [#2669](https://github.com/casey/just/issues/2669)), and the maintainer's focus on stability and a clean syntax is totally fair. Rather than forking or requesting changes, `just-x` takes the lightweight approach: a thin shell wrapper that adds expressive recipe names via convention-based character substitution.

```
just build            →  just build
just build!           →  just build-x        # force rebuild, clean & build

just lint:go          →  just lint--go
just lint:frontend    →  just lint--frontend  # namespaced recipe groups

just deploy           →  just deploy
just deploy?          →  just deploy-q       # dry-run, what would happen?
```

Write expressive recipe names in your terminal. `just-x` translates them to valid justfile names behind the scenes.

## Installation

### One-liner

```bash
curl -fsSL https://raw.githubusercontent.com/amberpixels/just-x/main/install.sh | bash
```

### From a local clone

```bash
git clone https://github.com/amberpixels/just-x.git
cd just-x
./install.sh
```

Both methods will:
- Install `just-x` to `~/.local/bin/`
- Add `~/.local/bin` to your `PATH` (if needed)
- Add `eval "$(just-x init)"` to your shell rc file

After install, restart your shell and `just` will support extended recipe names.

### Custom function name

By default `just-x init` creates a function named `just` (overrides `just` itself). You can use a custom name instead:

```bash
eval "$(just-x init j)"    # use j as the command
eval "$(just-x init jj)"   # use jj as the command
```

## How it works

| You type     | Justfile recipe | Default mapping |
|-------------|----------------|-----------------|
| `just dev!`      | `dev-x`        | `!` → `-x`     |
| `just ready?`    | `ready-q`      | `?` → `-q`     |
| `just app:build` | `app--build`   | `:` → `--`     |

## Examples

```bash
# In your justfile, name recipes with the convention:
# dev-x:        (maps from dev!)
# deploy-x:     (maps from deploy!)
# ready-q:      (maps from ready?)
# app--build:   (maps from app:build)

just dev!              # runs: just dev-x
just deploy! staging   # runs: just deploy-x staging
just ready?            # runs: just ready-q
just app:build         # runs: just app--build

# Regular recipes pass through unchanged
just test              # runs: just test
just --list            # runs: just --list
```

## Configuration

Override the default mappings with environment variables:

```bash
export JUST_X_BANG="-x"       # ! replacement (default: -x)
export JUST_X_QUESTION="-q"   # ? replacement (default: -q)
export JUST_X_COLON="--"      # : replacement (default: --)
```

## Commands

```bash
just-x init [name]         # Generate shell function (default: just)
just-x translate <recipe>  # Show translation (for debugging)
just-x version             # Print version
just-x help                # Show help
```

## License

MIT
