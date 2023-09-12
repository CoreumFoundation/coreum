package types

import (
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

// Type of messages for amino.
const (
	TypeMsgUpdateStakingParams = "update-staking-params"
)

var (
	_ sdk.Msg            = &MsgUpdateStakingParams{}
	_ legacytx.LegacyMsg = &MsgUpdateStakingParams{}
)

// RegisterLegacyAminoCodec registers the amino types and interfaces.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgUpdateStakingParams{}, fmt.Sprintf("%s/MsgUpdateStakingParams", ModuleName), nil)
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

// GetSigners returns the required signers of this message type.
func (m *MsgUpdateStakingParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgUpdateStakingParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgUpdateStakingParams) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgUpdateStakingParams) Type() string {
	return TypeMsgUpdateStakingParams
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
