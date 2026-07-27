#!/bin/sh
# Oh My Zsh install script — unattended install with env vars.
#
# This script is embedded via //go:embed assets and executed by the Oh My Zsh
# module through runner.Script(). It must be POSIX-compliant (NFR-10).
#
# https://ohmyz.sh

set -euo pipefail

echo "==> Downloading Oh My Zsh install script..."
sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"

echo "==> Oh My Zsh installation complete"
