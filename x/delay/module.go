package delay

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
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

	"github.com/CoreumFoundation/coreum/v2/x/delay/keeper"
	"github.com/CoreumFoundation/coreum/v2/x/delay/types"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the delay module.
type AppModuleBasic struct{}

// Name returns the delay module's name.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers the delay module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// DefaultGenesis returns default genesis state as raw bytes for the delay module.
func (amb AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the delay module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return errors.Wrapf(err, "failed to unmarshal %s genesis state", types.ModuleName)
	}
	return genState.Validate()
}

// RegisterRESTRoutes registers the REST routes for the delay module.
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the delay module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {}

// GetTxCmd returns the root tx command for the delay module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns no root query command for the delay module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// RegisterInterfaces registers interfaces and implementations of the delay module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {}

// AppModule implements an application module for the delay module.
type AppModule struct {
	AppModuleBasic

	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object.
func NewAppModule(keeper keeper.Keeper) AppModule {
	return AppModule{
		keeper: keeper,
	}
}

// Name returns the delay module's name.
func (AppModule) Name() string { return types.ModuleName }

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {}

// RegisterInvariants registers the delay module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// QuerierRoute returns the delay module's querier route name.
func (AppModule) QuerierRoute() string { return types.RouterKey }

// InitGenesis performs genesis initialization for the delay module.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	cdc.MustUnmarshalJSON(gs, &genState)

	InitGenesis(ctx, am.keeper, genState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the delay module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock executes delayed items.
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
	if err := am.keeper.ExecuteDelayedItems(ctx); err != nil {
		panic(err)
	}
}

// EndBlock returns the end blocker for the delay module.
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the delay module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {}

// ProposalContents doesn't return any content functions for governance proposals.
// FIXME(v47-legacy) try to remove/replace the usage.
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent { //nolint:staticcheck // we need to keep backward compatibility
	return nil
}

// RegisterStoreDecoder registers a decoder for supply module's types.
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
