#!/bin/sh
# Oh My Zsh install script — unattended install with env vars.
#
# This script is embedded via //go:embed assets and executed by the Oh My Zsh
# module through runner.Script(). It must be POSIX-compliant (NFR-10).
#
# https://ohmyz.sh

set -eu

if [ -d "${HOME:-}/.oh-my-zsh" ]; then
    echo "==> Oh My Zsh is already installed; skipping."
    exit 0
fi

if ! command -v curl >/dev/null 2>&1; then
    echo "Error: curl is required to install Oh My Zsh." >&2
    exit 1
fi

if ! command -v mktemp >/dev/null 2>&1; then
    echo "Error: mktemp is required to install Oh My Zsh." >&2
    exit 1
fi

if ! command -v sh >/dev/null 2>&1; then
    echo "Error: sh is required to install Oh My Zsh." >&2
    exit 1
fi

echo "==> Downloading Oh My Zsh install script..."
install_script=$(mktemp)
trap 'rm -f "$install_script"' 0 1 2 15

if ! curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh >"$install_script"; then
    echo "Error: failed to download the Oh My Zsh installer." >&2
    exit 1
fi

RUNZSH=no CHSH=no sh "$install_script"

echo "==> Oh My Zsh installation complete"
