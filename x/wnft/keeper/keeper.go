package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/x/nft"
	nftkeeper "github.com/CoreumFoundation/coreum/x/nft/keeper"
	"github.com/CoreumFoundation/coreum/x/wnft/types"
)

// Wrapper wraps the original nft keeper and intercepts its original methods if needed
type Wrapper struct {
	nftkeeper.Keeper
	nonFungibleTokenProvider types.NonFungibleTokenProvider
}

// NewWrappedNFTKeeper returns a new instance of the WrappedNFTKeeper
func NewWrappedNFTKeeper(originalKeeper nftkeeper.Keeper, provider types.NonFungibleTokenProvider) Wrapper {
	return Wrapper{
		Keeper:                   originalKeeper,
		nonFungibleTokenProvider: provider,
	}
}

// Send overwrites Send method of the original keeper.
// Copied from https://github.com/cosmos/cosmos-sdk/blob/a1143138716b64bc4fa0aa53c0f0fa59eb675bb7/x/nft/keeper/msg_server.go#L14
func (wk Wrapper) Send(goCtx context.Context, msg *nft.MsgSend) (*nft.MsgSendResponse, error) {
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

// Transfer overwrites the original transfer function to include our custom interceptor
func (wk Wrapper) Transfer(ctx sdk.Context, classID string, nftID string, receiver sdk.AccAddress) error {
	if err := wk.nonFungibleTokenProvider.BeforeTransfer(ctx, classID, nftID, receiver); err != nil {
		return err
	}

	return wk.Keeper.Transfer(ctx, classID, nftID, receiver)
}
