package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

var _ types.MsgServer = MsgServer{}

// MsgKeeper defines subscope of keeper methods required by msg service.
type MsgKeeper interface {
	IssueClass(ctx sdk.Context, settings types.IssueClassSettings) (string, error)
	Mint(ctx sdk.Context, settings types.MintSettings) error
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

// IssueClass issues new non-fungible token class.
func (ms MsgServer) IssueClass(ctx context.Context, req *types.MsgIssueClass) (*types.EmptyResponse, error) {
	issuer, err := sdk.AccAddressFromBech32(req.Issuer)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidInput, "invalid issuer in MsgIssueClass")
	}
	if _, err := ms.keeper.IssueClass(
		sdk.UnwrapSDKContext(ctx),
		types.IssueClassSettings{
			Issuer:      issuer,
			Name:        req.Name,
			Symbol:      req.Symbol,
			Description: req.Description,
			URI:         req.URI,
			URIHash:     req.URIHash,
			Data:        req.Data,
		},
	); err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// Mint mints non-fungible token.
func (ms MsgServer) Mint(ctx context.Context, req *types.MsgMint) (*types.EmptyResponse, error) {
	owner, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidInput, "invalid sender")
	}
	if err := ms.keeper.Mint(
		sdk.UnwrapSDKContext(ctx),
		types.MintSettings{
			Sender:  owner,
			ClassID: req.ClassID,
			ID:      req.ID,
			URI:     req.URI,
			URIHash: req.URIHash,
			Data:    req.Data,
		},
	); err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
