#!/bin/sh
set -e

# ghostty-linux.sh — Install Ghostty terminal on Linux (native or WSL2 via WSLg)
# Called by the MyDots installer (M-48).
# Downloads the latest Ghostty release from GitHub, extracts, and installs.

GHOSTTY_VERSION="1.0.0"
ARCH="$(uname -m)"
TARBALL="ghostty-${GHOSTTY_VERSION}-${ARCH}-linux.tar.gz"
URL="https://github.com/ghostty-org/ghostty/releases/download/v${GHOSTTY_VERSION}/${TARBALL}"

mkdir -p ~/.local/bin ~/.config/ghostty
curl -fsSL "$URL" -o "/tmp/${TARBALL}"
tar -xzf "/tmp/${TARBALL}" -C ~/.local/bin/
rm -f "/tmp/${TARBALL}"
