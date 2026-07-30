#!/bin/sh
set -eu

# Installs the documented community-maintained Ubuntu package; it is not an
# official Ghostty project binary.

for tool in curl mktemp bash; do
    if ! command -v "$tool" >/dev/null 2>&1; then
        echo "Error: $tool is required to install Ghostty on Ubuntu." >&2
        exit 1
    fi
done

installer=$(mktemp "${TMPDIR:-/tmp}/mydots-ghostty-ubuntu.XXXXXX")
trap 'rm -f "$installer"' 0 1 2 15

url="https://raw.githubusercontent.com/mkasberg/ghostty-ubuntu/HEAD/install.sh"
if ! curl -fsSL -o "$installer" "$url"; then
    echo "Error: failed to download the Ghostty Ubuntu installer." >&2
    exit 1
fi

if ! bash "$installer"; then
    echo "Error: Ghostty Ubuntu installer failed." >&2
    exit 1
fi

echo "==> Ghostty installation complete"
