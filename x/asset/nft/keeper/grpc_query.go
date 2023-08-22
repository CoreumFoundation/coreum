package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/v2/x/asset/nft/types"
)

var _ types.QueryServer = QueryService{}

// QueryKeeper defines subscope of keeper methods required by query service.
type QueryKeeper interface {
	GetParams(ctx sdk.Context) types.Params
	GetClass(ctx sdk.Context, classID string) (types.Class, error)
	GetClasses(ctx sdk.Context, issuer *sdk.AccAddress, pagination *query.PageRequest) ([]types.Class, *query.PageResponse, error)
	IsFrozen(ctx sdk.Context, classID, nftID string) (bool, error)
	IsWhitelisted(ctx sdk.Context, classID, nftID string, account sdk.AccAddress) (bool, error)
	GetWhitelistedAccountsForNFT(ctx sdk.Context, classID, nftID string, q *query.PageRequest) ([]string, *query.PageResponse, error)
	GetBurntByClass(ctx sdk.Context, classID string, q *query.PageRequest) (*query.PageResponse, []string, error)
	IsBurnt(ctx sdk.Context, classID, nftID string) (bool, error)
}

// QueryService serves grpc query requests for assetsnft module.
type QueryService struct {
	keeper QueryKeeper
}

// NewQueryService initiates the new instance of query service.
func NewQueryService(keeper QueryKeeper) QueryService {
	return QueryService{
		keeper: keeper,
	}
}

// Params queries the parameters of x/asset/nft module.
func (qs QueryService) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	return &types.QueryParamsResponse{
		Params: qs.keeper.GetParams(sdk.UnwrapSDKContext(ctx)),
	}, nil
}

// Class returns the asset NFT class.
func (qs QueryService) Class(ctx context.Context, req *types.QueryClassRequest) (*types.QueryClassResponse, error) {
	nftClass, err := qs.keeper.GetClass(sdk.UnwrapSDKContext(ctx), req.Id)
	if err != nil {
		return nil, err
	}

	return &types.QueryClassResponse{
		Class: nftClass,
	}, nil
}

// Classes returns the asset NFT classes.
func (qs QueryService) Classes(ctx context.Context, req *types.QueryClassesRequest) (*types.QueryClassesResponse, error) {
	var issuer *sdk.AccAddress

	if req.Issuer != "" {
		issuerAddress, err := sdk.AccAddressFromBech32(req.Issuer)
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrInvalidInput, "invalid issuer account")
		}
		issuer = &issuerAddress
	}

	classes, pageRes, err := qs.keeper.GetClasses(sdk.UnwrapSDKContext(ctx), issuer, req.Pagination)
	return &types.QueryClassesResponse{
		Pagination: pageRes,
		Classes:    classes,
	}, err
}

// Frozen returns whether NFT is frozen or not.
func (qs QueryService) Frozen(ctx context.Context, req *types.QueryFrozenRequest) (*types.QueryFrozenResponse, error) {
	frozen, err := qs.keeper.IsFrozen(sdk.UnwrapSDKContext(ctx), req.ClassId, req.Id)
	return &types.QueryFrozenResponse{
		Frozen: frozen,
	}, err
}

// Whitelisted checks to see if an account is whitelisted for an NFT.
func (qs QueryService) Whitelisted(ctx context.Context, req *types.QueryWhitelistedRequest) (*types.QueryWhitelistedResponse, error) {
	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidInput, "invalid account")
	}
	isWhitelisted, err := qs.keeper.IsWhitelisted(sdk.UnwrapSDKContext(ctx), req.ClassId, req.Id, account)
	if err != nil {
		return nil, err
	}

	return &types.QueryWhitelistedResponse{
		Whitelisted: isWhitelisted,
	}, nil
}

// WhitelistedAccountsForNFT returns the list of accounts which are whitelited to hold this NFT.
func (qs QueryService) WhitelistedAccountsForNFT(ctx context.Context, req *types.QueryWhitelistedAccountsForNFTRequest) (*types.QueryWhitelistedAccountsForNFTResponse, error) {
	accounts, pageRes, err := qs.keeper.GetWhitelistedAccountsForNFT(sdk.UnwrapSDKContext(ctx), req.ClassId, req.Id, req.Pagination)
	return &types.QueryWhitelistedAccountsForNFTResponse{
		Pagination: pageRes,
		Accounts:   accounts,
	}, err
}

// BurntNFT checks if an NFT is burnt or not.
func (qs QueryService) BurntNFT(ctx context.Context, req *types.QueryBurntNFTRequest) (*types.QueryBurntNFTResponse, error) {
	isBurnt, err := qs.keeper.IsBurnt(sdk.UnwrapSDKContext(ctx), req.ClassId, req.NftId)
	if err != nil {
		return nil, err
	}

	return &types.QueryBurntNFTResponse{
		Burnt: isBurnt,
	}, nil
}

// BurntNFTsInClass returns the list of burnt NFTs in a class.
func (qs QueryService) BurntNFTsInClass(ctx context.Context, req *types.QueryBurntNFTsInClassRequest) (*types.QueryBurntNFTsInClassResponse, error) {
	pageRes, list, err := qs.keeper.GetBurntByClass(sdk.UnwrapSDKContext(ctx), req.ClassId, req.Pagination)
	if err != nil {
		return nil, err
	}

	return &types.QueryBurntNFTsInClassResponse{
		Pagination: pageRes,
		NftIds:     list,
	}, nil
}
