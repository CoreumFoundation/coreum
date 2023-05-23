package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v4/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"

	"github.com/CoreumFoundation/coreum/x/wibctransfer/types"
)

// Wrapper is a wrapper of the IBC transfer keeper.
type Wrapper struct {
	ibctransferkeeper.Keeper
}

// NewKeeper returns a new Wrapper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	key sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	ics4Wrapper ibctransfertypes.ICS4Wrapper,
	channelKeeper ibctransfertypes.ChannelKeeper,
	portKeeper ibctransfertypes.PortKeeper,
	authKeeper ibctransfertypes.AccountKeeper,
	bankKeeper ibctransfertypes.BankKeeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
) Wrapper {
	return Wrapper{
		Keeper: ibctransferkeeper.NewKeeper(cdc, key, paramSpace, ics4Wrapper, channelKeeper, portKeeper, authKeeper,
			bankKeeper, scopedKeeper),
	}
}

// Transfer defines a rpc handler method for MsgTransfer.
func (k Wrapper) Transfer(goCtx context.Context, msg *ibctransfertypes.MsgTransfer) (*ibctransfertypes.MsgTransferResponse, error) {
	//nolint:contextcheck // it is fine to produce the context this way
	return k.Keeper.Transfer(sdk.WrapSDKContext(types.WithDirection(sdk.UnwrapSDKContext(goCtx), types.DirectionOut)), msg)
}
