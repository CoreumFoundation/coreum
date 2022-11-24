package asset

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/x/asset/keeper"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// NewHandler return tx handler of the asset module.
func NewHandler(ms keeper.MsgServer) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		goCtx := sdk.WrapSDKContext(ctx)

		switch msg := msg.(type) {
		case *types.MsgIssueFungibleToken:
			res, err := ms.IssueFungibleToken(goCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgFreezeFungibleToken:
			res, err := ms.FreezeFungibleToken(goCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgUnfreezeFungibleToken:
			res, err := ms.UnfreezeFungibleToken(goCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgGlobalFreezeFungibleToken:
			res, err := ms.GlobalFreezeFungibleToken(goCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgGlobalUnfreezeFungibleToken:
			res, err := ms.GlobalUnfreezeFungibleToken(goCtx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}
