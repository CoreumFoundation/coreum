package testing

import (
	"time"

	"github.com/CoreumFoundation/coreum/pkg/config"
)

const (
	// votingPeriod is the proposal voting period duration
	votingPeriod = time.Second * 15

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
		VotingPeriod:     votingPeriod.String(),
	}

	networkConfig.StakingConfig = config.StakingConfig{
		UnbondingTime: unbondingTime.String(),
		MaxValidators: 32,
	}

	networkConfig.FundedAccounts = nil
	networkConfig.GenTxs = nil

	return networkConfig, nil
}
