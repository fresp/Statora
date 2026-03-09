package embed

import "embed"

//go:embed all:dist/*
var Assets embed.FS
