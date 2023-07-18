package nft_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v2/x/nft"
)

func TestAmino(t *testing.T) {
	const address = "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"

	tests := []struct {
		name          string
		msg           legacytx.LegacyMsg
		wantAminoJSON string
	}{
		{
			name: nft.TypeMsgSend,
			msg: &nft.MsgSend{
				ClassId:  "class1",
				Id:       "id1",
				Sender:   address,
				Receiver: address,
			},
			wantAminoJSON: `{"type":"cnft/MsgSend","value":{"class_id":"class1","id":"id1","receiver":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantAminoJSON, string(tt.msg.GetSignBytes()))
		})
	}
}
