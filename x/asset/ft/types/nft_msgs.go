package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgIssueNonFungibleTokenClass{}
	_ sdk.Msg = &MsgMintNonFungibleToken{}
)

const (
	nftClassMaxNameLength        = 128
	nftClassMaxDescriptionLength = 256
	nftMaxURILength              = 256
	nftMaxURIHashLength          = 128
	nftMaxDataSize               = 5 * 1000 // 5kb
)

// ValidateBasic checks that message fields are valid.
func (msg *MsgIssueNonFungibleTokenClass) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Issuer); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid issuer account %s", msg.Issuer)
	}

	if len(msg.Name) > nftClassMaxNameLength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid name %q, the length must be less than or equal %d", msg.Name, nftClassMaxNameLength)
	}

	if err := ValidateNonFungibleTokenClassSymbol(msg.Symbol); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if len(msg.Description) > nftClassMaxDescriptionLength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid description %q, the length must be less than or equal %d", msg.Description, nftClassMaxDescriptionLength)
	}

	if len(msg.URI) > nftMaxURILength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid URI %q, the length must be less than or equal %d", len(msg.URI), nftMaxURILength)
	}

	if len(msg.URIHash) > nftMaxURIHashLength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid URI hash %q, the length must be less than or equal %d", len(msg.URIHash), nftMaxURIHashLength)
	}

	if msg.Data != nil && len(msg.Data.Value) > nftMaxDataSize {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid data, it's allowed to use %d bytes", nftMaxDataSize)
	}

	return nil
}

// GetSigners returns the required signers of this message type.
func (msg *MsgIssueNonFungibleTokenClass) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Issuer),
	}
}

// ValidateBasic checks that message fields are valid.
func (msg *MsgMintNonFungibleToken) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender account %s", msg.Sender)
	}

	if err := ValidateNonFungibleTokenID(msg.ID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if _, err := DeconstructNonFungibleTokenClassID(msg.ClassID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if len(msg.URI) > nftMaxURILength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid URI %q, the length must be less than or equal %d", len(msg.URI), nftMaxURILength)
	}

	if len(msg.URIHash) > nftMaxURIHashLength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid URI hash %q, the length must be less than or equal %d", len(msg.URIHash), nftMaxURIHashLength)
	}

	if msg.Data != nil && len(msg.Data.Value) > nftMaxDataSize {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid data, it's allowed to use %d bytes", nftMaxDataSize)
	}

	return nil
}

// GetSigners returns the required signers of this message type.
func (msg *MsgMintNonFungibleToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}
