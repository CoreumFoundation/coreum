package types_test

import (
	"strings"
	"testing"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v5/pkg/config"
	"github.com/CoreumFoundation/coreum/v5/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
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
		InitialAmount: sdkmath.NewInt(777),
		URI:           "https://my.invalid",
		URIHash:       "sha-hash",
		DEXSettings: &types.DEXSettings{
			UnifiedRefAmount:  lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("1.1")),
			WhitelistedDenoms: []string{"denom1", "denom2"},
		},
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
			name: "invalid_issuer_address",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.Issuer = "invalid"
				return msg
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
		{
			name: "invalid_missing_symbol",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.Symbol = ""
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_long_symbol",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.Symbol = string(make([]byte, 10000))
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_prohibited_chars_in_symbol",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.Symbol = "1BT"
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_nil_initial_amount",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.InitialAmount = sdkmath.Int{} // nil
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_negative_initial_amount",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.InitialAmount = sdkmath.NewInt(-100)
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_long_description",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.Description = string(make([]byte, 10000))
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_empty_subunit",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.Subunit = ""
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_empty_subunit",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.Subunit = ""
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_negative_burn_rate",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.BurnRate = sdkmath.LegacyMustNewDecFromStr("-0.1")
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_negative_send_commission_rate",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.SendCommissionRate = sdkmath.LegacyMustNewDecFromStr("-0.1")
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_duplicated_feature",
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
		{
			name: "invalid_uri",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.URI = string(make([]byte, 257))
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_uri_hash",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.URIHash = strings.Repeat("x", 129)
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_dex_settings",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.DEXSettings = &types.DEXSettings{
					UnifiedRefAmount: lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("0")),
				}
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_dex_whitelisted_denoms_duplicate",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.DEXSettings.WhitelistedDenoms = []string{"denom1", "denom1"}
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_dex_whitelisted_denoms_invalid_denom",
			messageFunc: func(msg types.MsgIssue) types.MsgIssue {
				msg.DEXSettings.WhitelistedDenoms = []string{"123!!!!!!!123", "denom2"}
				return msg
			},
			expectedError: types.ErrInvalidInput,
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
					Amount: sdkmath.NewInt(100),
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
					Amount: sdkmath.NewInt(100),
				},
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
		{
			name: "invalid account",
			message: types.MsgFreeze{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq+",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdkmath.NewInt(100),
				},
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			requireT := require.New(t)
			err := tc.message.ValidateBasic()
			if tc.expectedError == nil {
				requireT.NoError(err)
			} else {
				requireT.True(sdkerrors.IsOf(err, tc.expectedError))
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
					Amount: sdkmath.NewInt(100),
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
					Amount: sdkmath.NewInt(100),
				},
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
		{
			name: "invalid account",
			message: types.MsgUnfreeze{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq+",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdkmath.NewInt(100),
				},
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
		{
			name: "invalid denom",
			message: types.MsgUnfreeze{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Coin: sdk.Coin{
					Denom:  "0abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdkmath.NewInt(100),
				},
			},
			expectedErrorString: "invalid denom",
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			requireT := require.New(t)
			err := tc.message.ValidateBasic()
			switch {
			case tc.expectedError == nil && tc.expectedErrorString == "":
				requireT.NoError(err)
			case tc.expectedErrorString != "":
				requireT.Contains(err.Error(), tc.expectedErrorString)
			default:
				requireT.True(sdkerrors.IsOf(err, tc.expectedError))
			}
		})
	}
}

