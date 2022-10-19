package testing

import (
	"encoding/json"
	"time"

	"github.com/CoreumFoundation/coreum/pkg/config"
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

// NewNetworkConfig returns the network config used by integration tests.
func NewNetworkConfig() (config.NetworkConfig, error) {
	networkConfig, err := config.NetworkConfigByChainID(config.ChainIDDev)
	if err != nil {
		return config.NetworkConfig{}, err
	}
	networkConfig.GovConfig.ProposalConfig = config.GovProposalConfig{
		MinDepositAmount: "1000",
		MinDepositPeriod: minDepositPeriod.String(),
		VotingPeriod:     minVotingPeriod.String(),
	}

	networkConfig.StakingConfig = config.StakingConfig{
		UnbondingTime: unbondingTime.String(),
		MaxValidators: 32,
	}

	networkConfig.FundedAccounts = make([]config.FundedAccount, 0)
	networkConfig.GenTxs = make([]json.RawMessage, 0)

	return networkConfig, nil
}
