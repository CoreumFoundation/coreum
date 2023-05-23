package wibctransfer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v4/modules/apps/transfer"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v4/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v4/modules/core/exported"

	"github.com/CoreumFoundation/coreum/x/wibctransfer/types"
)

var _ porttypes.IBCModule = DirectionMiddleware{}

// DirectionMiddleware adds information about IBC transfer direction to the context.
type DirectionMiddleware struct {
	transfer.IBCModule
}

// NewDirectionMiddleware returns middleware adding direction to the context.
func NewDirectionMiddleware(module transfer.IBCModule) DirectionMiddleware {
	return DirectionMiddleware{
		IBCModule: module,
	}
}

// OnRecvPacket adds direction to the context and calls the upper implementation.
func (im DirectionMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	return im.IBCModule.OnRecvPacket(types.WithDirection(ctx, types.DirectionIn), packet, relayer)
}
