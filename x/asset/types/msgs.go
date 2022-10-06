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
	const maxCodeLength = 32
	const maxDescriptionLength = 200
	const maxPrecision = 32

	if _, err := sdk.AccAddressFromBech32(msg.From); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from %s", msg.From)
	}

	if _, err := sdk.AccAddressFromBech32(msg.Definition.Recipient); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid recipient %s", msg.Definition.Recipient)
	}

	if len(msg.Definition.Code) == 0 || len(msg.Definition.Code) > maxCodeLength {
		return sdkerrors.Wrapf(ErrInvalidAsset, "invalid code %s, the length must be greater than 0 and less than %d", msg.Definition.Code, maxCodeLength)
	}

	if len(msg.Definition.Description) > maxDescriptionLength {
		return sdkerrors.Wrapf(ErrInvalidAsset, "invalid description %s, the length must less than %d", msg.Definition.Description, maxDescriptionLength)
	}

	switch msg.Definition.Type {
	case AssetType_FT: //nolint:nosnakecase // protogen
		if msg.Definition.Ft.Precision > maxPrecision {
			return sdkerrors.Wrapf(ErrInvalidAsset, "invalid precision %d, must less than %d", msg.Definition.Ft.Precision, maxPrecision)
		}
		if msg.Definition.Ft.InitialAmount.IsNegative() {
			return sdkerrors.Wrapf(ErrInvalidAsset, "invalid initial amount %s, can't be negative", msg.Definition.Ft.InitialAmount.String())
		}
	case AssetType_NFT: //nolint:nosnakecase // protogen
		return sdkerrors.Wrapf(ErrInvalidAsset, "asset module doesn't support the NFT issuance yet")
	}
	return nil
}

// GetSigners returns the message signers.
func (msg MsgIssueAsset) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.From),
	}
}
