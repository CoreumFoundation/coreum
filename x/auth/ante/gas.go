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
			gasRequired += dgd.gasRequiredByMessage(msg)
		}

		if gasDeclared := feeTx.GetGas(); gasRequired > gasDeclared {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "deterministic messages in the transaction require %d units of gas in total, while only %d were allowed by the gas limit", gasRequired, gasDeclared)
		}
	}

	return next(ctx, tx, simulate)
}

func (dgd DeterministicGasDecorator) gasRequiredByMessage(msg sdk.Msg) uint64 {
	switch msg.(type) {
	case *banktypes.MsgSend:
		return dgd.requirements.BankSend
	default:
		return 0
	}
}
