package types

import (
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type extendedMsg interface {
	sdk.Msg
	sdk.HasValidateBasic
}

var _ extendedMsg = &MsgUpdateParams{}

// RegisterLegacyAminoCodec registers the amino types and interfaces.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, fmt.Sprintf("%s/MsgUpdateParams", ModuleName))
}

// ValidateBasic checks that message fields are valid.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return cosmoserrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	if err := m.Params.ValidateBasic(); err != nil {
		return cosmoserrors.ErrInvalidRequest.Wrapf("invalid params, errors: %s", err)
	}

	return nil
}
