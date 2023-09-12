package v1

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/CoreumFoundation/coreum/v2/x/feemodel/types"
)

// ParamsKeeper specifies methods of params keeper required by the migration.
type ParamsKeeper interface {
	GetSubspace(s string) (paramstypes.Subspace, bool)
}

// Keeper specifies methods of the keeper required by the migration.
type Keeper interface {
	SetParams(sdk.Context, types.Params) error
}

// MigrateParams migrates params of the ft module from the params module to inside the ft module.
func MigrateParams(ctx sdk.Context, keeper Keeper, paramsKeeper ParamsKeeper) error {
	sp, ok := paramsKeeper.GetSubspace(types.ModuleName)
	if !ok {
		return sdkerrors.Wrap(types.ErrInvalidState, "params subspace does not exist")
	}

	if !sp.HasKeyTable() {
		sp.WithKeyTable(paramstypes.NewKeyTable().RegisterParamSet(&types.Params{}))
	}

	var params types.Params
	sp.GetParamSet(ctx, &params)
	return keeper.SetParams(ctx, params)
}
