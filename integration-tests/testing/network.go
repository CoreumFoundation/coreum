package testing

import (
	"fmt"
	"time"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/x/auth"
	"github.com/CoreumFoundation/coreum/x/feemodel"
)

const (
	MinDepositPeriod = time.Second * 5
	MinVotingPeriod  = time.Second * 5
)

// NetworkConfig is the network config used by integration tests
var NetworkConfig = app.NetworkConfig{
	ChainID:       app.Devnet,
	Enabled:       true,
	GenesisTime:   time.Now(),
	AddressPrefix: "devcore",
	TokenSymbol:   app.TokenSymbolDev,
	Fee: app.FeeConfig{
		FeeModel:         feemodel.DefaultModel(),
		DeterministicGas: auth.DefaultDeterministicGasRequirements(),
	},
	GovConfig: app.GovConfig{
		ProposalConfig: app.GovProposalConfig{
			MinDepositAmount: "10000000",
			MinDepositPeriod: fmt.Sprintf("%ds", int(MinDepositPeriod.Seconds())),
			VotingPeriod:     fmt.Sprintf("%ds", int(MinVotingPeriod.Seconds())),
		},
	},
}
