package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v2/x/feemodel/types"
)

var _ types.MsgServer = MsgServer{}

// MsgKeeper defines an interface of keeper required by fee module.
type MsgKeeper interface {
	UpdateParams(ctx sdk.Context, authority string, params types.Params) error
}

// MsgServer serves grpc tx requests for the module.
type MsgServer struct {
	keeper MsgKeeper
}

// NewMsgServer returns a new instance of the MsgServer.
func NewMsgServer(keeper MsgKeeper) MsgServer {
	return MsgServer{
		keeper: keeper,
	}
}

// UpdateParams is a governance operation that sets parameters of the module.
func (ms MsgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.EmptyResponse, error) {
	if err := ms.keeper.UpdateParams(sdk.UnwrapSDKContext(goCtx), req.Authority, req.Params); err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
