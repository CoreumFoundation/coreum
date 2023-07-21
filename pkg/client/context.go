package client

// `Context` structure in this file is a wrapper around client context delivered by Cosmos SDK.
// The original code does not respect cancelable `ctx`, leading to situations when dead HTTP connection
// may halt the application.
// The purpose of wrapping it is to modify its `queryABCI` private method to pass `ctx` correctly to the gRPC client.
// Public methods present here simply redirect the calls to their original implementations in the base client context.

import (
	"context"
	"io"
	"reflect"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/module"
	protobufgrpc "github.com/gogo/protobuf/grpc"
	gogoproto "github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/CoreumFoundation/coreum/v2/pkg/config"
)

var protoCodec = encoding.GetCodec(proto.Name)

// ContextConfig stores context config.
type ContextConfig struct {
	GasConfig     GasConfig
	TimeoutConfig TimeoutConfig
}

// TimeoutConfig is the part of context config holding timeout parameters.
type TimeoutConfig struct {
	RequestTimeout           time.Duration
	TxTimeout                time.Duration
	TxStatusPollInterval     time.Duration
	TxNextBlocksTimeout      time.Duration
	TxNextBlocksPollInterval time.Duration
}

// GasConfig is the part of context config holding gas parameters.
type GasConfig struct {
	GasAdjustment      float64
	GasPriceAdjustment sdk.Dec
}

// DefaultContextConfig returns default context config.
func DefaultContextConfig() ContextConfig {
	return ContextConfig{
		GasConfig: GasConfig{
			GasAdjustment:      1.0,
			GasPriceAdjustment: sdk.MustNewDecFromStr("1.1"),
		},
		TimeoutConfig: TimeoutConfig{
			RequestTimeout:           10 * time.Second,
			TxTimeout:                time.Minute,
			TxStatusPollInterval:     500 * time.Millisecond,
			TxNextBlocksTimeout:      time.Minute,
			TxNextBlocksPollInterval: time.Second,
		},
	}
}

// NewContext returns new context.
func NewContext(contextConfig ContextConfig, modules module.BasicManager) Context {
	encodingConfig := config.NewEncodingConfig(modules)
	return Context{
		config: contextConfig,
		clientCtx: client.Context{}.
			WithCodec(encodingConfig.Codec).
			WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
			WithTxConfig(encodingConfig.TxConfig).
			WithLegacyAmino(encodingConfig.Amino),
	}
}

// Context exposes the functionality of SDK context in a way where we may intercept GRPC-related method (Invoke)
// to provide better implementation.
type Context struct {
	config     ContextConfig
	clientCtx  client.Context
	grpcClient protobufgrpc.ClientConn
}

// SDKContext returns original sdk client context required by some functions.
func (c Context) SDKContext() client.Context {
	return c.clientCtx
}

// ChainID returns chain ID.
func (c Context) ChainID() string {
	return c.clientCtx.ChainID
}

// WithChainID returns a copy of the context with an updated chain ID.
func (c Context) WithChainID(chainID string) Context {
	c.clientCtx = c.clientCtx.WithChainID(chainID)
	return c
}

// GasAdjustment returns gas adjustment.
func (c Context) GasAdjustment() float64 {
	return c.config.GasConfig.GasAdjustment
}

// GasPriceAdjustment returns gas price adjustment.
func (c Context) GasPriceAdjustment() sdk.Dec {
	return c.config.GasConfig.GasPriceAdjustment
}

// WithGasAdjustment returns context with new gas adjustment.
func (c Context) WithGasAdjustment(adj float64) Context {
	c.config.GasConfig.GasAdjustment = adj
	return c
}

// WithGasPriceAdjustment returns context with new gas price adjustment.
func (c Context) WithGasPriceAdjustment(adj sdk.Dec) Context {
	c.config.GasConfig.GasPriceAdjustment = adj
	return c
}

// WithRPCClient returns a copy of the context with an updated RPC client
// instance.
func (c Context) WithRPCClient(client rpcclient.Client) Context {
	c.clientCtx = c.clientCtx.WithClient(client)
	return c
}

