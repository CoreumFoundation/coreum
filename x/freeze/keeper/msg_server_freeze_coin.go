package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/x/freeze/types"
)

func (k msgServer) FreezeCoin(goCtx context.Context, msg *types.MsgFreezeCoin) (*types.MsgFreezeCoinResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Parse coin holder address
	holderAddr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, err
	}

	// TODO: Make sure the transaction sender can freeze the given token

	// Freeze coin
	k.Keeper.FreezeCoin(ctx, holderAddr, msg.Coin)

	return &types.MsgFreezeCoinResponse{}, nil
}
