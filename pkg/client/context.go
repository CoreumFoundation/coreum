package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

func NewClientContext(opts ...option) client.Context {
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

type option func(client.Context) client.Context

func WithRPCClient(cli rpcclient.Client) option {
	return func(ctx client.Context) client.Context {
		return ctx.WithClient(cli)
	}
}

func WithChainID(chainID string) option {
	return func(ctx client.Context) client.Context {
		return ctx.WithChainID(chainID)
	}
}

func WithBroadcastMode(mode string) option {
	return func(ctx client.Context) client.Context {
		return ctx.WithBroadcastMode(mode)
	}
}
