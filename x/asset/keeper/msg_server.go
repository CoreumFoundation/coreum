package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// MsgKeeper defines subscope of keeper methods required by msg service.
type MsgKeeper interface {
	IssueFungibleToken(ctx sdk.Context, settings types.IssueFungibleTokenSettings) (string, error)
}

// MsgServer serves grpc tx requests for assets module.
type MsgServer struct {
	keeper MsgKeeper
}

// NewMsgServer returns a new instance of the MsgServer.
func NewMsgServer(keeper MsgKeeper) MsgServer {
	return MsgServer{
		keeper: keeper,
	}
}

// IssueFungibleToken defines a tx handler to issue a new fungible token.
func (ms MsgServer) IssueFungibleToken(ctx context.Context, req *types.MsgIssueFungibleToken) (*types.MsgIssueFungibleTokenResponse, error) {
	issuer, err := sdk.AccAddressFromBech32(req.Issuer)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidFungibleToken, "invalid issuer in MsgIssueFungibleToken")
	}
	recipient, err := sdk.AccAddressFromBech32(req.Recipient)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidFungibleToken, "invalid recipient in MsgIssueFungibleToken")
	}
	_, err = ms.keeper.IssueFungibleToken(sdk.UnwrapSDKContext(ctx), types.IssueFungibleTokenSettings{
		Issuer:        issuer,
		Symbol:        req.Symbol,
		Description:   req.Description,
		Recipient:     recipient,
		InitialAmount: req.InitialAmount,
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgIssueFungibleTokenResponse{}, nil
}
