package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

var _ types.QueryServer = QueryService{}

// QueryKeeper defines subscope of keeper methods required by query service.
type QueryKeeper interface {
	GetClass(ctx sdk.Context, classID string) (types.Class, error)
	IsFrozen(ctx sdk.Context, classID, nftID string) (bool, error)
	IsWhitelisted(ctx sdk.Context, classID, nftID string, account sdk.AccAddress) (bool, error)
	GetAllWhitelistedAccountsForNFT(ctx sdk.Context, classID, nftID string, q *query.PageRequest) (*query.PageResponse, []string, error)
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

// Class reruns the asset NFT class.
func (q QueryService) Class(ctx context.Context, req *types.QueryClassRequest) (*types.QueryClassResponse, error) {
	nftClass, err := q.keeper.GetClass(sdk.UnwrapSDKContext(ctx), req.Id)
	if err != nil {
		return nil, err
	}

	return &types.QueryClassResponse{
		Class: nftClass,
	}, nil
}

// Frozen reruns whether NFT is frozen or not.
func (q QueryService) Frozen(ctx context.Context, req *types.QueryFrozenRequest) (*types.QueryFrozenResponse, error) {
	frozen, err := q.keeper.IsFrozen(sdk.UnwrapSDKContext(ctx), req.ClassId, req.Id)
	return &types.QueryFrozenResponse{
		Frozen: frozen,
	}, err
}

// Whitelisted checks to see if an account is whitelisted for an NFT.
func (q QueryService) Whitelisted(ctx context.Context, req *types.QueryWhitelistedRequest) (*types.QueryWhitelistedResponse, error) {
	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidInput, "invalid account")
	}
	isWhitelisted, err := q.keeper.IsWhitelisted(sdk.UnwrapSDKContext(ctx), req.ClassId, req.Id, account)
	if err != nil {
		return nil, err
	}

	return &types.QueryWhitelistedResponse{
		Whitelisted: isWhitelisted,
	}, nil
}

// WhitelistedAccountsForNFT returns the list of accounts which are whitelited to hold this NFT.
func (q QueryService) WhitelistedAccountsForNFT(ctx context.Context, req *types.QueryWhitelistedAccountsForNFTRequest) (*types.QueryWhitelistedAccountsForNFTResponse, error) {
	pageRes, accounts, err := q.keeper.GetAllWhitelistedAccountsForNFT(sdk.UnwrapSDKContext(ctx), req.ClassId, req.Id, req.Pagination)
	return &types.QueryWhitelistedAccountsForNFTResponse{
		Pagination: pageRes,
		Accounts:   accounts,
	}, err
}
