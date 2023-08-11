package v3

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/CoreumFoundation/coreum/v2/app/upgrade"
)

// juno reference: https://github.com/CosmosContracts/juno/pull/646/files#diff-8ae5168a16be54c5a00ba9dcf5e54cabc4d053c2f3d77ac700aeef3f3dffd87b

const Name = "v3"

func New(mm *module.Manager, configurator module.Configurator) upgrade.Upgrade {
	return upgrade.Upgrade{
		Name: Name,
		StoreUpgrades: storetypes.StoreUpgrades{
			Added: []string{},
		},
		Upgrade: func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
			return mm.RunMigrations(ctx, configurator, vm)
		},
	}
}
