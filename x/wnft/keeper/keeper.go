package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/x/nft"
	nftkeeper "cosmossdk.io/x/nft/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/v6/x/wnft/types"
)

// Wrapper wraps the original nft keeper and intercepts its original methods if needed.
type Wrapper struct {
	nftkeeper.Keeper
	nonFungibleTokenProvider types.NonFungibleTokenProvider
}

// NewWrappedNFTKeeper returns a new instance of the WrappedNFTKeeper.
func NewWrappedNFTKeeper(originalKeeper nftkeeper.Keeper, provider types.NonFungibleTokenProvider) Wrapper {
	return Wrapper{
		Keeper:                   originalKeeper,
		nonFungibleTokenProvider: provider,
	}
}

// Send overwrites Send method of the original keeper.
// Copied from
// https://github.com/cosmos/cosmos-sdk/blob/a1143138716b64bc4fa0aa53c0f0fa59eb675bb7/x/nft/keeper/msg_server.go#L14
// On each update we need to make sure it is up-to-date with original cosmos version of nft.
func (wk Wrapper) Send(ctx context.Context, msg *nft.MsgSend) (*nft.MsgSendResponse, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	owner := wk.GetOwner(ctx, msg.ClassId, msg.Id)
	if !owner.Equals(sender) {
		return nil, sdkerrors.Wrapf(cosmoserrors.ErrUnauthorized, "%s is not the owner of nft %s", sender, msg.Id)
	}

	receiver, err := sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return nil, err
	}

	if err := wk.Transfer(sdk.UnwrapSDKContext(ctx), msg.ClassId, msg.Id, receiver); err != nil {
		return nil, err
	}

	err = sdk.UnwrapSDKContext(ctx).EventManager().EmitTypedEvent(&nft.EventSend{
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

// Transfer overwrites the original transfer function to include our custom interceptor.
func (wk Wrapper) Transfer(ctx sdk.Context, classID, nftID string, receiver sdk.AccAddress) error {
	return wk.nonFungibleTokenProvider.Transfer(ctx, classID, nftID, receiver)
}
