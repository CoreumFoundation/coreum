#!/bin/sh

set -e

VERSION="v3.0.3"
BIN_PATH="$HOME/bin/cored"

URL_LINUX_AMD64="https://github.com/CoreumFoundation/coreum/releases/download/${VERSION}/cored-linux-amd64"
URL_LINUX_ARM64="https://github.com/CoreumFoundation/coreum/releases/download/${VERSION}/cored-linux-arm64"
URL_DARWIN_AMD64="https://github.com/CoreumFoundation/coreum/releases/download/${VERSION}/cored-client-darwin-amd64"
URL_DARWIN_ARM64="https://github.com/CoreumFoundation/coreum/releases/download/${VERSION}/cored-client-darwin-arm64"

PLATFORM="$(uname)/$(uname -m)"
URL=""

case "$PLATFORM" in
  "Linux/x86_64") URL=$URL_LINUX_AMD64 ;;
  "Linux/arm64") URL=$URL_LINUX_ARM64 ;;
  "Darwin/x86_64") URL=$URL_DARWIN_AMD64 ;;
  "Darwin/arm64") URL=$URL_DARWIN_ARM64 ;;
  *) echo "Unsupported platform $PLATFORM"; exit 1
esac

echo "Downloading Coreum client..."

curl -L "$URL" --output "$BIN_PATH" --create-dirs
chmod u+x "$BIN_PATH"

echo "Coreum client installed in ${BIN_PATH}"
