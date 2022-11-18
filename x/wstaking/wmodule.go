package wstaking

import (
	"github.com/CoreumFoundation/coreum/x/wstaking/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"
)

// AppWModule implements an application module for the wrapped staking module.
type AppWModule struct {
	staking.AppModule
	stakingKeeper stakingkeeper.Keeper
	keeper        keeper.Keeper
}

// NewAppWModule creates a new AppWModule object
func NewAppWModule(cdc codec.Codec, stakingKeeper stakingkeeper.Keeper, ak stakingtypes.AccountKeeper, bk stakingtypes.BankKeeper, keeper keeper.Keeper) AppWModule {
	stakingAppModule := staking.NewAppModule(cdc, stakingKeeper, ak, bk)
	return AppWModule{
		AppModule:     stakingAppModule,
		stakingKeeper: stakingKeeper,
		keeper:        keeper,
	}
}

// Route returns the message routing key for the staking module.

// FIXME add test to check consensus version

// RegisterServices registers module services.
func (am AppWModule) RegisterServices(cfg module.Configurator) {
	stakingKeeperMsgSrv := stakingkeeper.NewMsgServerImpl(am.stakingKeeper)
	// wrap the staking keeper message servet to intersect the messages
	stakingtypes.RegisterMsgServer(cfg.MsgServer(), keeper.NewWMsgServerImpl(stakingKeeperMsgSrv, am.keeper))
	querier := stakingkeeper.Querier{Keeper: am.stakingKeeper}
	stakingtypes.RegisterQueryServer(cfg.QueryServer(), querier)

	m := stakingkeeper.NewMigrator(am.stakingKeeper)
	err := cfg.RegisterMigration(stakingtypes.ModuleName, 1, m.Migrate1to2)
	if err != nil {
		panic(errors.Wrap(err, "can't register staing migration"))
	}
}
