package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestMsgIssueAsset_ValidateBasic(t *testing.T) {
	requireT := require.New(t)
	acc := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	msgF := func() types.MsgIssueAsset {
		return types.MsgIssueAsset{
			Definition: &types.AssetDefinition{
				Recipient:   acc.String(),
				Type:        types.AssetType_FT, //nolint:nosnakecase // protogen
				Code:        "BTC",
				Description: "BTC Description",
				Ft: &types.FTCustomDefinition{
					Precision:     6,
					InitialAmount: sdk.NewInt(777),
				},
			},
		}
	}

	msg := msgF()
	requireT.NoError(msg.ValidateBasic())

	msg = msgF()
	msg.From = "invalid"
	requireT.Error(msg.ValidateBasic())

	msg = msgF()
	msg.Definition.Recipient = "invalid"
	requireT.Error(msg.ValidateBasic())

	msg = msgF()
	msg.Definition.Code = ""
	requireT.Error(msg.ValidateBasic())

	msg = msgF()
	msg.Definition.Code = string(make([]byte, 10000))
	requireT.Error(msg.ValidateBasic())

	msg = msgF()
	msg.Definition.Description = string(make([]byte, 10000))
	requireT.Error(msg.ValidateBasic())

	msg = msgF()
	msg.Definition.Ft.Precision = 100
	requireT.Error(msg.ValidateBasic())

	msg = msgF()
	msg.Definition.Ft.InitialAmount = sdk.NewInt(-100)
	requireT.Error(msg.ValidateBasic())

	msg = msgF()
	msg.Definition.Type = types.AssetType_NFT //nolint:nosnakecase // protogen
	requireT.Error(msg.ValidateBasic())
}
