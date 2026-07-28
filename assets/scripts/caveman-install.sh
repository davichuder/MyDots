#!/bin/sh
# Caveman install script — wrapper for the caveman curl-script with --only openclaw.
#
# This script is embedded via //go:embed assets and executed by the Caveman
# module through runner.Script(). It must be POSIX-compliant (NFR-10).
#
# https://github.com/caveman-cli/caveman

set -eu

echo "==> Installing caveman (openclaw only)..."
curl -fsSL "https://caveman.sh/install" | sh -s -- --only openclaw

echo "==> Caveman installation complete"
