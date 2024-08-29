package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

var _ types.QueryServer = QueryService{}

// QueryKeeper defines subscope of keeper methods required by query service.
type QueryKeeper interface {
	GetParams(ctx sdk.Context) types.Params
	GetIssuerTokens(
		ctx sdk.Context,
		issuer sdk.AccAddress,
		pagination *query.PageRequest,
	) ([]types.Token, *query.PageResponse, error)
	GetToken(ctx sdk.Context, denom string) (types.Token, error)
	GetTokenUpgradeStatuses(ctx sdk.Context, denom string) types.TokenUpgradeStatuses
	GetFrozenBalances(
		ctx sdk.Context,
		addr sdk.AccAddress,
		pagination *query.PageRequest,
	) (sdk.Coins, *query.PageResponse, error)
	GetFrozenBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetWhitelistedBalances(
		ctx sdk.Context,
		addr sdk.AccAddress,
		pagination *query.PageRequest,
	) (sdk.Coins, *query.PageResponse, error)
	GetWhitelistedBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetDEXLockedBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetDEXSettings(ctx sdk.Context, denom string) (types.DEXSettings, error)
}

// BankKeeper represents required methods of bank keeper.
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	LockedCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
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
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "issuer is required and must be valid account address")
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
func (qs QueryService) TokenUpgradeStatuses(
	ctx context.Context,
	req *types.QueryTokenUpgradeStatusesRequest,
) (*types.QueryTokenUpgradeStatusesResponse, error) {
	tokenUpgradeStatuses := qs.keeper.GetTokenUpgradeStatuses(sdk.UnwrapSDKContext(ctx), req.GetDenom())

	return &types.QueryTokenUpgradeStatusesResponse{
		Statuses: tokenUpgradeStatuses,
	}, nil
}

// Balance returns balance of the denom for the account.
func (qs QueryService) Balance(
	ctx context.Context,
	req *types.QueryBalanceRequest,
) (*types.QueryBalanceResponse, error) {
	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
	}

	denom := req.GetDenom()
	vestingLocked := qs.bankKeeper.LockedCoins(ctx, account).AmountOf(denom)
	dexLocked := qs.keeper.GetDEXLockedBalance(sdk.UnwrapSDKContext(ctx), account, denom).Amount

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &types.QueryBalanceResponse{
		Balance:         qs.bankKeeper.GetBalance(ctx, account, denom).Amount,
		Whitelisted:     qs.keeper.GetWhitelistedBalance(sdkCtx, account, denom).Amount,
		Frozen:          qs.keeper.GetFrozenBalance(sdkCtx, account, denom).Amount,
		Locked:          vestingLocked.Add(dexLocked),
		LockedInVesting: vestingLocked,
		LockedInDEX:     dexLocked,
	}, nil
}

// FrozenBalances lists frozen balances on a given account.
func (qs QueryService) FrozenBalances(
	goCtx context.Context,
	req *types.QueryFrozenBalancesRequest,
) (*types.QueryFrozenBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
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
func (qs QueryService) FrozenBalance(
	goCtx context.Context,
	req *types.QueryFrozenBalanceRequest,
) (*types.QueryFrozenBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
	}
	balance := qs.keeper.GetFrozenBalance(ctx, account, req.GetDenom())

	return &types.QueryFrozenBalanceResponse{
		Balance: balance,
	}, nil
}

// WhitelistedBalances lists whitelisted balances on a given account.
func (qs QueryService) WhitelistedBalances(
	goCtx context.Context,
	req *types.QueryWhitelistedBalancesRequest,
) (*types.QueryWhitelistedBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
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
func (qs QueryService) WhitelistedBalance(
	goCtx context.Context,
	req *types.QueryWhitelistedBalanceRequest,
) (*types.QueryWhitelistedBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
	}
	balance := qs.keeper.GetWhitelistedBalance(ctx, account, req.GetDenom())

	return &types.QueryWhitelistedBalanceResponse{
		Balance: balance,
	}, nil
}

// DEXSettings returns DEX settings.
func (qs QueryService) DEXSettings(
	goCtx context.Context,
	req *types.QueryDEXSettingsRequest,
) (*types.QueryDEXSettingsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	settings, err := qs.keeper.GetDEXSettings(ctx, req.Denom)
	if err != nil {
		return nil, err
	}

	return &types.QueryDEXSettingsResponse{
		DEXSettings: settings,
	}, nil
}
