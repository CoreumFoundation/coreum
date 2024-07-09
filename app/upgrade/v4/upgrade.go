package v4

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	ibchookstypes "github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"

	"github.com/CoreumFoundation/coreum/v4/app/upgrade"
)

// Name defines the upgrade name.
const (
	Name      = "v4"
	NameAlias = "Coreum V4"
)

// New makes an upgrade handler for v4 upgrade.
func New(name string, mm *module.Manager, configurator module.Configurator,
	consensusParamKeeper consensusparamkeeper.Keeper,
) upgrade.Upgrade {
	return upgrade.Upgrade{
		Name: name,
		StoreUpgrades: store.StoreUpgrades{
			Added: []string{
				ibchookstypes.StoreKey,
				packetforwardtypes.StoreKey,
				icacontrollertypes.StoreKey,
				icahosttypes.StoreKey,
			},
		},
		Upgrade: func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
			vmap, err := mm.RunMigrations(ctx, configurator, vm)
			if err != nil {
				return nil, err
			}

			consensusParams, err := consensusParamKeeper.Get(ctx)
			if err != nil {
				return nil, err
			}
			consensusParams.Block.MaxBytes = 6_291_456
			consensusParamKeeper.Set(ctx, consensusParams)

			return vmap, nil
		},
	}
}
