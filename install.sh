#!/usr/bin/env bash
set -e

# codi installer
# Usage: curl -fsSL https://raw.githubusercontent.com/rakuv3r/codi/main/install.sh | bash

VERSION="latest"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
REPO="rakuv3r/codi"

echo "🔧 Installing codi..."

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux*)
        OS_NAME="linux"
        ;;
    Darwin*)
        OS_NAME="darwin"
        ;;
    *)
        echo "❌ Unsupported OS: $OS"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64|amd64)
        ARCH_NAME="amd64"
        ;;
    arm64|aarch64)
        ARCH_NAME="arm64"
        ;;
    *)
        echo "❌ Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Determine download URL
if [ "$VERSION" = "latest" ]; then
    DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/codi-${OS_NAME}-${ARCH_NAME}"
else
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/codi-${OS_NAME}-${ARCH_NAME}"
fi

echo "  Platform: ${OS_NAME}/${ARCH_NAME}"
echo "  URL: ${DOWNLOAD_URL}"

# Create install directory
mkdir -p "$INSTALL_DIR"

# Download binary
TEMP_FILE="$(mktemp)"
echo "  Downloading..."
if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "$TEMP_FILE" "$DOWNLOAD_URL"
elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$TEMP_FILE" "$DOWNLOAD_URL"
else
    echo "❌ Neither curl nor wget found. Please install one of them."
    exit 1
fi

# Install binary
chmod +x "$TEMP_FILE"
mv "$TEMP_FILE" "$INSTALL_DIR/codi"

echo "✅ codi installed to $INSTALL_DIR/codi"

# Check if install directory is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "⚠️  $INSTALL_DIR is not in your PATH"
    echo "   Add it to your shell profile:"
    echo ""
    if [ -f "$HOME/.zshrc" ]; then
        echo "   echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.zshrc"
    elif [ -f "$HOME/.bashrc" ]; then
        echo "   echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.bashrc"
    else
        echo "   export PATH=\"\$PATH:$INSTALL_DIR\""
    fi
    echo ""
fi

# Verify installation
echo ""
echo "🎉 Installation complete!"
echo "   Run 'codi --help' to get started"
