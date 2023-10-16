package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// NFTKeeper defines the expected NFT interface.
//
//nolint:interfacebloat
type NFTKeeper interface {
	SaveClass(ctx sdk.Context, class nft.Class) error
	GetClass(ctx sdk.Context, classID string) (nft.Class, bool)
	UpdateClass(ctx sdk.Context, class nft.Class) error
	GetNFTsOfClass(ctx sdk.Context, classID string) []nft.NFT
	HasClass(ctx sdk.Context, classID string) bool
	HasNFT(ctx sdk.Context, classID, id string) bool
	Mint(ctx sdk.Context, token nft.NFT, receiver sdk.AccAddress) error
	Burn(ctx sdk.Context, classID, nftID string) error
	Update(ctx sdk.Context, n nft.NFT) error
	GetOwner(ctx sdk.Context, classID, nftID string) sdk.AccAddress
	Transfer(ctx sdk.Context, classID string, nftID string, receiver sdk.AccAddress) error
}

// BankKeeper defines the expected bank interface.
type BankKeeper interface {
	BurnCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

// WasmKeeper represents the expected method from the wasm keeper.
type WasmKeeper interface {
	HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool
}

// ParamsKeeper specifies expected methods of params keeper.
type ParamsKeeper interface {
	GetSubspace(s string) (paramstypes.Subspace, bool)
}

// AccountKeeper defines the account contract that must be fulfilled when
// creating an nft keeper.
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}
