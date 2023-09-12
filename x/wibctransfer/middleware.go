package wibctransfer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/CoreumFoundation/coreum/v3/x/wibctransfer/types"
)

var _ porttypes.IBCModule = PurposeMiddleware{}

// PurposeMiddleware adds information about IBC transfer purpose to the context.
type PurposeMiddleware struct {
	transfer.IBCModule
}

// NewPurposeMiddleware returns middleware adding purpose to the context.
func NewPurposeMiddleware(module transfer.IBCModule) PurposeMiddleware {
	return PurposeMiddleware{
		IBCModule: module,
	}
}

// OnRecvPacket adds purpose-in to the context and calls the upper implementation.
func (im PurposeMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	return im.IBCModule.OnRecvPacket(types.WithPurpose(ctx, types.PurposeIn), packet, relayer)
}

// OnAcknowledgementPacket adds purpose-ack to the context and calls the upper implementation.
func (im PurposeMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return im.IBCModule.OnAcknowledgementPacket(types.WithPurpose(ctx, types.PurposeAck), packet, acknowledgement, relayer)
}

// OnTimeoutPacket adds purpose-timeout to the context and calls the upper implementation.
func (im PurposeMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return im.IBCModule.OnTimeoutPacket(types.WithPurpose(ctx, types.PurposeTimeout), packet, relayer)
}
