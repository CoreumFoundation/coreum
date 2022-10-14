package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgIssueFungibleToken{}
)

// ValidateBasic validates the message.
func (msg MsgIssueFungibleToken) ValidateBasic() error {
	const maxSymbolLength = 32
	const maxDescriptionLength = 200

	if _, err := sdk.AccAddressFromBech32(msg.Issuer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid issuer %s", msg.Issuer)
	}

	if len(msg.Symbol) == 0 || len(msg.Symbol) > maxSymbolLength {
		return sdkerrors.Wrapf(ErrInvalidFungibleToken, "invalid symbol %s, the length must be greater than 0 and less than %d", msg.Symbol, maxSymbolLength)
	}

	if err := sdk.ValidateDenom(msg.Symbol); err != nil {
		return sdkerrors.Wrapf(ErrInvalidFungibleToken, "invalid symbol %s, the symbol must follow the rule: [a-zA-Z][a-zA-Z0-9/-]", msg.Symbol)
	}

	if len(msg.Description) > maxDescriptionLength {
		return sdkerrors.Wrapf(ErrInvalidFungibleToken, "invalid description %q, the length must less than %d", msg.Description, maxDescriptionLength)
	}

	if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid recipient %s", msg.Recipient)
	}

	if msg.InitialAmount.IsNil() || msg.InitialAmount.IsNegative() {
		return sdkerrors.Wrapf(ErrInvalidFungibleToken, "invalid initial amount %s, can't be negative", msg.InitialAmount.String())
	}

	return nil
}

// GetSigners MsgIssueFungibleToken the message signers.
func (msg MsgIssueFungibleToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Issuer),
	}
}
