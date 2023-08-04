package types

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/samber/lo"
)

// Type of messages for amino.
const (
	TypeMsgIssueClass          = "issue-class"
	TypeMsgMint                = "mint"
	TypeMsgBurn                = "burn"
	TypeMsgFreeze              = "freeze"
	TypeMsgUnfreeze            = "unfreeze"
	TypeMsgAddToWhitelist      = "whitelist"
	TypeMsgRemoveFromWhitelist = "remove-from-whitelist"
)

var (
	_ sdk.Msg            = &MsgIssueClass{}
	_ legacytx.LegacyMsg = &MsgIssueClass{}
	_ sdk.Msg            = &MsgMint{}
	_ legacytx.LegacyMsg = &MsgMint{}
	_ sdk.Msg            = &MsgBurn{}
	_ legacytx.LegacyMsg = &MsgBurn{}
	_ sdk.Msg            = &MsgFreeze{}
	_ legacytx.LegacyMsg = &MsgFreeze{}
	_ sdk.Msg            = &MsgUnfreeze{}
	_ legacytx.LegacyMsg = &MsgUnfreeze{}
	_ sdk.Msg            = &MsgAddToWhitelist{}
	_ legacytx.LegacyMsg = &MsgAddToWhitelist{}
	_ sdk.Msg            = &MsgRemoveFromWhitelist{}
	_ legacytx.LegacyMsg = &MsgRemoveFromWhitelist{}
)

// Constraints.
const (
	ClassMaxNameLength        = 128
	ClassMaxDescriptionLength = 256
	MaxURILength              = 256
	MaxURIHashLength          = 128
	MaxDataSize               = 5 * 1024 // 5KB
)

// RegisterLegacyAminoCodec registers the amino types and interfaces.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgIssueClass{}, fmt.Sprintf("%s/MsgIssueClass", ModuleName), nil)
	cdc.RegisterConcrete(&MsgMint{}, fmt.Sprintf("%s/MsgMint", ModuleName), nil)
	cdc.RegisterConcrete(&MsgBurn{}, fmt.Sprintf("%s/MsgBurn", ModuleName), nil)
	cdc.RegisterConcrete(&MsgFreeze{}, fmt.Sprintf("%s/MsgFreeze", ModuleName), nil)
	cdc.RegisterConcrete(&MsgUnfreeze{}, fmt.Sprintf("%s/MsgUnfreeze", ModuleName), nil)
	cdc.RegisterConcrete(&MsgAddToWhitelist{}, fmt.Sprintf("%s/MsgAddToWhitelist", ModuleName), nil)
	cdc.RegisterConcrete(&MsgRemoveFromWhitelist{}, fmt.Sprintf("%s/MsgRemoveFromWhitelist", ModuleName), nil)
}

// ValidateBasic checks that message fields are valid.
func (m *MsgIssueClass) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Issuer); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid issuer account %s", m.Issuer)
	}

	if len(m.Name) > ClassMaxNameLength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid name %q, the length must be less than or equal %d", m.Name, ClassMaxNameLength)
	}

	if err := ValidateClassSymbol(m.Symbol); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if err := ValidateData(m.Data); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if len(m.Description) > ClassMaxDescriptionLength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid description %q, the length must be less than or equal %d", m.Description, ClassMaxDescriptionLength)
	}

	if len(m.URI) > MaxURILength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid URI %q, the length must be less than or equal %d", len(m.URI), MaxURILength)
	}

	if err := ValidateRoyaltyRate(m.RoyaltyRate); err != nil {
		return err
	}

	if len(m.URIHash) > MaxURIHashLength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid URI hash %q, the length must be less than or equal %d", len(m.URIHash), MaxURIHashLength)
	}

	duplicates := lo.FindDuplicates(m.Features)
	if len(duplicates) != 0 {
		return sdkerrors.Wrapf(ErrInvalidInput, "duplicated features in the class features list, duplicates: %v", duplicates)
	}

	return nil
}

// GetSigners returns the required signers of this message type.
func (m *MsgIssueClass) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Issuer),
	}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgIssueClass) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgIssueClass) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgIssueClass) Type() string {
	return TypeMsgIssueClass
}

// ValidateBasic checks that message fields are valid.
func (m *MsgMint) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid sender account %s", m.Sender)
	}

	if err := ValidateTokenID(m.ID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if err := ValidateData(m.Data); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if _, _, err := DeconstructClassID(m.ClassID); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, err.Error())
	}

	if len(m.URI) > MaxURILength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid URI %q, the length must be less than or equal %d", len(m.URI), MaxURILength)
	}

	if len(m.URIHash) > MaxURIHashLength {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid URI hash %q, the length must be less than or equal %d", len(m.URIHash), MaxURIHashLength)
	}

	return nil
}

// GetSigners returns the required signers of this message type.
func (m *MsgMint) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Sender),
	}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgMint) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgMint) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgMint) Type() string {
	return TypeMsgMint
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

// GetSigners returns the required signers of this message type.
func (m *MsgBurn) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Sender),
	}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgBurn) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgBurn) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgBurn) Type() string {
	return TypeMsgBurn
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

// GetSigners returns the required signers of this message type.
func (m *MsgFreeze) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Sender),
	}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgFreeze) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgFreeze) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgFreeze) Type() string {
	return TypeMsgFreeze
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

// GetSigners returns the required signers of this message type.
func (m *MsgUnfreeze) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Sender),
	}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgUnfreeze) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgUnfreeze) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgUnfreeze) Type() string {
	return TypeMsgUnfreeze
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

// GetSigners returns the required signers of this message type.
func (m *MsgAddToWhitelist) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Sender),
	}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgAddToWhitelist) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgAddToWhitelist) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgAddToWhitelist) Type() string {
	return TypeMsgAddToWhitelist
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

// GetSigners returns the required signers of this message type.
func (m *MsgRemoveFromWhitelist) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Sender),
	}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgRemoveFromWhitelist) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgRemoveFromWhitelist) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgRemoveFromWhitelist) Type() string {
	return TypeMsgRemoveFromWhitelist
}

var (
	amino          = codec.NewLegacyAmino()
	moduleAminoCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
