package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

// NewClientContext returns a new cosmos client context
func NewClientContext(opts ...Option) client.Context {
	encodingConfig := NewEncodingConfig()
	ctx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino)

	for _, o := range opts {
		ctx = o(ctx)
	}

	return ctx
}

// Option type allows to modify client context
type Option func(client.Context) client.Context

// WithRPCClient option sets rpc client
func WithRPCClient(cli rpcclient.Client) Option {
	return func(ctx client.Context) client.Context {
		return ctx.WithClient(cli)
	}
}

// WithChainID option sets chainID
func WithChainID(chainID string) Option {
	return func(ctx client.Context) client.Context {
		return ctx.WithChainID(chainID)
	}
}
