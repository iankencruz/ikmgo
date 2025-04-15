package ikmgo

import "embed"

//go:embed templates/* templates/admin/* templates/emails/* templates/partials/* static/**/*
var EmbeddedFiles embed.FS
