package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/airdrop/types"
)

// MsgKeeper defines subscope of keeper methods required by msg service.
type MsgKeeper interface {
	Create(ctx sdk.Context, airdropInfo types.AirdropInfo) error
	Claim(ctx sdk.Context, denom string, airdropID uint64, recipient sdk.AccAddress) error
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

func (ms MsgServer) Claim(ctx context.Context, req *types.MsgClaim) (*types.MsgClaimResponse, error) {
	recipient, err := sdk.AccAddressFromBech32(req.Recipient)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = ms.keeper.Claim(sdk.UnwrapSDKContext(ctx), req.Denom, req.Id, recipient)
	if err != nil {
		return nil, err
	}
	return &types.MsgClaimResponse{}, nil
}
