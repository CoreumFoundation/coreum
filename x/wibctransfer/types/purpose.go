package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Direction is the direction of the transfer.
type Direction string

const (
	// PurposeOut is used when IBC transfer from Coreum to peered chain is initialized by executing ibctransfertypes.MsgTransfer message.
	PurposeOut Direction = "ibcTransferOut"
	// PurposeIn is used when incoming IBC transfer comes from peered chain to Coreum.
	PurposeIn Direction = "ibcTransferIn"
	// PurposeAck is used when outgoing IBC transfer from Coreum is acknowledged by the peered chain.
	PurposeAck Direction = "ibcTransferAck"
	// PurposeTimeout is used when outgoing IBC transfer from Coreum times out.
	PurposeTimeout Direction = "ibcTransferTimeout"
)

type directionKey struct{}

// WithPurpose stores IBC transfer purpose inside SDK context.
func WithPurpose(ctx sdk.Context, direction Direction) sdk.Context {
	return ctx.WithValue(directionKey{}, direction)
}

// IsPurposeOut returns true if context is tagged with an outgoing transfer.
func IsPurposeOut(ctx sdk.Context) bool {
	d, ok := getPurpose(ctx.Context())
	return ok && d == PurposeOut
}

// IsPurposeIn returns true if context is tagged with an incoming transfer.
func IsPurposeIn(ctx sdk.Context) bool {
	d, ok := getPurpose(ctx.Context())
	return ok && d == PurposeIn
}

// IsPurposeAck returns true if context is tagged with an acknowledged transfer.
func IsPurposeAck(ctx sdk.Context) bool {
	d, ok := getPurpose(ctx.Context())
	return ok && d == PurposeAck
}

// IsPurposeTimeout returns true if context is tagged with timed-out transfer.
func IsPurposeTimeout(ctx sdk.Context) bool {
	d, ok := getPurpose(ctx.Context())
	return ok && d == PurposeTimeout
}

func getPurpose(ctx context.Context) (Direction, bool) {
	direction, ok := ctx.Value(directionKey{}).(Direction)
	return direction, ok
}
