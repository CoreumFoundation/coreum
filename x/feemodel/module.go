package feemodel

import (
	"context"
	"encoding/json"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/hashicorp/go-metrics"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/v6/x/feemodel/client/cli"
	"github.com/CoreumFoundation/coreum/v6/x/feemodel/keeper"
	"github.com/CoreumFoundation/coreum/v6/x/feemodel/types"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModule{}
	_ module.AppModuleSimulation = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasServices         = AppModule{}

	_ appmodule.HasEndBlocker = AppModule{}
	_ appmodule.AppModule     = AppModule{}
)

// Keeper defines an interface of keeper required by fee module.
//
//nolint:interfacebloat // the interface exposes all the method, breaking it down is not helpful.
type Keeper interface {
	TrackedGas(ctx sdk.Context) int64
	SetParams(ctx sdk.Context, params types.Params) error
	GetParams(ctx sdk.Context) (types.Params, error)
	GetShortEMAGas(ctx sdk.Context) int64
	SetShortEMAGas(ctx sdk.Context, emaGas int64) error
	GetLongEMAGas(ctx sdk.Context) int64
	SetLongEMAGas(ctx sdk.Context, emaGas int64) error
	GetMinGasPrice(ctx sdk.Context) sdk.DecCoin
	SetMinGasPrice(ctx sdk.Context, minGasPrice sdk.DecCoin) error
	CalculateEdgeGasPriceAfterBlocks(ctx sdk.Context, after uint32) (sdk.DecCoin, sdk.DecCoin, error)
	UpdateParams(ctx sdk.Context, authority string, params types.Params) error
}

// AppModuleBasic defines the basic application module used by the fee module.
type AppModuleBasic struct{}

// Name returns the fee module's name.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers the fee module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// DefaultGenesis returns default genesis state as raw bytes for the fee
// module.
func (amb AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the fee module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var genesis types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genesis); err != nil {
		return errors.Wrapf(err, "failed to unmarshal %s genesis state", types.ModuleName)
	}
	return genesis.Validate()
}

// RegisterRESTRoutes registers the REST routes for the fee module.
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the fee module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
	if err != nil {
		panic(err)
	}
}

// GetTxCmd returns the root tx command for the fee module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns no root query command for the fee module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// RegisterInterfaces registers interfaces and implementations of the fee module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// AppModule implements an application module for the fee module.
type AppModule struct {
	AppModuleBasic

	keeper       Keeper
	paramsKeeper types.ParamsKeeper
}

// NewAppModule creates a new AppModule object.
func NewAppModule(keeper Keeper, paramsKeeper types.ParamsKeeper) AppModule {
	return AppModule{
		keeper:       keeper,
		paramsKeeper: paramsKeeper,
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryService(am.keeper))

	m := keeper.NewMigrator(am.keeper, am.paramsKeeper)
	if err := cfg.RegisterMigration(types.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(errors.Errorf("can't register module %s migrations, err: %s", types.ModuleName, err))
	}
}

// Name returns the fee module's name.
func (AppModule) Name() string { return types.ModuleName }

// RegisterInvariants registers the fee module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// InitGenesis performs genesis initialization for the fee module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	genesis := &types.GenesisState{}
	cdc.MustUnmarshalJSON(data, genesis)

	if err := am.keeper.SetParams(ctx, genesis.Params); err != nil {
		panic(err)
	}
	if err := am.keeper.SetMinGasPrice(ctx, genesis.MinGasPrice); err != nil {
		panic(err)
	}
	if err := am.keeper.SetShortEMAGas(ctx, genesis.ShortEmaGas); err != nil {
		panic(err)
	}
	if err := am.keeper.SetLongEMAGas(ctx, genesis.LongEmaGas); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the exported genesis state as raw bytes for the fee
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	params, err := am.keeper.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	return cdc.MustMarshalJSON(&types.GenesisState{
		Params:      params,
		MinGasPrice: am.keeper.GetMinGasPrice(ctx),
		ShortEmaGas: am.keeper.GetShortEMAGas(ctx),
		LongEmaGas:  am.keeper.GetLongEMAGas(ctx),
	})
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 2 }

// EndBlock returns the end blocker for the fee module. It returns no validator
// updates.
func (am AppModule) EndBlock(c context.Context) error {
	ctx := sdk.UnwrapSDKContext(c)
	currentGasUsage := am.keeper.TrackedGas(ctx)
	params, err := am.keeper.GetParams(ctx)
	if err != nil {
		return err
	}
	model := types.NewModel(params.Model)
	previousMinGasPrice := am.keeper.GetMinGasPrice(ctx)

	newShortEMA := types.CalculateEMA(am.keeper.GetShortEMAGas(ctx), currentGasUsage,
		params.Model.ShortEmaBlockLength)
	newLongEMA := types.CalculateEMA(am.keeper.GetLongEMAGas(ctx), currentGasUsage,
		params.Model.LongEmaBlockLength)

	newMinGasPrice := model.CalculateNextGasPrice(newShortEMA, newLongEMA)

	if err := am.keeper.SetShortEMAGas(ctx, newShortEMA); err != nil {
		return err
	}
	if err := am.keeper.SetLongEMAGas(ctx, newLongEMA); err != nil {
		return err
	}
	if err := am.keeper.SetMinGasPrice(
		ctx,
		sdk.NewDecCoinFromDec(previousMinGasPrice.Denom, newMinGasPrice),
	); err != nil {
		return err
	}
	metrics.SetGauge([]string{"min_gas_price"}, float32(newMinGasPrice.MustFloat64()))

	return nil
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the fee module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// RegisterStoreDecoder registers a decoder for supply module's types.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
