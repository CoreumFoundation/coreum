package blockexplorer

import (
	_ "embed"
)

// BDJunoConfigTemplate contains bdjuno configuration template
//go:embed bdjuno/config/config.tpl.yaml
var BDJunoConfigTemplate string
