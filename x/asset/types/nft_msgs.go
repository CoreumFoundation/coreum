package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgCreateNonFungibleTokenClass{}
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
func (msg *MsgCreateNonFungibleTokenClass) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator account %s", msg.Creator)
	}

	if len(msg.Name) > nftClassMaxNameLength {
		return sdkerrors.Wrapf(ErrInvalidNonFungibleTokenClass, "invalid name %q, the length must be less than %d", msg.Name, nftClassMaxNameLength)
	}

	if err := ValidateNonFungibleTokenClassSymbol(msg.Symbol); err != nil {
		return sdkerrors.Wrap(ErrInvalidNonFungibleTokenClass, err.Error())
	}

	if len(msg.Description) > nftClassMaxDescriptionLength {
		return sdkerrors.Wrapf(ErrInvalidNonFungibleTokenClass, "invalid description %q, the length must be less than %d", msg.Description, nftClassMaxDescriptionLength)
	}

	if len(msg.Uri) > nftMaxURILength {
		return sdkerrors.Wrapf(ErrInvalidNonFungibleTokenClass, "invalid URI %q, the length must be less than %d", len(msg.Uri), nftMaxURILength)
	}

	if len(msg.UriHash) > nftMaxURIHashLength {
		return sdkerrors.Wrapf(ErrInvalidNonFungibleTokenClass, "invalid URI hash %q, the length must be less than %d", len(msg.UriHash), nftMaxURIHashLength)
	}

	if msg.Data != nil && len(msg.Data.Value) > nftMaxDataSize {
		return sdkerrors.Wrapf(ErrInvalidNonFungibleTokenClass, "invalid data, it's allowed to use %d bytes", nftMaxDataSize)
	}

	return nil
}

// GetSigners returns the required signers of this message type.
func (msg *MsgCreateNonFungibleTokenClass) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Creator),
	}
}

// ValidateBasic checks that message fields are valid.
func (msg *MsgMintNonFungibleToken) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender account %s", msg.Sender)
	}

	if err := ValidateNonFungibleTokenID(msg.Id); err != nil {
		return sdkerrors.Wrap(ErrInvalidNonFungibleToken, err.Error())
	}

	if _, err := DeconstructNonFungibleTokenClassID(msg.ClassId); err != nil {
		return sdkerrors.Wrap(ErrInvalidNonFungibleToken, err.Error())
	}

	if len(msg.Uri) > nftMaxURILength {
		return sdkerrors.Wrapf(ErrInvalidNonFungibleToken, "invalid URI %q, the length must be less than %d", len(msg.Uri), nftMaxURILength)
	}

	if len(msg.UriHash) > nftMaxURIHashLength {
		return sdkerrors.Wrapf(ErrInvalidNonFungibleToken, "invalid URI hash %q, the length must be less than %d", len(msg.UriHash), nftMaxURIHashLength)
	}

	if msg.Data != nil && len(msg.Data.Value) > nftMaxDataSize {
		return sdkerrors.Wrapf(ErrInvalidNonFungibleToken, "invalid data, it's allowed to use %d bytes", nftMaxDataSize)
	}

	return nil
}

// GetSigners returns the required signers of this message type.
func (msg *MsgMintNonFungibleToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}
