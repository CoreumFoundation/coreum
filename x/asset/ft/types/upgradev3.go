package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgEnableIBCRequest{}
	_ sdk.Msg = &MsgEnableIBCExecutor{}
)

// ValidateBasic checks that message fields are valid.
func (msg MsgEnableIBCRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	return nil
}

// GetSigners returns the required signers of this message type.
func (msg MsgEnableIBCRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}

// ValidateBasic checks that message fields are valid.
func (msg MsgEnableIBCExecutor) ValidateBasic() error {
	return nil
}

// GetSigners returns the required signers of this message type.
func (msg MsgEnableIBCExecutor) GetSigners() []sdk.AccAddress {
	return nil
}
