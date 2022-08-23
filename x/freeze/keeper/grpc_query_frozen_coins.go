package keeper

import (
	"context"

	"github.com/CoreumFoundation/coreum/x/freeze/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k BaseKeeper) FrozenCoins(goCtx context.Context, req *types.QueryFrozenCoinsRequest) (*types.QueryFrozenCoinsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Parse coin holder address
	holderAddr, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, err
	}

	frozenCoins, err := k.ListAccountFrozenCoins(ctx, holderAddr)
	if err != nil {
		return nil, err
	}

	return &types.QueryFrozenCoinsResponse{
		Coins: frozenCoins,
	}, nil
}
