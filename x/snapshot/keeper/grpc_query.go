package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/snapshot/types"
)

type QueryKeeper interface {
	GetPendingFreezeRequests(ctx sdk.Context, accAddress sdk.AccAddress) ([]types.FreezeRequest, error)
	GetFrozenSnapshots(ctx sdk.Context, accAddress sdk.AccAddress) ([]types.FrozenSnapshot, error)
}

type QueryService struct {
	keeper QueryKeeper
}

func NewQueryService(keeper QueryKeeper) QueryService {
	return QueryService{
		keeper: keeper,
	}
}

func (qs QueryService) PendingFreezeRequests(ctx context.Context, req *types.QueryPendingFreezeRequestsRequest) (*types.QueryPendingFreezeRequestsResponse, error) {
	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	requests, err := qs.keeper.GetPendingFreezeRequests(sdk.UnwrapSDKContext(ctx), address)
	if err != nil {
		return nil, err
	}

	return &types.QueryPendingFreezeRequestsResponse{
		Requests: requests,
	}, nil
}

func (qs QueryService) FrozenSnapshots(ctx context.Context, req *types.QueryFrozenSnapshotsRequest) (*types.QueryFrozenSnapshotsResponse, error) {
	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	snapshots, err := qs.keeper.GetFrozenSnapshots(sdk.UnwrapSDKContext(ctx), address)
	if err != nil {
		return nil, err
	}

	return &types.QueryFrozenSnapshotsResponse{
		Snapshots: snapshots,
	}, nil
}
