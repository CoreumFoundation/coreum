package wbank

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum/v2/x/wbank/keeper"
)

// AppModuleBasic defines the basic application module used by the wrapped bank module.
type AppModuleBasic struct {
	bank.AppModuleBasic
}

// AppModule implements an application module for the wrapped bank module.
type AppModule struct {
	bank.AppModule
	keeper keeper.BaseKeeperWrapper
}

// NewAppModule creates a new bank AppModule object.
func NewAppModule(cdc codec.Codec, keeper keeper.BaseKeeperWrapper, accountKeeper banktypes.AccountKeeper) AppModule {
	bankModule := bank.NewAppModule(cdc, keeper.BaseKeeper, accountKeeper)
	return AppModule{
		AppModule: bankModule,
		keeper:    keeper,
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// copied the bank's RegisterServices to replace with the keeper wrapper
	banktypes.RegisterMsgServer(cfg.MsgServer(), bankkeeper.NewMsgServerImpl(am.keeper))
	banktypes.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	m := bankkeeper.NewMigrator(am.keeper.BaseKeeper)
	if err := cfg.RegisterMigration(banktypes.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(err)
	}
}

// Route returns the message routing key for the bank module.
func (am AppModule) Route() sdk.Route {
	// we need to pass the wrapped keeper to the handler to use it instead of the default
	return sdk.NewRoute(banktypes.RouterKey, bank.NewHandler(am.keeper))
}
