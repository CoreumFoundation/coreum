package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/snapshot/types"
)

type QueryKeeper interface {
	GetPending(ctx sdk.Context, accAddress sdk.AccAddress) []types.SnapshotInfo
	GetSnapshots(ctx sdk.Context, accAddress sdk.AccAddress) []types.SnapshotInfo
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
	pending := qs.keeper.GetPending(sdk.UnwrapSDKContext(ctx), address)

	return &types.QueryPendingResponse{
		Pending: pending,
	}, nil
}

func (qs QueryService) Snapshots(ctx context.Context, req *types.QuerySnapshotsRequest) (*types.QuerySnapshotsResponse, error) {
	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	snapshots := qs.keeper.GetSnapshots(sdk.UnwrapSDKContext(ctx), address)

	return &types.QuerySnapshotsResponse{
		Snapshots: snapshots,
	}, nil
}
