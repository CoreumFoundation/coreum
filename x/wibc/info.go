package wibc

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v4/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v4/modules/core/exported"
)

var _ porttypes.IBCModule = InfoMiddleware{}

// Action is the action represented by the info object.
type Action string

// Info stores port and channel on which IBC packet is received.
type Info struct {
	Port    string
	Channel string
	Action  Action
}

type infoKey struct{}

// GetInfo returns IBC info stored inside context.
func GetInfo(ctx context.Context) (Info, bool) {
	info, ok := ctx.Value(infoKey{}).(Info)
	if !ok {
		return Info{}, false
	}

	return info, true
}

// WithSDKInfo stores IBC info inside SDK context.
func WithSDKInfo(ctx sdk.Context, info Info) sdk.Context {
	return ctx.WithValue(infoKey{}, info)
}

// WithGOInfo stores IBC info inside GO context.
func WithGOInfo(ctx context.Context, info Info) context.Context {
	return sdk.WrapSDKContext(WithSDKInfo(sdk.UnwrapSDKContext(ctx), info))
}

// NewInfoMiddleware returns new info middleware.
func NewInfoMiddleware(module porttypes.IBCModule, action Action) InfoMiddleware {
	return InfoMiddleware{
		module: module,
		action: action,
	}
}

// InfoMiddleware adds information about IBC port and channel of the received packet to the context.
type InfoMiddleware struct {
	module porttypes.IBCModule
	action Action
}

// OnRecvPacket adds port and channel info to the context and calls the upper implementation.
func (im InfoMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	return im.module.OnRecvPacket(WithSDKInfo(ctx, Info{
		Port:    packet.DestinationPort,
		Channel: packet.DestinationChannel,
		Action:  im.action,
	}), packet, relayer)
}

// OnChanOpenInit simply calls the implementation of the wrapped module.
func (im InfoMiddleware) OnChanOpenInit(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version string) (string, error) {
	return im.module.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, version)
}

// OnChanOpenTry simply calls the implementation of the wrapped module.
func (im InfoMiddleware) OnChanOpenTry(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, counterpartyVersion string) (version string, err error) {
	return im.module.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, counterpartyVersion)
}

// OnChanOpenAck simply calls the implementation of the wrapped module.
func (im InfoMiddleware) OnChanOpenAck(ctx sdk.Context, portID, channelID string, counterpartyChannelID string, counterpartyVersion string) error {
	return im.module.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm simply calls the implementation of the wrapped module.
func (im InfoMiddleware) OnChanOpenConfirm(ctx sdk.Context, portID, channelID string) error {
	return im.module.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit simply calls the implementation of the wrapped module.
func (im InfoMiddleware) OnChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	return im.module.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm simply calls the implementation of the wrapped module.
func (im InfoMiddleware) OnChanCloseConfirm(ctx sdk.Context, portID, channelID string) error {
	return im.module.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnAcknowledgementPacket simply calls the implementation of the wrapped module.
func (im InfoMiddleware) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	return im.module.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket simply calls the implementation of the wrapped module.
func (im InfoMiddleware) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	return im.module.OnTimeoutPacket(ctx, packet, relayer)
}
