#!/bin/sh
# Oh My Zsh install script — unattended install with env vars.
#
# This script is embedded via //go:embed assets and executed by the Oh My Zsh
# module through runner.Script(). It must be POSIX-compliant (NFR-10).
#
# https://ohmyz.sh

set -eu

echo "==> Downloading Oh My Zsh install script..."
install_script=$(mktemp)
trap 'rm -f "$install_script"' 0 1 2 15

if ! curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh >"$install_script"; then
    echo "Error: failed to download the Oh My Zsh installer." >&2
    exit 1
fi

sh "$install_script"

echo "==> Oh My Zsh installation complete"
