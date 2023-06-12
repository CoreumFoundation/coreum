package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// Keeper specifies methods of keeper required by the migration.
type Keeper interface {
	GetParamsV1(ctx sdk.Context) types.ParamsV1
	SetParams(ctx sdk.Context, params types.Params)
}

// MigrateParams migrates asset ft params state from v1 to v2.
func MigrateParams(ctx sdk.Context, keeper Keeper) {
	paramsV1 := keeper.GetParamsV1(ctx)
	keeper.SetParams(ctx, types.Params{
		IssueFee:                    paramsV1.IssueFee,
		TokenUpgradeDecisionTimeout: ctx.BlockTime().Add(types.DefaultTokenUpgradeDecisionPeriod),
		TokenUpgradeGracePeriod:     types.DefaultTokenUpgradeGracePeriod,
	})
}
