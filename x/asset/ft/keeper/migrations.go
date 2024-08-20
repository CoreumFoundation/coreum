package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v3 "github.com/CoreumFoundation/coreum/v4/x/asset/ft/migrations/v3"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	ftKeeper     Keeper
	paramsKeeper v3.ParamsKeeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(ftKeeper Keeper, paramsKeeper v3.ParamsKeeper) Migrator {
	return Migrator{
		ftKeeper:     ftKeeper,
		paramsKeeper: paramsKeeper,
	}
}

// Migrate3to4 migrates from version 3 to 4.
func (m Migrator) Migrate3to4(ctx sdk.Context) error {
	return v3.MigrateDefinitions(ctx, m.ftKeeper)
}
