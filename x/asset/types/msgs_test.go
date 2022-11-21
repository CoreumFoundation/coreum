package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestMain(m *testing.M) {
	n, err := config.NetworkByChainID(constant.ChainIDDev)
	if err != nil {
		panic(err)
	}
	n.SetSDKConfig()
	m.Run()
}

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

//nolint:dupl // test cases are identical between freeze and unfreeze, but reuse is not beneficial for tests
func TestMsgFreezeFungibleToken_ValidateBasic(t *testing.T) {
	testCases := []struct {
		name          string
		message       types.MsgFreezeFungibleToken
		expectedError error
	}{
		{
			name: "valid msg",
			message: types.MsgFreezeFungibleToken{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5-2J7e",
					Amount: sdk.NewInt(100),
				},
			},
		},
		{
			name: "invalid issuer address",
			message: types.MsgFreezeFungibleToken{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5+",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5-2J7e",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid account",
			message: types.MsgFreezeFungibleToken{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq+",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5-2J7e",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			assertT := assert.New(t)
			err := tc.message.ValidateBasic()
			if tc.expectedError == nil {
				assertT.NoError(err)
			} else {
				assertT.True(sdkerrors.IsOf(err, tc.expectedError))
			}
		})
	}
}

//nolint:dupl // test cases are identical between freeze and unfreeze, but reuse is not beneficial for tests
func TestMsgUnfreezeFungibleToken_ValidateBasic(t *testing.T) {
	testCases := []struct {
		name          string
		message       types.MsgUnfreezeFungibleToken
		expectedError error
	}{
		{
			name: "valid msg",
			message: types.MsgUnfreezeFungibleToken{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5-2J7e",
					Amount: sdk.NewInt(100),
				},
			},
		},
		{
			name: "invalid issuer address",
			message: types.MsgUnfreezeFungibleToken{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5+",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5-2J7e",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid account",
			message: types.MsgUnfreezeFungibleToken{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq+",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5-2J7e",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			assertT := assert.New(t)
			err := tc.message.ValidateBasic()
			if tc.expectedError == nil {
				assertT.NoError(err)
			} else {
				assertT.True(sdkerrors.IsOf(err, tc.expectedError))
			}
		})
	}
}
