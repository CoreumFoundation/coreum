package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/x/nft"
	nftkeeper "github.com/CoreumFoundation/coreum/x/nft/keeper"
	"github.com/CoreumFoundation/coreum/x/wnft/types"
)

// NFTKeeperWrapper wraps the original nft keeper and intercepts its original methods if needed
type NFTKeeperWrapper struct {
	nftkeeper.Keeper
	assetNFTProvider types.AssetNFTProvider
}

// NewWrappedNFTKeeper returns a new instance of the WrappedNFTKeeper
func NewWrappedNFTKeeper(originalKeeper nftkeeper.Keeper, provider types.AssetNFTProvider) NFTKeeperWrapper {
	return NFTKeeperWrapper{
		Keeper:           originalKeeper,
		assetNFTProvider: provider,
	}
}

// Send overwrites Send method of the original keeper.
func (wk NFTKeeperWrapper) Send(goCtx context.Context, msg *nft.MsgSend) (*nft.MsgSendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	owner := wk.GetOwner(ctx, msg.ClassId, msg.Id)
	if !owner.Equals(sender) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not the owner of nft %s", sender, msg.Id)
	}

	receiver, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return nil, err
	}

	if err := wk.Transfer(ctx, msg.ClassId, msg.Id, receiver); err != nil {
		return nil, err
	}

	err = ctx.EventManager().EmitTypedEvent(&nft.EventSend{
		ClassId:  msg.ClassId,
		Id:       msg.Id,
		Sender:   msg.Sender,
		Receiver: msg.Receiver,
	})
	if err != nil {
		return nil, err
	}

	return &nft.MsgSendResponse{}, nil
}

// Transfer overwrites the original transfer function to include the
func (wk NFTKeeperWrapper) Transfer(ctx sdk.Context, classID string, nftID string, receiver sdk.AccAddress) error {
	if err := wk.assetNFTProvider.BeforeTransfer(ctx, classID, nftID, receiver); err != nil {
		return err
	}

	return wk.Keeper.Transfer(ctx, classID, nftID, receiver)
}
