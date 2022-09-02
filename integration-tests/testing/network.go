package testing

import (
	"time"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/x/auth"
	"github.com/CoreumFoundation/coreum/x/feemodel"
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
			MinDepositPeriod: "20s",
			VotingPeriod:     "20s",
		},
	},
}
