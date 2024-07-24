package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

var _ types.QueryServer = QueryService{}

// QueryKeeper defines subscope of keeper methods required by query service.
type QueryKeeper interface {
	GetOrderByAddressAndID(ctx sdk.Context, acc sdk.AccAddress, orderID string) (types.Order, error)
}

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

// Order queries order by account and ID.
func (qs QueryService) Order(ctx context.Context, req *types.QueryOrderRequest) (*types.QueryOrderResponse, error) {
	accAddr, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", req.Account)
	}
	order, err := qs.keeper.GetOrderByAddressAndID(sdk.UnwrapSDKContext(ctx), accAddr, req.Id)
	if err != nil {
		return nil, err
	}

	return &types.QueryOrderResponse{
		Order: order,
	}, nil
}
