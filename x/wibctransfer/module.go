package wibctransfer

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/CoreumFoundation/coreum/v3/x/wibctransfer/keeper"
)

// AppModuleBasic defines the basic application module used by the wrapped IBC transfer module.
type AppModuleBasic struct {
	transfer.AppModuleBasic
}

// AppModule implements an application module for the wrapped IBC trnasfer module.
type AppModule struct {
	transfer.AppModule
	keeper keeper.TransferKeeperWrapper
}

// NewAppModule creates a new IBC transfer module AppModule object.
func NewAppModule(keeper keeper.TransferKeeperWrapper) AppModule {
	return AppModule{
		AppModule: transfer.NewAppModule(keeper.Keeper),
		keeper:    keeper,
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// copied from the IBC transfer module RegisterServices to replace with the keeper wrapper
	ibctransfertypes.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	ibctransfertypes.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	m := ibctransferkeeper.NewMigrator(am.keeper.Keeper)
	if err := cfg.RegisterMigration(ibctransfertypes.ModuleName, 1, m.MigrateTraces); err != nil {
		panic(fmt.Sprintf("failed to migrate transfer app from version 1 to 2: %v", err))
	}

	if err := cfg.RegisterMigration(ibctransfertypes.ModuleName, 2, m.MigrateTotalEscrowForDenom); err != nil {
		panic(fmt.Sprintf("failed to migrate transfer app from version 2 to 3: %v", err))
	}
}
