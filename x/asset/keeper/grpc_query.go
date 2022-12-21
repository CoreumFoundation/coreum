package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

var _ types.QueryServer = QueryService{}

// QueryKeeper defines subscope of keeper methods required by query service.
type QueryKeeper interface {
	GetFungibleToken(ctx sdk.Context, denom string) (types.FungibleToken, error)
	GetFrozenBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetFrozenBalances(ctx sdk.Context, addr sdk.AccAddress, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error)
	GetWhitelistedBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetWhitelistedBalances(ctx sdk.Context, addr sdk.AccAddress, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error)
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

// FungibleToken queries an fungible token.
func (qs QueryService) FungibleToken(ctx context.Context, req *types.QueryFungibleTokenRequest) (*types.QueryFungibleTokenResponse, error) {
	token, err := qs.keeper.GetFungibleToken(sdk.UnwrapSDKContext(ctx), req.GetDenom())
	if err != nil {
		return nil, err
	}

	return &types.QueryFungibleTokenResponse{
		FungibleToken: token,
	}, nil
}

// FrozenBalances lists frozen balances on a given account
func (qs QueryService) FrozenBalances(goCtx context.Context, req *types.QueryFrozenBalancesRequest) (*types.QueryFrozenBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid account address")
	}
	balances, pageRes, err := qs.keeper.GetFrozenBalances(ctx, account, req.Pagination)
	if err != nil {
		return nil, err
	}

	return &types.QueryFrozenBalancesResponse{
		Balances:   balances,
		Pagination: pageRes,
	}, nil
}

// FrozenBalance lists frozen balance of a denom on a given account
func (qs QueryService) FrozenBalance(goCtx context.Context, req *types.QueryFrozenBalanceRequest) (*types.QueryFrozenBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid account address")
	}
	balance := qs.keeper.GetFrozenBalance(ctx, account, req.GetDenom())

	return &types.QueryFrozenBalanceResponse{
		Balance: balance,
	}, nil
}

// WhitelistedBalances lists whitelisted balances on a given account
func (qs QueryService) WhitelistedBalances(goCtx context.Context, req *types.QueryWhitelistedBalancesRequest) (*types.QueryWhitelistedBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid account address")
	}
	balances, pageRes, err := qs.keeper.GetWhitelistedBalances(ctx, account, req.Pagination)
	if err != nil {
		return nil, err
	}

	return &types.QueryWhitelistedBalancesResponse{
		Balances:   balances,
		Pagination: pageRes,
	}, nil
}

// WhitelistedBalance lists whitelisted balance of a denom on a given account
func (qs QueryService) WhitelistedBalance(goCtx context.Context, req *types.QueryWhitelistedBalanceRequest) (*types.QueryWhitelistedBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid account address")
	}
	balance := qs.keeper.GetWhitelistedBalance(ctx, account, req.GetDenom())

	return &types.QueryWhitelistedBalanceResponse{
		Balance: balance,
	}, nil
}
