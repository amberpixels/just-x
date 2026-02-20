#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="${HOME}/.local/bin"
REPO="amberpixels/just-x"
BRANCH="main"

# --- Detect shell and rc file ---
detect_rc_file() {
  case "${SHELL:-}" in
    */zsh)  echo "${HOME}/.zshrc" ;;
    */bash) echo "${HOME}/.bashrc" ;;
    *)      echo "" ;;
  esac
}

RC_FILE="$(detect_rc_file)"

# --- Install binary ---
mkdir -p "$INSTALL_DIR"

echo "Installing just-x to ${INSTALL_DIR}/just-x..."

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" 2>/dev/null && pwd)" || SCRIPT_DIR=""

if [[ -n "$SCRIPT_DIR" && -f "${SCRIPT_DIR}/just-x" ]]; then
  cp "${SCRIPT_DIR}/just-x" "${INSTALL_DIR}/just-x"
elif command -v curl &>/dev/null; then
  curl -fsSL "https://raw.githubusercontent.com/${REPO}/${BRANCH}/just-x" -o "${INSTALL_DIR}/just-x"
elif command -v wget &>/dev/null; then
  wget -qO "${INSTALL_DIR}/just-x" "https://raw.githubusercontent.com/${REPO}/${BRANCH}/just-x"
else
  echo "Error: curl or wget required" >&2
  exit 1
fi

chmod +x "${INSTALL_DIR}/just-x"

# --- Ensure ~/.local/bin is in PATH ---
if [[ -n "$RC_FILE" ]] && ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
  if ! grep -qF 'export PATH="$HOME/.local/bin:$PATH"' "$RC_FILE" 2>/dev/null; then
    echo '' >> "$RC_FILE"
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$RC_FILE"
    echo "Added ~/.local/bin to PATH in ${RC_FILE}"
  fi
  export PATH="${INSTALL_DIR}:${PATH}"
fi

# --- Add eval line to rc file ---
EVAL_LINE='eval "$(just-x init)"'

if [[ -n "$RC_FILE" ]]; then
  if ! grep -qF "$EVAL_LINE" "$RC_FILE" 2>/dev/null; then
    echo '' >> "$RC_FILE"
    echo "$EVAL_LINE" >> "$RC_FILE"
    echo "Added just-x init to ${RC_FILE}"
  fi
else
  echo "Could not detect shell rc file. Manually add to your rc file:"
  echo "  $EVAL_LINE"
fi

# --- Summary ---
echo ""
echo "Done! just-x installed successfully."
echo ""
echo "Restart your shell or run:"
echo "  source ${RC_FILE:-~/.zshrc}"
