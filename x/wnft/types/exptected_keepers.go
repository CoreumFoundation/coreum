package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// NonFungibleTokenProvider defines the interface to intercept within nft method calls
type NonFungibleTokenProvider interface {
	BeforeTransfer(ctx sdk.Context, classID string, nftID string, receiver sdk.AccAddress) error
}
