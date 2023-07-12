package v1

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/CoreumFoundation/coreum/v2/app/upgrade"
	"github.com/CoreumFoundation/coreum/v2/pkg/config"
	assetnftkeeper "github.com/CoreumFoundation/coreum/v2/x/asset/nft/keeper"
	assetnfttypes "github.com/CoreumFoundation/coreum/v2/x/asset/nft/types"
	"github.com/CoreumFoundation/coreum/v2/x/nft"
)

// Name defines the upgrade name.
const Name = "v1"

// NewV1Upgrade makes an upgrade handler for v1 upgrade.
func NewV1Upgrade(mm *module.Manager, configurator module.Configurator, chosenNetwork config.NetworkConfig, assetNFTKeeper assetnftkeeper.Keeper) upgrade.Upgrade {
	return upgrade.Upgrade{
		Name: Name,
		StoreUpgrades: storetypes.StoreUpgrades{
			Added: []string{assetnfttypes.ModuleName, nft.ModuleName},
		},
		Upgrade: func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
			afterVM, err := mm.RunMigrations(ctx, configurator, vm)
			if err != nil {
				return nil, err
			}

			params := assetNFTKeeper.GetParams(ctx)
			params.MintFee = sdk.NewInt64Coin(chosenNetwork.Denom(), 0)
			assetNFTKeeper.SetParams(ctx, params)

			return afterVM, nil
		},
	}
}
