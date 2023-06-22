package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v1 "github.com/CoreumFoundation/coreum/x/asset/ft/legacy/v1"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	paramsKeeper v1.ParamsKeeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(paramsKeeper v1.ParamsKeeper) Migrator {
	return Migrator{paramsKeeper: paramsKeeper}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v1.MigrateParams(ctx, m.paramsKeeper)
}
