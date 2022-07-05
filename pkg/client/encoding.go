package client

import (
	"github.com/CoreumFoundation/coreum/app"
	"github.com/ignite-hq/cli/ignite/pkg/cosmoscmd"
)

// NewEncodingConfig returns the encoding config
func NewEncodingConfig() cosmoscmd.EncodingConfig {
	return cosmoscmd.MakeEncodingConfig(app.ModuleBasics)
}
