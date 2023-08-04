package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/CoreumFoundation/coreum/v2/x/wibctransfer/types"
)

// TransferKeeperWrapper is a wrapper of the IBC transfer keeper.
type TransferKeeperWrapper struct {
	ibctransferkeeper.Keeper
}

// NewTransferKeeperWrapper returns a new TransferKeeperWrapper instance.
func NewTransferKeeperWrapper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	ics4Wrapper porttypes.ICS4Wrapper,
	channelKeeper ibctransfertypes.ChannelKeeper,
	portKeeper ibctransfertypes.PortKeeper,
	authKeeper ibctransfertypes.AccountKeeper,
	bankKeeper ibctransfertypes.BankKeeper,
	scopedKeeper exported.ScopedKeeper,
) TransferKeeperWrapper {
	return TransferKeeperWrapper{
		Keeper: ibctransferkeeper.NewKeeper(cdc, key, paramSpace, ics4Wrapper, channelKeeper, portKeeper, authKeeper,
			bankKeeper, scopedKeeper),
	}
}

// Transfer defines a rpc handler method for MsgTransfer.
func (k TransferKeeperWrapper) Transfer(goCtx context.Context, msg *ibctransfertypes.MsgTransfer) (*ibctransfertypes.MsgTransferResponse, error) {
	goCtx = sdk.WrapSDKContext(types.WithPurpose(sdk.UnwrapSDKContext(goCtx), types.PurposeOut))
	//nolint:contextcheck // it is fine to produce the context this way
	return k.Keeper.Transfer(goCtx, msg)
}
