package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v1 "github.com/CoreumFoundation/coreum/v3/x/customparams/migrations/v1"
	"github.com/CoreumFoundation/coreum/v3/x/customparams/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper       Keeper
	paramsKeeper types.ParamsKeeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper, paramsKeeper types.ParamsKeeper) Migrator {
	return Migrator{
		keeper:       keeper,
		paramsKeeper: paramsKeeper,
	}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v1.MigrateParams(ctx, m.keeper, m.paramsKeeper)
}
