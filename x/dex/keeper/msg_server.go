package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

var _ types.MsgServer = MsgServer{}

// MsgKeeper defines subscope of keeper methods required by msg service.
type MsgKeeper interface {
	UpdateParams(ctx sdk.Context, authority string, params types.Params) error
	PlaceOrder(ctx sdk.Context, order types.Order) error
	CancelOrder(ctx sdk.Context, acc sdk.AccAddress, orderID string) error
	CancelOrdersByDenom(ctx sdk.Context, admin, acc sdk.AccAddress, denom string) error
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

// UpdateParams is a governance operation that sets parameters of the module.
func (ms MsgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.EmptyResponse, error) {
	if err := ms.keeper.UpdateParams(sdk.UnwrapSDKContext(goCtx), req.Authority, req.Params); err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// PlaceOrder places an order on orderbook.
func (ms MsgServer) PlaceOrder(ctx context.Context, msg *types.MsgPlaceOrder) (*types.EmptyResponse, error) {
	order, err := types.NewOrderFromMsgPlaceOrder(*msg)
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

// CancelOrdersByDenom cancels all orders by denom and account.
func (ms MsgServer) CancelOrdersByDenom(
	ctx context.Context, msg *types.MsgCancelOrdersByDenom,
) (*types.EmptyResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender")
	}

	acc, err := sdk.AccAddressFromBech32(msg.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account")
	}

	return &types.EmptyResponse{}, ms.keeper.CancelOrdersByDenom(sdk.UnwrapSDKContext(ctx), sender, acc, msg.Denom)
}
