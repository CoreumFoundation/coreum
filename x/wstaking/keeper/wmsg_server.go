package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// WMsgServer is wrapper staking keeper message server.
type WMsgServer struct {
	stakingtypes.MsgServer
	keeper Keeper
}

// NewWMsgServerImpl returns an implementation of the bank WMsgServer interface
// for the provided Keeper.
func NewWMsgServerImpl(stakingMsgSrv stakingtypes.MsgServer, keeper Keeper) stakingtypes.MsgServer {
	return WMsgServer{
		MsgServer: stakingMsgSrv,
		keeper:    keeper,
	}
}

// CreateValidator defines wrapped method for creating a new validator
func (s WMsgServer) CreateValidator(goCtx context.Context, msg *stakingtypes.MsgCreateValidator) (*stakingtypes.MsgCreateValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	expectedMinSelfDelegation := s.keeper.GetParams(ctx).MinSelfDelegation
	if expectedMinSelfDelegation.GT(msg.MinSelfDelegation) {
		return nil, sdkerrors.Wrapf(
			stakingtypes.ErrSelfDelegationBelowMinimum, "expected %s, got %s", expectedMinSelfDelegation, msg.MinSelfDelegation,
		)
	}

	return s.MsgServer.CreateValidator(goCtx, msg)
}
