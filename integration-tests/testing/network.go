package testing

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/app"
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
		FeeModel: feemodel.Model{
			InitialGasPrice:         sdk.NewInt(1500),
			MaxGasPrice:             sdk.NewInt(1500000),
			MaxDiscount:             sdk.MustNewDecFromStr("0.5"),
			EscalationStartBlockGas: 37500000, // 300 * BankSend message
			MaxBlockGas:             50000000, // 400 * BankSend message
			ShortAverageBlockLength: 10,
			LongAverageBlockLength:  1000,
		},
		DeterministicGas: app.DeterministicGasConfig{
			BankSend: 125000,
		},
	},
}
