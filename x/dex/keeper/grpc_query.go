package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

var _ types.QueryServer = QueryService{}

// QueryKeeper defines subscope of keeper methods required by query service.
type QueryKeeper interface {
	GetParams(ctx sdk.Context) (types.Params, error)
	GetOrderByAddressAndID(ctx sdk.Context, acc sdk.AccAddress, orderID string) (types.Order, error)
	GetOrders(
		ctx sdk.Context,
		creator sdk.AccAddress,
		pagination *query.PageRequest,
	) ([]types.Order, *query.PageResponse, error)
	GetOrderBooks(
		ctx sdk.Context,
		pagination *query.PageRequest,
	) ([]types.OrderBookData, *query.PageResponse, error)
	GetOrderBook(
		ctx sdk.Context,
		baseDenom, quoteDenom string,
	) (*types.Price, *sdkmath.Int, error)
	GetOrderBookOrders(
		ctx sdk.Context,
		baseDenom, quoteDenom string,
		side types.Side,
		pagination *query.PageRequest,
	) ([]types.Order, *query.PageResponse, error)
	GetAccountDenomOrdersCount(
		ctx sdk.Context,
		acc sdk.AccAddress,
		denom string,
	) (uint64, error)
}

// QueryService serves grpc query requests for the module.
type QueryService struct {
	keeper QueryKeeper
}

// Params queries the parameters of x/asset/ft module.
func (qs QueryService) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := qs.keeper.GetParams(sdk.UnwrapSDKContext(ctx))
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: params}, nil
}

// Orders returns creator orders.
func (qs QueryService) Orders(
	ctx context.Context,
	req *types.QueryOrdersRequest,
) (*types.QueryOrdersResponse, error) {
	creatorAddr, err := sdk.AccAddressFromBech32(req.Creator)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", req.Creator)
	}

	orders, pageRes, err := qs.keeper.GetOrders(sdk.UnwrapSDKContext(ctx), creatorAddr, req.Pagination)
	if err != nil {
		return nil, err
	}

	return &types.QueryOrdersResponse{
		Orders:     orders,
		Pagination: pageRes,
	}, nil
}

// OrderBooks queries order books.
func (qs QueryService) OrderBooks(
	ctx context.Context,
	req *types.QueryOrderBooksRequest,
) (*types.QueryOrderBooksResponse, error) {
	orderBooks, pageRes, err := qs.keeper.GetOrderBooks(sdk.UnwrapSDKContext(ctx), req.Pagination)
	if err != nil {
		return nil, err
	}

	return &types.QueryOrderBooksResponse{
		OrderBooks: orderBooks,
		Pagination: pageRes,
	}, nil
}

// OrderBook queries order book details.
func (qs QueryService) OrderBook(
	ctx context.Context,
	req *types.QueryOrderBookRequest,
) (*types.QueryOrderBookResponse, error) {
	priceTick, quantityStep, err := qs.keeper.GetOrderBook(sdk.UnwrapSDKContext(ctx), req.BaseDenom, req.QuoteDenom)
	if err != nil {
		return nil, err
	}

	return &types.QueryOrderBookResponse{
		PriceTick:    *priceTick,
		QuantityStep: *quantityStep,
	}, nil
}

// OrderBookOrders queries order book orders.
func (qs QueryService) OrderBookOrders(
	ctx context.Context,
	req *types.QueryOrderBookOrdersRequest,
) (*types.QueryOrderBookOrdersResponse, error) {
	orders, pageRes, err := qs.keeper.GetOrderBookOrders(
		sdk.UnwrapSDKContext(ctx), req.BaseDenom, req.QuoteDenom, req.Side, req.Pagination,
	)
	if err != nil {
		return nil, err
	}

	return &types.QueryOrderBookOrdersResponse{
		Orders:     orders,
		Pagination: pageRes,
	}, nil
}

// NewQueryService initiates the new instance of query service.
func NewQueryService(keeper QueryKeeper) QueryService {
	return QueryService{
		keeper: keeper,
	}
}

// Order queries order by creator and ID.
func (qs QueryService) Order(ctx context.Context, req *types.QueryOrderRequest) (*types.QueryOrderResponse, error) {
	creatorAddr, err := sdk.AccAddressFromBech32(req.Creator)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", req.Creator)
	}
	order, err := qs.keeper.GetOrderByAddressAndID(sdk.UnwrapSDKContext(ctx), creatorAddr, req.Id)
	if err != nil {
		return nil, err
	}

	return &types.QueryOrderResponse{
		Order: order,
	}, nil
}

// AccountDenomOrdersCount queries account orders count.
func (qs QueryService) AccountDenomOrdersCount(
	ctx context.Context,
	req *types.QueryAccountDenomOrdersCountRequest,
) (*types.QueryAccountDenomOrdersCountResponse, error) {
	acc, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", req.Account)
	}
	count, err := qs.keeper.GetAccountDenomOrdersCount(sdk.UnwrapSDKContext(ctx), acc, req.Denom)
	if err != nil {
		return nil, err
	}

	return &types.QueryAccountDenomOrdersCountResponse{
		Count: count,
	}, nil
}
