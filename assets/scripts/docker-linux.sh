#!/bin/sh
set -e

# docker-linux.sh — Install Docker Engine on Linux (Native or WSL2)
# Called by the MyDots installer (M-38).
# Adds Docker's official apt repository and installs docker-ce family.

# Pre-req: curl, sudo, apt-get available
curl -fsSL https://get.docker.com -o /tmp/get-docker.sh
sudo sh /tmp/get-docker.sh
sudo usermod -aG docker "$USER"
