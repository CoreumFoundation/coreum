package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v3/x/asset/nft/types"
)

var _ types.MsgServer = MsgServer{}

// MsgKeeper defines subscope of keeper methods required by msg service.
type MsgKeeper interface {
	IssueClass(ctx sdk.Context, settings types.IssueClassSettings) (string, error)
	Mint(ctx sdk.Context, settings types.MintSettings) error
	Burn(ctx sdk.Context, owner sdk.AccAddress, classID, ID string) error
	Freeze(ctx sdk.Context, sender sdk.AccAddress, classID, nftID string) error
	Unfreeze(ctx sdk.Context, sender sdk.AccAddress, classID, nftID string) error
	AddToWhitelist(ctx sdk.Context, classID, nftID string, sender, account sdk.AccAddress) error
	RemoveFromWhitelist(ctx sdk.Context, classID, nftID string, sender, account sdk.AccAddress) error
	UpdateParams(ctx sdk.Context, authority string, params types.Params) error
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
			Features:    req.Features,
			RoyaltyRate: req.RoyaltyRate,
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

// Burn burns the non-fungible token.
func (ms MsgServer) Burn(ctx context.Context, req *types.MsgBurn) (*types.EmptyResponse, error) {
	owner, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidInput, "invalid sender")
	}

	if err := ms.keeper.Burn(
		sdk.UnwrapSDKContext(ctx),
		owner,
		req.ClassID,
		req.ID,
	); err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// Freeze freeze the non-fungible token.
func (ms MsgServer) Freeze(ctx context.Context, req *types.MsgFreeze) (*types.EmptyResponse, error) {
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidInput, "invalid sender")
	}

	err = ms.keeper.Freeze(sdk.UnwrapSDKContext(ctx), sender, req.ClassID, req.ID)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// Unfreeze unfreezes the non-fungible token.
func (ms MsgServer) Unfreeze(ctx context.Context, req *types.MsgUnfreeze) (*types.EmptyResponse, error) {
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidInput, "invalid sender")
	}

	err = ms.keeper.Unfreeze(sdk.UnwrapSDKContext(ctx), sender, req.ClassID, req.ID)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// AddToWhitelist adds an account to the whitelisted list of accounts for the NFT.
func (ms MsgServer) AddToWhitelist(ctx context.Context, req *types.MsgAddToWhitelist) (*types.EmptyResponse, error) {
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidInput, "invalid sender")
	}

	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidInput, "invalid account")
	}

	if err := ms.keeper.AddToWhitelist(sdk.UnwrapSDKContext(ctx), req.ClassID, req.ID, sender, account); err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// RemoveFromWhitelist removes an account from the whitelisted list of accounts for the NFT.
func (ms MsgServer) RemoveFromWhitelist(ctx context.Context, req *types.MsgRemoveFromWhitelist) (*types.EmptyResponse, error) {
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidInput, "invalid sender")
	}

	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidInput, "invalid account")
	}

	if err := ms.keeper.RemoveFromWhitelist(sdk.UnwrapSDKContext(ctx), req.ClassID, req.ID, sender, account); err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// UpdateParams is a governance operation that sets parameters of the module.
func (ms MsgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.EmptyResponse, error) {
	if err := ms.keeper.UpdateParams(sdk.UnwrapSDKContext(goCtx), req.Authority, req.Params); err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
