package ft

import (
	"context"
	"encoding/json"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/CoreumFoundation/coreum/x/asset/ft/client/cli"
	"github.com/CoreumFoundation/coreum/x/asset/ft/keeper"
	v1 "github.com/CoreumFoundation/coreum/x/asset/ft/legacy/v1"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic implements the AppModuleBasic interface for the asset ft module.
type AppModuleBasic struct {
	cdc codec.BinaryCodec
}

// NewAppModuleBasic return the asset ft AppModuleBasic.
func NewAppModuleBasic(cdc codec.BinaryCodec) AppModuleBasic {
	return AppModuleBasic{
		cdc: cdc,
	}
}

// Name returns the asset ft module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the legacy codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterInterfaces registers the module's interface types.
func (a AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

// DefaultGenesis returns the asset ft module's default genesis state.
func (a AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the asset ft module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return errors.Wrapf(err, "failed to unmarshal %s genesis state", types.ModuleName)
	}
	return genState.Validate()
}

// RegisterRESTRoutes registers the asset ft module's REST service handlers.
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the asset ft module's root tx command.
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns the asset ft module's root query command.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface for the asset ft module.
type AppModule struct {
	AppModuleBasic

	keeper       keeper.Keeper
	bankKeeper   types.BankKeeper
	paramsKeeper v1.ParamsKeeper
}

// NewAppModule returns the new instance of the AppModule.
func NewAppModule(
	cdc codec.Codec,
	keeper keeper.Keeper,
	bankKeeper types.BankKeeper,
	paramsKeeper v1.ParamsKeeper,
) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		keeper:         keeper,
		bankKeeper:     bankKeeper,
		paramsKeeper:   paramsKeeper,
	}
}

// Name returns the asset ft module's name.
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

// Route returns the asset ft module's message routing key.
func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

// QuerierRoute returns the asset ft module's query routing key.
func (AppModule) QuerierRoute() string { return types.QuerierRoute }

// LegacyQuerierHandler returns the asset ft module's Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return nil
}

// RegisterServices registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryService(am.keeper, am.bankKeeper))

	m := keeper.NewMigrator(am.keeper, am.paramsKeeper)
	err := cfg.RegisterMigration(types.ModuleName, 1, m.Migrate1to2)
	if err != nil {
		panic(err)
	}
}

// RegisterInvariants registers the asset ft module's invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper)
}

// InitGenesis performs the asset ft module's genesis initialization It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	// Initialize global index to index in genesis state
	cdc.MustUnmarshalJSON(gs, &genState)

	InitGenesis(ctx, am.keeper, genState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the asset ft module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion implements ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 2 }

// BeginBlock executes all ABCI BeginBlock logic respective to the asset ft module.
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock executes all ABCI EndBlock logic respective to the asset ft module. It
// returns no validator updates.
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the asset ft module.
func (AppModule) GenerateGenesisState(_ *module.SimulationState) {}

// ProposalContents doesn't return any content functions for governance proposals.
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized asset ft param changes for the simulator.
func (AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {
	return nil
}

// RegisterStoreDecoder registers a decoder for asset ft module's types.
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the asset ft module operations with their respective weights.
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