// WithGRPCClient returns a copy of the context with an updated GRPCClient client.
func (c Context) WithGRPCClient(grpcClient protobufgrpc.ClientConn) Context {
	c.grpcClient = grpcClient
	return c
}

// WithBroadcastMode returns a copy of the context with an updated broadcast
// mode.
func (c Context) WithBroadcastMode(mode string) Context {
	c.clientCtx = c.clientCtx.WithBroadcastMode(mode)
	return c
}

// TxConfig returns TxConfig of SDK context.
func (c Context) TxConfig() client.TxConfig {
	return c.clientCtx.TxConfig
}

// WithFromName returns a copy of the context with an updated from account name.
func (c Context) WithFromName(name string) Context {
	c.clientCtx = c.clientCtx.WithFromName(name)
	return c
}

// WithFromAddress returns a copy of the context with an updated from account address.
func (c Context) WithFromAddress(addr sdk.AccAddress) Context {
	c.clientCtx = c.clientCtx.WithFromAddress(addr)
	return c
}

// WithFeeGranterAddress returns a copy of the context with an updated fee granter account
// address.
func (c Context) WithFeeGranterAddress(addr sdk.AccAddress) Context {
	c.clientCtx = c.clientCtx.WithFeeGranterAddress(addr)
	return c
}

// NewStream implements the grpc ClientConn.NewStream method.
func (c Context) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	if c.RPCClient() != nil {
		return nil, errors.New("streaming rpc not supported")
	}

	if c.GRPCClient() != nil {
		return c.grpcClient.NewStream(ctx, desc, method, opts...)
	}

	return nil, errors.New("neither RPC nor GRPC client is set")
}

// FeeGranterAddress returns the fee granter address from the context.
func (c Context) FeeGranterAddress() sdk.AccAddress {
	return c.clientCtx.GetFeeGranterAddress()
}

// FromName returns the key name for the current context.
func (c Context) FromName() string {
	return c.clientCtx.GetFromName()
}

// FromAddress returns the from address from the context's name.
func (c Context) FromAddress() sdk.AccAddress {
	return c.clientCtx.GetFromAddress()
}

// BroadcastMode returns configured tx broadcast mode.
func (c Context) BroadcastMode() string {
	return c.clientCtx.BroadcastMode
}

// RPCClient returns RPC client.
func (c Context) RPCClient() rpcclient.Client {
	return c.clientCtx.Client
}

// GRPCClient returns GRPCClient client.
func (c Context) GRPCClient() protobufgrpc.ClientConn {
	return c.grpcClient
}

// InterfaceRegistry returns interface registry of SDK context.
func (c Context) InterfaceRegistry() codectypes.InterfaceRegistry {
	return c.clientCtx.InterfaceRegistry
}

// Keyring returns keyring.
func (c Context) Keyring() keyring.Keyring {
	return c.clientCtx.Keyring
}

// WithKeyring returns a copy of the context with an updated keyring.
func (c Context) WithKeyring(k keyring.Keyring) Context {
	c.clientCtx = c.clientCtx.WithKeyring(k)
	return c
}

// WithKeyringOptions returns a copy of the context with an updated keyring.
func (c Context) WithKeyringOptions(opts ...keyring.Option) Context {
	c.clientCtx = c.clientCtx.WithKeyringOptions(opts...)
	return c
}

// WithInput returns a copy of the context with an updated input.
func (c Context) WithInput(r io.Reader) Context {
	c.clientCtx = c.clientCtx.WithInput(r)
	return c
}

// WithCodec returns a copy of the Context with an updated Codec.
func (c Context) WithCodec(cdc codec.Codec) Context {
	c.clientCtx = c.clientCtx.WithCodec(cdc)
	return c
}

// Codec returns the registered Codec.
func (c Context) Codec() codec.Codec {
	return c.clientCtx.Codec
}

// WithOutput returns a copy of the context with an updated output writer (e.g. stdout).
func (c Context) WithOutput(w io.Writer) Context {
	c.clientCtx = c.clientCtx.WithOutput(w)
	return c
}

