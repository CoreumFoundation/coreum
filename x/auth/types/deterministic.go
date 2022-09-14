package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// DefaultDeterministicGasRequirements returns default config for deterministic gas
func DefaultDeterministicGasRequirements() DeterministicGasRequirements {
	return DeterministicGasRequirements{
		FixedGas: 50000,
		BankSend: 30000,
	}
}

const (
	// FreeBytes defines how many tx bytes are stored for free (included in `FixedGas` price)
	FreeBytes = 2048

	// FreeSignatures defines how many secp256k1 signatures are verified for free (included in `FixedGas` price)
	FreeSignatures = 1
)

// DeterministicGasRequirements specifies gas required by some transaction types
type DeterministicGasRequirements struct {
	// FixedGas is the fixed amount of gas charged on each transaction as a payment for executing ante handler. This includes:
	// - most of the stuff done by ante decorators
	// - `FreeSignatures` secp256k1 signature verifications
	// - `FreeBytes` bytes of transaction
	FixedGas uint64
	BankSend uint64
}

// GasRequiredByMessage returns gas required by a sdk.Msg.
// If fixed gas is not specified for the message type it returns 0.
func (dgr DeterministicGasRequirements) GasRequiredByMessage(msg sdk.Msg) uint64 {
	switch msg.(type) {
	case *banktypes.MsgSend:
		return dgr.BankSend
	default:
		return 0
	}
}
