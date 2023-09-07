package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v1 "github.com/CoreumFoundation/coreum/v2/x/asset/nft/legacy/v1"
	v2 "github.com/CoreumFoundation/coreum/v2/x/asset/nft/legacy/v2"
	"github.com/CoreumFoundation/coreum/v2/x/asset/nft/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper       Keeper
	nftKeeper    types.NFTKeeper
	wasmKeeper   types.WasmKeeper
	paramsKeeper types.ParamsKeeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper, nftKeeper types.NFTKeeper, wasmKeeper types.WasmKeeper, paramsKeeper types.ParamsKeeper) Migrator {
	return Migrator{
		keeper:       keeper,
		nftKeeper:    nftKeeper,
		wasmKeeper:   wasmKeeper,
		paramsKeeper: paramsKeeper,
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

// Migrate2to3 migrates from version 2 to 3.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v2.MigrateParams(ctx, m.keeper, m.paramsKeeper)
}
