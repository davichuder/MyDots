#!/bin/sh
# docker-linux.sh — Install Docker Engine on supported Debian/Ubuntu systems.
# The official Docker installer configures the APT repository and installs the
# Docker Engine package family.

set -eu

if [ -z "${USER:-}" ]; then
    echo "Error: USER must be set to configure Docker group membership." >&2
    exit 1
fi

if command -v docker >/dev/null 2>&1; then
    echo "==> Docker is already installed; skipping package installation."
else
    for command_name in curl mktemp sudo sh id getent usermod; do
        if ! command -v "$command_name" >/dev/null 2>&1; then
            echo "Error: $command_name is required to install Docker." >&2
            exit 1
        fi
    done

    echo "==> Downloading the official Docker APT installer..."
    install_script=$(mktemp "${TMPDIR:-/tmp}/mydots-docker.XXXXXX")
    trap 'rm -f "$install_script"' 0 1 2 15

    if ! curl -fsSL https://get.docker.com >"$install_script"; then
        echo "Error: failed to download the Docker installer." >&2
        exit 1
    fi

    sudo sh "$install_script"
fi

if ! groups=$(id -nG "$USER"); then
    echo "Error: unable to determine Docker group membership for $USER." >&2
    exit 1
fi

case " $groups " in
    *" docker "*)
        echo "==> $USER is already in the docker group; skipping group update."
        ;;
    *)
        sudo usermod -aG docker "$USER"
        ;;
esac

echo "==> Docker installation complete"
