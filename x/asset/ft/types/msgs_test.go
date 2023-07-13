package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v2/pkg/config"
	"github.com/CoreumFoundation/coreum/v2/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/v2/x/asset/ft/types"
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
	acc := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	validMessage := types.MsgIssue{
		Issuer:        acc.String(),
		Symbol:        "BTC",
		Subunit:       "btc",
		Precision:     1,
		Description:   "BTC Description",
		InitialAmount: sdk.NewInt(777),
	}

	testCases := []struct {
		name          string
		messageFunc   func(types.MsgIssue) types.MsgIssue
		expectedError error
	}{
		{
			name: "valid",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				return msg
			},
		},
		{
			name: "invalid issuer address",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.Issuer = "invalid"
				return msg
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid missing symbol",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.Symbol = ""
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid long symbol",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.Symbol = string(make([]byte, 10000))
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid prohibited chars in symbol",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.Symbol = "1BT"
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid nil initial amount",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.InitialAmount = sdk.Int{} // nil
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid negative initial amount",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.InitialAmount = sdk.NewInt(-100)
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid long description",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.Description = string(make([]byte, 10000))
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid empty subunit",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.Subunit = ""
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid empty subunit",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.Subunit = ""
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid negative burn rate",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.BurnRate = sdk.MustNewDecFromStr("-0.1")
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid negative send commission rate",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.SendCommissionRate = sdk.MustNewDecFromStr("-0.1")
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid duplicated feature",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.Features = []types.Feature{
					types.Feature_burning,
					types.Feature_whitelisting,
					types.Feature_burning,
				}
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			assertT := assert.New(t)
			err := tc.messageFunc(validMessage).ValidateBasic()
			if tc.expectedError == nil {
				assertT.NoError(err)
			} else {
				assertT.True(sdkerrors.IsOf(err, tc.expectedError))
			}
		})
	}
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
		{
			name: "issuer freezing",
			message: types.MsgFreeze{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: sdkerrors.ErrUnauthorized,
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
			Coin:   sdk.NewCoin("abc"+"-"+acc.String(), sdk.NewInt(100)),
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
			Coin:   sdk.NewCoin("abc"+"-"+acc.String(), sdk.NewInt(100)),
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
		{
			name: "issuer whitelisting",
			message: types.MsgSetWhitelistedLimit{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: sdkerrors.ErrUnauthorized,
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

func TestAmino(t *testing.T) {
	const address = "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"
	coin := sdk.NewInt64Coin("my-denom", 1)

	tests := []struct {
		name          string
		msg           legacytx.LegacyMsg
		wantAminoJSON string
	}{
		{
			name: types.TypeMsgIssue,
			msg: &types.MsgIssue{
				Issuer: address,
				Symbol: "ABC",
			},
			wantAminoJSON: `{"type":"assetft/MsgIssue","value":{"burn_rate":"0","initial_amount":"0","issuer":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5","send_commission_rate":"0","symbol":"ABC"}}`,
		},
		{
			name: types.TypeMsgMint,
			msg: &types.MsgMint{
				Sender: address,
				Coin:   coin,
			},
			wantAminoJSON: `{"type":"assetft/MsgMint","value":{"coin":{"amount":"1","denom":"my-denom"},"sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: types.TypeMsgBurn,
			msg: &types.MsgBurn{
				Sender: address,
				Coin:   coin,
			},
			wantAminoJSON: `{"type":"assetft/MsgBurn","value":{"coin":{"amount":"1","denom":"my-denom"},"sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: types.TypeMsgFreeze,
			msg: &types.MsgFreeze{
				Sender: address,
				Coin:   coin,
			},
			wantAminoJSON: `{"type":"assetft/MsgFreeze","value":{"coin":{"amount":"1","denom":"my-denom"},"sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: types.TypeMsgUnfreeze,
			msg: &types.MsgUnfreeze{
				Sender: address,
				Coin:   coin,
			},
			wantAminoJSON: `{"type":"assetft/MsgUnfreeze","value":{"coin":{"amount":"1","denom":"my-denom"},"sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: types.TypeMsgUnfreeze,
			msg: &types.MsgUnfreeze{
				Sender: address,
				Coin:   coin,
			},
			wantAminoJSON: `{"type":"assetft/MsgUnfreeze","value":{"coin":{"amount":"1","denom":"my-denom"},"sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: types.TypeMsgGloballyFreeze,
			msg: &types.MsgGloballyFreeze{
				Sender: address,
				Denom:  coin.Denom,
			},
			wantAminoJSON: `{"type":"assetft/MsgGloballyFreeze","value":{"denom":"my-denom","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: types.TypeMsgGloballyUnfreeze,
			msg: &types.MsgGloballyUnfreeze{
				Sender: address,
				Denom:  coin.Denom,
			},
			wantAminoJSON: `{"type":"assetft/MsgGloballyUnfreeze","value":{"denom":"my-denom","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: types.TypeMsgUnfreeze,
			msg: &types.MsgUnfreeze{
				Sender: address,
				Coin:   coin,
			},
			wantAminoJSON: `{"type":"assetft/MsgUnfreeze","value":{"coin":{"amount":"1","denom":"my-denom"},"sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: types.TypeMsgSetWhitelistedLimit,
			msg: &types.MsgSetWhitelistedLimit{
				Sender:  address,
				Account: address,
				Coin:    coin,
			},
			wantAminoJSON: `{"type":"assetft/MsgUnMsgSetWhitelistedLimitfreeze","value":{"account":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5","coin":{"amount":"1","denom":"my-denom"},"sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: types.TypeMsgUpgradeTokenV1,
			msg: &types.MsgUpgradeTokenV1{
				Sender:     address,
				Denom:      coin.Denom,
				IbcEnabled: false,
			},
			wantAminoJSON: `{"type":"assetft/MsgUpgradeTokenV1","value":{"denom":"my-denom","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantAminoJSON, string(tt.msg.GetSignBytes()))
		})
	}
}
