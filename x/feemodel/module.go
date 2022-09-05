package feemodel

import (
	"encoding/json"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/CoreumFoundation/coreum/x/feemodel/types"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

// Keeper defines an interface of keeper required by fee module
type Keeper interface {
	TrackedGas(ctx sdk.Context) int64
	SetParams(ctx sdk.Context, params types.Params)
	GetParams(ctx sdk.Context) types.Params
	GetShortEMAGas(ctx sdk.Context) int64
	SetShortEMAGas(ctx sdk.Context, emaGas int64)
	GetLongEMAGas(ctx sdk.Context) int64
	SetLongEMAGas(ctx sdk.Context, emaGas int64)
	SetMinGasPrice(ctx sdk.Context, minGasPrice sdk.Coin)
}

// AppModuleBasic defines the basic application module used by the fee module.
type AppModuleBasic struct {
}

// Name returns the fee module's name.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers the fee module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// DefaultGenesis returns default genesis state as raw bytes for the fee
// module.
func (amb AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&types.GenesisState{Params: types.DefaultParams()})
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
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {}

// GetTxCmd returns the root tx command for the fee module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns no root query command for the fee module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// RegisterInterfaces registers interfaces and implementations of the fee module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {}

// AppModule implements an application module for the fee module.
type AppModule struct {
	AppModuleBasic

	keeper   Keeper
	feeDenom string // FIXME (wojtek): store it in genesis
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {}

// NewAppModule creates a new AppModule object
func NewAppModule(
	keeper Keeper,
	feeDenom string,
) AppModule {
	return AppModule{
		keeper:   keeper,
		feeDenom: feeDenom,
	}
}

// Name returns the fee module's name.
func (AppModule) Name() string { return types.ModuleName }

// RegisterInvariants registers the fee module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// Route returns the message routing key for the fee module.
func (am AppModule) Route() sdk.Route { return sdk.Route{} }

// QuerierRoute returns the fee module's querier route name.
func (AppModule) QuerierRoute() string { return "" }

// LegacyQuerierHandler returns the fee module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return nil
}

// InitGenesis performs genesis initialization for the fee module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesis types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesis)

	// FIXME (wojtek): Remove this after crust imports new version of coreum
	if genesis.Validate() != nil {
		genesis = types.GenesisState{
			Params: types.DefaultParams(),
		}
	}

	am.keeper.SetParams(ctx, genesis.Params)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the fee
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	return nil
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock performs a no-op.
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock returns the end blocker for the fee module. It returns no validator
// updates.
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	// TODO (wojtek): add simulation tests
	currentGasUsage := am.keeper.TrackedGas(ctx)
	params := am.keeper.GetParams(ctx)
	model := types.NewModel(params)

	newShortEMA := types.CalculateEMA(am.keeper.GetShortEMAGas(ctx), currentGasUsage,
		params.ShortEmaBlockLength)
	newLongEMA := types.CalculateEMA(am.keeper.GetLongEMAGas(ctx), currentGasUsage,
		params.LongEmaBlockLength)

	minGasPrice := model.CalculateNextGasPrice(newShortEMA, newLongEMA)

	am.keeper.SetShortEMAGas(ctx, newShortEMA)
	am.keeper.SetLongEMAGas(ctx, newLongEMA)
	am.keeper.SetMinGasPrice(ctx, sdk.NewCoin(am.feeDenom, minGasPrice))

	return []abci.ValidatorUpdate{}
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the fee module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {}

// ProposalContents doesn't return any content functions for governance proposals.
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized fee param changes for the simulator.
func (AppModule) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	return nil
}

// RegisterStoreDecoder registers a decoder for supply module's types
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
