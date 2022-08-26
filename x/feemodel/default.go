package feemodel

import sdk "github.com/cosmos/cosmos-sdk/types"

// DefaultModel returns model with default values
func DefaultModel() Model {
	return Model{
		// TODO: Find good parameters before lunching mainnet
		InitialGasPrice:         sdk.NewInt(1500),
		MaxGasPrice:             sdk.NewInt(1500000),
		MaxDiscount:             sdk.MustNewDecFromStr("0.5"),
		EscalationStartBlockGas: 37500000, // 300 * BankSend message
		MaxBlockGas:             50000000, // 400 * BankSend message
		ShortAverageBlockLength: 10,
		LongAverageBlockLength:  1000,
	}
}
