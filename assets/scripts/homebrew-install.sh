#!/bin/sh
# Homebrew install script — fetches and runs the official Homebrew installer.
#
# This script is embedded via //go:embed assets and executed by the Homebrew
# module through runner.Script(). It must be POSIX-compliant (NFR-10).
#
# https://brew.sh

set -euo pipefail

echo "==> Downloading Homebrew install script..."
NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

echo "==> Homebrew installation complete"
