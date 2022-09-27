package grpc

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	abci "github.com/tendermint/tendermint/abci/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var protoCodec = encoding.GetCodec(proto.Name)

// NewClient creates new GRPC client
func NewClient(clientCtx cosmosclient.Context) Client {
	return Client{
		clientCtx: clientCtx,
	}
}

// Client contains code copied from https://github.com/cosmos/cosmos-sdk/blob/35ae2c4c72d4aeb33447d5a7af23ca47f786606e/client/grpc_query.go
// We improved it to respect received context instead of passing context.Background().
// Original code supports both querying and broadcasting transactions. We removed broadcasting part because doing it
// correctly (with context) would require copying big chunks of code while that functionality is not needed because another functions
// are used for broadcasting transactions.
type Client struct {
	clientCtx cosmosclient.Context
}

// Invoke invokes GRPC method
func (client Client) Invoke(ctx context.Context, method string, req, reply interface{}, opts ...grpc.CallOption) (err error) {
	if reflect.ValueOf(req).IsNil() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "request cannot be nil")
	}

	reqBz, err := protoCodec.Marshal(req)
	if err != nil {
		return err
	}

	// parse height header
	md, _ := metadata.FromOutgoingContext(ctx)
	height := client.clientCtx.Height
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

	res, err := client.queryABCI(ctx, abciReq)
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

	if client.clientCtx.InterfaceRegistry != nil {
		return types.UnpackInterfaces(reply, client.clientCtx.InterfaceRegistry)
	}

	return nil
}

// NewStream implements the grpc ClientConn.NewStream method
func (client Client) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("streaming rpc not supported")
}

func (client Client) queryABCI(ctx context.Context, req abci.RequestQuery) (abci.ResponseQuery, error) {
	node, err := client.clientCtx.GetNode()
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
