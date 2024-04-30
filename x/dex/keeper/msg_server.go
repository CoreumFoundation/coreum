package keeper

import (
	"context"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

var _ types.MsgServer = MsgServer{}

// MsgKeeper defines subscope of keeper methods required by msg service.
type MsgKeeper interface{}

// MsgServer serves grpc tx requests for dex module.
type MsgServer struct {
	keeper MsgKeeper
}

// NewMsgServer returns a new instance of the MsgServer.
func NewMsgServer(keeper MsgKeeper) MsgServer {
	return MsgServer{
		keeper: keeper,
	}
}

// PlaceOrder places an order on orderbook.
func (ms MsgServer) PlaceOrder(ctx context.Context, msg *types.MsgPlaceOrder) (*types.EmptyResponse, error) {
	// TODO(dex) implement
	return &types.EmptyResponse{}, nil
}
