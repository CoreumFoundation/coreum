package docs

import "embed"

// Docs embeds openapi doc.a
//
//go:embed static
var Docs embed.FS
