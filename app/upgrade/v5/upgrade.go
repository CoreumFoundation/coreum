package v5

import (
	"context"
	"time"

	sdkmath "cosmossdk.io/math"
	store "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govparamkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/samber/lo"

	"github.com/CoreumFoundation/coreum/v4/app/upgrade"
	"github.com/CoreumFoundation/coreum/v4/pkg/config"
	dextypes "github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

// Name defines the upgrade name.
const Name = "v5"

// New makes an upgrade handler for v5 upgrade.
func New(mm *module.Manager, configurator module.Configurator,
	chosenNetwork config.NetworkConfig, govParamKeeper govparamkeeper.Keeper,
) upgrade.Upgrade {
	return upgrade.Upgrade{
		Name: Name,
		StoreUpgrades: store.StoreUpgrades{
			Added: []string{
				dextypes.StoreKey,
			},
		},
		Upgrade: func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
			vmap, err := mm.RunMigrations(ctx, configurator, vm)
			if err != nil {
				return nil, err
			}

			govParams, err := govParamKeeper.Params.Get(ctx)
			if err != nil {
				return nil, err
			}

			govParams.ProposalCancelRatio = sdkmath.LegacyMustNewDecFromStr("0.5").String()
			govParams.ProposalCancelDest = ""
			govParams.ExpeditedVotingPeriod = lo.ToPtr(24 * time.Hour)
			govParams.ExpeditedThreshold = sdkmath.LegacyMustNewDecFromStr("0.667").String()
			govParams.ExpeditedMinDeposit = sdk.NewCoins(
				sdk.NewCoin(chosenNetwork.Denom(), sdkmath.NewInt(4_000_000_000)),
			)
			govParams.MinDepositRatio = sdkmath.LegacyMustNewDecFromStr("0.01").String()

			err = govParamKeeper.Params.Set(ctx, govParams)
			if err != nil {
				return nil, err
			}

			return vmap, nil
		},
	}
}
