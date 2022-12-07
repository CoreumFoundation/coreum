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
	GetFungibleToken(ctx sdk.Context, denom string) (types.FungibleToken, error)
	FreezeFungibleToken(ctx sdk.Context, sender sdk.AccAddress, addr sdk.AccAddress, coin sdk.Coin) error
	UnfreezeFungibleToken(ctx sdk.Context, sender sdk.AccAddress, addr sdk.AccAddress, coin sdk.Coin) error
	MintFungibleToken(ctx sdk.Context, sender sdk.AccAddress, coin sdk.Coin) error
	BurnFungibleToken(ctx sdk.Context, sender sdk.AccAddress, coin sdk.Coin) error
	GloballyFreezeFungibleToken(ctx sdk.Context, sender sdk.AccAddress, denom string) error
	GloballyUnfreezeFungibleToken(ctx sdk.Context, sender sdk.AccAddress, denom string) error
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
func (ms MsgServer) IssueFungibleToken(ctx context.Context, req *types.MsgIssueFungibleToken) (*types.EmptyResponse, error) {
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
		Features:      req.Features,
	})
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// FreezeFungibleToken freezes coins on an account.
func (ms MsgServer) FreezeFungibleToken(goCtx context.Context, req *types.MsgFreezeFungibleToken) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid account address")
	}

	err = ms.keeper.FreezeFungibleToken(ctx, sender, account, req.Coin)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// UnfreezeFungibleToken unfreezes coins on an account.
func (ms MsgServer) UnfreezeFungibleToken(goCtx context.Context, req *types.MsgUnfreezeFungibleToken) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid account address")
	}

	err = ms.keeper.UnfreezeFungibleToken(ctx, sender, account, req.Coin)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// MintFungibleToken mints new fungible tokens.
func (ms MsgServer) MintFungibleToken(goCtx context.Context, req *types.MsgMintFungibleToken) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	err = ms.keeper.MintFungibleToken(ctx, sender, req.Coin)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// BurnFungibleToken a part of the token
func (ms MsgServer) BurnFungibleToken(goCtx context.Context, req *types.MsgBurnFungibleToken) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	err = ms.keeper.BurnFungibleToken(ctx, sender, req.Coin)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// GloballyFreezeFungibleToken globally freezes fungible token
func (ms MsgServer) GloballyFreezeFungibleToken(goCtx context.Context, req *types.MsgGloballyFreezeFungibleToken) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	if err := ms.keeper.GloballyFreezeFungibleToken(ctx, sender, req.Denom); err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// GloballyUnfreezeFungibleToken globally unfreezes fungible token
func (ms MsgServer) GloballyUnfreezeFungibleToken(goCtx context.Context, req *types.MsgGloballyUnfreezeFungibleToken) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	if err := ms.keeper.GloballyUnfreezeFungibleToken(ctx, sender, req.Denom); err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
