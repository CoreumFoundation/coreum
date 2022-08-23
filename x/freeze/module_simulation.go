package freeze

import (
	"math/rand"

	"github.com/CoreumFoundation/coreum/testutil/sample"
	freezesimulation "github.com/CoreumFoundation/coreum/x/freeze/simulation"
	"github.com/CoreumFoundation/coreum/x/freeze/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = freezesimulation.FindAccount
	_ = simappparams.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	opWeightMsgFreezeCoin = "op_weight_msg_freeze_coin"
	// TODO: Determine the simulation weight value
	defaultWeightMsgFreezeCoin int = 100

	opWeightMsgUnfreezeCoin = "op_weight_msg_unfreeze_coin"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUnfreezeCoin int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	freezeGenesis := types.GenesisState{
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&freezeGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized  param changes for the simulator
func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {

	return []simtypes.ParamChange{}
}

// RegisterStoreDecoder registers a decoder
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgFreezeCoin int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgFreezeCoin, &weightMsgFreezeCoin, nil,
		func(_ *rand.Rand) {
			weightMsgFreezeCoin = defaultWeightMsgFreezeCoin
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgFreezeCoin,
		freezesimulation.SimulateMsgFreezeCoin(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUnfreezeCoin int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgUnfreezeCoin, &weightMsgUnfreezeCoin, nil,
		func(_ *rand.Rand) {
			weightMsgUnfreezeCoin = defaultWeightMsgUnfreezeCoin
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUnfreezeCoin,
		freezesimulation.SimulateMsgUnfreezeCoin(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
