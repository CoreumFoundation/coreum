package keeper

import (
	"context"

	"cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"

	"github.com/CoreumFoundation/coreum/v6/x/wibctransfer/types"
)

// TransferKeeperWrapper is a wrapper of the IBC transfer keeper.
type TransferKeeperWrapper struct {
	ibctransferkeeper.Keeper
}

// NewTransferKeeperWrapper returns a new TransferKeeperWrapper instance.
func NewTransferKeeperWrapper(
	cdc codec.BinaryCodec,
	kvStoreService store.KVStoreService,
	paramSpace paramtypes.Subspace,
	ics4Wrapper porttypes.ICS4Wrapper,
	channelKeeper ibctransfertypes.ChannelKeeper,
	appMsgServiceRouter *baseapp.MsgServiceRouter,
	authKeeper ibctransfertypes.AccountKeeper,
	bankKeeper ibctransfertypes.BankKeeper,
	authority string,
) TransferKeeperWrapper {
	return TransferKeeperWrapper{
		Keeper: ibctransferkeeper.NewKeeper(
			cdc,
			kvStoreService,
			paramSpace,
			ics4Wrapper,
			channelKeeper,
			appMsgServiceRouter,
			authKeeper,
			bankKeeper,
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
