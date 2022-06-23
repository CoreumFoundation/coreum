package blockexplorer

import (
	_ "embed"
)

// HasuraMetadataTemplate contains hasura metadata template
//go:embed hasura/metadata/metadata.tmpl.json
var HasuraMetadataTemplate string
