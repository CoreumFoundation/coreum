package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
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
	msg.Description = string(make([]byte, 10000))
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
				Issuer:  "cosmos1jgfmlcywhqwctenjeljs2huxw7l4p6apgaclyn",
				Account: "cosmos15jlvsclyuk7ezdzylarma225phfwv8me044yn0",
				Coin: sdk.Coin{
					Denom:  "symbol-cosmos1jgfmlcywhqwctenjeljs2huxw7l4p6apgaclyn-JiWf",
					Amount: sdk.NewInt(100),
				},
			},
		},
		{
			name: "invalid issuer address",
			message: types.MsgFreezeFungibleToken{
				Issuer:  "cosmos1jgfmlcywhqwctenjeljs2huxw7l4p6apgaclyn+",
				Account: "cosmos15jlvsclyuk7ezdzylarma225phfwv8me044yn0",
				Coin: sdk.Coin{
					Denom:  "symbol-cosmos1jgfmlcywhqwctenjeljs2huxw7l4p6apgaclyn+-JiWf",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid account",
			message: types.MsgFreezeFungibleToken{
				Issuer:  "cosmos1jgfmlcywhqwctenjeljs2huxw7l4p6apgaclyn",
				Account: "cosmos15jlvsclyuk7ezdzylarma225phfwv8me044yn0+",
				Coin: sdk.Coin{
					Denom:  "symbol-cosmos1jgfmlcywhqwctenjeljs2huxw7l4p6apgaclyn-JiWf",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid symbol",
			message: types.MsgFreezeFungibleToken{
				Issuer:  "cosmos1ms0g509yc38dvefpvq8nw8llv6j8mc0y8lx3yr",
				Account: "cosmos1upnqz6yhlxx3dnrvrkraz823ces39ee345zerw",
				Coin: sdk.Coin{
					Denom:  "symbol+-cosmos1ms0g509yc38dvefpvq8nw8llv6j8mc0y8lx3yr-LHxh",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: types.ErrInvalidDenom,
		},
		{
			name: "invalid denom checksum",
			message: types.MsgFreezeFungibleToken{
				Issuer:  "cosmos1jgfmlcywhqwctenjeljs2huxw7l4p6apgaclyn",
				Account: "cosmos15jlvsclyuk7ezdzylarma225phfwv8me044yn0",
				Coin: sdk.Coin{
					Denom:  "symbol-posmos1jgfmlcywhqwctenjeljs2huxw7l4p6apgaclyn-JiWf",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: types.ErrInvalidDenom,
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
				Issuer:  "cosmos1jgfmlcywhqwctenjeljs2huxw7l4p6apgaclyn",
				Account: "cosmos15jlvsclyuk7ezdzylarma225phfwv8me044yn0",
				Coin: sdk.Coin{
					Denom:  "symbol-cosmos1jgfmlcywhqwctenjeljs2huxw7l4p6apgaclyn-JiWf",
					Amount: sdk.NewInt(100),
				},
			},
		},
		{
			name: "invalid issuer address",
			message: types.MsgUnfreezeFungibleToken{
				Issuer:  "cosmos1jgfmlcywhqwctenjeljs2huxw7l4p6apgaclyn+",
				Account: "cosmos15jlvsclyuk7ezdzylarma225phfwv8me044yn0",
				Coin: sdk.Coin{
					Denom:  "symbol-cosmos1jgfmlcywhqwctenjeljs2huxw7l4p6apgaclyn+-JiWf",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid account",
			message: types.MsgUnfreezeFungibleToken{
				Issuer:  "cosmos1jgfmlcywhqwctenjeljs2huxw7l4p6apgaclyn",
				Account: "cosmos15jlvsclyuk7ezdzylarma225phfwv8me044yn0+",
				Coin: sdk.Coin{
					Denom:  "symbol-cosmos1jgfmlcywhqwctenjeljs2huxw7l4p6apgaclyn-JiWf",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid symbol",
			message: types.MsgUnfreezeFungibleToken{
				Issuer:  "cosmos1ms0g509yc38dvefpvq8nw8llv6j8mc0y8lx3yr",
				Account: "cosmos1upnqz6yhlxx3dnrvrkraz823ces39ee345zerw",
				Coin: sdk.Coin{
					Denom:  "symbol+-cosmos1ms0g509yc38dvefpvq8nw8llv6j8mc0y8lx3yr-LHxh",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: types.ErrInvalidDenom,
		},
		{
			name: "invalid denom checksum",
			message: types.MsgUnfreezeFungibleToken{
				Issuer:  "cosmos1jgfmlcywhqwctenjeljs2huxw7l4p6apgaclyn",
				Account: "cosmos15jlvsclyuk7ezdzylarma225phfwv8me044yn0",
				Coin: sdk.Coin{
					Denom:  "symbol-posmos1jgfmlcywhqwctenjeljs2huxw7l4p6apgaclyn-JiWf",
					Amount: sdk.NewInt(100),
				},
			},
			expectedError: types.ErrInvalidDenom,
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
