#!/bin/sh
set -e

REPO="frank-chris/burrow"
BIN="burrow"
INSTALL_DIR="/usr/local/bin"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  darwin) OS="darwin" ;;
  linux)  OS="linux" ;;
  *)
    echo "Unsupported OS: $OS"
    echo "Download manually: https://github.com/$REPO/releases"
    exit 1
    ;;
esac

# Detect arch
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)          ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    echo "Download manually: https://github.com/$REPO/releases"
    exit 1
    ;;
esac

# Get latest version tag
VERSION=$(curl -sSf "https://api.github.com/repos/$REPO/releases/latest" \
  | grep '"tag_name"' | cut -d'"' -f4)

if [ -z "$VERSION" ]; then
  echo "Could not fetch latest version from GitHub."
  exit 1
fi

FILENAME="${BIN}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$FILENAME"

echo "Installing $BIN $VERSION ($OS/$ARCH)..."

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -sSfL "$URL" -o "$TMP/$FILENAME"
tar -xzf "$TMP/$FILENAME" -C "$TMP"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/$BIN" "$INSTALL_DIR/$BIN"
else
  sudo mv "$TMP/$BIN" "$INSTALL_DIR/$BIN"
fi

echo ""
echo "Installed to $INSTALL_DIR/$BIN"
echo "Run 'burrow --version' to verify."
