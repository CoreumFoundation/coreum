package types_test

import (
	"testing"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/pkg/config"
	"github.com/CoreumFoundation/coreum/v4/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func TestMain(m *testing.M) {
	n, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	if err != nil {
		panic(err)
	}
	n.SetSDKConfig()
	m.Run()
}

func TestMsgIssue_ValidateBasic(t *testing.T) {
	// TODO: Implement

	acc := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	validMessage := types.MsgCreateLimitOrder{
		Issuer:        acc.String(),
		Symbol:        "BTC",
		Subunit:       "btc",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(777),
	}

	testCases := []struct {
		name          string
		messageFunc   func(order types.MsgCreateLimitOrder) types.MsgCreateLimitOrder
		expectedError error
	}{
		{
			name: "valid",
			messageFunc: func(msg types.MsgCreateLimitOrder) types.MsgCreateLimitOrder {
				return msg
			},
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			requireT := require.New(t)
			err := tc.messageFunc(validMessage).ValidateBasic()
			if tc.expectedError == nil {
				requireT.NoError(err)
			} else {
				requireT.True(sdkerrors.IsOf(err, tc.expectedError))
			}
		})
	}
}

//nolint:lll // we don't care about test strings
func TestAmino(t *testing.T) {
	const address = "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"

	tests := []struct {
		name          string
		msg           legacytx.LegacyMsg
		wantAminoJSON string
	}{
		{
			name: types.TypeMsgCreateLimitOrder,
			msg: &types.MsgCreateLimitOrder{
				Issuer: address,
				Symbol: "ABC",
			},
			wantAminoJSON: `{"type":"dex/MsgIssue","value":{"burn_rate":"0","initial_amount":"0","issuer":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5","send_commission_rate":"0","symbol":"ABC"}}`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantAminoJSON, string(tt.msg.GetSignBytes()))
		})
	}
}
