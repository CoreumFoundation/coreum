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
	TypeMsgIssue               = "issue"
	TypeMsgMint                = "mint"
	TypeMsgBurn                = "burn"
	TypeMsgFreeze              = "freeze"
	TypeMsgUnfreeze            = "unfreeze"
	TypeMsgGloballyFreeze      = "globally-freeze"
	TypeMsgGloballyUnfreeze    = "globally-unfreeze"
	TypeMsgClawback            = "clawback"
	TypeMsgSetWhitelistedLimit = "set-whitelisted-limit"
	TypeMsgUpgradeTokenV1      = "upgrade-token-v1"
	TypeMsgUpdateParams        = "update-params"
)

const (
	// MaxDescriptionLength is max description length.
	MaxDescriptionLength = 200
	// MaxURILength is max URI length.
	MaxURILength = 256
	// MaxURIHashLength is max URIHash length.
	MaxURIHashLength = 128
)

var (
	_ sdk.Msg            = &MsgIssue{}
	_ legacytx.LegacyMsg = &MsgIssue{}
	_ sdk.Msg            = &MsgMint{}
	_ legacytx.LegacyMsg = &MsgMint{}
	_ sdk.Msg            = &MsgBurn{}
	_ legacytx.LegacyMsg = &MsgBurn{}
	_ sdk.Msg            = &MsgFreeze{}
	_ legacytx.LegacyMsg = &MsgFreeze{}
	_ sdk.Msg            = &MsgUnfreeze{}
	_ legacytx.LegacyMsg = &MsgUnfreeze{}
	_ sdk.Msg            = &MsgSetFrozen{}
	_ legacytx.LegacyMsg = &MsgSetFrozen{}
	_ sdk.Msg            = &MsgGloballyFreeze{}
	_ legacytx.LegacyMsg = &MsgGloballyFreeze{}
	_ sdk.Msg            = &MsgGloballyUnfreeze{}
	_ legacytx.LegacyMsg = &MsgGloballyUnfreeze{}
	_ sdk.Msg            = &MsgClawback{}
	_ legacytx.LegacyMsg = &MsgClawback{}
	_ sdk.Msg            = &MsgSetWhitelistedLimit{}
	_ legacytx.LegacyMsg = &MsgSetWhitelistedLimit{}
	_ sdk.Msg            = &MsgUpgradeTokenV1{}
	_ legacytx.LegacyMsg = &MsgUpgradeTokenV1{}
	_ sdk.Msg            = &MsgUpdateParams{}
	_ legacytx.LegacyMsg = &MsgUpdateParams{}
)

// RegisterLegacyAminoCodec registers the amino types and interfaces.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgIssue{}, fmt.Sprintf("%s/MsgIssue", ModuleName), nil)
	cdc.RegisterConcrete(&MsgMint{}, fmt.Sprintf("%s/MsgMint", ModuleName), nil)
	cdc.RegisterConcrete(&MsgBurn{}, fmt.Sprintf("%s/MsgBurn", ModuleName), nil)
	cdc.RegisterConcrete(&MsgFreeze{}, fmt.Sprintf("%s/MsgFreeze", ModuleName), nil)
	cdc.RegisterConcrete(&MsgUnfreeze{}, fmt.Sprintf("%s/MsgUnfreeze", ModuleName), nil)
	cdc.RegisterConcrete(&MsgSetFrozen{}, fmt.Sprintf("%s/MsgSetFrozen", ModuleName), nil)
	cdc.RegisterConcrete(&MsgGloballyFreeze{}, fmt.Sprintf("%s/MsgGloballyFreeze", ModuleName), nil)
	cdc.RegisterConcrete(&MsgGloballyUnfreeze{}, fmt.Sprintf("%s/MsgGloballyUnfreeze", ModuleName), nil)
	cdc.RegisterConcrete(&MsgSetWhitelistedLimit{}, fmt.Sprintf("%s/MsgSetWhitelistedLimit", ModuleName), nil)
	cdc.RegisterConcrete(&MsgUpgradeTokenV1{}, fmt.Sprintf("%s/MsgUpgradeTokenV1", ModuleName), nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, fmt.Sprintf("%s/MsgUpdateParams", ModuleName), nil)
}

// ValidateBasic validates the message.
func (m MsgIssue) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Issuer); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid issuer %s", m.Issuer)
	}

	if err := ValidateSubunit(m.Subunit); err != nil {
		return err
	}

	if err := ValidateSymbol(m.Symbol); err != nil {
		return err
	}

	if err := ValidateBurnRate(m.BurnRate); err != nil {
		return err
	}

	if err := ValidateSendCommissionRate(m.SendCommissionRate); err != nil {
		return err
	}

	if err := ValidatePrecision(m.Precision); err != nil {
		return err
	}

	// we allow zero initial amount, in that case we won't mint it initially
	if m.InitialAmount.IsNil() || m.InitialAmount.IsNegative() {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid initial amount %s, can't be negative", m.InitialAmount.String())
	}

	if len(m.Description) > MaxDescriptionLength {
		return sdkerrors.Wrapf(
			ErrInvalidInput,
			"invalid description %q, the length must be less than %d",
			m.Description,
			MaxDescriptionLength,
		)
	}

	duplicates := lo.FindDuplicates(m.Features)
	if len(duplicates) != 0 {
		return sdkerrors.Wrapf(ErrInvalidInput, "duplicated features in the features list, duplicates: %v", duplicates)
	}

	if len(m.URI) > MaxURILength {
		return sdkerrors.Wrapf(
			ErrInvalidInput,
			"invalid URI %q, the length must be less than or equal %d",
			len(m.URI),
			MaxURILength,
		)
	}

	if len(m.URIHash) > MaxURIHashLength {
		return sdkerrors.Wrapf(
			ErrInvalidInput,
			"invalid URI hash %q, the length must be less than or equal %d",
			len(m.URIHash),
			MaxURIHashLength,
		)
	}

	return nil
}

