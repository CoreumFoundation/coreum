package v4

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	ibchookstypes "github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/types"

	"github.com/CoreumFoundation/coreum/v4/app/upgrade"
	dextypes "github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

// Name defines the upgrade name.
const Name = "v4"

// New makes an upgrade handler for v4 upgrade.
func New(mm *module.Manager, configurator module.Configurator) upgrade.Upgrade {
	return upgrade.Upgrade{
		Name: Name,
		StoreUpgrades: store.StoreUpgrades{
			Added: []string{
				ibchookstypes.StoreKey,
				packetforwardtypes.StoreKey,
				dextypes.StoreKey,
			},
		},
		Upgrade: func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
			return mm.RunMigrations(ctx, configurator, vm)
		},
	}
}
