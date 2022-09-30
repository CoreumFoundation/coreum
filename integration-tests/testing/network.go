package testing

import (
	"time"

	"github.com/CoreumFoundation/coreum/pkg/config"
	feemodeltypes "github.com/CoreumFoundation/coreum/x/feemodel/types"
)

const (
	// minDepositPeriod is the proposal deposit period duration. Deposit should be made together with the proposal
	// so not needed to spend more time to make extra deposits.
	minDepositPeriod = time.Millisecond * 500

	// minVotingPeriod is the proposal voting period duration
	minVotingPeriod = time.Second * 15

	// unbondingTime is the coins unbonding time
	unbondingTime = time.Second * 5
)

// NetworkConfig is the network config used by integration tests
var NetworkConfig = config.NetworkConfig{
	ChainID:       config.Devnet,
	Enabled:       true,
	GenesisTime:   time.Now(),
	AddressPrefix: "devcore",
	TokenSymbol:   config.TokenSymbolDev,
	Fee: config.FeeConfig{
		FeeModel:         feemodeltypes.DefaultModel(),
		DeterministicGas: config.DefaultDeterministicGasRequirements(),
	},
	GovConfig: config.GovConfig{
		ProposalConfig: config.GovProposalConfig{
			MinDepositAmount: "1000",
			MinDepositPeriod: minDepositPeriod.String(),
			VotingPeriod:     minVotingPeriod.String(),
		},
	},
	StakingConfig: config.StakingConfig{
		UnbondingTime: unbondingTime.String(),
		MaxValidators: 32,
	},
}
