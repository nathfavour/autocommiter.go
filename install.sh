#!/bin/bash

# Autocommiter Universal Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/nathfavour/autocommiter.go/master/install.sh | bash

set -e

REPO="nathfavour/autocommiter.go"
GITHUB_URL="https://github.com/$REPO"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

if [ "$OS" = "darwin" ]; then
    OS="darwin"
elif [ "$OS" = "linux" ]; then
    # Check for Android (Termux)
    if [ -n "$TERMUX_VERSION" ] || [ -d "/data/data/com.termux" ]; then
        OS="android"
    else
        OS="linux"
    fi
else
    echo "Unsupported OS: $OS"
    exit 1
fi

BINARY_NAME="autocommiter-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    BINARY_NAME+=".exe"
fi

echo "Detected Platform: $OS/$ARCH"

# Get latest release tag
echo "Fetching release metadata..."
LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep -Po '"tag_name": "\K.*?(?=")' || echo "latest")

if [ -z "$LATEST_TAG" ]; then
    echo "Error: Failed to fetch latest release from GitHub."
    exit 1
fi

echo "Resolved version: $LATEST_TAG"

# Download binary
DOWNLOAD_URL="$GITHUB_URL/releases/download/$LATEST_TAG/$BINARY_NAME"

echo "Downloading $BINARY_NAME ($LATEST_TAG)..."
curl -fsSL "$DOWNLOAD_URL" -o autocommiter

# Install binary
INSTALL_DIR="$HOME/.local/bin"

if [ ! -d "$INSTALL_DIR" ]; then
    echo "Creating installation directory: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR"
fi

mv autocommiter "$INSTALL_DIR/autocommiter"
chmod +x "$INSTALL_DIR/autocommiter"

echo "Successfully installed autocommiter to $INSTALL_DIR/autocommiter"

# Ensure the install dir is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "⚠️  Note: $INSTALL_DIR is not in your PATH."
    echo "You may need to add it to your shell configuration (.bashrc, .zshrc, etc.):"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
fi

# Verify
"$INSTALL_DIR/autocommiter" version || "$INSTALL_DIR/autocommiter" --help || true

echo ""
echo "Run 'autocommiter --help' to get started!"
