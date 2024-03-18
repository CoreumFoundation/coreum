package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

var _ types.MsgServer = MsgServer{}

// MsgKeeper defines subscope of keeper methods required by msg service.
type MsgKeeper interface {
	StoreTransientOrder(ctx sdk.Context, order Order) error
}

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

// CreateLimitOrder handles MsgCreateLimitOrder message.
func (ms MsgServer) CreateLimitOrder(
	ctx context.Context,
	msg *types.MsgCreateLimitOrder,
) (*types.EmptyResponse, error) {
	order := types.OrderLimit(*msg)
	err := ms.keeper.StoreTransientOrder(sdk.UnwrapSDKContext(ctx), &order)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
