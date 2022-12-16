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
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

func TestMain(m *testing.M) {
	n, err := config.NetworkByChainID(constant.ChainIDDev)
	if err != nil {
		panic(err)
	}
	n.SetSDKConfig()
	m.Run()
}

func TestMsgIssue_ValidateBasic(t *testing.T) {
	requireT := require.New(t)
	acc := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	msgF := func() types.MsgIssue {
		return types.MsgIssue{
			Issuer:        acc.String(),
			Symbol:        "BTC",
			Subunit:       "btc",
			Description:   "BTC Description",
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
	msg.InitialAmount = sdk.Int{}
	requireT.Error(msg.ValidateBasic())

	msg = msgF()
	msg.InitialAmount = sdk.NewInt(-100)
	requireT.Error(msg.ValidateBasic())

	msg = msgF()
	msg.Description = string(make([]byte, 10000))
	requireT.Error(msg.ValidateBasic())

	msg = msgF()
	msg.Subunit = ""
	requireT.Error(msg.ValidateBasic())
}

func TestMsgFreeze_ValidateBasic(t *testing.T) {
	testCases := []struct {
		name          string
		message       types.MsgFreeze
		expectedError error
	}{
		{
			name: "valid msg",
			message: types.MsgFreeze{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdk.NewInt(100),
				},
			},
		},
		{
			name: "invalid sender address",
			message: types.MsgFreeze{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5+",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid account",
			message: types.MsgFreeze{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq+",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
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
func TestMsgUnfreeze_ValidateBasic(t *testing.T) {
	testCases := []struct {
		name                string
		message             types.MsgUnfreeze
		expectedError       error
		expectedErrorString string
	}{
		{
			name: "valid msg",
			message: types.MsgUnfreeze{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdk.NewInt(100),
				},
			},
		},
		{
			name: "invalid sender address",
			message: types.MsgUnfreeze{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5+",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid account",
			message: types.MsgUnfreeze{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq+",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid denom",
			message: types.MsgUnfreeze{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Coin: sdk.Coin{
					Denom:  "0abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdk.NewInt(100),
				},
			},
			expectedErrorString: "invalid denom",
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			assertT := assert.New(t)
			err := tc.message.ValidateBasic()
			switch {
			case tc.expectedError == nil && tc.expectedErrorString == "":
				assertT.NoError(err)
			case tc.expectedErrorString != "":
				assertT.Contains(err.Error(), tc.expectedErrorString)
			default:
				assertT.True(sdkerrors.IsOf(err, tc.expectedError))
			}
		})
	}
}

//nolint:dupl // tests and mint tests are identical, but merging them is not beneficial
func TestMsgMint_ValidateBasic(t *testing.T) {
	type M = types.MsgMint

	acc := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	defaultMsg := func() M {
		return M{
			Sender: acc.String(),
			Coin:   sdk.NewCoin("ABC"+"-"+acc.String(), sdk.NewInt(100)),
		}
	}

	testCases := []struct {
		name        string
		modifyMsg   func(M) M
		expectError bool
	}{
		{
			name:      "all is good",
			modifyMsg: func(m M) M { return m },
		},
		{
			name:        "invalid sender address",
			modifyMsg:   func(m M) M { m.Sender = "invalid sender"; return m },
			expectError: true,
		},
		{
			name:        "invalid coin",
			modifyMsg:   func(m M) M { m.Coin = sdk.Coin{}; return m },
			expectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			requireT := require.New(t)
			msg := tc.modifyMsg(defaultMsg())
			if tc.expectError {
				requireT.Error(msg.ValidateBasic())
			} else {
				requireT.NoError(msg.ValidateBasic())
			}
		})
	}
}

//nolint:dupl // tests and mint tests are identical, but merging them is not beneficial
func TestMsgBurn_ValidateBasic(t *testing.T) {
	type M = types.MsgBurn

	acc := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	defaultMsg := func() M {
		return M{
			Sender: acc.String(),
			Coin:   sdk.NewCoin("ABC"+"-"+acc.String(), sdk.NewInt(100)),
		}
	}

	testCases := []struct {
		name        string
		modifyMsg   func(M) M
		expectError bool
	}{
		{
			name:      "all is good",
			modifyMsg: func(m M) M { return m },
		},
		{
			name:        "invalid sender address",
			modifyMsg:   func(m M) M { m.Sender = "invalid sender"; return m },
			expectError: true,
		},
		{
			name:        "invalid coin",
			modifyMsg:   func(m M) M { m.Coin = sdk.Coin{}; return m },
			expectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			requireT := require.New(t)
			msg := tc.modifyMsg(defaultMsg())
			if tc.expectError {
				requireT.Error(msg.ValidateBasic())
			} else {
				requireT.NoError(msg.ValidateBasic())
			}
		})
	}
}

//nolint:dupl // test cases are identical between freeze and unfreeze, but reuse is not beneficial for tests
func TestMsgSetWhitelistedLimit_ValidateBasic(t *testing.T) {
	testCases := []struct {
		name                string
		message             types.MsgSetWhitelistedLimit
		expectedError       error
		expectedErrorString string
	}{
		{
			name: "valid msg",
			message: types.MsgSetWhitelistedLimit{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdk.NewInt(100),
				},
			},
		},
		{
			name: "invalid sender address",
			message: types.MsgSetWhitelistedLimit{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5+",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid account",
			message: types.MsgSetWhitelistedLimit{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq+",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid denom",
			message: types.MsgSetWhitelistedLimit{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Coin: sdk.Coin{
					Denom:  "0abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdk.NewInt(100),
				},
			},
			expectedErrorString: "invalid denom",
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			assertT := assert.New(t)
			err := tc.message.ValidateBasic()
			switch {
			case tc.expectedError == nil && tc.expectedErrorString == "":
				assertT.NoError(err)
			case tc.expectedErrorString != "":
				assertT.Contains(err.Error(), tc.expectedErrorString)
			default:
				assertT.True(sdkerrors.IsOf(err, tc.expectedError))
			}
		})
	}
}
