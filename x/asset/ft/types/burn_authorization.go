//nolint:dupl // this code is identical to the mint part, but they should not be merged.
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

var _ authz.Authorization = &BurnAuthorization{}

// NewBurnAuthorization returns a new BurnAuthorization object.
func NewBurnAuthorization(burnLimit sdk.Coins) *BurnAuthorization {
	return &BurnAuthorization{
		BurnLimit: burnLimit,
	}
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a BurnAuthorization) MsgTypeURL() string {
	return sdk.MsgTypeURL(&MsgBurn{})
}

// Accept implements Authorization.Accept.
func (a BurnAuthorization) Accept(ctx sdk.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
	mBurn, ok := msg.(*MsgBurn)
	if !ok {
		return authz.AcceptResponse{}, sdkerrors.ErrInvalidType.Wrap("type mismatch")
	}

	limitLeft, isNegative := a.BurnLimit.SafeSub(mBurn.Coin)
	if isNegative {
		return authz.AcceptResponse{}, sdkerrors.ErrUnauthorized.Wrapf("requested amount is more than burn limit")
	}

	return authz.AcceptResponse{
		Accept:  true,
		Delete:  limitLeft.IsZero(),
		Updated: &BurnAuthorization{BurnLimit: limitLeft},
	}, nil
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a BurnAuthorization) ValidateBasic() error {
	if !a.BurnLimit.IsAllPositive() {
		return sdkerrors.ErrInvalidCoins.Wrapf("burn limit must be positive")
	}

	return nil
}
