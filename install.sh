#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Check for Go
if ! command -v go &>/dev/null; then
    if command -v brew &>/dev/null; then
        read -rp "Go is not installed. Install via Homebrew? [y/N] " answer
        if [[ "$answer" =~ ^[Yy]$ ]]; then
            brew install go
        else
            echo "Aborted. Install Go manually from https://go.dev/dl/"
            exit 1
        fi
    else
        echo "Error: Go is not installed. Install it from https://go.dev/dl/ or via Homebrew (brew install go)."
        exit 1
    fi
fi

# Build
echo "Building st..."
cd "$(dirname "$0")"
go build -o st .

# Install
if [ -w "$INSTALL_DIR" ]; then
    mv st "$INSTALL_DIR/st"
else
    echo "Installing to $INSTALL_DIR (requires sudo)..."
    sudo mv st "$INSTALL_DIR/st"
fi

echo "Installed st to $INSTALL_DIR/st"
echo "Run 'st --help' to get started."
