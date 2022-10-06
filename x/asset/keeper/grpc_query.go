package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// QueryKeeper defines subscope of keeper methods required by query service.
type QueryKeeper interface {
	GetAsset(ctx sdk.Context, id uint64) (types.Asset, error)
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
	asset, err := qs.keeper.GetAsset(sdk.UnwrapSDKContext(ctx), req.GetId())
	if err != nil {
		return nil, err
	}

	return &types.QueryAssetResponse{
		Asset: asset,
	}, nil
}
