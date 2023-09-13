package v1

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/CoreumFoundation/coreum/v3/x/customparams/types"
)

// ParamsKeeper specifies methods of params keeper required by the migration.
type ParamsKeeper interface {
	GetSubspace(s string) (paramstypes.Subspace, bool)
}

// Keeper specifies methods of the keeper required by the migration.
type Keeper interface {
	SetStakingParams(sdk.Context, types.StakingParams) error
}

// MigrateParams migrates params of the ft module from the params module to inside the ft module.
func MigrateParams(ctx sdk.Context, keeper Keeper, paramsKeeper ParamsKeeper) error {
	sp, ok := paramsKeeper.GetSubspace(types.CustomParamsStaking)
	// set KeyTable if it has not already been set
	if !sp.HasKeyTable() {
		sp = sp.WithKeyTable(types.StakingParamKeyTable())
	}

	if !ok {
		return sdkerrors.Wrap(types.ErrInvalidState, "params subspace does not exist")
	}

	var params types.StakingParams
	sp.GetParamSet(ctx, &params)
	return keeper.SetStakingParams(ctx, params)
}
