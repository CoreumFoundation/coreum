package app

import (
	"github.com/cosmos/cosmos-sdk/client"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// NewDefaultClientContext returns a new cosmos client context
func NewDefaultClientContext() client.Context {
	encodingConfig := NewEncodingConfig()
	return client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithLegacyAmino(encodingConfig.Amino)
}
