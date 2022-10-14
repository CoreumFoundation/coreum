package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/x/airdrop/types"
)

// MsgKeeper defines subscope of keeper methods required by msg service.
type MsgKeeper interface {
	Create(ctx sdk.Context, airdropInfo types.AirdropInfo) error
}

// MsgServer serves grpc tx requests for assets module.
type MsgServer struct {
	keeper MsgKeeper
}

// NewMsgServer returns a new instance of the MsgServer.
func NewMsgServer(keeper MsgKeeper) MsgServer {
	return MsgServer{
		keeper: keeper,
	}
}

func (ms MsgServer) Create(ctx context.Context, req *types.MsgCreate) (*types.MsgCreateResponse, error) {
	err := ms.keeper.Create(sdk.UnwrapSDKContext(ctx), types.AirdropInfo{
		Sender:        req.Sender,
		Height:        req.Height,
		Description:   req.Description,
		RequiredDenom: req.RequiredDenom,
		Offer:         req.Offer,
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateResponse{}, nil
}
