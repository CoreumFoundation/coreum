package types

import (
	context "context"

	"cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// NFTKeeper defines the expected NFT interface.
//
//nolint:interfacebloat
type NFTKeeper interface {
	SaveClass(ctx context.Context, class nft.Class) error
	GetClass(ctx context.Context, classID string) (nft.Class, bool)
	UpdateClass(ctx context.Context, class nft.Class) error
	GetNFTsOfClass(ctx context.Context, classID string) []nft.NFT
	HasClass(ctx context.Context, classID string) bool
	GetNFT(ctx context.Context, classID, nftID string) (nft.NFT, bool)
	HasNFT(ctx context.Context, classID, id string) bool
	Mint(ctx context.Context, token nft.NFT, receiver sdk.AccAddress) error
	Burn(ctx context.Context, classID, nftID string) error
	Update(ctx context.Context, n nft.NFT) error
	GetOwner(ctx context.Context, classID, nftID string) sdk.AccAddress
	Transfer(ctx context.Context, classID, nftID string, receiver sdk.AccAddress) error
}

// BankKeeper defines the expected bank interface.
type BankKeeper interface {
	BurnCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

// WasmKeeper represents the expected method from the wasm keeper.
type WasmKeeper interface {
	HasContractInfo(ctx context.Context, contractAddress sdk.AccAddress) bool
}

// ParamsKeeper specifies expected methods of params keeper.
type ParamsKeeper interface {
	GetSubspace(s string) (paramstypes.Subspace, bool)
}
