package v1

import (
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
)

// InitialTokenUpgradeDecisionPeriod is the period applied on top of the current block time to produce initial value of upgrade decision timeout.
const InitialTokenUpgradeDecisionPeriod = time.Hour * 24 * 21

// ParamsKeeper specifies methods of params keeper required by the migration.
type ParamsKeeper interface {
	GetSubspace(s string) (paramstypes.Subspace, bool)
}

// MigrateParams migrates asset ft params state from v1 to v2.
func MigrateParams(ctx sdk.Context, paramsKeeper ParamsKeeper) error {
	ftSubspace, ok := paramsKeeper.GetSubspace(types.ModuleName)
	if !ok {
		return sdkerrors.Wrap(types.ErrInvalidState, "params subspace does not exist")
	}
	// set KeyTable if it has not already been set
	if !ftSubspace.HasKeyTable() {
		ftSubspace.WithKeyTable(paramstypes.NewKeyTable().RegisterParamSet(&types.Params{}))
	}

	ftSubspace.Set(ctx, types.KeyTokenUpgradeDecisionTimeout, ctx.BlockTime().Add(InitialTokenUpgradeDecisionPeriod))
	ftSubspace.Set(ctx, types.KeyTokenUpgradeGracePeriod, types.DefaultTokenUpgradeGracePeriod)

	return nil
}
