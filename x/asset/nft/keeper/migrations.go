package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v1 "github.com/CoreumFoundation/coreum/x/asset/nft/legacy/v1"
	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper     Keeper
	nftKeeper  types.NFTKeeper
	wasmKeeper types.WasmKeeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper, nftKeeper types.NFTKeeper, wasmKeeper types.WasmKeeper) Migrator {
	return Migrator{
		keeper:     keeper,
		nftKeeper:  nftKeeper,
		wasmKeeper: wasmKeeper,
	}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	if err := v1.MigrateStore(ctx, m.keeper.storeKey); err != nil {
		return err
	}

	if err := v1.MigrateClassFeatures(ctx, m.keeper); err != nil {
		return err
	}

	return v1.MigrateWasmCreatedNFTData(ctx, m.nftKeeper, m.keeper, m.wasmKeeper)
}
