package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/CoreumFoundation/coreum/v2/x/customparams/types"
)

var _ types.MsgServer = MsgServer{}

// MsgKeeper defines an interface of keeper required by fee module.
type MsgKeeper interface {
	SetStakingParams(ctx sdk.Context, params types.StakingParams) error
	GetAuthority() string
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

// UpdateStakingParams is a governance operation that sets staking parameters.
func (m MsgServer) UpdateStakingParams(ctx context.Context, req *types.MsgUpdateStakingParams) (*types.EmptyResponse, error) {
	if m.keeper.GetAuthority() != req.Authority {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.keeper.GetAuthority(), req.Authority)
	}

	err := m.keeper.SetStakingParams(sdk.UnwrapSDKContext(ctx), req.StakingParams)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
