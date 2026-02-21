#!/usr/bin/env bash
set -eo pipefail

INSTALL_DIR="${HOME}/.local/bin"

# --- Detect shell and rc file ---
detect_rc_file() {
  case "${SHELL:-}" in
    */zsh)  echo "${HOME}/.zshrc" ;;
    */bash) echo "${HOME}/.bashrc" ;;
    *)      echo "" ;;
  esac
}

RC_FILE="$(detect_rc_file)"

# --- Remove binary ---
if [[ -f "${INSTALL_DIR}/just-x" ]]; then
  rm "${INSTALL_DIR}/just-x"
else
  printf '\033[33m⚠ just-x not found in %s\033[0m\n' "$INSTALL_DIR"
  exit 0
fi

# --- Remove eval line from rc file ---
EVAL_LINE='eval "$(just-x init)"'

if [[ -n "$RC_FILE" && -f "$RC_FILE" ]]; then
  if grep -qF "$EVAL_LINE" "$RC_FILE" 2>/dev/null; then
    grep -vF "$EVAL_LINE" "$RC_FILE" > "${RC_FILE}.tmp"
    mv "${RC_FILE}.tmp" "$RC_FILE"
  fi
else
  printf '\033[33m⚠ Could not detect shell. Manually remove:\033[0m %s\n' "$EVAL_LINE"
  exit 0
fi

# --- Summary ---
printf '\033[32m✓\033[0m just-x uninstalled\n'
printf '  Restart your shell or run: \033[1msource %s\033[0m\n' "$RC_FILE"
