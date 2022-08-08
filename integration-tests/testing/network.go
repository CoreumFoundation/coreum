package testing

import (
	"math/big"
	"time"

	"github.com/CoreumFoundation/coreum/app"
)

// NetworkConfig is the network config used by integration tests
var NetworkConfig = app.NetworkConfig{
	ChainID:       app.Devnet,
	Enabled:       true,
	GenesisTime:   time.Now(),
	AddressPrefix: "devcore",
	TokenSymbol:   app.TokenSymbolDev,
	Fee: app.FeeConfig{
		InitialGasPrice:       big.NewInt(1500),
		MinDiscountedGasPrice: big.NewInt(1000),
		DeterministicGas: app.DeterministicGasConfig{
			BankSend: 120000,
		},
	},
}
