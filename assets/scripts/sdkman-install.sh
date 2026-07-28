#!/bin/sh
# SDKMAN install script — unattended install.
#
# This script is embedded via //go:embed assets and executed by the Sdkman
# module through runner.Script(). It must be POSIX-compliant (NFR-10).
#
# https://sdkman.io

set -eu

echo "==> Downloading SDKMAN install script..."
install_script=$(mktemp)
trap 'rm -f "$install_script"' 0 1 2 15

if ! curl -fsSL "https://get.sdkman.io" >"$install_script"; then
    echo "Error: failed to download the SDKMAN installer." >&2
    exit 1
fi

bash "$install_script"

echo "==> SDKMAN installation complete"
