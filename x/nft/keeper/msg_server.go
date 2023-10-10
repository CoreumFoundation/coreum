package keeper

import (
	"context"

	cosmosnft "github.com/cosmos/cosmos-sdk/x/nft"

	"github.com/CoreumFoundation/coreum/v3/x/nft"
)

var _ nft.MsgServer = Keeper{}

// Send implement Send method of the types.MsgServer.
func (k Keeper) Send(goCtx context.Context, msg *nft.MsgSend) (*nft.MsgSendResponse, error) {
	_, err := k.wkeeper.Send(goCtx, &cosmosnft.MsgSend{
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
