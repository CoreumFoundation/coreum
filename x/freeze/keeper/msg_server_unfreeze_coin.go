package keeper

import (
	"context"

    "github.com/CoreumFoundation/coreum/x/freeze/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)


func (k msgServer) UnfreezeCoin(goCtx context.Context,  msg *types.MsgUnfreezeCoin) (*types.MsgUnfreezeCoinResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

    // TODO: Handling the message
    _ = ctx

	return &types.MsgUnfreezeCoinResponse{}, nil
}
