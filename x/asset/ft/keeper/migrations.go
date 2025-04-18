package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v4 "github.com/CoreumFoundation/coreum/v6/x/asset/ft/migrations/v4"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	ftKeeper     Keeper
	paramsKeeper v4.ParamsKeeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(ftKeeper Keeper, paramsKeeper v4.ParamsKeeper) Migrator {
	return Migrator{
		ftKeeper:     ftKeeper,
		paramsKeeper: paramsKeeper,
	}
}

// Migrate5to6 migrates from version 5 to 6.
func (m Migrator) Migrate5to6(ctx sdk.Context) error {
	return nil
}
