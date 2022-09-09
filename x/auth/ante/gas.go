package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// DeterministicGasRequirements specifies gas required by some transaction types
type DeterministicGasRequirements struct {
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

// DeterministicGasDecorator verifies that declared gas limit meets the requirements
// of messages for which deterministic gas amount is defined.
// CONTRACT: Tx must implement FeeTx to use DeterministicGasDecorator
type DeterministicGasDecorator struct {
	requirements DeterministicGasRequirements
}

// NewDeterministicGasDecorator creates ante decorator refusing transactions which does not offer enough gas limit
// covering requirements of messages for which deterministic gas amount is defined.
func NewDeterministicGasDecorator(requirements DeterministicGasRequirements) DeterministicGasDecorator {
	return DeterministicGasDecorator{
		requirements: requirements,
	}
}

// AnteHandle handles transaction in ante decorator
func (dgd DeterministicGasDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if ctx.IsCheckTx() && !simulate {
		var gasRequired uint64
		for _, msg := range tx.GetMsgs() {
			gasRequired += dgd.requirements.GasRequiredByMessage(msg)
		}

		if gasDeclared := feeTx.GetGas(); gasRequired > gasDeclared {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "deterministic messages in the transaction require %d units of gas in total, while only %d were allowed by the gas limit", gasRequired, gasDeclared)
		}
	}

	return next(ctx, tx, simulate)
}
