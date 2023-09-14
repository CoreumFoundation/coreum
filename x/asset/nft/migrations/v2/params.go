package v2

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/CoreumFoundation/coreum/v3/x/asset/nft/types"
)

// ParamsKeeper specifies methods of params keeper required by the migration.
type ParamsKeeper interface {
	GetSubspace(s string) (paramstypes.Subspace, bool)
}

// NFTKeeper specifies methods of the nft keeper required by the migration.
type NFTKeeper interface {
	SetParams(sdk.Context, types.Params) error
}

// MigrateParams migrates params of the nft module from the params module to inside the nft module.
func MigrateParams(ctx sdk.Context, keeper NFTKeeper, paramsKeeper ParamsKeeper) error {
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
