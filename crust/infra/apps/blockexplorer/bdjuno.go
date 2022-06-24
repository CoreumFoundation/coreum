package blockexplorer

import (
	_ "embed"
)

// BDJunoConfigTemplate contains bdjuno configuration template
//go:embed bdjuno/config/config.tmpl.yaml
var BDJunoConfigTemplate string
