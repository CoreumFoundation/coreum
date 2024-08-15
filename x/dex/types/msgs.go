package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type extendedMsg interface {
	sdk.Msg
	sdk.HasValidateBasic
}

var _ extendedMsg = &MsgPlaceOrder{}

// RegisterLegacyAminoCodec registers the amino types and interfaces.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgPlaceOrder{}, fmt.Sprintf("%s/MsgPlaceOrder", ModuleName))
}

// ValidateBasic validates the message.
func (m MsgPlaceOrder) ValidateBasic() error {
	if _, err := NewOrderFormMsgPlaceOrder(m); err != nil {
		return err
	}

	return nil
}
