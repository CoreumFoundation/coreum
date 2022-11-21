package testing

import (
	"time"

	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
)

const (
	// votingPeriod is the proposal voting period duration
	votingPeriod = time.Second * 15
)

// NewNetworkConfig returns the network config used by integration tests.
func NewNetworkConfig() (config.NetworkConfig, error) {
	networkConfig, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	if err != nil {
		return config.NetworkConfig{}, err
	}
	networkConfig.GovConfig.ProposalConfig = config.GovProposalConfig{
		MinDepositAmount: "1000",
		VotingPeriod:     votingPeriod.String(),
	}

	networkConfig.FundedAccounts = nil
	networkConfig.GenTxs = nil

	return networkConfig, nil
}
