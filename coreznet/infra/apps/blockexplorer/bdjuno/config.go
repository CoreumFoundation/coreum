package bdjuno

import (
	_ "embed"
)

// ConfigTemplate contains bdjuno configuration template
//go:embed config/config.yaml
var ConfigTemplate string
