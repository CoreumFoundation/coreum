package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

var _ types.QueryServer = QueryService{}

// QueryKeeper defines subscope of keeper methods required by query service.
type QueryKeeper interface {
	GetParams(ctx sdk.Context) types.Params
	GetIssuerTokens(ctx sdk.Context, issuer sdk.AccAddress, pagination *query.PageRequest) ([]types.Token, *query.PageResponse, error)
	GetToken(ctx sdk.Context, denom string) (types.Token, error)
	GetTokenUpgradeStatuses(ctx sdk.Context, denom string) (types.TokenUpgradeStatuses, error)
	GetFrozenBalances(ctx sdk.Context, addr sdk.AccAddress, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error)
	GetFrozenBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetWhitelistedBalances(ctx sdk.Context, addr sdk.AccAddress, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error)
	GetWhitelistedBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// BankKeeper represents required methods of bank keeper.
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	LockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// QueryService serves grpc query requests for assets module.
type QueryService struct {
	keeper     QueryKeeper
	bankKeeper BankKeeper
}

// NewQueryService initiates the new instance of query service.
func NewQueryService(keeper QueryKeeper, bankKeeper BankKeeper) QueryService {
	return QueryService{
		keeper:     keeper,
		bankKeeper: bankKeeper,
	}
}

// Params queries the parameters of x/asset/ft module.
func (qs QueryService) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	return &types.QueryParamsResponse{
		Params: qs.keeper.GetParams(sdk.UnwrapSDKContext(ctx)),
	}, nil
}

// Tokens returns fungible tokens query result.
func (qs QueryService) Tokens(ctx context.Context, req *types.QueryTokensRequest) (*types.QueryTokensResponse, error) {
	issuer, err := sdk.AccAddressFromBech32(req.Issuer)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "issuer is required and must be valid account address")
	}
	tokens, pageRes, err := qs.keeper.GetIssuerTokens(sdk.UnwrapSDKContext(ctx), issuer, req.Pagination)
	if err != nil {
		return nil, err
	}

	return &types.QueryTokensResponse{
		Pagination: pageRes,
		Tokens:     tokens,
	}, nil
}

// Token queries an fungible token.
func (qs QueryService) Token(ctx context.Context, req *types.QueryTokenRequest) (*types.QueryTokenResponse, error) {
	token, err := qs.keeper.GetToken(sdk.UnwrapSDKContext(ctx), req.GetDenom())
	if err != nil {
		return nil, err
	}

	return &types.QueryTokenResponse{
		Token: token,
	}, nil
}

// TokenUpgradeStatuses returns the token upgrade statuses of a specified denom.
func (qs QueryService) TokenUpgradeStatuses(ctx context.Context, req *types.QueryTokenUpgradeStatusesRequest) (*types.QueryTokenUpgradeStatusesResponse, error) {
	tokenUpgradeStatuses, err := qs.keeper.GetTokenUpgradeStatuses(sdk.UnwrapSDKContext(ctx), req.GetDenom())
	if err != nil {
		return nil, err
	}

	return &types.QueryTokenUpgradeStatusesResponse{
		Statuses: tokenUpgradeStatuses,
	}, nil
}

// Balance returns balance of the denom for the account.
func (qs QueryService) Balance(goCtx context.Context, req *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid account address")
	}

	locked := qs.bankKeeper.LockedCoins(ctx, account)

	denom := req.GetDenom()
	return &types.QueryBalanceResponse{
		Balance:     qs.bankKeeper.GetBalance(ctx, account, denom).Amount,
		Whitelisted: qs.keeper.GetWhitelistedBalance(ctx, account, denom).Amount,
		Frozen:      qs.keeper.GetFrozenBalance(ctx, account, denom).Amount,
		Locked:      locked.AmountOf(denom),
	}, nil
}

// FrozenBalances lists frozen balances on a given account.
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

// FrozenBalance lists frozen balance of a denom on a given account.
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

// WhitelistedBalances lists whitelisted balances on a given account.
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

// WhitelistedBalance lists whitelisted balance of a denom on a given account.
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
