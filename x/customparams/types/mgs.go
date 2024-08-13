package types

import (
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Type of messages for amino.
const (
	TypeMsgUpdateStakingParams = "update-staking-params"
)

type extendedMsg interface {
	sdk.Msg
	sdk.HasValidateBasic
}

var (
	_ extendedMsg = &MsgUpdateStakingParams{}
)

// RegisterLegacyAminoCodec registers the amino types and interfaces.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgUpdateStakingParams{}, fmt.Sprintf("%s/MsgUpdateStakingParams", ModuleName))
}

// ValidateBasic checks that message fields are valid.
func (m *MsgUpdateStakingParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return cosmoserrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	if err := m.StakingParams.ValidateBasic(); err != nil {
		return cosmoserrors.ErrInvalidRequest.Wrapf("invalid params, err: %s", err)
	}

	return nil
}