func TestMsgMint_ValidateBasic(t *testing.T) {
	type M = types.MsgMint

	acc := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	defaultMsg := func() M {
		return M{
			Sender: acc.String(),
			Coin:   sdk.NewCoin("abc"+"-"+acc.String(), sdkmath.NewInt(100)), //nolint:goconst
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

func TestMsgBurn_ValidateBasic(t *testing.T) {
	type M = types.MsgBurn

	acc := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	defaultMsg := func() M {
		return M{
			Sender: acc.String(),
			Coin:   sdk.NewCoin("abc"+"-"+acc.String(), sdkmath.NewInt(100)),
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

func TestMsgClawback_ValidateBasic(t *testing.T) {
	testCases := []struct {
		name          string
		message       types.MsgClawback
		expectedError error
	}{
		{
			name: "valid msg",
			message: types.MsgClawback{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdkmath.NewInt(100),
				},
			},
		},
		{
			name: "invalid sender address",
			message: types.MsgClawback{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5+",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdkmath.NewInt(100),
				},
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
		{
			name: "invalid account",
			message: types.MsgClawback{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq+",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdkmath.NewInt(100),
				},
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			requireT := require.New(t)
			err := tc.message.ValidateBasic()
			if tc.expectedError == nil {
				requireT.NoError(err)
			} else {
				requireT.True(sdkerrors.IsOf(err, tc.expectedError))
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
					Amount: sdkmath.NewInt(100),
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
					Amount: sdkmath.NewInt(100),
				},
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
		{
			name: "invalid account",
			message: types.MsgSetWhitelistedLimit{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq+",
				Coin: sdk.Coin{
					Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdkmath.NewInt(100),
				},
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
		{
			name: "invalid denom",
			message: types.MsgSetWhitelistedLimit{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Coin: sdk.Coin{
					Denom:  "0abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
					Amount: sdkmath.NewInt(100),
				},
			},
			expectedErrorString: "invalid denom",
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			requireT := require.New(t)
			err := tc.message.ValidateBasic()
			switch {
			case tc.expectedError == nil && tc.expectedErrorString == "":
				requireT.NoError(err)
			case tc.expectedErrorString != "":
				requireT.Contains(err.Error(), tc.expectedErrorString)
			default:
				requireT.True(sdkerrors.IsOf(err, tc.expectedError))
			}
		})
	}
}

func TestMsgTransferAdmin_ValidateBasic(t *testing.T) {
	testCases := []struct {
		name                string
		message             types.MsgTransferAdmin
		expectedError       error
		expectedErrorString string
	}{
		{
			name: "valid msg",
			message: types.MsgTransferAdmin{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Denom:   "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
			},
		},
		{
			name: "invalid sender address",
			message: types.MsgTransferAdmin{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5+",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq",
				Denom:   "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
		{
			name: "invalid account",
			message: types.MsgTransferAdmin{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore1k3mke3gyf9apyd8vxveutgp9h4j2e80e05yfuq+",
				Denom:   "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
		{
			name: "invalid denom",
			message: types.MsgTransferAdmin{
				Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Account: "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Denom:   "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5+",
			},
			expectedErrorString: "invalid denom",
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			requireT := require.New(t)
			err := tc.message.ValidateBasic()
			switch {
			case tc.expectedError == nil && tc.expectedErrorString == "":
				requireT.NoError(err)
			case tc.expectedErrorString != "":
				requireT.Contains(err.Error(), tc.expectedErrorString)
			default:
				requireT.ErrorIs(err, tc.expectedError)
			}
		})
	}
}

func TestMsgClearAdmin_ValidateBasic(t *testing.T) {
	testCases := []struct {
		name                string
		message             types.MsgClearAdmin
		expectedError       error
		expectedErrorString string
	}{
		{
			name: "valid msg",
			message: types.MsgClearAdmin{
				Sender: "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
			},
		},
		{
			name: "invalid sender address",
			message: types.MsgClearAdmin{
				Sender: "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5+",
				Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
		{
			name: "invalid denom",
			message: types.MsgClearAdmin{
				Sender: "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
				Denom:  "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5+",
			},
			expectedErrorString: "invalid denom",
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			requireT := require.New(t)
			err := tc.message.ValidateBasic()
			switch {
			case tc.expectedError == nil && tc.expectedErrorString == "":
				requireT.NoError(err)
			case tc.expectedErrorString != "":
				requireT.Contains(err.Error(), tc.expectedErrorString)
			default:
				requireT.ErrorIs(err, tc.expectedError)
			}
		})
	}
}

func TestMsgUpdateDEXUnifiedRefAmount_ValidateBasic(t *testing.T) {
	validMessage := types.MsgUpdateDEXUnifiedRefAmount{
		Sender:           sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address()).String(),
		Denom:            "dnm",
		UnifiedRefAmount: sdkmath.LegacyMustNewDecFromStr("1.3"),
	}

	testCases := []struct {
		name          string
		messageFunc   func(msg types.MsgUpdateDEXUnifiedRefAmount) types.MsgUpdateDEXUnifiedRefAmount
		expectedError error
	}{
		{
			name: "valid",
			messageFunc: func(msg types.MsgUpdateDEXUnifiedRefAmount) types.MsgUpdateDEXUnifiedRefAmount {
				return msg
			},
		},
		{
			name: "invalid_sender",
			messageFunc: func(msg types.MsgUpdateDEXUnifiedRefAmount) types.MsgUpdateDEXUnifiedRefAmount {
				msg.Sender = "invalid"
				return msg
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
		{
			name: "invalid_unified_ref_amount",
			messageFunc: func(msg types.MsgUpdateDEXUnifiedRefAmount) types.MsgUpdateDEXUnifiedRefAmount {
				msg.UnifiedRefAmount = sdkmath.LegacyMustNewDecFromStr("-1")
				return msg
			},
			expectedError: types.ErrInvalidInput,
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

func TestMsgUpdateDEXWhitelistedDenoms_ValidateBasic(t *testing.T) {
	validMessage := types.MsgUpdateDEXWhitelistedDenoms{
		Sender:            sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address()).String(),
		Denom:             "dnm",
		WhitelistedDenoms: []string{"denom1", "denom2"},
	}

	testCases := []struct {
		name          string
		messageFunc   func(msg types.MsgUpdateDEXWhitelistedDenoms) types.MsgUpdateDEXWhitelistedDenoms
		expectedError error
	}{
		{
			name: "valid",
			messageFunc: func(msg types.MsgUpdateDEXWhitelistedDenoms) types.MsgUpdateDEXWhitelistedDenoms {
				return msg
			},
		},
		{
			name: "invalid_dex_whitelisted_denoms_duplicate",
			messageFunc: func(msg types.MsgUpdateDEXWhitelistedDenoms) types.MsgUpdateDEXWhitelistedDenoms {
				msg.WhitelistedDenoms = []string{"denom1", "denom1"}
				return msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_dex_whitelisted_denoms_invalid_denom",
			messageFunc: func(msg types.MsgUpdateDEXWhitelistedDenoms) types.MsgUpdateDEXWhitelistedDenoms {
				msg.WhitelistedDenoms = []string{"denom1", "1!!!!denom1"}
				return msg
			},
			expectedError: types.ErrInvalidInput,
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
	coin := sdk.NewInt64Coin("my-denom", 1)

	tests := []struct {
		name          string
		msg           sdk.Msg
		wantAminoJSON string
	}{
		{
			name: sdk.MsgTypeURL(&types.MsgIssue{}),
			msg: &types.MsgIssue{
				Issuer: address,
				Symbol: "ABC",
			},
			wantAminoJSON: `{"type":"assetft/MsgIssue","value":{"burn_rate":"0","initial_amount":"0","issuer":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5","send_commission_rate":"0","symbol":"ABC"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgMint{}),
			msg: &types.MsgMint{
				Sender: address,
				Coin:   coin,
			},
			wantAminoJSON: `{"type":"assetft/MsgMint","value":{"coin":{"amount":"1","denom":"my-denom"},"sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgBurn{}),
			msg: &types.MsgBurn{
				Sender: address,
				Coin:   coin,
			},
			wantAminoJSON: `{"type":"assetft/MsgBurn","value":{"coin":{"amount":"1","denom":"my-denom"},"sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgFreeze{}),
			msg: &types.MsgFreeze{
				Sender: address,
				Coin:   coin,
			},
			wantAminoJSON: `{"type":"assetft/MsgFreeze","value":{"coin":{"amount":"1","denom":"my-denom"},"sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgUnfreeze{}),
			msg: &types.MsgUnfreeze{
				Sender: address,
				Coin:   coin,
			},
			wantAminoJSON: `{"type":"assetft/MsgUnfreeze","value":{"coin":{"amount":"1","denom":"my-denom"},"sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgSetFrozen{}),
			msg: &types.MsgSetFrozen{
				Sender:  address,
				Account: address,
				Coin:    coin,
			},
			wantAminoJSON: `{"type":"assetft/MsgSetFrozen","value":{"account":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5","coin":{"amount":"1","denom":"my-denom"},"sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgGloballyFreeze{}),
			msg: &types.MsgGloballyFreeze{
				Sender: address,
				Denom:  coin.Denom,
			},
			wantAminoJSON: `{"type":"assetft/MsgGloballyFreeze","value":{"denom":"my-denom","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgGloballyUnfreeze{}),
			msg: &types.MsgGloballyUnfreeze{
				Sender: address,
				Denom:  coin.Denom,
			},
			wantAminoJSON: `{"type":"assetft/MsgGloballyUnfreeze","value":{"denom":"my-denom","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgSetWhitelistedLimit{}),
			msg: &types.MsgSetWhitelistedLimit{
				Sender:  address,
				Account: address,
				Coin:    coin,
			},
			wantAminoJSON: `{"type":"assetft/MsgSetWhitelistedLimit","value":{"account":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5","coin":{"amount":"1","denom":"my-denom"},"sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgUpgradeTokenV1{}),
			msg: &types.MsgUpgradeTokenV1{
				Sender:     address,
				Denom:      coin.Denom,
				IbcEnabled: false,
			},
			wantAminoJSON: `{"type":"assetft/MsgUpgradeTokenV1","value":{"denom":"my-denom","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgUpdateDEXUnifiedRefAmount{}),
			msg: &types.MsgUpdateDEXUnifiedRefAmount{
				Sender:           address,
				Denom:            coin.Denom,
				UnifiedRefAmount: sdkmath.LegacyMustNewDecFromStr("1.3"),
			},
			wantAminoJSON: `{"type":"assetft/MsgUpdateDEXUnifiedRefAmount","value":{"denom":"my-denom","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5","unified_ref_amount":"1.300000000000000000"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgUpdateDEXWhitelistedDenoms{}),
			msg: &types.MsgUpdateDEXWhitelistedDenoms{
				Sender:            address,
				Denom:             coin.Denom,
				WhitelistedDenoms: []string{"denom2", "denom3"},
			},
			wantAminoJSON: `{"type":"assetft/MsgUpdateDEXWhitelistedDenoms","value":{"denom":"my-denom","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5","whitelisted_denoms":["denom2","denom3"]}}`,
		},
	}

	legacyAmino := codec.NewLegacyAmino()
	types.RegisterLegacyAminoCodec(legacyAmino)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			generatedJSON := legacyAmino.Amino.MustMarshalJSON(tt.msg)
			require.Equal(t, tt.wantAminoJSON, string(sdk.MustSortJSON(generatedJSON)))
		})
	}
}
