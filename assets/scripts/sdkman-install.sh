#!/bin/sh
# SDKMAN install script — unattended install.
#
# This script is embedded via //go:embed assets and executed by the Sdkman
# module through runner.Script(). It must be POSIX-compliant (NFR-10).
#
# https://sdkman.io

set -eu

if [ -f "${HOME:-}/.sdkman/bin/sdkman-init.sh" ]; then
    echo "==> SDKMAN is already installed; skipping."
    exit 0
fi

if ! command -v curl >/dev/null 2>&1; then
    echo "Error: curl is required to install SDKMAN." >&2
    exit 1
fi

if ! command -v mktemp >/dev/null 2>&1; then
    echo "Error: mktemp is required to install SDKMAN." >&2
    exit 1
fi

if ! command -v bash >/dev/null 2>&1; then
    echo "Error: bash is required to install SDKMAN." >&2
    exit 1
fi

bash_version=$(bash --version 2>/dev/null || true)
case "$bash_version" in
    *" version "*)
        bash_version=${bash_version#*" version "}
        bash_version=${bash_version%% *}
        bash_major=${bash_version%%.*}
        ;;
    *)
        bash_major=""
        ;;
esac

case "$bash_major" in
    ''|*[!0-9]*)
        echo "Error: unable to determine the Bash version. SDKMAN requires Bash 4 or newer; install Bash 4+ and ensure it is first on PATH." >&2
        exit 1
        ;;
esac

if [ "$bash_major" -lt 4 ]; then
    echo "Error: Bash 4 or newer is required to install SDKMAN; found Bash $bash_major. Upgrade Bash and ensure it is first on PATH." >&2
    exit 1
fi

echo "==> Downloading SDKMAN install script..."
install_script=$(mktemp)
trap 'rm -f "$install_script"' 0 1 2 15

if ! curl -fsSL "https://get.sdkman.io" >"$install_script"; then
    echo "Error: failed to download the SDKMAN installer." >&2
    exit 1
fi

SDKMAN_DIR="${HOME}/.sdkman" bash "$install_script"

echo "==> SDKMAN installation complete"
