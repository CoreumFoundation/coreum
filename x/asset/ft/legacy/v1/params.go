package v1

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// InitialTokenUpgradeDecisionPeriod is the period applied on top of the current block time to produce initial value of upgrade decision timeout.
const InitialTokenUpgradeDecisionPeriod = time.Hour * 24 * 30

// ParamsKeeper specifies methods of params keeper required by the migration.
type ParamsKeeper interface {
	GetSubspace(s string) (paramstypes.Subspace, bool)
}

// MigrateParams migrates asset ft params state from v1 to v2.
func MigrateParams(ctx sdk.Context, paramsKeeper ParamsKeeper) error {
	ftSubspace, ok := paramsKeeper.GetSubspace(types.ModuleName)
	if !ok {
		return errors.New("params subspace does not exist")
	}

	ftSubspace.Set(ctx, types.KeyTokenUpgradeDecisionTimeout, ctx.BlockTime().Add(InitialTokenUpgradeDecisionPeriod))
	ftSubspace.Set(ctx, types.KeyTokenUpgradeGracePeriod, types.DefaultTokenUpgradeGracePeriod)

	return nil
}
