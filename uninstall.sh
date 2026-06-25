#!/usr/bin/env bash
set -eo pipefail

# justx uninstaller — removes the alias block from your rc file and deletes the
# installed binary.

GREEN='\033[32m'; YELLOW='\033[33m'; BOLD='\033[1m'; NC='\033[0m'

# --- Detect rc file ---
SHELL_NAME="$(basename "${SHELL:-}")"
case "$SHELL_NAME" in
  zsh)  RC_FILE="${HOME}/.zshrc" ;;
  bash) RC_FILE="${HOME}/.bashrc" ;;
  *)    RC_FILE="" ;;
esac

# --- Remove the alias block ---
if [[ -n "$RC_FILE" && -f "$RC_FILE" ]] && grep -qF '# >>> justx >>>' "$RC_FILE" 2>/dev/null; then
  awk '
    /# >>> justx >>>/ { skip = 1 }
    !skip            { print }
    /# <<< justx <<</ { skip = 0 }
  ' "$RC_FILE" > "${RC_FILE}.tmp" && mv "${RC_FILE}.tmp" "$RC_FILE"
  printf "${GREEN}✓${NC} removed justx aliases from %s\n" "$RC_FILE"
else
  printf "${YELLOW}⚠ no justx alias block found in your rc file${NC}\n"
fi

# --- Remove the binary ---
if command -v go >/dev/null 2>&1; then
  GOBIN="$(go env GOBIN)"
  [[ -z "$GOBIN" ]] && GOBIN="$(go env GOPATH)/bin"
  if [[ -x "${GOBIN}/justx" ]]; then
    rm -f "${GOBIN}/justx"
    printf "${GREEN}✓${NC} removed %s/justx\n" "$GOBIN"
  fi
fi

printf "  Restart your shell to finish. ${BOLD}Done.${NC}\n"
