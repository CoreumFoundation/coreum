package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v3/x/customparams/types"
)

var _ types.MsgServer = MsgServer{}

// MsgKeeper defines an interface of keeper required by fee module.
type MsgKeeper interface {
	UpdateStakingParams(ctx sdk.Context, authority string, params types.StakingParams) error
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

// UpdateStakingParams is a governance operation that sets staking parameters.
func (m MsgServer) UpdateStakingParams(ctx context.Context, req *types.MsgUpdateStakingParams) (*types.EmptyResponse, error) {
	if err := m.keeper.UpdateStakingParams(sdk.UnwrapSDKContext(ctx), req.Authority, req.StakingParams); err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
