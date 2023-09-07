package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/CoreumFoundation/coreum/v2/x/feemodel/types"
)

var _ types.MsgServer = MsgServer{}

// MsgKeeper defines an interface of keeper required by fee module.
type MsgKeeper interface {
	SetParams(ctx sdk.Context, params types.Params) error
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

// UpdateParams is a governance operation that sets parameters of the module.
func (m MsgServer) UpdateParams(ctx context.Context, req *types.MsgUpdateParams) (*types.EmptyResponse, error) {
	if m.keeper.GetAuthority() != req.Authority {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.keeper.GetAuthority(), req.Authority)
	}

	err := m.keeper.SetParams(sdk.UnwrapSDKContext(ctx), req.Params)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
