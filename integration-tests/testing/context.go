package testing

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	googlegrpc "google.golang.org/grpc"

	"github.com/CoreumFoundation/coreum/pkg/grpc"
)

func NewContext(clientCtx client.Context, grpcClient grpc.Client) Context {
	return Context{
		clientCtx:  clientCtx,
		grpcClient: grpcClient,
	}
}

type Context struct {
	clientCtx  client.Context
	grpcClient grpc.Client
}

func (c Context) ClientContext() client.Context {
	return c.clientCtx
}

func (c Context) TxConfig() client.TxConfig {
	return c.clientCtx.TxConfig
}

func (c Context) WithFromName(name string) Context {
	c.clientCtx = c.clientCtx.WithFromName(name)
	return c
}

func (c Context) WithFromAddress(addr sdk.AccAddress) Context {
	c.clientCtx = c.clientCtx.WithFromAddress(addr)
	return c
}

func (c Context) Invoke(ctx context.Context, method string, req, reply interface{}, opts ...googlegrpc.CallOption) (err error) {
	return c.grpcClient.Invoke(ctx, method, req, reply)
}

func (c Context) NewStream(ctx context.Context, desc *googlegrpc.StreamDesc, method string, opts ...googlegrpc.CallOption) (googlegrpc.ClientStream, error) {
	return c.grpcClient.NewStream(ctx, desc, method, opts...)
}
