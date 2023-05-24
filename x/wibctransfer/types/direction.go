package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Direction is the direction of the transfer.
type Direction string

const (
	// DirectionOut is used when IBC transfer from Coreum to peered chain is initialized by executing ibctransfertypes.MsgTransfer message.
	DirectionOut Direction = "ibcTransferOut"
	// DirectionIn is used when incoming IBC transfer comes from peered chain to Coreum.
	DirectionIn Direction = "ibcTransferIn"
)

type directionKey struct{}

// WithDirection stores IBC transfer direction inside SDK context.
func WithDirection(ctx sdk.Context, direction Direction) sdk.Context {
	return ctx.WithValue(directionKey{}, direction)
}

// IsDirectionOut returns true if context is tagged with an outgoing transfer.
func IsDirectionOut(ctx sdk.Context) bool {
	d, ok := getDirection(ctx.Context())
	return ok && d == DirectionOut
}

// IsDirectionIn returns true if context is tagged with an incoming transfer.
func IsDirectionIn(ctx sdk.Context) bool {
	d, ok := getDirection(ctx.Context())
	return ok && d == DirectionIn
}

func getDirection(ctx context.Context) (Direction, bool) {
	direction, ok := ctx.Value(directionKey{}).(Direction)
	return direction, ok
}
