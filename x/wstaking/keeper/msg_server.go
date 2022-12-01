package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	wstakingtypes "github.com/CoreumFoundation/coreum/x/wstaking/types"
)

// MsgServer is wrapper staking customParamsKeeper message server.
type MsgServer struct {
	stakingtypes.MsgServer
	customParamsKeeper wstakingtypes.CustomParamsKeeper
}

// NewMsgServerImpl returns an implementation of the staking wrapped MsgServer.
func NewMsgServerImpl(stakingMsgSrv stakingtypes.MsgServer, customParamsKeeper wstakingtypes.CustomParamsKeeper) stakingtypes.MsgServer {
	return MsgServer{
		MsgServer:          stakingMsgSrv,
		customParamsKeeper: customParamsKeeper,
	}
}

// CreateValidator defines wrapped method for creating a new validator.
func (s MsgServer) CreateValidator(goCtx context.Context, msg *stakingtypes.MsgCreateValidator) (*stakingtypes.MsgCreateValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	expectedMinSelfDelegation := s.customParamsKeeper.GetStakingParams(ctx).MinSelfDelegation
	if expectedMinSelfDelegation.GT(msg.MinSelfDelegation) {
		return nil, sdkerrors.Wrapf(
			stakingtypes.ErrSelfDelegationBelowMinimum, "min self delegation must be greater or equal than global min self delegation: %s", msg.MinSelfDelegation,
		)
	}

	return s.MsgServer.CreateValidator(goCtx, msg)
}
