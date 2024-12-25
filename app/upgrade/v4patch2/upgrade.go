package v4patch2

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/CoreumFoundation/coreum/v4/app/upgrade"
)

// Name defines the upgrade name.
const (
	Name = "v4patch2"
)

// New makes an upgrade handler for v4patch1 upgrade.
func New(name string, mm *module.Manager, configurator module.Configurator,
) upgrade.Upgrade {
	return upgrade.Upgrade{
		Name: name,
		StoreUpgrades: store.StoreUpgrades{
			Added: []string{},
		},
		Upgrade: func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
			return mm.RunMigrations(ctx, configurator, vm)
		},
	}
}
