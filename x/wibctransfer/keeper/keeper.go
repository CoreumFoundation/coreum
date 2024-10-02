package keeper

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/CoreumFoundation/coreum/v5/x/wibctransfer/types"
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
	authority string,
) TransferKeeperWrapper {
	return TransferKeeperWrapper{
		Keeper: ibctransferkeeper.NewKeeper(
			cdc,
			key,
			paramSpace,
			ics4Wrapper,
			channelKeeper,
			portKeeper,
			authKeeper,
			bankKeeper,
			scopedKeeper,
			authority,
		),
	}
}

// Transfer defines a rpc handler method for MsgTransfer.
func (k TransferKeeperWrapper) Transfer(
	ctx context.Context, msg *ibctransfertypes.MsgTransfer,
) (*ibctransfertypes.MsgTransferResponse, error) {
	ctx = types.WithPurpose(sdk.UnwrapSDKContext(ctx), types.PurposeOut)
	//nolint:contextcheck // this is correct context passing
	return k.Keeper.Transfer(ctx, msg)
}
