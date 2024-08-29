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

const (
	// MaxDescriptionLength is max description length.
	MaxDescriptionLength = 200
	// MaxURILength is max URI length.
	MaxURILength = 256
	// MaxURIHashLength is max URIHash length.
	MaxURIHashLength = 128
)

// extendedMsg is sdk.Msg with extended functions.
type extendedMsg interface {
	sdk.Msg
	sdk.HasValidateBasic
}

var (
	_ extendedMsg = &MsgIssue{}
	_ extendedMsg = &MsgMint{}
	_ extendedMsg = &MsgBurn{}
	_ extendedMsg = &MsgFreeze{}
	_ extendedMsg = &MsgUnfreeze{}
	_ extendedMsg = &MsgSetFrozen{}
	_ extendedMsg = &MsgGloballyFreeze{}
	_ extendedMsg = &MsgGloballyUnfreeze{}
	_ extendedMsg = &MsgClawback{}
	_ extendedMsg = &MsgSetWhitelistedLimit{}
	_ extendedMsg = &MsgTransferAdmin{}
	_ extendedMsg = &MsgUpgradeTokenV1{}
	_ extendedMsg = &MsgUpdateParams{}
	_ extendedMsg = &MsgUpdateDEXSettings{}
)

// RegisterLegacyAminoCodec registers the amino types and interfaces.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgIssue{}, fmt.Sprintf("%s/MsgIssue", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgMint{}, fmt.Sprintf("%s/MsgMint", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgBurn{}, fmt.Sprintf("%s/MsgBurn", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgFreeze{}, fmt.Sprintf("%s/MsgFreeze", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgUnfreeze{}, fmt.Sprintf("%s/MsgUnfreeze", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgSetFrozen{}, fmt.Sprintf("%s/MsgSetFrozen", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgGloballyFreeze{}, fmt.Sprintf("%s/MsgGloballyFreeze", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgGloballyUnfreeze{}, fmt.Sprintf("%s/MsgGloballyUnfreeze", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgSetWhitelistedLimit{}, fmt.Sprintf("%s/MsgSetWhitelistedLimit", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgUpgradeTokenV1{}, fmt.Sprintf("%s/MsgUpgradeTokenV1", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, fmt.Sprintf("%s/MsgUpdateParams", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgUpdateDEXSettings{}, fmt.Sprintf("%s/MsgUpdateDEXSettings", ModuleName))
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

	if err := ValidateFeatures(m.Features); err != nil {
		return err
	}

	// we allow zero initial amount, in that case we won't mint it initially
	if m.InitialAmount.IsNil() || m.InitialAmount.IsNegative() {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid initial amount %s, can't be negative", m.InitialAmount.String())
	}

	if m.DEXSettings != nil {
		if err := ValidateDEXSettings(*m.DEXSettings); err != nil {
			return err
		}
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

// ValidateBasic checks that message fields are valid.
func (m MsgFreeze) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
	}

	_, _, err := DeconstructDenom(m.Coin.Denom)
	if err != nil {
		return err
	}

	return m.Coin.Validate()
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

// ValidateBasic checks that message fields are valid.
func (m MsgSetFrozen) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
	}

	_, _, err := DeconstructDenom(m.Coin.Denom)
	if err != nil {
		return err
	}

	return m.Coin.Validate()
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

// ValidateBasic checks that message fields are valid.
func (m MsgClawback) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
	}

	_, _, err := DeconstructDenom(m.Coin.Denom)
	if err != nil {
		return err
	}

	return m.Coin.Validate()
}

// ValidateBasic checks that message fields are valid.
func (m MsgSetWhitelistedLimit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
	}

	_, _, err := DeconstructDenom(m.Coin.Denom)
	if err != nil {
		return err
	}

	return m.Coin.Validate()
}

// ValidateBasic checks that message fields are valid.
func (m MsgTransferAdmin) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid account address")
	}

	_, _, err := DeconstructDenom(m.Denom)
	if err != nil {
		return err
	}

	return nil
}

// ValidateBasic checks that message fields are valid.
func (m MsgClearAdmin) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidAddress, "invalid sender address")
	}

	_, _, err := DeconstructDenom(m.Denom)
	if err != nil {
		return err
	}

	return nil
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

// ValidateBasic checks that message fields are valid.
func (m MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return cosmoserrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	return m.Params.ValidateBasic()
}

// ValidateBasic checks that message fields are valid.
func (m MsgUpdateDEXSettings) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return cosmoserrors.ErrInvalidAddress.Wrapf("invalid sender address: %s", err)
	}

	return ValidateDEXSettings(m.DEXSettings)
}
