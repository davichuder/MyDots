#!/bin/sh
# Caveman install script — wrapper for the caveman curl-script with --only openclaw.
#
# This script is embedded via //go:embed assets and executed by the Caveman
# module through runner.Script(). It must be POSIX-compliant (NFR-10).
#
# https://github.com/caveman-cli/caveman

set -eu

echo "==> Installing caveman (openclaw only)..."
install_script=$(mktemp)
trap 'rm -f "$install_script"' 0 1 2 15

if ! curl -fsSL "https://caveman.sh/install" >"$install_script"; then
    echo "Error: failed to download the Caveman installer." >&2
    exit 1
fi

sh "$install_script" --only openclaw

echo "==> Caveman installation complete"
