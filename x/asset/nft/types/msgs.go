package types

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/samber/lo"
)

type extendedMsg interface {
	sdk.Msg
	sdk.HasValidateBasic
}

var (
	_ extendedMsg = &MsgIssueClass{}
	_ extendedMsg = &MsgMint{}
	_ extendedMsg = &MsgUpdateData{}
	_ extendedMsg = &MsgBurn{}
	_ extendedMsg = &MsgFreeze{}
	_ extendedMsg = &MsgUnfreeze{}
	_ extendedMsg = &MsgAddToWhitelist{}
	_ extendedMsg = &MsgRemoveFromWhitelist{}
	_ extendedMsg = &MsgAddToClassWhitelist{}
	_ extendedMsg = &MsgRemoveFromClassWhitelist{}
	_ extendedMsg = &MsgClassFreeze{}
	_ extendedMsg = &MsgClassUnfreeze{}
	_ extendedMsg = &MsgUpdateParams{}
)

// Constraints.
const (
	ClassMaxNameLength        = 128
	ClassMaxDescriptionLength = 256
	MaxURILength              = 256
	MaxURIHashLength          = 128
)

// RegisterLegacyAminoCodec registers the amino types and interfaces.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgIssueClass{}, fmt.Sprintf("%s/MsgIssueClass", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgMint{}, fmt.Sprintf("%s/MsgMint", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgUpdateData{}, fmt.Sprintf("%s/MsgUpdateData", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgBurn{}, fmt.Sprintf("%s/MsgBurn", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgFreeze{}, fmt.Sprintf("%s/MsgFreeze", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgUnfreeze{}, fmt.Sprintf("%s/MsgUnfreeze", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgAddToWhitelist{}, fmt.Sprintf("%s/MsgAddToWhitelist", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgRemoveFromWhitelist{}, fmt.Sprintf("%s/MsgRemoveFromWhitelist", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgAddToClassWhitelist{}, fmt.Sprintf("%s/MsgAddToClassWhitelist", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgRemoveFromClassWhitelist{}, fmt.Sprintf("%s/MsgRemoveFromClassWhitelist", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgClassFreeze{}, fmt.Sprintf("%s/MsgClassFreeze", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgClassUnfreeze{}, fmt.Sprintf("%s/MsgClassUnfreeze", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, fmt.Sprintf("%s/MsgUpdateParams", ModuleName))
}

// ValidateBasic checks that message fields are valid.
func (m *MsgIssueClass) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Issuer); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid issuer account %s", m.Issuer)
	}

	if len(m.Name) > ClassMaxNameLength {
		return sdkerrors.Wrapf(
			ErrInvalidInput,
			"invalid name %q, the length must be less than or equal %d",
			m.Name, ClassMaxNameLength,
		)
	}

	if err := ValidateClassSymbol(m.Symbol); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if err := ValidateClassData(m.Data); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if len(m.Description) > ClassMaxDescriptionLength {
		return sdkerrors.Wrapf(
			ErrInvalidInput,
			"invalid description %q, the length must be less than or equal %d",
			m.Description,
			ClassMaxDescriptionLength,
		)
	}

	if len(m.URI) > MaxURILength {
		return sdkerrors.Wrapf(
			ErrInvalidInput,
			"invalid URI %q, the length must be less than or equal %d",
			len(m.URI),
			MaxURILength,
		)
	}

	if err := ValidateRoyaltyRate(m.RoyaltyRate); err != nil {
		return err
	}

	if len(m.URIHash) > MaxURIHashLength {
		return sdkerrors.Wrapf(
			ErrInvalidInput,
			"invalid URI hash %q, the length must be less than or equal %d",
			len(m.URIHash), MaxURIHashLength,
		)
	}

	duplicates := lo.FindDuplicates(m.Features)
	if len(duplicates) != 0 {
		return sdkerrors.Wrapf(ErrInvalidInput, "duplicated features in the class features list, duplicates: %v", duplicates)
	}

	return nil
}

