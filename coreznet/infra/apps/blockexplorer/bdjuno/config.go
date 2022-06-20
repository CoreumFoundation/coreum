package bdjuno

import (
	_ "embed"
)

// Config contains bdjuno configuration
//go:embed config/config.yaml
var Config string
