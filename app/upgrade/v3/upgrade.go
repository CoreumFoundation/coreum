package v3

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibccoreexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/samber/lo"

	"github.com/CoreumFoundation/coreum/v2/app/upgrade"
	assetfttypes "github.com/CoreumFoundation/coreum/v2/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/v2/x/asset/nft/types"
	customparamstypes "github.com/CoreumFoundation/coreum/v2/x/customparams/types"
	feemodeltypes "github.com/CoreumFoundation/coreum/v2/x/feemodel/types"
)

// juno reference: https://github.com/CosmosContracts/juno/pull/646/files#diff-8ae5168a16be54c5a00ba9dcf5e54cabc4d053c2f3d77ac700aeef3f3dffd87b

// upgrade v45->v46: https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/UPGRADING.md
// upgrade v46->v47: https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/UPGRADING.md

const Name = "v3"

func New(mm *module.Manager, configurator module.Configurator, paramsKeeper paramskeeper.Keeper, consensusParamsKeeper consensusparamkeeper.Keeper) upgrade.Upgrade {
	return upgrade.Upgrade{
		Name: Name,
		StoreUpgrades: storetypes.StoreUpgrades{
			Added: []string{
				// Migration of SDK modules away from x/params:

				// https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/UPGRADING.md#xcrisis
				crisistypes.ModuleName,
				// https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/UPGRADING.md#xconsensus
				consensustypes.ModuleName,
			},
		},
		Upgrade: func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
			logger := ctx.Logger().With("upgrade", Name)

			// https://github.com/cosmos/cosmos-sdk/pull/12363/files
			// Set param key table for x/params module migration
			for _, subspace := range paramsKeeper.GetSubspaces() {
				subspace := subspace

				if lo.Contains([]string{
					// TODO(migration-away-from-x/params): Add migration of params for Coreum modules.
					feemodeltypes.ModuleName,
					assetfttypes.ModuleName,
					assetnfttypes.ModuleName,
					customparamstypes.CustomParamsStaking,

					// TODO: What should we do with ibc modules?
					ibccoreexported.ModuleName,
				}, subspace.Name()) {
					continue
				}

				keyTable, ok := map[string]paramstypes.KeyTable{
					// cosmos-sdk:
					authtypes.ModuleName:     authtypes.ParamKeyTable(),
					banktypes.ModuleName:     banktypes.ParamKeyTable(),
					stakingtypes.ModuleName:  stakingtypes.ParamKeyTable(),
					distrtypes.ModuleName:    distrtypes.ParamKeyTable(),
					slashingtypes.ModuleName: slashingtypes.ParamKeyTable(),
					govtypes.ModuleName:      govv1.ParamKeyTable(),
					crisistypes.ModuleName:   crisistypes.ParamKeyTable(),
					minttypes.ModuleName:     minttypes.ParamKeyTable(),

					// ibc:
					ibctransfertypes.ModuleName: ibctransfertypes.ParamKeyTable(),
					// TODO: do we use ICA ?
					icacontrollertypes.SubModuleName: icacontrollertypes.ParamKeyTable(),
					icahosttypes.SubModuleName:       icahosttypes.ParamKeyTable(),

					// wasm:
					wasmtypes.ModuleName: wasmtypes.ParamKeyTable(),

					// coreum:
					// TODO(migration-away-from-x/params): Add migration of params for Coreum modules.
				}[subspace.Name()]

				if !ok {
					return nil, fmt.Errorf("no keyTable defined for subspace: %s", subspace.Name())
				}

				if !subspace.HasKeyTable() {
					subspace.WithKeyTable(keyTable)
				}
			}

			// Migrate Tendermint consensus parameters from x/params module to a deprecated x/consensus module.
			// The old params module is required to still be imported in your app.go in order to handle this migration.
			baseAppLegacySS := paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
			baseapp.MigrateParams(ctx, baseAppLegacySS, &consensusParamsKeeper)

			// Run migrations
			logger.Info(fmt.Sprintf("pre migrate version map: %v", vm))
			versionMap, err := mm.RunMigrations(ctx, configurator, vm)
			if err != nil {
				return nil, err
			}
			logger.Info(fmt.Sprintf("post migrate version map: %v", versionMap))

			return versionMap, nil
		},
	}
}
