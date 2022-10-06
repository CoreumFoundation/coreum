package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// QueryKeeper defines subscope of keeper methods required by query service.
type QueryKeeper interface {
	GetAsset(ctx sdk.Context, id string) types.Asset
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

// Asset queries an asset.
func (qs QueryService) Asset(ctx context.Context, req *types.QueryAssetRequest) (*types.QueryAssetResponse, error) {
	return &types.QueryAssetResponse{
		Asset: qs.keeper.GetAsset(sdk.UnwrapSDKContext(ctx), req.GetId()),
	}, nil
}
