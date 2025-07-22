package v6

import (
	"context"

	store "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/CoreumFoundation/coreum/v6/app/upgrade"
	wbankkeeper "github.com/CoreumFoundation/coreum/v6/x/wbank/keeper"
)

// Name defines the upgrade name.
const Name = "v6"

// New makes an upgrade handler for v6 upgrade.
func New(
	mm *module.Manager, configurator module.Configurator, bankKeeper wbankkeeper.BaseKeeperWrapper,
) upgrade.Upgrade {
	return upgrade.Upgrade{
		Name: Name,
		StoreUpgrades: store.StoreUpgrades{
			Added: []string{},
		},
		Upgrade: func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
			vmap, err := mm.RunMigrations(ctx, configurator, vm)
			if err != nil {
				return nil, err
			}

			if err := migrateDenomSymbol(ctx, bankKeeper); err != nil {
				return nil, err
			}

			return vmap, nil
		},
	}
}
