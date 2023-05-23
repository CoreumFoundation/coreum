package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Direction is the direction of the transfer.
type Direction string

const (
	// DirectionOut is used when IBC transfer to another chain is initialized by executing ibctransfertypes.MsgTransfer message.
	DirectionOut Direction = "ibcTransferOut"
	// DirectionIn is used when incoming IBC transfer comes to the target chain.
	DirectionIn Direction = "ibcTransferIn"
)

type directionKey struct{}

// GetDirection returns IBC transfer direction stored inside context.
func GetDirection(ctx context.Context) (Direction, bool) {
	direction, ok := ctx.Value(directionKey{}).(Direction)
	if !ok {
		return "", false
	}

	return direction, true
}

// WithDirection stores IBC transfer direction inside SDK context.
func WithDirection(ctx sdk.Context, direction Direction) sdk.Context {
	return ctx.WithValue(directionKey{}, direction)
}
