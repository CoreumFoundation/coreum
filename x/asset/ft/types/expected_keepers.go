package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum/x/nft"
)

// BankKeeper defines the expected bank interface.
type BankKeeper interface {
	GetDenomMetaData(ctx sdk.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata)
	MintCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// NFTKeeper defines the expected NFT interface.
type NFTKeeper interface {
	SaveClass(ctx sdk.Context, class nft.Class) error
	HasClass(ctx sdk.Context, classID string) bool
	GetClasses(ctx sdk.Context) (classes []*nft.Class)
	HasNFT(ctx sdk.Context, classID, id string) bool
	Mint(ctx sdk.Context, token nft.NFT, receiver sdk.AccAddress) error
}