// WithFrom returns a copy of the context with an updated from address or name.
func (c Context) WithFrom(from string) Context {
	c.clientCtx = c.clientCtx.WithFrom(from)
	return c
}

// WithOutputFormat returns a copy of the context with an updated OutputFormat field.
func (c Context) WithOutputFormat(format string) Context {
	c.clientCtx = c.clientCtx.WithOutputFormat(format)
	return c
}

// WithNodeURI returns a copy of the context with an updated node URI.
func (c Context) WithNodeURI(nodeURI string) Context {
	c.clientCtx = c.clientCtx.WithNodeURI(nodeURI)
	return c
}

// WithHeight returns a copy of the context with an updated height.
func (c Context) WithHeight(height int64) Context {
	c.clientCtx = c.clientCtx.WithHeight(height)
	return c
}

// WithUseLedger returns a copy of the context with an updated UseLedger flag.
func (c Context) WithUseLedger(useLedger bool) Context {
	c.clientCtx = c.clientCtx.WithUseLedger(useLedger)
	return c
}

// WithHomeDir returns a copy of the Context with HomeDir set.
func (c Context) WithHomeDir(dir string) Context {
	c.clientCtx = c.clientCtx.WithHomeDir(dir)
	return c
}

// WithKeyringDir returns a copy of the Context with KeyringDir set.
func (c Context) WithKeyringDir(dir string) Context {
	c.clientCtx = c.clientCtx.WithKeyringDir(dir)
	return c
}

// WithGenerateOnly returns a copy of the context with updated GenerateOnly value.
func (c Context) WithGenerateOnly(generateOnly bool) Context {
	c.clientCtx = c.clientCtx.WithGenerateOnly(generateOnly)
	return c
}

// WithSimulation returns a copy of the context with updated Simulate value.
func (c Context) WithSimulation(simulate bool) Context {
	c.clientCtx = c.clientCtx.WithSimulation(simulate)
	return c
}

// WithOffline returns a copy of the context with updated Offline value.
func (c Context) WithOffline(offline bool) Context {
	c.clientCtx = c.clientCtx.WithOffline(offline)
	return c
}

// WithSignModeStr returns a copy of the context with an updated SignMode
// value.
func (c Context) WithSignModeStr(signModeStr string) Context {
	c.clientCtx = c.clientCtx.WithSignModeStr(signModeStr)
	return c
}

// WithSkipConfirmation returns a copy of the context with an updated SkipConfirm
// value.
func (c Context) WithSkipConfirmation(skip bool) Context {
	c.clientCtx = c.clientCtx.WithSkipConfirmation(skip)
	return c
}

// WithTxConfig returns the context with an updated TxConfig.
func (c Context) WithTxConfig(generator client.TxConfig) Context {
	c.clientCtx = c.clientCtx.WithTxConfig(generator)
	return c
}

// WithAccountRetriever returns the context with an updated AccountRetriever.
func (c Context) WithAccountRetriever(retriever client.AccountRetriever) Context {
	c.clientCtx = c.clientCtx.WithAccountRetriever(retriever)
	return c
}

// WithInterfaceRegistry returns the context with an updated InterfaceRegistry.
func (c Context) WithInterfaceRegistry(interfaceRegistry codectypes.InterfaceRegistry) Context {
	c.clientCtx = c.clientCtx.WithInterfaceRegistry(interfaceRegistry)
	return c
}

// WithViper returns the context with Viper field. This Viper instance is used to read
// client-side config from the config file.
func (c Context) WithViper(prefix string) Context {
	c.clientCtx = c.clientCtx.WithViper(prefix)
	return c
}

// PrintString prints the raw string to ctx.Output if it's defined, otherwise to os.Stdout.
func (c Context) PrintString(str string) error {
	return c.clientCtx.PrintBytes([]byte(str))
}

