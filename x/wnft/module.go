package wnft

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/CoreumFoundation/coreum/x/nft"
	nftmodule "github.com/CoreumFoundation/coreum/x/nft/module"
	"github.com/CoreumFoundation/coreum/x/wnft/keeper"
)

// AppModuleBasic implements the basic application module for the wrapped nft module.
type AppModuleBasic struct {
	nftmodule.AppModuleBasic
}

// AppModule implements an application module for the wnft module.
type AppModule struct {
	nftmodule.AppModule
	keeper keeper.Wrapper
}

// NewAppModule creates a new wnft AppModule object.
func NewAppModule(cdc codec.Codec, keeper keeper.Wrapper, ak nft.AccountKeeper, bk nft.BankKeeper, registry codectypes.InterfaceRegistry) AppModule {
	bankModule := nftmodule.NewAppModule(cdc, keeper.Keeper, ak, bk, registry)
	return AppModule{
		AppModule: bankModule,
		keeper:    keeper,
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	nft.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	nft.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}
