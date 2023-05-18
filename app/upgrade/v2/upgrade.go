package v2

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibchost "github.com/cosmos/ibc-go/v4/modules/core/24-host"

	"github.com/CoreumFoundation/coreum/app/upgrade"
)

// Name defines the upgrade name.
const Name = "v2"

// NewV2Upgrade makes an upgrade handler for v2 upgrade.
func NewV2Upgrade(mm *module.Manager, configurator module.Configurator) upgrade.Upgrade {
	return upgrade.Upgrade{
		Name: Name,
		StoreUpgrades: storetypes.StoreUpgrades{
			Added: []string{
				ibchost.StoreKey,
				ibctransfertypes.StoreKey,
			},
		},
		Upgrade: func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
			return mm.RunMigrations(ctx, configurator, vm)
		},
	}
}