// GetSigners returns the message signers.
func (m MsgIssue) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Issuer),
	}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgIssue) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgIssue) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgIssue) Type() string {
	return TypeMsgIssue
}

// ValidateBasic checks that message fields are valid.
func (m MsgMint) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, _, err := DeconstructDenom(m.Coin.Denom); err != nil {
		return err
	}

	return m.Coin.Validate()
}

// GetSigners returns the required signers of this message type.
func (m MsgMint) GetSigners() []sdk.AccAddress {
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
func (m MsgBurn) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, _, err := DeconstructDenom(m.Coin.Denom); err != nil {
		return err
	}

	return m.Coin.Validate()
}

// GetSigners returns the required signers of this message type.
func (m MsgBurn) GetSigners() []sdk.AccAddress {
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
func (m MsgFreeze) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
	}

	_, issuer, err := DeconstructDenom(m.Coin.Denom)
	if err != nil {
		return err
	}

	if issuer.String() == m.Account {
		return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "issuer's balance can't be frozen")
	}

	return m.Coin.Validate()
}

// GetSigners returns the required signers of this message type.
func (m MsgFreeze) GetSigners() []sdk.AccAddress {
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
func (m MsgUnfreeze) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
	}

	if _, _, err := DeconstructDenom(m.Coin.Denom); err != nil {
		return err
	}

	return m.Coin.Validate()
}

// GetSigners returns the required signers of this message type.
func (m MsgUnfreeze) GetSigners() []sdk.AccAddress {
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
func (m MsgSetFrozen) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
	}

	_, issuer, err := DeconstructDenom(m.Coin.Denom)
	if err != nil {
		return err
	}

	if issuer.String() == m.Account {
		return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "issuer's balance can't be frozen")
	}

	return m.Coin.Validate()
}

// GetSigners returns the required signers of this message type.
func (m MsgSetFrozen) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Sender),
	}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgSetFrozen) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgSetFrozen) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgSetFrozen) Type() string {
	return TypeMsgFreeze
}

// ValidateBasic checks that message fields are valid.
func (m MsgGloballyFreeze) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, _, err := DeconstructDenom(m.Denom); err != nil {
		return err
	}

	return nil
}

// GetSigners returns the required signers of this message type.
func (m MsgGloballyFreeze) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Sender),
	}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgGloballyFreeze) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgGloballyFreeze) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgGloballyFreeze) Type() string {
	return TypeMsgGloballyFreeze
}

// ValidateBasic checks that message fields are valid.
func (m MsgGloballyUnfreeze) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, _, err := DeconstructDenom(m.Denom); err != nil {
		return err
	}

	return nil
}

// GetSigners returns the required signers of this message type.
func (m MsgGloballyUnfreeze) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Sender),
	}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgGloballyUnfreeze) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgGloballyUnfreeze) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgGloballyUnfreeze) Type() string {
	return TypeMsgGloballyUnfreeze
}

func (m MsgClawback) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
	}

	_, issuer, err := DeconstructDenom(m.Coin.Denom)
	if err != nil {
		return err
	}

	if issuer.String() == m.Account {
		return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "issuer's balance can't be clawed back")
	}

	if issuer.String() != m.Sender {
		return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "only issuer can claw back balance")
	}

	return m.Coin.Validate()
}

func (m MsgClawback) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Sender),
	}
}

func (m MsgClawback) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

func (m MsgClawback) Route() string {
	return RouterKey
}

func (m MsgClawback) Type() string {
	return TypeMsgClawback
}

// ValidateBasic checks that message fields are valid.
func (m MsgSetWhitelistedLimit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
	}

	_, issuer, err := DeconstructDenom(m.Coin.Denom)
	if err != nil {
		return err
	}

	if issuer.String() == m.Account {
		return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "issuer's balance can't be whitelisted")
	}

	return m.Coin.Validate()
}

// GetSigners returns the required signers of this message type.
func (m MsgSetWhitelistedLimit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Sender),
	}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgSetWhitelistedLimit) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgSetWhitelistedLimit) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgSetWhitelistedLimit) Type() string {
	return TypeMsgSetWhitelistedLimit
}

// ValidateBasic checks that message fields are valid.
func (m MsgUpgradeTokenV1) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	_, issuer, err := DeconstructDenom(m.Denom)
	if err != nil {
		return err
	}

	if issuer.String() != m.Sender {
		return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "only issuer can upgrade the denom")
	}

	return nil
}

// GetSigners returns the required signers of this message type.
func (m MsgUpgradeTokenV1) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Sender),
	}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgUpgradeTokenV1) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgUpgradeTokenV1) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgUpgradeTokenV1) Type() string {
	return TypeMsgUpgradeTokenV1
}

// ValidateBasic checks that message fields are valid.
func (m MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return cosmoserrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	return m.Params.ValidateBasic()
}

// GetSigners returns the required signers of this message type.
func (m MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgUpdateParams) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgUpdateParams) Type() string {
	return TypeMsgUpdateParams
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
