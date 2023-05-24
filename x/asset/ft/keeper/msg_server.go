package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

var _ types.MsgServer = MsgServer{}

// MsgKeeper defines subscope of keeper methods required by msg service.
type MsgKeeper interface {
	Issue(ctx sdk.Context, settings types.IssueSettings) (string, error)
	Mint(ctx sdk.Context, sender sdk.AccAddress, coin sdk.Coin) error
	Burn(ctx sdk.Context, sender sdk.AccAddress, coin sdk.Coin) error
	Freeze(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error
	Unfreeze(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error
	GloballyFreeze(ctx sdk.Context, sender sdk.AccAddress, denom string) error
	GloballyUnfreeze(ctx sdk.Context, sender sdk.AccAddress, denom string) error
	SetWhitelistedBalance(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error
}

// UpgradeV3Keeper defines method required from keeper managing v3 upgrade.
type UpgradeV3Keeper interface {
	StoreEnableIBCRequest(ctx sdk.Context, sender sdk.AccAddress, denom string) error
	EnableIBC(ctx sdk.Context, denom string) error
}

// MsgServer serves grpc tx requests for assets module.
type MsgServer struct {
	keeper          MsgKeeper
	upgradeV3Keeper UpgradeV3Keeper
}

// NewMsgServer returns a new instance of the MsgServer.
func NewMsgServer(keeper MsgKeeper, upgradeV3Keeper UpgradeV3Keeper) MsgServer {
	return MsgServer{
		keeper:          keeper,
		upgradeV3Keeper: upgradeV3Keeper,
	}
}

// Issue defines a tx handler to issue a new fungible token.
func (ms MsgServer) Issue(ctx context.Context, req *types.MsgIssue) (*types.EmptyResponse, error) {
	issuer, err := sdk.AccAddressFromBech32(req.Issuer)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidInput, "invalid issuer in MsgIssue")
	}
	_, err = ms.keeper.Issue(sdk.UnwrapSDKContext(ctx), types.IssueSettings{
		Issuer:             issuer,
		Symbol:             req.Symbol,
		Subunit:            req.Subunit,
		Precision:          req.Precision,
		Description:        req.Description,
		InitialAmount:      req.InitialAmount,
		Features:           req.Features,
		BurnRate:           req.BurnRate,
		SendCommissionRate: req.SendCommissionRate,
	})
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// Mint mints new fungible tokens.
func (ms MsgServer) Mint(goCtx context.Context, req *types.MsgMint) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	err = ms.keeper.Mint(ctx, sender, req.Coin)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// Burn a part of the token.
func (ms MsgServer) Burn(goCtx context.Context, req *types.MsgBurn) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	err = ms.keeper.Burn(ctx, sender, req.Coin)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// Freeze freezes coins on an account.
func (ms MsgServer) Freeze(goCtx context.Context, req *types.MsgFreeze) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid account address")
	}

	err = ms.keeper.Freeze(ctx, sender, account, req.Coin)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// Unfreeze unfreezes coins on an account.
func (ms MsgServer) Unfreeze(goCtx context.Context, req *types.MsgUnfreeze) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid account address")
	}

	err = ms.keeper.Unfreeze(ctx, sender, account, req.Coin)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// GloballyFreeze globally freezes fungible token.
func (ms MsgServer) GloballyFreeze(goCtx context.Context, req *types.MsgGloballyFreeze) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	if err := ms.keeper.GloballyFreeze(ctx, sender, req.Denom); err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// GloballyUnfreeze globally unfreezes fungible token.
func (ms MsgServer) GloballyUnfreeze(goCtx context.Context, req *types.MsgGloballyUnfreeze) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	if err := ms.keeper.GloballyUnfreeze(ctx, sender, req.Denom); err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// SetWhitelistedLimit sets the limit of how many tokens account may hold.
func (ms MsgServer) SetWhitelistedLimit(goCtx context.Context, req *types.MsgSetWhitelistedLimit) (*types.EmptyResponse, error) {
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

// StoreEnableIBCRequest stores a request to enable IBC transfer.
func (ms MsgServer) StoreEnableIBCRequest(goCtx context.Context, req *types.MsgEnableIBCRequest) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	err = ms.upgradeV3Keeper.StoreEnableIBCRequest(ctx, sender, req.Denom)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// EnableIBC enables IBC for the token.
func (ms MsgServer) EnableIBC(goCtx context.Context, req *types.MsgEnableIBCExecutor) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := ms.upgradeV3Keeper.EnableIBC(ctx, req.Denom)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}
