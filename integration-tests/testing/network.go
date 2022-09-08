package testing

import (
	"time"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/x/auth"
	feemodeltypes "github.com/CoreumFoundation/coreum/x/feemodel/types"
)

const (
	// MinDepositPeriod is the proposal deposit period duration. Deposit should be made together with the proposal
	// so not needed to spend more time to make extra deposits.
	MinDepositPeriod = time.Second / 2

	// MinVotingPeriod is the proposal voting period duration
	MinVotingPeriod = time.Second * 5
)

// NetworkConfig is the network config used by integration tests
var NetworkConfig = app.NetworkConfig{
	ChainID:       app.Devnet,
	Enabled:       true,
	GenesisTime:   time.Now(),
	AddressPrefix: "devcore",
	TokenSymbol:   app.TokenSymbolDev,
	Fee: app.FeeConfig{
		FeeModel:         feemodeltypes.DefaultModel(),
		DeterministicGas: auth.DefaultDeterministicGasRequirements(),
	},
	GovConfig: app.GovConfig{
		ProposalConfig: app.GovProposalConfig{
			MinDepositAmount: "1000",
			MinDepositPeriod: MinDepositPeriod.String(),
			VotingPeriod:     MinVotingPeriod.String(),
		},
	},
}
