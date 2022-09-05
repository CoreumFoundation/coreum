package testing

import (
	"time"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/x/auth"
	feemodeltypes "github.com/CoreumFoundation/coreum/x/feemodel/types"
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
}
