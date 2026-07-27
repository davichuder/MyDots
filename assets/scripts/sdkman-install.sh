#!/bin/sh
# SDKMAN install script — unattended install.
#
# This script is embedded via //go:embed assets and executed by the Sdkman
# module through runner.Script(). It must be POSIX-compliant (NFR-10).
#
# https://sdkman.io

set -euo pipefail

echo "==> Downloading SDKMAN install script..."
curl -fsSL "https://get.sdkman.io" | bash

echo "==> SDKMAN installation complete"
