package fee

import (
	"encoding/json"
	"math"
	"math/big"
	"math/rand"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/CoreumFoundation/coreum/x/fee/keeper"
	"github.com/CoreumFoundation/coreum/x/fee/types"
)

// Model stores parameters defining fee model of coreum blockchain
type Model struct {
	FeeDenom                             string
	InitialGasPrice                      sdk.Int
	MaxGasPrice                          sdk.Int
	MaxDiscount                          float64
	EscalationStartBlockGas              int64
	MaxBlockGas                          int64
	EscalationInertia                    float64
	NumOfBlocksForCurrentAverageBlockGas uint
	NumOfBlocksForAverageBlockGas        uint
}

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the fee module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the fee module's name.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers the fee module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// DefaultGenesis returns default genesis state as raw bytes for the fee
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return nil
}

// ValidateGenesis performs genesis state validation for the fee module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
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

	keeper   keeper.Keeper
	feeModel Model
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec,
	keeper keeper.Keeper,
	feeModel Model) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		feeModel:       feeModel,
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
	am.keeper.SetMinGasPrice(ctx, sdk.NewCoin(am.feeModel.FeeDenom, am.feeModel.InitialGasPrice))
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the fee
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	return nil
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 2 }

// BeginBlock performs a no-op.
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock returns the end blocker for the fee module. It returns no validator
// updates.
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	previousCurrentAverage := uint64(am.keeper.GetCurrentAverageGas(ctx))
	previousAverage := uint64(am.keeper.GetAverageGas(ctx))
	currentGasUsage := uint64(am.keeper.TrackedGas(ctx))

	newCurrentAverage := int64((uint64(am.feeModel.NumOfBlocksForCurrentAverageBlockGas-1)*previousCurrentAverage + currentGasUsage) / uint64(am.feeModel.NumOfBlocksForCurrentAverageBlockGas))
	newAverage := int64((uint64(am.feeModel.NumOfBlocksForAverageBlockGas-1)*previousAverage + currentGasUsage) / uint64(am.feeModel.NumOfBlocksForAverageBlockGas))

	minGasPrice := calculateNextGasPrice(am.feeModel, newCurrentAverage, newAverage)

	am.keeper.SetCurrentAverageGas(ctx, newCurrentAverage)
	am.keeper.SetAverageGas(ctx, newAverage)
	am.keeper.SetMinGasPrice(ctx, sdk.NewCoin(am.feeModel.FeeDenom, sdk.NewIntFromBigInt(minGasPrice)))

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

func calculateNextGasPrice(feeModel Model, currentAverageGas int64, averageGas int64) *big.Int {
	switch {
	case currentAverageGas >= feeModel.MaxBlockGas:
		return feeModel.MaxGasPrice.BigInt()
	case currentAverageGas > feeModel.EscalationStartBlockGas:
		maxDiscountedGasPrice := computeMaxDiscountedGasPrice(feeModel.InitialGasPrice.BigInt(), feeModel.MaxDiscount)

		height := new(big.Int).Sub(feeModel.MaxGasPrice.BigInt(), maxDiscountedGasPrice)
		width := float64(feeModel.MaxBlockGas - feeModel.EscalationStartBlockGas)
		x := float64(currentAverageGas - feeModel.EscalationStartBlockGas)

		escalationOffsetFloat := new(big.Float).SetInt(height)
		escalationOffsetFloat.Mul(escalationOffsetFloat, new(big.Float).SetFloat64(math.Pow(x/width, feeModel.EscalationInertia)))
		escalationOffset, _ := escalationOffsetFloat.Int(nil)

		return maxDiscountedGasPrice.Add(maxDiscountedGasPrice, escalationOffset)
	case currentAverageGas >= averageGas:
		return computeMaxDiscountedGasPrice(feeModel.InitialGasPrice.BigInt(), feeModel.MaxDiscount)
	case averageGas > 0:
		discountFactor := math.Pow(1.-feeModel.MaxDiscount, float64(currentAverageGas)/float64(averageGas))

		gasPriceFloat := big.NewFloat(0).SetInt(feeModel.InitialGasPrice.BigInt())
		gasPriceFloat.Mul(gasPriceFloat, big.NewFloat(discountFactor))
		minGasPrice, _ := gasPriceFloat.Int(nil)

		return minGasPrice
	default:
		return feeModel.InitialGasPrice.BigInt()
	}
}

func computeMaxDiscountedGasPrice(initialGasPrice *big.Int, maxDiscount float64) *big.Int {
	maxDiscountedGasPriceFloat := big.NewFloat(0).SetInt(initialGasPrice)
	maxDiscountedGasPriceFloat.Mul(maxDiscountedGasPriceFloat, big.NewFloat(1.-maxDiscount))
	maxDiscountedGasPrice, _ := maxDiscountedGasPriceFloat.Int(nil)

	return maxDiscountedGasPrice
}
