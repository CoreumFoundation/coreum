package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/x/freeze/types"
)

func (k msgServer) UnfreezeCoin(goCtx context.Context, msg *types.MsgUnfreezeCoin) (*types.MsgUnfreezeCoinResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Parse coin holder address
	holderAddr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}

	// TODO: Make sure the transaction sender can unfreeze the given token

	// Unfreeze coin
	if err = k.Keeper.UnfreezeCoin(ctx, holderAddr, msg.Coin); err != nil {
		return nil, err
	}

	return &types.MsgUnfreezeCoinResponse{}, nil
}
