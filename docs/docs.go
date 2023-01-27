package docs

import "embed"

// Docs embeds openapi doc.
//
//go:embed static
var Docs embed.FS
