package keeper

import (
	"context"

    "github.com/CoreumFoundation/coreum/x/freeze/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)


func (k msgServer) FreezeCoin(goCtx context.Context,  msg *types.MsgFreezeCoin) (*types.MsgFreezeCoinResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

    // TODO: Handling the message
    _ = ctx

	return &types.MsgFreezeCoinResponse{}, nil
}
