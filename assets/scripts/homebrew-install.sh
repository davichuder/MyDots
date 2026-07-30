#!/bin/sh
# Homebrew install script — fetches and runs the official Homebrew installer.
#
# This script is embedded via //go:embed assets and executed by the Homebrew
# module through runner.Script(). It must be POSIX-compliant (NFR-10).
#
# https://brew.sh

set -eu

echo "==> Downloading Homebrew install script..."
install_script=$(mktemp)
trap 'rm -f "$install_script"' 0 1 2 15

if ! curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh >"$install_script"; then
    echo "Error: failed to download the Homebrew installer." >&2
    exit 1
fi

NONINTERACTIVE=1 /bin/bash "$install_script"

echo "==> Homebrew installation complete"
