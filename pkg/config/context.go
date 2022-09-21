package config

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// NewClientContext returns a new cosmos client context
func NewClientContext(modules module.BasicManager) client.Context {
	encodingConfig := NewEncodingConfig(modules)
	return client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino)
}
