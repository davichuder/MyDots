#!/bin/sh
# Install a Nerd Font archive on Linux. This script is embedded and invoked by
# M-47 with FONT_NAME set to a supported release archive name.

set -eu

: "${FONT_NAME:?FONT_NAME must name a Nerd Font zip archive}"

case "$FONT_NAME" in
    *[!A-Za-z0-9._-]* | .* | *..* | *"/"*)
        echo "Error: FONT_NAME must be a safe archive filename." >&2
        exit 1
        ;;
esac

if ! command -v unzip >/dev/null 2>&1; then
    echo "Error: unzip is required to install Nerd Fonts." >&2
    exit 1
fi

font_dir="$HOME/.local/share/fonts"
archive=$(mktemp "${TMPDIR:-/tmp}/mydots-font.XXXXXX")
trap 'rm -f "$archive"' 0 1 2 15

mkdir -p "$font_dir"

url="https://github.com/ryanoasis/nerd-fonts/releases/latest/download/$FONT_NAME"
if ! curl -fL -o "$archive" "$url"; then
    echo "Error: failed to download Nerd Font archive: $FONT_NAME" >&2
    exit 1
fi

unzip -o "$archive" -d "$font_dir"
fc-cache -fv "$font_dir"

echo "==> Nerd Font installation complete"
