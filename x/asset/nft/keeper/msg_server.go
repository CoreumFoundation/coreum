package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

var _ types.MsgServer = MsgServer{}

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
	SetWhitelistedBalance(ctx sdk.Context, sender sdk.AccAddress, addr sdk.AccAddress, coin sdk.Coin) error
}

// NonFungibleTokeMsgKeeper defines subscope of non-fungible toke keeper methods required by msg service.
type NonFungibleTokeMsgKeeper interface {
	IssueClass(ctx sdk.Context, settings types.IssueNonFungibleTokenClassSettings) (string, error)
	Mint(ctx sdk.Context, settings types.MintNonFungibleTokenSettings) error
}

// MsgServer serves grpc tx requests for assets module.
type MsgServer struct {
	keeper    MsgKeeper
	nftKeeper NonFungibleTokeMsgKeeper
}

// NewMsgServer returns a new instance of the MsgServer.
func NewMsgServer(keeper MsgKeeper, nftKeeper NonFungibleTokeMsgKeeper) MsgServer {
	return MsgServer{
		keeper:    keeper,
		nftKeeper: nftKeeper,
	}
}

// IssueFungibleToken defines a tx handler to issue a new fungible token.
func (ms MsgServer) IssueFungibleToken(ctx context.Context, req *types.MsgIssueFungibleToken) (*types.EmptyResponse, error) {
	issuer, err := sdk.AccAddressFromBech32(req.Issuer)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidInput, "invalid issuer in MsgIssueFungibleToken")
	}
	_, err = ms.keeper.IssueFungibleToken(sdk.UnwrapSDKContext(ctx), types.IssueFungibleTokenSettings{
		Issuer:        issuer,
		Symbol:        req.Symbol,
		Subunit:       req.Subunit,
		Precision:     req.Precision,
		Description:   req.Description,
		InitialAmount: req.InitialAmount,
		Features:      req.Features,
		BurnRate:      req.BurnRate,
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

// SetWhitelistedLimitFungibleToken sets the limit of how many tokens account may hold
func (ms MsgServer) SetWhitelistedLimitFungibleToken(goCtx context.Context, req *types.MsgSetWhitelistedLimitFungibleToken) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid account address")
	}

	err = ms.keeper.SetWhitelistedBalance(ctx, sender, account, req.Coin)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// IssueNonFungibleTokenClass issues new non-fungible token class.
func (ms MsgServer) IssueNonFungibleTokenClass(ctx context.Context, req *types.MsgIssueNonFungibleTokenClass) (*types.EmptyResponse, error) {
	issuer, err := sdk.AccAddressFromBech32(req.Issuer)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidInput, "invalid issuer in MsgIssueNonFungibleTokenClass")
	}
	if _, err := ms.nftKeeper.IssueClass(
		sdk.UnwrapSDKContext(ctx),
		types.IssueNonFungibleTokenClassSettings{
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

// MintNonFungibleToken mints non-fungible token.
func (ms MsgServer) MintNonFungibleToken(ctx context.Context, req *types.MsgMintNonFungibleToken) (*types.EmptyResponse, error) {
	owner, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidInput, "invalid sender")
	}
	if err := ms.nftKeeper.Mint(
		sdk.UnwrapSDKContext(ctx),
		types.MintNonFungibleTokenSettings{
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
