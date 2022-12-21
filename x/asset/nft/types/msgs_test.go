package types_test

import (
	"strings"
	"testing"

	codetypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

func TestMain(m *testing.M) {
	n, err := config.NetworkByChainID(constant.ChainIDDev)
	if err != nil {
		panic(err)
	}
	n.SetSDKConfig()
	m.Run()
}

//nolint:funlen // many test cases
func TestMsgIssueClass_ValidateBasic(t *testing.T) {
	requireT := require.New(t)

	dataString := "metadata"
	dataValue, err := codetypes.NewAnyWithValue(&gogotypes.BytesValue{Value: []byte(dataString)})
	requireT.NoError(err)

	validMessage := types.MsgIssueClass{
		Issuer:      "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		Name:        "name",
		Symbol:      "Symbol",
		Description: "description",
		URI:         "https://my.invalid",
		URIHash:     "sha-hash",
		Data:        dataValue,
	}
	testCases := []struct {
		name          string
		messageFunc   func() *types.MsgIssueClass
		expectedError error
	}{
		{
			name: "valid msg",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				return &msg
			},
		},
		{
			name: "invalid issuer",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.Issuer = "devcore172rc5sz2uc"
				return &msg
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid name",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.Name = strings.Repeat("x", 129)
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid empty symbol",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.Symbol = ""
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid char symbol",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.Symbol = "#x#"
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid description",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.Description = string(make([]byte, 257))
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid uri",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.URI = string(make([]byte, 257))
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid uri hash",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.URIHash = strings.Repeat("x", 129)
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid data",
			messageFunc: func() *types.MsgIssueClass {
				longDataString := string(make([]byte, 5001))
				longDataValue, err := codetypes.NewAnyWithValue(&gogotypes.BytesValue{Value: []byte(longDataString)})
				requireT.NoError(err)
				msg := validMessage
				msg.Data = longDataValue
				return &msg
			},
			expectedError: types.ErrInvalidInput,
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

func TestMsgMint_ValidateBasic(t *testing.T) {
	requireT := require.New(t)

	dataString := "metadata"
	dataValue, err := codetypes.NewAnyWithValue(&gogotypes.BytesValue{Value: []byte(dataString)})
	requireT.NoError(err)

	validMessage := types.MsgMint{
		Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		ID:      "my-id",
		ClassID: "symbol-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		URI:     "https://my.invalid",
		URIHash: "content-hash",
		Data:    dataValue,
	}
	testCases := []struct {
		name          string
		messageFunc   func() *types.MsgMint
		expectedError error
	}{
		{
			name: "valid msg",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				return &msg
			},
		},
		{
			name: "invalid id",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.ID = "id?"
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid sender",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.Sender = "devcore172rc5sz2uc"
				return &msg
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid classID",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.ClassID = "x"
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid uri",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.URI = string(make([]byte, 257))
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid uri hash",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.URIHash = strings.Repeat("x", 129)
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid data",
			messageFunc: func() *types.MsgMint {
				longDataString := string(make([]byte, 5001))
				longDataValue, err := codetypes.NewAnyWithValue(&gogotypes.BytesValue{Value: []byte(longDataString)})
				requireT.NoError(err)
				msg := validMessage
				msg.Data = longDataValue
				return &msg
			},
			expectedError: types.ErrInvalidInput,
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
