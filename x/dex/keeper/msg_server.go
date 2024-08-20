package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

var _ types.MsgServer = MsgServer{}

// MsgKeeper defines subscope of keeper methods required by msg service.
type MsgKeeper interface {
	PlaceOrder(ctx sdk.Context, order types.Order) error
	CancelOrder(ctx sdk.Context, acc sdk.AccAddress, orderID string) error
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

// PlaceOrder places an order on orderbook.
func (ms MsgServer) PlaceOrder(ctx context.Context, msg *types.MsgPlaceOrder) (*types.EmptyResponse, error) {
	order, err := types.NewOrderFormMsgPlaceOrder(*msg)
	if err != nil {
		return nil, err
	}
	if err := ms.keeper.PlaceOrder(sdk.UnwrapSDKContext(ctx), order); err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// CancelOrder cancels order and unlock locked balance.
func (ms MsgServer) CancelOrder(ctx context.Context, msg *types.MsgCancelOrder) (*types.EmptyResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender")
	}

	return &types.EmptyResponse{}, ms.keeper.CancelOrder(sdk.UnwrapSDKContext(ctx), sender, msg.ID)
}
