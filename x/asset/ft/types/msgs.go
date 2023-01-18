package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgIssue{}
	_ sdk.Msg = &MsgMint{}
	_ sdk.Msg = &MsgBurn{}
	_ sdk.Msg = &MsgFreeze{}
	_ sdk.Msg = &MsgUnfreeze{}
	_ sdk.Msg = &MsgSetWhitelistedLimit{}
)

// ValidateBasic validates the message.
func (msg MsgIssue) ValidateBasic() error {
	const maxDescriptionLength = 200

	if _, err := sdk.AccAddressFromBech32(msg.Issuer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid issuer %s", msg.Issuer)
	}

	if err := ValidateSubunit(msg.Subunit); err != nil {
		return err
	}

	if err := ValidateSymbol(msg.Symbol); err != nil {
		return err
	}

	if err := ValidateBurnRate(msg.BurnRate); err != nil {
		return err
	}

	if err := ValidateSendCommissionRate(msg.SendCommissionRate); err != nil {
		return err
	}

	// we allow zero initial amount, in that case we won't mint it initially
	if msg.InitialAmount.IsNil() || msg.InitialAmount.IsNegative() {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid initial amount %s, can't be negative", msg.InitialAmount.String())
	}

	if len(msg.Description) > maxDescriptionLength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid description %q, the length must be less than %d", msg.Description, maxDescriptionLength)
	}

	return nil
}

// GetSigners returns the message signers.
func (msg MsgIssue) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Issuer),
	}
}

// ValidateBasic checks that message fields are valid
func (msg MsgMint) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, _, err := DeconstructDenom(msg.Coin.Denom); err != nil {
		return err
	}

	return msg.Coin.Validate()
}

// GetSigners returns the required signers of this message type
func (msg MsgMint) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}

// ValidateBasic checks that message fields are valid
func (msg MsgBurn) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, _, err := DeconstructDenom(msg.Coin.Denom); err != nil {
		return err
	}

	return msg.Coin.Validate()
}

// GetSigners returns the required signers of this message type
func (msg MsgBurn) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}

// ValidateBasic checks that message fields are valid
func (msg MsgFreeze) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Account); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid account address")
	}

	if _, _, err := DeconstructDenom(msg.Coin.Denom); err != nil {
		return err
	}

	return msg.Coin.Validate()
}

// GetSigners returns the required signers of this message type
func (msg MsgFreeze) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}

// ValidateBasic checks that message fields are valid
func (msg MsgUnfreeze) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Account); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid account address")
	}

	if _, _, err := DeconstructDenom(msg.Coin.Denom); err != nil {
		return err
	}

	return msg.Coin.Validate()
}

// GetSigners returns the required signers of this message type
func (msg MsgUnfreeze) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}

// ValidateBasic checks that message fields are valid
func (msg MsgGloballyFreeze) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, _, err := DeconstructDenom(msg.Denom); err != nil {
		return err
	}

	return nil
}

// GetSigners returns the required signers of this message type
func (msg MsgGloballyFreeze) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}

// ValidateBasic checks that message fields are valid
func (msg MsgGloballyUnfreeze) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, _, err := DeconstructDenom(msg.Denom); err != nil {
		return err
	}

	return nil
}

// GetSigners returns the required signers of this message type
func (msg MsgGloballyUnfreeze) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}

// ValidateBasic checks that message fields are valid
func (msg MsgSetWhitelistedLimit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Account); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid account address")
	}

	if _, _, err := DeconstructDenom(msg.Coin.Denom); err != nil {
		return err
	}

	return msg.Coin.Validate()
}

// GetSigners returns the required signers of this message type
func (msg MsgSetWhitelistedLimit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}
