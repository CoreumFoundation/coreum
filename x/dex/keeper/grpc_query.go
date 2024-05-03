package keeper

import (
	"context"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

var _ types.QueryServer = QueryService{}

// QueryKeeper defines subscope of keeper methods required by query service.
type QueryKeeper interface{}

// QueryService serves grpc query requests for the module.
type QueryService struct {
	keeper QueryKeeper
}

// NewQueryService initiates the new instance of query service.
func NewQueryService(keeper QueryKeeper) QueryService {
	return QueryService{
		keeper: keeper,
	}
}

// Orders returns a lif of current orders.
func (qs QueryService) Orders(ctx context.Context, req *types.QueryOrdersRequest) (*types.QueryOrdersResponse, error) {
	return &types.QueryOrdersResponse{}, nil
}
