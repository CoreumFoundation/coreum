package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeforeTransfer includes logic that will be run before the Tranfer method of the nft module.
func (k Keeper) BeforeTransfer(ctx sdk.Context, classID, nftID string, receiver sdk.AccAddress) error {
	return k.isNFTSendable(ctx, classID, nftID)
}
