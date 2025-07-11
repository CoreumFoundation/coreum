package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// NonFungibleTokenProvider defines the interface to intercept within nft method calls.
type NonFungibleTokenProvider interface {
	Transfer(ctx sdk.Context, classID, nftID string, receiver sdk.AccAddress) error
}
