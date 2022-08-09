package keeper

import (
	"context"
	"fmt"

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

	// Check if the token is already frozen
	if k.Keeper.IsFrozenCoin(ctx, holderAddr, msg.Denom) {
		return nil, fmt.Errorf("coin %s is already frozen", msg.Denom)
	}

	// Freeze coin
	if err = k.Keeper.FreezeCoin(ctx, holderAddr, msg.Denom); err != nil {
		return nil, err
	}

	return &types.MsgFreezeCoinResponse{}, nil
}
