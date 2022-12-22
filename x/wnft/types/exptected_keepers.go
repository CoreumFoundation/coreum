package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// AssetNFTProvider defines the interface to intercept within nft method calls
type AssetNFTProvider interface {
	BeforeTransfer(ctx sdk.Context, classID string, nftID string, receiver sdk.AccAddress) error
}
