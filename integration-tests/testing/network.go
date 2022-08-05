package testing

import (
	"time"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// NetworkConfig is the network config used by integration tests
var NetworkConfig = app.NetworkConfig{
	ChainID:       app.Devnet,
	Enabled:       true,
	GenesisTime:   time.Now(),
	AddressPrefix: "devcore",
	TokenSymbol:   app.TokenSymbolDev,
	Fee: app.FeeConfig{
		InitialGasPrice:       types.NewInt(1500),
		MinDiscountedGasPrice: types.NewInt(1000),
		DeterministicGas: app.DeterministicGasConfig{
			BankSend: 120000,
		},
	},
}
