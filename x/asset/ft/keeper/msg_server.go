package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

var _ types.MsgServer = MsgServer{}

// MsgKeeper defines subscope of keeper methods required by msg service.
//
//nolint:interfacebloat // We accept the fact that this interface declares more than 10 methods.
type MsgKeeper interface {
	Issue(ctx sdk.Context, settings types.IssueSettings) (string, error)
	Mint(ctx sdk.Context, sender, recipient sdk.AccAddress, coin sdk.Coin) error
	Burn(ctx sdk.Context, sender sdk.AccAddress, coin sdk.Coin) error
	Freeze(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error
	Unfreeze(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error
	SetFrozen(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error
	GloballyFreeze(ctx sdk.Context, sender sdk.AccAddress, denom string) error
	GloballyUnfreeze(ctx sdk.Context, sender sdk.AccAddress, denom string) error
	Clawback(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error
	SetWhitelistedBalance(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error
	AddDelayedTokenUpgradeV1(ctx sdk.Context, sender sdk.AccAddress, denom string, ibcEnabled bool) error
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
		URI:                req.URI,
		URIHash:            req.URIHash,
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
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	recipient := sender
	if req.Recipient != "" {
		recipient, err = sdk.AccAddressFromBech32(req.Recipient)
		if err != nil {
			return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid recipient address")
		}
	}

	err = ms.keeper.Mint(ctx, sender, recipient, req.Coin)
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
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
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
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
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
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
	}

	err = ms.keeper.Unfreeze(ctx, sender, account, req.Coin)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// SetFrozen sets the frozen amount on an account.
func (ms MsgServer) SetFrozen(goCtx context.Context, req *types.MsgSetFrozen) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
	}

	err = ms.keeper.SetFrozen(ctx, sender, account, req.Coin)
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
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	if err := ms.keeper.GloballyFreeze(ctx, sender, req.Denom); err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// GloballyUnfreeze globally unfreezes fungible token.
func (ms MsgServer) GloballyUnfreeze(
	goCtx context.Context,
	req *types.MsgGloballyUnfreeze,
) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	if err := ms.keeper.GloballyUnfreeze(ctx, sender, req.Denom); err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// Clawback confiscates a part of fungible tokens from an account to the issuer.
func (ms MsgServer) Clawback(goCtx context.Context, req *types.MsgClawback) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
	}

	err = ms.keeper.Clawback(ctx, sender, account, req.Coin)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// SetWhitelistedLimit sets the limit of how many tokens account may hold.
func (ms MsgServer) SetWhitelistedLimit(
	goCtx context.Context,
	req *types.MsgSetWhitelistedLimit,
) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	account, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
	}

	err = ms.keeper.SetWhitelistedBalance(ctx, sender, account, req.Coin)
	if err != nil {
		return nil, err
	}

	return &types.EmptyResponse{}, nil
}

// UpgradeTokenV1 stores a request to upgrade token to V1.
func (ms MsgServer) UpgradeTokenV1(goCtx context.Context, req *types.MsgUpgradeTokenV1) (*types.EmptyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	err = ms.keeper.AddDelayedTokenUpgradeV1(ctx, sender, req.Denom, req.IbcEnabled)
	if err != nil {
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
