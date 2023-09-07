package v3

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclientkeeper "github.com/cosmos/ibc-go/v7/modules/core/02-client/keeper"
	ibccoreexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctmmigrations "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint/migrations"
	"github.com/samber/lo"

	"github.com/CoreumFoundation/coreum/v2/app/upgrade"
	assetfttypes "github.com/CoreumFoundation/coreum/v2/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/v2/x/asset/nft/types"
	customparamstypes "github.com/CoreumFoundation/coreum/v2/x/customparams/types"
	feemodeltypes "github.com/CoreumFoundation/coreum/v2/x/feemodel/types"
)

// References:
// upgrade v45->v46: https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/UPGRADING.md
// upgrade v46->v47: https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/UPGRADING.md

// Name defines the upgrade name.
const Name = "v3"

// New makes an upgrade handler for v3 upgrade.
//
//nolint:funlen
func New(
	mm *module.Manager,
	configurator module.Configurator,
	appCoded codec.Codec,
	paramsKeeper paramskeeper.Keeper,
	consensusParamsKeeper consensusparamkeeper.Keeper,
	ibcClientKeeper ibcclientkeeper.Keeper,
	govKeeper govkeeper.Keeper,
	stakingKeeper stakingkeeper.Keeper,
) upgrade.Upgrade {
	return upgrade.Upgrade{
		Name: Name,
		StoreUpgrades: storetypes.StoreUpgrades{
			Added: []string{
				// Migration of SDK modules away from x/params:

				// https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/UPGRADING.md#xcrisis
				crisistypes.StoreKey,
				// https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/UPGRADING.md#xconsensus
				consensustypes.StoreKey,
				customparamstypes.StoreKey,
			},
		},
		Upgrade: func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
			logger := ctx.Logger().With("upgrade", Name)

			moduleNameKeyTableMap := map[string]paramstypes.KeyTable{
				// cosmos-sdk:
				authtypes.ModuleName:     authtypes.ParamKeyTable(), //nolint:staticcheck
				banktypes.ModuleName:     banktypes.ParamKeyTable(), //nolint:staticcheck
				stakingtypes.ModuleName:  stakingtypes.ParamKeyTable(),
				distrtypes.ModuleName:    distrtypes.ParamKeyTable(),    //nolint:staticcheck
				slashingtypes.ModuleName: slashingtypes.ParamKeyTable(), //nolint:staticcheck
				govtypes.ModuleName:      govv1.ParamKeyTable(),         //nolint:staticcheck
				crisistypes.ModuleName:   crisistypes.ParamKeyTable(),   //nolint:staticcheck
				minttypes.ModuleName:     minttypes.ParamKeyTable(),     //nolint:staticcheck

				// ibc:
				ibctransfertypes.ModuleName: ibctransfertypes.ParamKeyTable(),

				// wasm:
				wasmtypes.ModuleName: wasmtypes.ParamKeyTable(), //nolint:staticcheck

				// coreum:
				// TODO(migration-away-from-x/params): Add migration of params for Coreum modules. Skipping them for now.
			}

			// https://github.com/cosmos/cosmos-sdk/pull/12363/files
			// Set param key table for x/params module migration
			for _, subspace := range paramsKeeper.GetSubspaces() {
				subspace := subspace

				if lo.Contains([]string{
					// TODO(migration-away-from-x/params): Add migration of params for Coreum modules. Skipping them for now.
					feemodeltypes.ModuleName,
					assetfttypes.ModuleName,
					assetnfttypes.ModuleName,
					customparamstypes.CustomParamsStaking,

					// We skip ibc module intentionally, because each module inside ibc has its own params.
					ibccoreexported.ModuleName,
				}, subspace.Name()) {
					continue
				}

				keyTable, ok := moduleNameKeyTableMap[subspace.Name()]

				if !ok {
					return nil, fmt.Errorf("no keyTable defined for subspace: %s", subspace.Name())
				}

				// Reference x/mint inside cosmos-sdk: https://github.com/cosmos/cosmos-sdk/pull/12363/files#diff-eff63269a2122bd0bc1c08f3d029aa99812aa47ce9fd85ef531dc3e04327ffc5L36
				if !subspace.HasKeyTable() {
					subspace.WithKeyTable(keyTable)
				}
			}

			// Migrate Tendermint consensus parameters from x/params module to a dedicated x/consensus module.
			// The old params module is required to still be imported in your app.go in order to handle this migration.
			baseAppLegacySS := paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
			baseapp.MigrateParams(ctx, baseAppLegacySS, &consensusParamsKeeper)

			// Run migrations
			logger.Info(fmt.Sprintf("pre migrate version map: %v", vm))
			vmPost, err := mm.RunMigrations(ctx, configurator, vm)
			if err != nil {
				return nil, err
			}
			logger.Info(fmt.Sprintf("post migrate version map: %v", vmPost))

			// IBC:
			// v4 -> v5: https://github.com/cosmos/ibc-go/blob/main/docs/migrations/v4-to-v5.md#chains
			// No upgrade needed.

			// v5 -> v6: https://github.com/cosmos/ibc-go/blob/main/docs/migrations/v5-to-v6.md#chains
			// Skipped for now. Might be needed if we decide to integrate interchain accounts (ICS27).

			// v6 -> v7: https://github.com/cosmos/ibc-go/blob/main/docs/migrations/v6-to-v7.md#chains
			// prune expired tendermint consensus states to save storage space
			_, err = ibctmmigrations.PruneExpiredConsensusStates(ctx, appCoded, ibcClientKeeper)
			if err != nil {
				return nil, err
			}

			// v7 -> v7.1: https://github.com/cosmos/ibc-go/blob/main/docs/migrations/v7-to-v7_1.md#chains
			// explicitly update the IBC 02-client params, adding the localhost client type
			params := ibcClientKeeper.GetParams(ctx)
			params.AllowedClients = append(params.AllowedClients, ibccoreexported.Localhost)
			ibcClientKeeper.SetParams(ctx, params)

			// TODO(new-gov-params): Discuss new values for the following params with the team and set here & inside genesis.v3.json.
			// min_initial_deposit_ratio, burn_vote_quorum, burn_proposal_deposit_prevote, burn_vote_veto
			govParams := govKeeper.GetParams(ctx)
			govParams.MinInitialDepositRatio = sdk.NewDec(0).Quo(sdk.NewDec(100)).String()
			govParams.BurnVoteQuorum = false
			govParams.BurnProposalDepositPrevote = false
			govParams.BurnVoteVeto = true
			if err := govKeeper.SetParams(ctx, govParams); err != nil {
				return nil, err
			}

			// TODO(new-staking-params): Discuss new values for the following params with the team and set here & inside genesis.v3.json.
			// min_commission_rate
			stakingParams := stakingKeeper.GetParams(ctx)
			stakingParams.MinCommissionRate = sdk.ZeroDec()
			err = stakingKeeper.SetParams(ctx, stakingParams)
			if err != nil {
				return nil, err
			}

			return vmPost, nil
		},
	}
}
