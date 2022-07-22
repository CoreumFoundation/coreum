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
		FeeModel: app.FeeModel{
			InitialGasPrice:                      big.NewInt(1500),
			MaxDiscount:                          0.15,
			OptimalBlockGas:                      43750000, // 350 * BankSend transactions
			MaxBlockGas:                          50000000, // 400 * BankSend transactions
			NumOfBlocksForCurrentAverageBlockGas: 10,
			NumOfBlocksForAverageBlockGas:        1000,
		},
		DeterministicGas: app.DeterministicGasConfig{
			BankSend: 125000,
		},
	},
}
