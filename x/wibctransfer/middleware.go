package wibctransfer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"

	"github.com/CoreumFoundation/coreum/v6/x/wibctransfer/types"
)

var _ porttypes.IBCModule = PurposeMiddleware{}

// PurposeMiddleware adds information about IBC transfer purpose to the context.
type PurposeMiddleware struct {
	porttypes.IBCModule
}

// NewPurposeMiddleware returns middleware adding purpose to the context.
func NewPurposeMiddleware(module porttypes.IBCModule) PurposeMiddleware {
	return PurposeMiddleware{
		IBCModule: module,
	}
}

// OnRecvPacket adds purpose-in to the context and calls the upper implementation.
func (im PurposeMiddleware) OnRecvPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	ctx = sdk.UnwrapSDKContext(types.WithPurpose(ctx, types.PurposeIn))
	return im.IBCModule.OnRecvPacket(ctx, channelVersion, packet, relayer)
}

// OnAcknowledgementPacket adds purpose-ack to the context and calls the upper implementation.
func (im PurposeMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	ctx = sdk.UnwrapSDKContext(types.WithPurpose(ctx, types.PurposeAck))
	return im.IBCModule.OnAcknowledgementPacket(ctx, channelVersion, packet, acknowledgement, relayer)
}

// OnTimeoutPacket adds purpose-timeout to the context and calls the upper implementation.
func (im PurposeMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	channelVersion string,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	ctx = sdk.UnwrapSDKContext(types.WithPurpose(ctx, types.PurposeAck))
	return im.IBCModule.OnTimeoutPacket(ctx, channelVersion, packet, relayer)
}
