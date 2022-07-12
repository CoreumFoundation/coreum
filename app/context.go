package app

import (
	"github.com/cosmos/cosmos-sdk/client"
)

// NewDefaultClientContext returns a new cosmos client context
func NewDefaultClientContext() client.Context {
	encodingConfig := NewEncodingConfig()
	return client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino)
}
