package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v1 "github.com/CoreumFoundation/coreum/v2/x/asset/ft/legacy/v1"
	v2 "github.com/CoreumFoundation/coreum/v2/x/asset/ft/legacy/v2"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	ftKeeper     Keeper
	paramsKeeper v1.ParamsKeeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(ftKeeper Keeper, paramsKeeper v1.ParamsKeeper) Migrator {
	return Migrator{
		ftKeeper:     ftKeeper,
		paramsKeeper: paramsKeeper,
	}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	if err := v1.MigrateParams(ctx, m.paramsKeeper); err != nil {
		return err
	}
	return v1.MigrateFeatures(ctx, m.ftKeeper)
}

// Migrate2to3 migrates from version 2 to 3.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	return v2.MigrateParams(ctx, m.ftKeeper, m.paramsKeeper)
}
