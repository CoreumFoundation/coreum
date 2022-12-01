package types_test

import (
	"testing"

	codetypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

//nolint:funlen // many test cases
func TestMsgCreateNonFungibleTokenClass_ValidateBasic(t *testing.T) {
	requireT := require.New(t)

	dataString := "metadata"
	dataValue, err := codetypes.NewAnyWithValue(&gogotypes.BytesValue{Value: []byte(dataString)})
	requireT.NoError(err)

	validaMessage := types.MsgCreateNonFungibleTokenClass{
		Creator:     "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		Name:        "name",
		Symbol:      "symbol",
		Description: "description",
		Uri:         "https://my.uri",
		UriHash:     "sha-hash",
		Data:        dataValue,
	}
	testCases := []struct {
		name          string
		messageFunc   func() *types.MsgCreateNonFungibleTokenClass
		expectedError error
	}{
		{
			name: "valid msg",
			messageFunc: func() *types.MsgCreateNonFungibleTokenClass {
				msg := validaMessage
				return &msg
			},
		},
		{
			name: "invalid creator",
			messageFunc: func() *types.MsgCreateNonFungibleTokenClass {
				msg := validaMessage
				msg.Creator = "devcore172rc5sz2uc"
				return &msg
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid name",
			messageFunc: func() *types.MsgCreateNonFungibleTokenClass {
				msg := validaMessage
				msg.Name = string(make([]byte, 129))
				return &msg
			},
			expectedError: types.ErrInvalidNonFungibleTokenClass,
		},
		{
			name: "invalid empty symbol",
			messageFunc: func() *types.MsgCreateNonFungibleTokenClass {
				msg := validaMessage
				msg.Symbol = ""
				return &msg
			},
			expectedError: types.ErrInvalidNonFungibleTokenClass,
		},
		{
			name: "invalid char symbol",
			messageFunc: func() *types.MsgCreateNonFungibleTokenClass {
				msg := validaMessage
				msg.Symbol = "#x#"
				return &msg
			},
			expectedError: types.ErrInvalidNonFungibleTokenClass,
		},
		{
			name: "invalid description",
			messageFunc: func() *types.MsgCreateNonFungibleTokenClass {
				msg := validaMessage
				msg.Description = string(make([]byte, 257))
				return &msg
			},
			expectedError: types.ErrInvalidNonFungibleTokenClass,
		},
		{
			name: "invalid uri",
			messageFunc: func() *types.MsgCreateNonFungibleTokenClass {
				msg := validaMessage
				msg.Uri = string(make([]byte, 257))
				return &msg
			},
			expectedError: types.ErrInvalidNonFungibleTokenClass,
		},
		{
			name: "invalid uri hash",
			messageFunc: func() *types.MsgCreateNonFungibleTokenClass {
				msg := validaMessage
				msg.UriHash = string(make([]byte, 129))
				return &msg
			},
			expectedError: types.ErrInvalidNonFungibleTokenClass,
		},
		{
			name: "invalid data",
			messageFunc: func() *types.MsgCreateNonFungibleTokenClass {
				longDataString := string(make([]byte, 5001))
				longDataValue, err := codetypes.NewAnyWithValue(&gogotypes.BytesValue{Value: []byte(longDataString)})
				requireT.NoError(err)
				msg := validaMessage
				msg.Data = longDataValue
				return &msg
			},
			expectedError: types.ErrInvalidNonFungibleTokenClass,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			assertT := assert.New(t)
			err := tc.messageFunc().ValidateBasic()
			if tc.expectedError == nil {
				assertT.NoError(err)
			} else {
				assertT.True(sdkerrors.IsOf(err, tc.expectedError))
			}
		})
	}
}

func TestMsgMintNonFungibleToken_ValidateBasic(t *testing.T) {
	requireT := require.New(t)

	dataString := "metadata"
	dataValue, err := codetypes.NewAnyWithValue(&gogotypes.BytesValue{Value: []byte(dataString)})
	requireT.NoError(err)

	validaMessage := types.MsgMintNonFungibleToken{
		Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		Id:      "my-id",
		ClassId: "symbol-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		Uri:     "https://my.uri",
		UriHash: "sha-hash",
		Data:    dataValue,
	}
	testCases := []struct {
		name          string
		messageFunc   func() *types.MsgMintNonFungibleToken
		expectedError error
	}{
		{
			name: "valid msg",
			messageFunc: func() *types.MsgMintNonFungibleToken {
				msg := validaMessage
				return &msg
			},
		},
		{
			name: "invalid id",
			messageFunc: func() *types.MsgMintNonFungibleToken {
				msg := validaMessage
				msg.Id = "id?"
				return &msg
			},
			expectedError: types.ErrInvalidNonFungibleToken,
		},
		{
			name: "invalid sender",
			messageFunc: func() *types.MsgMintNonFungibleToken {
				msg := validaMessage
				msg.Sender = "devcore172rc5sz2uc"
				return &msg
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid classID",
			messageFunc: func() *types.MsgMintNonFungibleToken {
				msg := validaMessage
				msg.ClassId = "x"
				return &msg
			},
			expectedError: types.ErrInvalidNonFungibleToken,
		},
		{
			name: "invalid uri",
			messageFunc: func() *types.MsgMintNonFungibleToken {
				msg := validaMessage
				msg.Uri = string(make([]byte, 257))
				return &msg
			},
			expectedError: types.ErrInvalidNonFungibleToken,
		},
		{
			name: "invalid uri hash",
			messageFunc: func() *types.MsgMintNonFungibleToken {
				msg := validaMessage
				msg.UriHash = string(make([]byte, 129))
				return &msg
			},
			expectedError: types.ErrInvalidNonFungibleToken,
		},
		{
			name: "invalid data",
			messageFunc: func() *types.MsgMintNonFungibleToken {
				longDataString := string(make([]byte, 5001))
				longDataValue, err := codetypes.NewAnyWithValue(&gogotypes.BytesValue{Value: []byte(longDataString)})
				requireT.NoError(err)
				msg := validaMessage
				msg.Data = longDataValue
				return &msg
			},
			expectedError: types.ErrInvalidNonFungibleToken,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			assertT := assert.New(t)
			err := tc.messageFunc().ValidateBasic()
			if tc.expectedError == nil {
				assertT.NoError(err)
			} else {
				assertT.True(sdkerrors.IsOf(err, tc.expectedError))
			}
		})
	}
}
