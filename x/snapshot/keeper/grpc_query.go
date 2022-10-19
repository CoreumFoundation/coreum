package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/snapshot/types"
)

type QueryKeeper interface {
	GetPending(ctx sdk.Context, accAddress sdk.AccAddress) ([]types.SnapshotRequest, error)
	GetSnapshots(ctx sdk.Context, accAddress sdk.AccAddress) ([]types.Snapshot, error)
}

type QueryService struct {
	keeper QueryKeeper
}

func NewQueryService(keeper QueryKeeper) QueryService {
	return QueryService{
		keeper: keeper,
	}
}

func (qs QueryService) Pending(ctx context.Context, req *types.QueryPendingRequest) (*types.QueryPendingResponse, error) {
	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	pending, err := qs.keeper.GetPending(sdk.UnwrapSDKContext(ctx), address)
	if err != nil {
		return nil, err
	}

	return &types.QueryPendingResponse{
		Pending: pending,
	}, nil
}

func (qs QueryService) Snapshots(ctx context.Context, req *types.QuerySnapshotsRequest) (*types.QuerySnapshotsResponse, error) {
	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	snapshots, err := qs.keeper.GetSnapshots(sdk.UnwrapSDKContext(ctx), address)
	if err != nil {
		return nil, err
	}

	return &types.QuerySnapshotsResponse{
		Snapshots: snapshots,
	}, nil
}
