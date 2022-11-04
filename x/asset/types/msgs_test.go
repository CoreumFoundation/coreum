package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestMsgIssueFungibleToken_ValidateBasic(t *testing.T) {
	requireT := require.New(t)
	acc := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	msgF := func() types.MsgIssueFungibleToken {
		return types.MsgIssueFungibleToken{
			Issuer:        acc.String(),
			Symbol:        "BTC",
			Description:   "BTC Description",
			Recipient:     acc.String(),
			InitialAmount: sdk.NewInt(777),
		}
	}

	msg := msgF()
	requireT.NoError(msg.ValidateBasic())

	msg = msgF()
	msg.Issuer = "invalid"
	requireT.Error(msg.ValidateBasic())

	msg = msgF()
	msg.Symbol = ""
	requireT.Error(msg.ValidateBasic())

	msg = msgF()
	msg.Symbol = string(make([]byte, 10000))
	requireT.Error(msg.ValidateBasic())

	msg = msgF()
	msg.Symbol = "1BT"
	requireT.Error(msg.ValidateBasic())

	msg = msgF()
	msg.Recipient = "invalid"
	requireT.Error(msg.ValidateBasic())

	msg = msgF()
	msg.InitialAmount = sdk.Int{}
	requireT.Error(msg.ValidateBasic())

	msg = msgF()
	msg.InitialAmount = sdk.NewInt(-100)
	requireT.Error(msg.ValidateBasic())

	msg = msgF()
	msg.Description = string(make([]byte, 10000))
	requireT.Error(msg.ValidateBasic())
}
