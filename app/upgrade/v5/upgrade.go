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
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/samber/lo"

	"github.com/CoreumFoundation/coreum/v6/app/upgrade"
	"github.com/CoreumFoundation/coreum/v6/pkg/config"
	dexkeeper "github.com/CoreumFoundation/coreum/v6/x/dex/keeper"
	dextypes "github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

// Name defines the upgrade name.
const Name = "v5"

// New makes an upgrade handler for v5 upgrade.
func New(mm *module.Manager, configurator module.Configurator,
	chosenNetwork config.NetworkConfig,
	govParamKeeper govparamkeeper.Keeper,
	dexKeeper dexkeeper.Keeper,
	mintKeeper mintkeeper.Keeper,
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

			govParams.ProposalCancelRatio = sdkmath.LegacyMustNewDecFromStr("1.0").String()
			govParams.ProposalCancelDest = ""
			govParams.ExpeditedVotingPeriod = lo.ToPtr(24 * time.Hour)
			govParams.ExpeditedThreshold = sdkmath.LegacyMustNewDecFromStr("0.667").String()
			govParams.ExpeditedMinDeposit = sdk.NewCoins(
				sdk.NewCoin(chosenNetwork.Denom(), sdkmath.NewInt(20_000_000_000)),
			)
			govParams.MinDepositRatio = sdkmath.LegacyMustNewDecFromStr("0.01").String()
			govParams.BurnVoteQuorum = true

			err = govParamKeeper.Params.Set(ctx, govParams)
			if err != nil {
				return nil, err
			}

			sdkCtx := sdk.UnwrapSDKContext(ctx)
			//nolint:contextcheck // this is correct context passing.
			dexParams, err := dexKeeper.GetParams(sdkCtx)
			if err != nil {
				return nil, err
			}
			// 10core
			dexParams.OrderReserve = sdk.NewInt64Coin(chosenNetwork.Denom(), 10_000_000)
			//nolint:contextcheck // this is correct context passing.
			if err = dexKeeper.SetParams(sdkCtx, dexParams); err != nil {
				return nil, err
			}

			return vmap, nil
		},
	}
}