// ValidateBasic checks that message fields are valid.
func (m *MsgMint) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid sender account %s", m.Sender)
	}

	if err := ValidateTokenID(m.ID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if err := ValidateNFTData(m.Data); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if _, _, err := DeconstructClassID(m.ClassID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if len(m.URI) > MaxURILength {
		return sdkerrors.Wrapf(
			ErrInvalidInput,
			"invalid URI %q, the length must be less than or equal %d",
			len(m.URI), MaxURILength,
		)
	}

	if len(m.URIHash) > MaxURIHashLength {
		return sdkerrors.Wrapf(
			ErrInvalidInput,
			"invalid URI hash %q, the length must be less than or equal %d",
			len(m.URIHash), MaxURIHashLength,
		)
	}

	return nil
}

// ValidateBasic checks that message fields are valid.
func (m *MsgUpdateData) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid sender account %s", m.Sender)
	}
	if err := ValidateTokenID(m.ID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if _, _, err := DeconstructClassID(m.ClassID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if len(m.Items) == 0 {
		return sdkerrors.Wrap(ErrInvalidInput, "nothing to update")
	}

	duplicates := lo.FindDuplicates(lo.Map(m.Items,
		func(item DataDynamicIndexedItem, _ int) uint32 {
			return item.Index
		},
	))
	if len(duplicates) != 0 {
		return sdkerrors.Wrapf(ErrInvalidInput, "duplicated index of DataDynamicIndexedItem, duplicates: %v", duplicates)
	}

	return nil
}

// ValidateBasic checks that message fields are valid.
func (m *MsgBurn) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid sender account %s", m.Sender)
	}

	if err := ValidateTokenID(m.ID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if _, _, err := DeconstructClassID(m.ClassID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	return nil
}

// ValidateBasic checks that message fields are valid.
func (m *MsgFreeze) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid sender account %s", m.Sender)
	}

	if err := ValidateTokenID(m.ID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if _, _, err := DeconstructClassID(m.ClassID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	return nil
}

// ValidateBasic checks that message fields are valid.
func (m *MsgUnfreeze) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid sender account %s", m.Sender)
	}

	if err := ValidateTokenID(m.ID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if _, _, err := DeconstructClassID(m.ClassID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	return nil
}

// ValidateBasic checks that message fields are valid.
func (m *MsgAddToWhitelist) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid sender account %s", m.Sender)
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid account %s", m.Sender)
	}

	if err := ValidateTokenID(m.ID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if _, _, err := DeconstructClassID(m.ClassID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	return nil
}

// ValidateBasic checks that message fields are valid.
func (m *MsgRemoveFromWhitelist) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid sender account %s", m.Sender)
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid account %s", m.Sender)
	}

	if err := ValidateTokenID(m.ID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if _, _, err := DeconstructClassID(m.ClassID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	return nil
}

// ValidateBasic checks that message fields are valid.
func (m *MsgClassFreeze) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid sender account %s", m.Sender)
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid account %s", m.Sender)
	}

	if _, _, err := DeconstructClassID(m.ClassID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	return nil
}

// ValidateBasic checks that message fields are valid.
func (m *MsgClassUnfreeze) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid sender account %s", m.Sender)
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid account %s", m.Sender)
	}

	if _, _, err := DeconstructClassID(m.ClassID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	return nil
}

// ValidateBasic checks that message fields are valid.
func (m MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return cosmoserrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	return m.Params.ValidateBasic()
}

// ValidateBasic checks that message fields are valid.
func (m *MsgAddToClassWhitelist) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid sender account %s", m.Sender)
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid account %s", m.Sender)
	}

	if _, _, err := DeconstructClassID(m.ClassID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	return nil
}

// ValidateBasic checks that message fields are valid.
func (m *MsgRemoveFromClassWhitelist) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid sender account %s", m.Sender)
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid account %s", m.Sender)
	}

	if _, _, err := DeconstructClassID(m.ClassID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	return nil
}
