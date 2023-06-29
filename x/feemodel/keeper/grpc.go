package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/CoreumFoundation/coreum/x/feemodel/types"
)

// QueryKeeper defines subscope of keeper methods required by query service.
type QueryKeeper interface {
	GetParams(ctx sdk.Context) types.Params
	GetMinGasPrice(ctx sdk.Context) sdk.DecCoin
	EstimateFutureGasPrice(ctx sdk.Context, after uint32) (sdk.DecCoin, sdk.DecCoin)
}

// NewQueryService creates query service.
func NewQueryService(keeper QueryKeeper) QueryService {
	return QueryService{
		keeper: keeper,
	}
}

// QueryService serves grpc requests for fee model.
type QueryService struct {
	keeper QueryKeeper
}

// MinGasPrice returns current minimum gas price required by the network.
func (qs QueryService) MinGasPrice(ctx context.Context, req *types.QueryMinGasPriceRequest) (*types.QueryMinGasPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	return &types.QueryMinGasPriceResponse{
		MinGasPrice: qs.keeper.GetMinGasPrice(sdk.UnwrapSDKContext(ctx)),
	}, nil
}

// MinGasPrice returns current minimum gas price required by the network.
func (qs QueryService) RecommendedGasPrice(ctx context.Context, req *types.QueryRecommendedGasPriceRequest) (*types.QueryRecommendedGasPriceResponse, error) {
	const maxAfterBlocks uint32 = 50
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.AfterBlocks > maxAfterBlocks {
		return nil, status.Errorf(codes.InvalidArgument, "after blocks must be lower than %d", maxAfterBlocks)
	}

	low, high := qs.keeper.EstimateFutureGasPrice(sdk.UnwrapSDKContext(ctx), req.AfterBlocks)
	med := sdk.NewDecCoinFromDec(low.Denom, low.Amount.Add(high.Amount).QuoInt64(2))
	return &types.QueryRecommendedGasPriceResponse{
		Low:  low,
		Med:  med,
		High: high,
	}, nil
}

// Params returns params of fee model.
func (qs QueryService) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	return &types.QueryParamsResponse{
		Params: qs.keeper.GetParams(sdk.UnwrapSDKContext(ctx)),
	}, nil
}
