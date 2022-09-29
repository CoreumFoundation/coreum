package config

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// DefaultDeterministicGasRequirements returns default config for deterministic gas
func DefaultDeterministicGasRequirements() DeterministicGasRequirements {
	return DeterministicGasRequirements{
		FixedGas:       50000,
		FreeBytes:      2048,
		FreeSignatures: 1,
		BankSend:       30000,
	}
}

// DeterministicGasRequirements specifies gas required by some transaction types
type DeterministicGasRequirements struct {
	// FixedGas is the fixed amount of gas charged on each transaction as a payment for executing ante handler. This includes:
	// - most of the stuff done by ante decorators
	// - `FreeSignatures` secp256k1 signature verifications
	// - `FreeBytes` bytes of transaction
	FixedGas uint64

	// FreeBytes defines how many tx bytes are stored for free (included in `FixedGas` price)
	FreeBytes uint64

	// FreeSignatures defines how many secp256k1 signatures are verified for free (included in `FixedGas` price)
	FreeSignatures uint64

	BankSend uint64
}

// GasRequiredByMessage returns gas required by a sdk.Msg.
// If fixed gas is not specified for the message type it returns 0.
func (dgr DeterministicGasRequirements) GasRequiredByMessage(msg sdk.Msg) (uint64, bool) {
	// Following is the list of messages having deterministic gas amount defined.
	// To test the real gas usage return `false` and run an integration test which reports the used gas.
	// Then define a reasonable value for the message and return `true` again.

	switch msg.(type) {
	case *banktypes.MsgSend:
		return dgr.BankSend, true
	default:
		return 0, false
	}
}

// TxBaseGas is the free gas we give to every transaction to cover costs of tx size and signature verification
func (dgr DeterministicGasRequirements) TxBaseGas(params authtypes.Params) uint64 {
	return dgr.FreeBytes*params.TxSizeCostPerByte + dgr.FreeSignatures*params.SigVerifyCostSecp256k1
}
