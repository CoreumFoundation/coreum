package docs

import "embed"

// Docs embeds openapi doc
//nolint:gofmt,goimports // Looks like gofmt linter has a bug and it produces error because of go:embed
//go:embed static
var Docs embed.FS
