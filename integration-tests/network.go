package integrationtests

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

	networkConfig.CustomParamsConfig = config.CustomParamsConfig{
		Staking: config.CustomParamsStakingConfig{
			MinSelfDelegation: sdk.NewInt(10_000_000), // 10 core
		},
	}

	return networkConfig, nil
}
