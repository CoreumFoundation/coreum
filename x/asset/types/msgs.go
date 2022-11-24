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
	const maxDescriptionLength = 200

	if _, err := sdk.AccAddressFromBech32(msg.Issuer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid issuer %s", msg.Issuer)
	}

	if err := ValidateSymbol(msg.Symbol); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid recipient %s", msg.Recipient)
	}

	// we allow zero initial amount, in that case we won't mint it initially
	if msg.InitialAmount.IsNil() || msg.InitialAmount.IsNegative() {
		return sdkerrors.Wrapf(ErrInvalidFungibleToken, "invalid initial amount %s, can't be negative", msg.InitialAmount.String())
	}

	if len(msg.Description) > maxDescriptionLength {
		return sdkerrors.Wrapf(ErrInvalidFungibleToken, "invalid description %q, the length must less than %d", msg.Description, maxDescriptionLength)
	}

	return nil
}

// GetSigners MsgIssueFungibleToken the message signers.
func (msg MsgIssueFungibleToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Issuer),
	}
}

// ValidateBasic checks that message fields are valid
func (msg MsgFreezeFungibleToken) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Account); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid account address")
	}

	if _, _, err := DeconstructFungibleTokenDenom(msg.Coin.Denom); err != nil {
		return err
	}

	return msg.Coin.Validate()
}

// GetSigners returns the required signers of this message type
func (msg MsgFreezeFungibleToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}

// ValidateBasic checks that message fields are valid
func (msg MsgUnfreezeFungibleToken) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Account); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid account address")
	}

	if _, _, err := DeconstructFungibleTokenDenom(msg.Coin.Denom); err != nil {
		return err
	}

	return msg.Coin.Validate()
}

// GetSigners returns the required signers of this message type
func (msg MsgUnfreezeFungibleToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}

// ValidateBasic checks that message fields are valid
func (msg MsgMintFungibleToken) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, _, err := DeconstructFungibleTokenDenom(msg.Coin.Denom); err != nil {
		return err
	}

	return msg.Coin.Validate()
}

// GetSigners returns the required signers of this message type
func (msg MsgMintFungibleToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}

// ValidateBasic checks that message fields are valid
func (msg MsgBurnFungibleToken) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, _, err := DeconstructFungibleTokenDenom(msg.Coin.Denom); err != nil {
		return err
	}

	return msg.Coin.Validate()
}

// GetSigners returns the required signers of this message type
func (msg MsgBurnFungibleToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}
