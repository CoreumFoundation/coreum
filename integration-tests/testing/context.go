package testing

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	googlegrpc "google.golang.org/grpc"

	"github.com/CoreumFoundation/coreum/pkg/grpc"
)

// NewContext returns new context
func NewContext(clientCtx client.Context, grpcClient grpc.Client) Context {
	return Context{
		clientCtx:  clientCtx,
		grpcClient: grpcClient,
	}
}

// Context exposes the functionality of SDK context in a way where we may intercept GRPC-related method (Invoke)
// to provide better implementation
type Context struct {
	clientCtx  client.Context
	grpcClient grpc.Client
}

// TxConfig returns TxConfig of SDK context
func (c Context) TxConfig() client.TxConfig {
	return c.clientCtx.TxConfig
}

// WithFromName returns a copy of the context with an updated from account name
func (c Context) WithFromName(name string) Context {
	c.clientCtx = c.clientCtx.WithFromName(name)
	return c
}

// WithFromAddress returns a copy of the context with an updated from account address
func (c Context) WithFromAddress(addr sdk.AccAddress) Context {
	c.clientCtx = c.clientCtx.WithFromAddress(addr)
	return c
}

// Invoke invokes GRPC method
func (c Context) Invoke(ctx context.Context, method string, req, reply interface{}, opts ...googlegrpc.CallOption) (err error) {
	return c.grpcClient.Invoke(ctx, method, req, reply)
}

// NewStream implements the grpc ClientConn.NewStream method
func (c Context) NewStream(ctx context.Context, desc *googlegrpc.StreamDesc, method string, opts ...googlegrpc.CallOption) (googlegrpc.ClientStream, error) {
	return c.grpcClient.NewStream(ctx, desc, method, opts...)
}

// GetFeeGranterAddress returns the fee granter address from the context
func (c Context) GetFeeGranterAddress() sdk.AccAddress {
	return c.clientCtx.GetFeeGranterAddress()
}

// GetFromName returns the key name for the current context.
func (c Context) GetFromName() string {
	return c.clientCtx.GetFromName()
}

// GetFromAddress returns the from address from the context's name.
func (c Context) GetFromAddress() sdk.AccAddress {
	return c.clientCtx.GetFromAddress()
}

// BroadcastMode returns configured tx broadcast mode
func (c Context) BroadcastMode() string {
	return c.clientCtx.BroadcastMode
}

// Client returns RPC client
func (c Context) Client() rpcclient.Client {
	return c.clientCtx.Client
}

// InterfaceRegistry returns interface registry of SDK context
func (c Context) InterfaceRegistry() codectypes.InterfaceRegistry {
	return c.clientCtx.InterfaceRegistry
}
