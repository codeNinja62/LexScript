#!/usr/bin/env bash
# LexScript — First-time setup script (Linux / macOS / WSL)
# Run this once after cloning the repository.
#
#   bash setup.sh
#
# What it does:
#   1. Builds the lexs binary
#   2. Installs the LexScript VS Code extension

set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ""
echo "LexScript Setup"
echo "==============="
echo ""

# ── Step 1: Build the Go binary ──────────────────────────────────────────────
echo "[1/2] Building lexs binary..."
(cd "$ROOT" && go build -o bin/lexs .)
echo "      OK: bin/lexs"

# ── Step 2: Install the VS Code extension ────────────────────────────────────
echo "[2/2] Installing LexScript VS Code extension..."

VSIX="$ROOT/vscode-extension/lexscript-0.4.0.vsix"
if [ ! -f "$VSIX" ]; then
    echo "      VSIX not found at $VSIX"
    echo "      Rebuild it with: cd vscode-extension && npm install && npm run package"
    exit 1
fi

if ! command -v code &>/dev/null; then
    echo "      'code' command not found. Add VS Code to PATH, then run:"
    echo "      code --install-extension \"$VSIX\""
    exit 1
fi

code --install-extension "$VSIX"
echo "      OK: LexScript extension installed"

echo ""
echo "Setup complete!"
echo "Reload VS Code (Ctrl+Shift+P -> 'Developer: Reload Window') then open a .lxs file."
echo ""
