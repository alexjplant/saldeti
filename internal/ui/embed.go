package ui

import "embed"

//go:embed templates/*.html templates/**/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS
