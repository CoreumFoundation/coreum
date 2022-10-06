package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// Keeper is the asset module keeper.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey sdk.StoreKey
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// IssueAsset issues new asset.
func (k Keeper) IssueAsset(ctx sdk.Context, name string) string {
	// TODO(dhil) replace with real implementation
	return "id1"
}

// GetAsset return the asset by its id.
func (k Keeper) GetAsset(ctx sdk.Context, id string) types.Asset {
	// TODO(dhil) replace with real implementation
	return types.Asset{
		Id:   "id1",
		Name: "name1",
	}
}

// Logger returns the Keeper logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
