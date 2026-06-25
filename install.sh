#!/usr/bin/env bash
set -eo pipefail

# justx installer — builds the binary with `go install` and wires the shell
# aliases (`j` and `just`) needed to shadow real `just` and to dodge zsh's `?`
# globbing via `noglob`.

REPO="github.com/amberpixels/just-x"
CMD_PKG="${REPO}/cmd/justx"

GREEN='\033[32m'; YELLOW='\033[33m'; RED='\033[31m'; BOLD='\033[1m'; NC='\033[0m'

# --- Require Go ---
if ! command -v go >/dev/null 2>&1; then
  printf "${RED}✗ Go is required${NC} — install it from https://go.dev/dl/ and retry.\n" >&2
  exit 1
fi

# --- Detect shell + rc file ---
SHELL_NAME="$(basename "${SHELL:-}")"
case "$SHELL_NAME" in
  zsh)  RC_FILE="${HOME}/.zshrc" ;;
  bash) RC_FILE="${HOME}/.bashrc" ;;
  *)    RC_FILE="" ;;
esac

# --- Install the binary ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" 2>/dev/null && pwd)" || SCRIPT_DIR=""
if [[ -n "$SCRIPT_DIR" && -d "${SCRIPT_DIR}/cmd/justx" ]]; then
  printf "Installing justx from source (%s)…\n" "$SCRIPT_DIR"
  (cd "$SCRIPT_DIR" && go install ./cmd/justx)
else
  printf "Installing justx via go install (%s@latest)…\n" "$CMD_PKG"
  go install "${CMD_PKG}@latest"
fi

# --- Locate the Go bin dir ---
GOBIN="$(go env GOBIN)"
[[ -z "$GOBIN" ]] && GOBIN="$(go env GOPATH)/bin"

if [[ ! -x "${GOBIN}/justx" ]]; then
  printf "${RED}✗ justx was not found in %s after install${NC}\n" "$GOBIN" >&2
  exit 1
fi

# --- Build the alias block (noglob only makes sense in zsh) ---
PREFIX=""
[[ "$SHELL_NAME" == "zsh" ]] && PREFIX="noglob "

read -r -d '' ALIAS_BLOCK <<EOF || true
# >>> justx >>>
alias j='${PREFIX}justx'
alias just='${PREFIX}justx'
# <<< justx <<<
EOF

if [[ -z "$RC_FILE" ]]; then
  printf "${YELLOW}⚠ Could not detect your shell.${NC} Add these aliases manually:\n%s\n" "$ALIAS_BLOCK"
  exit 0
fi

# --- Ensure the Go bin dir is on PATH ---
if ! printf '%s' "$PATH" | tr ':' '\n' | grep -qx "$GOBIN"; then
  if ! grep -qF "export PATH=\"${GOBIN}:\$PATH\"" "$RC_FILE" 2>/dev/null; then
    {
      echo ''
      echo "export PATH=\"${GOBIN}:\$PATH\""
    } >> "$RC_FILE"
  fi
fi

# --- Add the alias block (idempotent) ---
if ! grep -qF '# >>> justx >>>' "$RC_FILE" 2>/dev/null; then
  {
    echo ''
    echo "$ALIAS_BLOCK"
  } >> "$RC_FILE"
fi

printf "${GREEN}✓${NC} justx installed to %s\n" "$GOBIN"
printf "  Restart your shell or run: ${BOLD}source %s${NC}\n" "$RC_FILE"
printf "  Then try: ${BOLD}j @init${NC}\n"
