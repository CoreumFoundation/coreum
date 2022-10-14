package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// QueryKeeper defines subscope of keeper methods required by query service.
type QueryKeeper interface {
	GetFungibleToken(ctx sdk.Context, denom string) (types.FungibleToken, error)
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
