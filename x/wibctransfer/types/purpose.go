package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Purpose is the purpose of the transfer.
type Purpose string

const (
	// PurposeOut is used when IBC transfer from Coreum to peered chain is initialized by executing
	// ibctransfertypes.MsgTransfer message.
	PurposeOut Purpose = "out"
	// PurposeIn is used when incoming IBC transfer comes from peered chain to Coreum.
	PurposeIn Purpose = "in"
	// PurposeAck is used when outgoing IBC transfer from Coreum is acknowledged by the peered chain.
	PurposeAck Purpose = "ack"
	// PurposeTimeout is used when outgoing IBC transfer from Coreum times out.
	PurposeTimeout Purpose = "timeout"
)

type purposeKey struct{}

// WithPurpose stores IBC transfer purpose inside SDK context.
func WithPurpose(ctx sdk.Context, direction Purpose) sdk.Context {
	return ctx.WithValue(purposeKey{}, direction)
}

// IsPurposeOut returns true if context is tagged with an outgoing transfer.
func IsPurposeOut(ctx sdk.Context) bool {
	d, ok := GetPurpose(ctx.Context())
	return ok && d == PurposeOut
}

// IsPurposeIn returns true if context is tagged with an incoming transfer.
func IsPurposeIn(ctx sdk.Context) bool {
	d, ok := GetPurpose(ctx.Context())
	return ok && d == PurposeIn
}

// IsPurposeAck returns true if context is tagged with an acknowledged transfer.
func IsPurposeAck(ctx sdk.Context) bool {
	d, ok := GetPurpose(ctx.Context())
	return ok && d == PurposeAck
}

// IsPurposeTimeout returns true if context is tagged with timed-out transfer.
func IsPurposeTimeout(ctx sdk.Context) bool {
	d, ok := GetPurpose(ctx.Context())
	return ok && d == PurposeTimeout
}

// GetPurpose returns the ibc purpose from the context.
func GetPurpose(ctx context.Context) (Purpose, bool) {
	purpose, ok := ctx.Value(purposeKey{}).(Purpose)
	return purpose, ok
}
