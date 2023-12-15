package v3patch2

// This patch is supposed to be used on testnet only for upgrading from v3.0.1 (v3patch1 plan)
// to v3.0.2 (v3patch2 plan).

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/CoreumFoundation/coreum/v4/app/upgrade"
)

// Name defines the upgrade name.
const Name = "v3patch2"

// New makes an upgrade handler for v3patch2 upgrade.
func New(mm *module.Manager, configurator module.Configurator) upgrade.Upgrade {
	return upgrade.Upgrade{
		Name: Name,
		Upgrade: func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
			return mm.RunMigrations(ctx, configurator, vm)
		},
	}
}
