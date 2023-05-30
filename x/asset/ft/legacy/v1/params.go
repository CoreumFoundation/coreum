package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// Keeper specifies methods of keeper required by the migration.
type Keeper interface {
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params)
}

// MigrateParams migrates asset ft params state from v1 to v2.
func MigrateParams(ctx sdk.Context, keeper Keeper) error {
	params := keeper.GetParams(ctx)
	params.TokenUpgradeDecisionTimeout = ctx.BlockTime().Add(types.DefaultTokenUpgradeDecisionPeriod)
	params.TokenUpgradeGracePeriod = types.DefaultTokenUpgradeGracePeriod
	keeper.SetParams(ctx, params)

	return nil
}
