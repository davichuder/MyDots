package main

import "embed"

//go:embed assets
var assets embed.FS

var _ = assets
