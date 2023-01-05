package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgIssueClass{}
	_ sdk.Msg = &MsgMint{}
	_ sdk.Msg = &MsgBurn{}
)

// Constraints
const (
	ClassMaxNameLength        = 128
	ClassMaxDescriptionLength = 256
	MaxURILength              = 256
	MaxURIHashLength          = 128
	MaxDataSize               = 5 * 1024 // 5KB
)

// ValidateBasic checks that message fields are valid.
func (msg *MsgIssueClass) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Issuer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid issuer account %s", msg.Issuer)
	}

	if len(msg.Name) > ClassMaxNameLength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid name %q, the length must be less than or equal %d", msg.Name, ClassMaxNameLength)
	}

	if err := ValidateClassSymbol(msg.Symbol); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if err := ValidateData(msg.Data); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if len(msg.Description) > ClassMaxDescriptionLength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid description %q, the length must be less than or equal %d", msg.Description, ClassMaxDescriptionLength)
	}

	if len(msg.URI) > MaxURILength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid URI %q, the length must be less than or equal %d", len(msg.URI), MaxURILength)
	}

	if len(msg.URIHash) > MaxURIHashLength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid URI hash %q, the length must be less than or equal %d", len(msg.URIHash), MaxURIHashLength)
	}

	return nil
}

// GetSigners returns the required signers of this message type.
func (msg *MsgIssueClass) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Issuer),
	}
}

// ValidateBasic checks that message fields are valid.
func (msg *MsgMint) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender account %s", msg.Sender)
	}

	if err := ValidateTokenID(msg.ID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if err := ValidateData(msg.Data); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if _, err := DeconstructClassID(msg.ClassID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if len(msg.URI) > MaxURILength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid URI %q, the length must be less than or equal %d", len(msg.URI), MaxURILength)
	}

	if len(msg.URIHash) > MaxURIHashLength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid URI hash %q, the length must be less than or equal %d", len(msg.URIHash), MaxURIHashLength)
	}

	return nil
}

// GetSigners returns the required signers of this message type.
func (msg *MsgMint) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}

// ValidateBasic checks that message fields are valid.
func (msg *MsgBurn) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender account %s", msg.Sender)
	}

	if err := ValidateTokenID(msg.ID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if _, err := DeconstructClassID(msg.ClassID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	return nil
}

// GetSigners returns the required signers of this message type.
func (msg *MsgBurn) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}
