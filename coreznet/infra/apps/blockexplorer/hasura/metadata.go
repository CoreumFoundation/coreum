package hasura

import (
	_ "embed"
)

// MetadataTemplate contains hasura metadata template
//go:embed metadata/metadata.tpl.json
var MetadataTemplate string
