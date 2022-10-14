package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/x/airdrop/types"
)

// QueryKeeper defines subscope of keeper methods required by query service.
type QueryKeeper interface {
	List(ctx sdk.Context, denom string) []types.Airdrop
}

// QueryService serves grpc query requests for assets module.
type QueryService struct {
	keeper QueryKeeper
}

// NewQueryService initiates the new instance of query service.
func NewQueryService(keeper QueryKeeper) QueryService {
	return QueryService{
		keeper: keeper,
	}
}

func (qs QueryService) List(ctx context.Context, req *types.QueryListRequest) (*types.QueryListResponse, error) {
	return &types.QueryListResponse{
		Airdrops: qs.keeper.List(sdk.UnwrapSDKContext(ctx), req.Denom),
	}, nil
}
