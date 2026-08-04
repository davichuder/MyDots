#!/bin/sh
# Caveman install script — wrapper for the caveman curl-script with --only openclaw.
#
# This script is embedded via //go:embed assets and executed by the Caveman
# module through runner.Script(). It must be POSIX-compliant (NFR-10).
#
# https://github.com/JuliusBrussee/caveman

set -eu

openclaw_workspace=${OPENCLAW_WORKSPACE:-"$HOME/.openclaw/workspace"}
caveman_skill="$openclaw_workspace/skills/caveman/SKILL.md"
caveman_soul="$openclaw_workspace/SOUL.md"

if [ -f "$caveman_skill" ] && [ -s "$caveman_skill" ] && [ -f "$caveman_soul" ]; then
	    caveman_begin_count=0
	    caveman_end_count=0
	    caveman_end_after_begin=false
	    while IFS= read -r line || [ -n "$line" ]; do
	        case "$line" in
	            "<!-- caveman-begin -->") caveman_begin_count=$((caveman_begin_count + 1)) ;;
	            "<!-- caveman-end -->")
	                caveman_end_count=$((caveman_end_count + 1))
	                if [ "$caveman_begin_count" -gt 0 ]; then
	                    caveman_end_after_begin=true
	                fi
	                ;;
	        esac
	    done <"$caveman_soul"
	    if [ "$caveman_begin_count" -eq 1 ] && [ "$caveman_end_count" -eq 1 ] && [ "$caveman_end_after_begin" = true ]; then
        echo "==> Caveman is already installed in the OpenClaw workspace; skipping."
        exit 0
    fi
fi

for tool in curl mktemp bash node npx; do
    if ! command -v "$tool" >/dev/null 2>&1; then
        if [ "$tool" = "node" ]; then
            echo "Error: Node.js 18 or newer is required to install Caveman." >&2
        elif [ "$tool" = "npx" ]; then
            echo "Error: npx is required to install Caveman. Reinstall Node.js 18 or newer." >&2
        else
            echo "Error: $tool is required to install Caveman." >&2
        fi
        exit 1
    fi
done

node_major=$(node -p "process.versions.node.split('.')[0]" 2>/dev/null) || {
    echo "Error: unable to determine the Node.js version. Caveman requires Node.js 18 or newer." >&2
    exit 1
}
case "$node_major" in
    ''|*[!0-9]*)
        echo "Error: unable to determine the Node.js version. Caveman requires Node.js 18 or newer." >&2
        exit 1
        ;;
esac
if [ "$node_major" -lt 18 ]; then
    echo "Error: Node.js 18 or newer is required to install Caveman; found Node $node_major. Upgrade Node.js: https://nodejs.org" >&2
    exit 1
fi

echo "==> Installing caveman (openclaw only)..."
install_script=$(mktemp "${TMPDIR:-/tmp}/mydots-caveman.XXXXXX")
trap 'rm -f "$install_script"' 0 1 2 15

if ! curl -fsSL "https://raw.githubusercontent.com/JuliusBrussee/caveman/main/install.sh" >"$install_script"; then
    echo "Error: failed to download the Caveman installer." >&2
    exit 1
fi

if ! bash "$install_script" --only openclaw --force; then
    echo "Error: Caveman installer failed." >&2
    exit 1
fi

echo "==> Caveman installation complete"
