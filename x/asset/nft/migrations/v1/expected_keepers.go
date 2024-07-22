package v1

import (
	"context"

	"cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v4/x/asset/nft/types"
)

// AssetNFTKeeper represents the assetnft keeper.
type AssetNFTKeeper interface {
	IterateAllClassDefinitions(ctx sdk.Context, cb func(types.ClassDefinition) (bool, error)) error
	SetClassDefinition(ctx sdk.Context, definition types.ClassDefinition) error
}

// NFTKeeper represents the expected methods from the nft keeper.
type NFTKeeper interface {
	GetClass(ctx context.Context, classID string) (nft.Class, bool)
	UpdateClass(ctx context.Context, class nft.Class) error
	GetNFTsOfClass(ctx context.Context, classID string) []nft.NFT
	Update(ctx context.Context, n nft.NFT) error
}

// WasmKeeper represents the expected method from the wasm keeper.
type WasmKeeper interface {
	HasContractInfo(ctx context.Context, contractAddress sdk.AccAddress) bool
}
