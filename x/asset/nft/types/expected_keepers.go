package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/x/nft"
)

// NFTKeeper defines the expected NFT interface.
type NFTKeeper interface {
	SaveClass(ctx sdk.Context, class nft.Class) error
	GetClass(ctx sdk.Context, classID string) (nft.Class, bool)
	HasClass(ctx sdk.Context, classID string) bool
	GetClasses(ctx sdk.Context) (classes []*nft.Class)
	HasNFT(ctx sdk.Context, classID, id string) bool
	Mint(ctx sdk.Context, token nft.NFT, receiver sdk.AccAddress) error
	Burn(ctx sdk.Context, classID, nftID string) error
	GetOwner(ctx sdk.Context, classID, nftID string) sdk.AccAddress
}

// BankKeeper defines the expected bank interface.
type BankKeeper interface {
	BurnCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
