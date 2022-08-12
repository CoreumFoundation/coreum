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
			MaxGasPrice:                          big.NewInt(15000),
			MaxDiscount:                          0.5,
			EscalationStartBlockGas:              37500000, // 300 * BankSend message
			MaxBlockGas:                          50000000, // 400 * BankSend message
			EscalationInertia:                    2.5,
			NumOfBlocksForCurrentAverageBlockGas: 10,
			NumOfBlocksForAverageBlockGas:        1000,
		},
		DeterministicGas: app.DeterministicGasConfig{
			BankSend: 125000,
		},
	},
}
