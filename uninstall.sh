#!/usr/bin/env bash
set -euo pipefail

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
  echo "Removed ${INSTALL_DIR}/just-x"
else
  echo "just-x not found in ${INSTALL_DIR}, skipping."
fi

# --- Remove eval line from rc file ---
EVAL_LINE='eval "$(just-x init)"'

if [[ -n "$RC_FILE" && -f "$RC_FILE" ]]; then
  if grep -qF "$EVAL_LINE" "$RC_FILE" 2>/dev/null; then
    grep -vF "$EVAL_LINE" "$RC_FILE" > "${RC_FILE}.tmp"
    mv "${RC_FILE}.tmp" "$RC_FILE"
    echo "Removed just-x init from ${RC_FILE}"
  fi
else
  echo "Could not detect shell rc file. Manually remove this line from your rc file:"
  echo "  $EVAL_LINE"
fi

# --- Summary ---
echo ""
echo "Done! just-x has been uninstalled."
echo ""
echo "Restart your shell or run:"
echo "  source ${RC_FILE:-~/.zshrc}"
