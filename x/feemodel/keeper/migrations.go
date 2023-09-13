package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v1 "github.com/CoreumFoundation/coreum/v3/x/feemodel/migrations/v1"
	"github.com/CoreumFoundation/coreum/v3/x/feemodel/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper       FeeModelMigrationKeeper
	paramsKeeper types.ParamsKeeper
}

// FeeModelMigrationKeeper specifies the methods of the keeper needed by migration.
type FeeModelMigrationKeeper interface {
	SetParams(sdk.Context, types.Params) error
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper FeeModelMigrationKeeper, paramsKeeper types.ParamsKeeper) Migrator {
	return Migrator{
		keeper:       keeper,
		paramsKeeper: paramsKeeper,
	}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v1.MigrateParams(ctx, m.keeper, m.paramsKeeper)
}
