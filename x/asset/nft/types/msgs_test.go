package types_test

import (
	"bytes"
	"strings"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

const (
	invalidNFTID   = "invalid-id?"
	invalidAccount = "devcore172rx"
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
	dataValue, err := codectypes.NewAnyWithValue(&types.DataBytes{Data: []byte(dataString)})
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
			name: "valid msg with max data size",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.Data = &codectypes.Any{
					TypeUrl: "/" + proto.MessageName((*types.DataBytes)(nil)),
					Value:   bytes.Repeat([]byte{0x01}, types.MaxDataSize),
				}
				return &msg
			},
		},
		{
			name: "valid msg with nil data",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.Data = nil
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
			name: "invalid data - too long",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.Data = &codectypes.Any{
					TypeUrl: "/" + proto.MessageName((*types.DataBytes)(nil)),
					Value:   bytes.Repeat([]byte{0x01}, types.MaxDataSize+1),
				}
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid data - wrong type",
			messageFunc: func() *types.MsgIssueClass {
				dataValue, err := codectypes.NewAnyWithValue(&types.MsgIssueClass{})
				requireT.NoError(err)
				msg := validMessage
				msg.Data = dataValue
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

//nolint:funlen
func TestMsgMint_ValidateBasic(t *testing.T) {
	requireT := require.New(t)

	dataString := "metadata"
	dataValue, err := codectypes.NewAnyWithValue(&types.DataBytes{Data: []byte(dataString)})
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
			name: "valid msg with max data size",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.Data = &codectypes.Any{
					TypeUrl: "/" + proto.MessageName((*types.DataBytes)(nil)),
					Value:   bytes.Repeat([]byte{0x01}, types.MaxDataSize),
				}
				return &msg
			},
		},
		{
			name: "valid msg with nil data",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.Data = nil
				return &msg
			},
		},
		{
			name: "invalid id",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.ID = invalidNFTID
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid sender",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.Sender = invalidAccount
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
			name: "invalid data - too long",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.Data = &codectypes.Any{
					TypeUrl: "/" + proto.MessageName((*types.DataBytes)(nil)),
					Value:   bytes.Repeat([]byte{0x01}, types.MaxDataSize+1),
				}
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid data - wrong type",
			messageFunc: func() *types.MsgMint {
				dataValue, err := codectypes.NewAnyWithValue(&types.MsgIssueClass{})
				requireT.NoError(err)
				msg := validMessage
				msg.Data = dataValue
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

//nolint:dupl // test case duplicates are ok
func TestMsgBurn_ValidateBasic(t *testing.T) {
	validMessage := types.MsgBurn{
		Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		ClassID: "symbol-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		ID:      "my-id",
	}
	testCases := []struct {
		name          string
		messageFunc   func() *types.MsgBurn
		expectedError error
	}{
		{
			name: "valid msg",
			messageFunc: func() *types.MsgBurn {
				msg := validMessage
				return &msg
			},
		},
		{
			name: "invalid id",
			messageFunc: func() *types.MsgBurn {
				msg := validMessage
				msg.ID = invalidNFTID
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid sender",
			messageFunc: func() *types.MsgBurn {
				msg := validMessage
				msg.Sender = invalidAccount
				return &msg
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid classID",
			messageFunc: func() *types.MsgBurn {
				msg := validMessage
				msg.ClassID = "x"
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

//nolint:dupl // test case duplicates are ok
func TestMsgFreeze_ValidateBasic(t *testing.T) {
	validMessage := types.MsgFreeze{
		Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		ClassID: "symbol-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		ID:      "my-id",
	}
	testCases := []struct {
		name          string
		messageFunc   func() *types.MsgFreeze
		expectedError error
	}{
		{
			name: "valid msg",
			messageFunc: func() *types.MsgFreeze {
				msg := validMessage
				return &msg
			},
		},
		{
			name: "invalid id",
			messageFunc: func() *types.MsgFreeze {
				msg := validMessage
				msg.ID = invalidNFTID
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid sender",
			messageFunc: func() *types.MsgFreeze {
				msg := validMessage
				msg.Sender = invalidAccount
				return &msg
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid classID",
			messageFunc: func() *types.MsgFreeze {
				msg := validMessage
				msg.ClassID = "x"
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

//nolint:dupl // test case duplicates are ok
func TestMsgUnfreeze_ValidateBasic(t *testing.T) {
	validMessage := types.MsgUnfreeze{
		Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		ClassID: "symbol-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		ID:      "my-id",
	}
	testCases := []struct {
		name          string
		messageFunc   func() *types.MsgUnfreeze
		expectedError error
	}{
		{
			name: "valid msg",
			messageFunc: func() *types.MsgUnfreeze {
				msg := validMessage
				return &msg
			},
		},
		{
			name: "invalid id",
			messageFunc: func() *types.MsgUnfreeze {
				msg := validMessage
				msg.ID = invalidNFTID
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid sender",
			messageFunc: func() *types.MsgUnfreeze {
				msg := validMessage
				msg.Sender = invalidAccount
				return &msg
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid classID",
			messageFunc: func() *types.MsgUnfreeze {
				msg := validMessage
				msg.ClassID = "x"
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

//nolint:dupl // test case duplicates are ok
func TestMsgAddToWhitelist_ValidateBasic(t *testing.T) {
	validMessage := types.MsgAddToWhitelist{
		Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		Account: "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		ClassID: "symbol-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		ID:      "my-id",
	}
	testCases := []struct {
		name          string
		messageFunc   func() *types.MsgAddToWhitelist
		expectedError error
	}{
		{
			name: "valid msg",
			messageFunc: func() *types.MsgAddToWhitelist {
				msg := validMessage
				return &msg
			},
		},
		{
			name: "invalid id",
			messageFunc: func() *types.MsgAddToWhitelist {
				msg := validMessage
				msg.ID = invalidNFTID
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid sender",
			messageFunc: func() *types.MsgAddToWhitelist {
				msg := validMessage
				msg.Sender = invalidAccount
				return &msg
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid account",
			messageFunc: func() *types.MsgAddToWhitelist {
				msg := validMessage
				msg.Account = "devcore172"
				return &msg
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid classID",
			messageFunc: func() *types.MsgAddToWhitelist {
				msg := validMessage
				msg.ClassID = "x"
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

//nolint:dupl // test case duplicates are ok
func TestMsgRemoveFromWhitelist_ValidateBasic(t *testing.T) {
	validMessage := types.MsgRemoveFromWhitelist{
		Sender:  "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		Account: "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		ClassID: "symbol-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5",
		ID:      "my-id",
	}
	testCases := []struct {
		name          string
		messageFunc   func() *types.MsgRemoveFromWhitelist
		expectedError error
	}{
		{
			name: "valid msg",
			messageFunc: func() *types.MsgRemoveFromWhitelist {
				msg := validMessage
				return &msg
			},
		},
		{
			name: "invalid id",
			messageFunc: func() *types.MsgRemoveFromWhitelist {
				msg := validMessage
				msg.ID = invalidNFTID
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid sender",
			messageFunc: func() *types.MsgRemoveFromWhitelist {
				msg := validMessage
				msg.Sender = invalidAccount
				return &msg
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid account",
			messageFunc: func() *types.MsgRemoveFromWhitelist {
				msg := validMessage
				msg.Account = "devcore172"
				return &msg
			},
			expectedError: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid classID",
			messageFunc: func() *types.MsgRemoveFromWhitelist {
				msg := validMessage
				msg.ClassID = "x"
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
