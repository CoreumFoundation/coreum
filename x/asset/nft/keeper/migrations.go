package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v2 "github.com/CoreumFoundation/coreum/v6/x/asset/nft/migrations/v2"
	"github.com/CoreumFoundation/coreum/v6/x/asset/nft/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper       Keeper
	nftKeeper    types.NFTKeeper
	wasmKeeper   types.WasmKeeper
	paramsKeeper types.ParamsKeeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(
	keeper Keeper, nftKeeper types.NFTKeeper, wasmKeeper types.WasmKeeper, paramsKeeper types.ParamsKeeper,
) Migrator {
	return Migrator{
		keeper:       keeper,
		nftKeeper:    nftKeeper,
		wasmKeeper:   wasmKeeper,
		paramsKeeper: paramsKeeper,
	}
}

// Migrate2to3 migrates from version 2 to 3.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v2.MigrateParams(ctx, m.keeper, m.paramsKeeper)
}
