package hasura

import (
	_ "embed"
)

// Metadata contains hasura metadata
//go:embed metadata/metadata.json
var Metadata string
