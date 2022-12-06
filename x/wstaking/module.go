package wstaking

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/wstaking/keeper"
	wstakingtypes "github.com/CoreumFoundation/coreum/x/wstaking/types"
)

// AppModule implements an application module for the wrapped staking module.
type AppModule struct {
	staking.AppModule
	stakingKeeper      stakingkeeper.Keeper
	customParamsKeeper wstakingtypes.CustomParamsKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	cdc codec.Codec,
	stakingKeeper stakingkeeper.Keeper,
	ak stakingtypes.AccountKeeper,
	bk stakingtypes.BankKeeper,
	customParamsKeeper wstakingtypes.CustomParamsKeeper,
) AppModule {
	stakingAppModule := staking.NewAppModule(cdc, stakingKeeper, ak, bk)
	return AppModule{
		AppModule:          stakingAppModule,
		stakingKeeper:      stakingKeeper,
		customParamsKeeper: customParamsKeeper,
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	stakingKeeperMsgSrv := stakingkeeper.NewMsgServerImpl(am.stakingKeeper)
	// wrap the staking keeper message server to intersect the messages
	stakingtypes.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(stakingKeeperMsgSrv, am.customParamsKeeper))
	querier := stakingkeeper.Querier{Keeper: am.stakingKeeper}
	stakingtypes.RegisterQueryServer(cfg.QueryServer(), querier)

	m := stakingkeeper.NewMigrator(am.stakingKeeper)
	err := cfg.RegisterMigration(stakingtypes.ModuleName, 1, m.Migrate1to2)
	if err != nil {
		panic(errors.Wrap(err, "can't register staing migration"))
	}
}
