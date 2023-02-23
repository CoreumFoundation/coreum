package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
)

// DenyMessagesDecorator denies transactions containing configured messages.
type DenyMessagesDecorator struct {
	deniedMessages map[string]struct{}
}

// NewDenyMessagesDecorator creates new DenyMessagesDecorator.
func NewDenyMessagesDecorator(msgs ...sdk.Msg) DenyMessagesDecorator {
	deniedMessages := map[string]struct{}{}
	for _, msg := range msgs {
		deniedMessages[proto.MessageName(msg)] = struct{}{}
	}
	return DenyMessagesDecorator{
		deniedMessages: deniedMessages,
	}
}

// AnteHandle resets the gas limit inside GasMeter.
func (dmd DenyMessagesDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	for _, msg := range tx.GetMsgs() {
		msgName := proto.MessageName(msg)
		if _, exists := dmd.deniedMessages[msgName]; exists {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "message %q is disabled", msgName)
		}
	}
	return next(ctx.WithGasMeter(sdk.NewInfiniteGasMeter()), tx, simulate)
}
