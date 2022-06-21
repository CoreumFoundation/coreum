package hasura

import (
	_ "embed"
)

// MetadataTemplate contains hasura metadata template
//go:embed metadata/metadata.json
var MetadataTemplate string
