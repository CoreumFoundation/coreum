package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

var _ authz.Authorization = &MintAuthorization{}

// NewMintAuthorization returns a new MintAuthorization object.
func NewMintAuthorization(mintLimit sdk.Coin) *MintAuthorization {
	return &MintAuthorization{
		MintLimit: mintLimit,
	}
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a MintAuthorization) MsgTypeURL() string {
	return sdk.MsgTypeURL(&MsgMint{})
}

// Accept implements Authorization.Accept.
func (a MintAuthorization) Accept(ctx sdk.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
	mMint, ok := msg.(*MsgMint)
	if !ok {
		return authz.AcceptResponse{}, sdkerrors.ErrInvalidType.Wrap("type mismatch")
	}

	limitLeft, err := a.MintLimit.SafeSub(mMint.Coin)
	if err != nil {
		return authz.AcceptResponse{}, sdkerrors.ErrUnauthorized.Wrapf("requested amount is more than mint limit")
	}

	return authz.AcceptResponse{
		Accept:  true,
		Delete:  limitLeft.IsZero(),
		Updated: &MintAuthorization{MintLimit: limitLeft},
	}, nil
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a MintAuthorization) ValidateBasic() error {
	if !a.MintLimit.IsPositive() {
		return sdkerrors.ErrInvalidCoins.Wrapf("spend limit must be positive")
	}
	return nil
}