// PrintBytes prints the raw bytes to ctx.Output if it's defined, otherwise to os.Stdout.
// NOTE: for printing a complex state object, you should use ctx.PrintOutput.
func (c Context) PrintBytes(o []byte) error {
	return c.clientCtx.PrintBytes(o)
}

// PrintProto outputs toPrint to the ctx.Output based on ctx.OutputFormat which is
// either text or json. If text, toPrint will be YAML encoded. Otherwise, toPrint
// will be JSON encoded using ctx.Codec. An error is returned upon failure.
func (c Context) PrintProto(toPrint gogoproto.Message) error {
	return c.clientCtx.PrintProto(toPrint)
}

// Invoke invokes GRPC method.
func (c Context) Invoke(ctx context.Context, method string, req, reply interface{}, opts ...grpc.CallOption) (err error) {
	if c.GRPCClient() != nil {
		return c.GRPCClient().Invoke(ctx, method, req, reply, opts...)
	}

	if c.RPCClient() != nil {
		return c.invokeRPC(ctx, method, req, reply, opts)
	}

	return errors.New("neither RPC nor GRPC client is set")
}

func (c Context) invokeRPC(ctx context.Context, method string, req, reply interface{}, opts []grpc.CallOption) error {
	if reflect.ValueOf(req).IsNil() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "request cannot be nil")
	}

	reqBz, err := protoCodec.Marshal(req)
	if err != nil {
		return err
	}

	// parse height header
	md, _ := metadata.FromOutgoingContext(ctx)
	height := c.clientCtx.Height
	if heights := md.Get(grpctypes.GRPCBlockHeightHeader); len(heights) > 0 {
		var err error
		height, err = strconv.ParseInt(heights[0], 10, 64)
		if err != nil {
			return err
		}
		if height < 0 {
			return sdkerrors.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"client.Context.Invoke: height (%d) from %q must be >= 0", height, grpctypes.GRPCBlockHeightHeader)
		}
	}

	abciReq := abci.RequestQuery{
		Path:   method,
		Data:   reqBz,
		Height: height,
	}

	res, err := c.queryABCI(ctx, abciReq)
	if err != nil {
		return err
	}

	err = protoCodec.Unmarshal(res.Value, reply)
	if err != nil {
		return err
	}

	// Create header metadata. For now the headers contain:
	// - block height
	// We then parse all the call options, if the call option is a
	// HeaderCallOption, then we manually set the value of that header to the
	// metadata.
	md = metadata.Pairs(grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(res.Height, 10))
	for _, callOpt := range opts {
		header, ok := callOpt.(grpc.HeaderCallOption)
		if !ok {
			continue
		}

		*header.HeaderAddr = md
	}

	if c.clientCtx.InterfaceRegistry != nil {
		return codectypes.UnpackInterfaces(reply, c.clientCtx.InterfaceRegistry)
	}

	return nil
}

func (c Context) queryABCI(ctx context.Context, req abci.RequestQuery) (abci.ResponseQuery, error) {
	node, err := c.clientCtx.GetNode()
	if err != nil {
		return abci.ResponseQuery{}, err
	}

	opts := rpcclient.ABCIQueryOptions{
		Height: req.Height,
		Prove:  req.Prove,
	}

	result, err := node.ABCIQueryWithOptions(ctx, req.Path, req.Data, opts)
	if err != nil {
		return abci.ResponseQuery{}, err
	}

	if !result.Response.IsOK() {
		return abci.ResponseQuery{}, sdkErrorToGRPCError(result.Response)
	}

	return result.Response, nil
}

func sdkErrorToGRPCError(resp abci.ResponseQuery) error {
	switch resp.Code {
	case sdkerrors.ErrInvalidRequest.ABCICode():
		return status.Error(codes.InvalidArgument, resp.Log)
	case sdkerrors.ErrUnauthorized.ABCICode():
		return status.Error(codes.Unauthenticated, resp.Log)
	case sdkerrors.ErrKeyNotFound.ABCICode():
		return status.Error(codes.NotFound, resp.Log)
	default:
		return status.Error(codes.Unknown, resp.Log)
	}
}
