package testing

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
			InitialGasPrice:         sdk.NewInt(1500),
			MaxGasPrice:             sdk.NewInt(1500000),
			MaxDiscount:             0.5,
			EscalationStartBlockGas: 37500000, // 300 * BankSend message
			MaxBlockGas:             50000000, // 400 * BankSend message
			ShortAverageInertia:     10,
			LongAverageInertia:      1000,
		},
		DeterministicGas: app.DeterministicGasConfig{
			BankSend: 125000,
		},
	},
}
