package v2

// For testnet we use v2.0.0 binary for this plan.
// For mainnet we use v2.0.2 binary for this plan.

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/CoreumFoundation/coreum/v3/app/upgrade"
	delaytypes "github.com/CoreumFoundation/coreum/v3/x/delay/types"
)

// Name defines the upgrade name.
const Name = "v2"

// New makes an upgrade handler for v2 upgrade.
func New(mm *module.Manager, configurator module.Configurator) upgrade.Upgrade {
	return upgrade.Upgrade{
		Name: Name,
		StoreUpgrades: storetypes.StoreUpgrades{
			Added: []string{
				ibcexported.StoreKey,
				ibctransfertypes.StoreKey,
				delaytypes.StoreKey,
			},
		},
		Upgrade: func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
			return mm.RunMigrations(ctx, configurator, vm)
		},
	}
}
