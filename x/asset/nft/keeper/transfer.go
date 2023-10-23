package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Transfer wraps the original transfer function of the nft keeper to include our custom interceptor.
func (k Keeper) Transfer(ctx sdk.Context, classID, nftID string, receiver sdk.AccAddress) error {
	if err := k.beforeTransfer(ctx, classID, nftID, receiver); err != nil {
		return err
	}

	return k.nftKeeper.Transfer(ctx, classID, nftID, receiver)
}

func (k Keeper) beforeTransfer(ctx sdk.Context, classID, nftID string, receiver sdk.AccAddress) error {
	if err := k.isNFTSendable(ctx, classID, nftID); err != nil {
		return err
	}

	return k.isNFTReceivable(ctx, classID, nftID, receiver)
}
