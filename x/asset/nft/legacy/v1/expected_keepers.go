package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v2/x/asset/nft/types"
	"github.com/CoreumFoundation/coreum/v2/x/nft"
)

// AssetNFTKeeper represents the assetnft keeper.
type AssetNFTKeeper interface {
	IterateAllClassDefinitions(ctx sdk.Context, cb func(types.ClassDefinition) (bool, error)) error
	SetClassDefinition(ctx sdk.Context, definition types.ClassDefinition) error
}

// NFTKeeper represents the expected methods from the nft keeper.
type NFTKeeper interface {
	GetClass(ctx sdk.Context, classID string) (nft.Class, bool)
	UpdateClass(ctx sdk.Context, class nft.Class) error
	GetNFTsOfClass(ctx sdk.Context, classID string) []nft.NFT
	Update(ctx sdk.Context, n nft.NFT) error
}

// WasmKeeper represents the expected method from the wasm keeper.
type WasmKeeper interface {
	HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool
}
