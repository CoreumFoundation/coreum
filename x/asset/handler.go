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
		var (
			goCtx = sdk.WrapSDKContext(ctx.WithEventManager(sdk.NewEventManager()))
			res   *types.EmptyResponse
			err   error
		)
		switch msg := msg.(type) {
		case *types.MsgIssueFungibleToken:
			res, err = ms.IssueFungibleToken(goCtx, msg)
		case *types.MsgFreezeFungibleToken:
			res, err = ms.FreezeFungibleToken(goCtx, msg)
		case *types.MsgUnfreezeFungibleToken:
			res, err = ms.UnfreezeFungibleToken(goCtx, msg)
		case *types.MsgGloballyFreezeFungibleToken:
			res, err = ms.GloballyFreezeFungibleToken(goCtx, msg)
		case *types.MsgGloballyUnfreezeFungibleToken:
			res, err = ms.GloballyUnfreezeFungibleToken(goCtx, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}

		return sdk.WrapServiceResult(ctx, res, err)
	}
}
