package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgIssueAsset{}
)

// ValidateBasic validates the message.
func (msg MsgIssueAsset) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.From); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.From)
	}

	if msg.Name == "" {
		return sdkerrors.Wrap(ErrInvalidAsset, "name is empty")
	}

	return nil
}

// GetSigners returns the message signers.
func (msg MsgIssueAsset) GetSigners() []sdk.AccAddress {
	from, _ := sdk.AccAddressFromBech32(msg.From)
	return []sdk.AccAddress{from}
}
