package types_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/pkg/config"
	"github.com/CoreumFoundation/coreum/v4/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/v4/x/asset/nft/types"
)

const (
	invalidNFTID   = "invalid-id?"
	invalidAccount = "devcore172rx"
)

func TestMain(m *testing.M) {
	n, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	if err != nil {
		panic(err)
	}
	n.SetSDKConfig()
	m.Run()
}

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
			name: "valid_msg",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				return &msg
			},
		},
		{
			name: "valid_msg_with_nil_data",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.Data = nil
				return &msg
			},
		},
		{
			name: "invalid_issuer",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.Issuer = "devcore172rc5sz2uc"
				return &msg
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
		{
			name: "invalid_name",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.Name = strings.Repeat("x", 129)
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_empty_symbol",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.Symbol = ""
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_char_symbol",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.Symbol = "#x#"
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_description",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.Description = string(make([]byte, 257))
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_uri",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.URI = string(make([]byte, 257))
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_uri_hash",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.URIHash = strings.Repeat("x", 129)
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_data_wrong_type",
			messageFunc: func() *types.MsgIssueClass {
				dataValue, err := codectypes.NewAnyWithValue(&types.DataDynamic{})
				requireT.NoError(err)
				msg := validMessage
				msg.Data = dataValue
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_duplicated_class_feature",
			messageFunc: func() *types.MsgIssueClass {
				msg := validMessage
				msg.Features = []types.ClassFeature{
					types.ClassFeature_burning,
					types.ClassFeature_whitelisting,
					types.ClassFeature_burning,
				}
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			requireT := require.New(t)
			err := tc.messageFunc().ValidateBasic()
			if tc.expectedError == nil {
				requireT.NoError(err)
			} else {
				requireT.True(sdkerrors.IsOf(err, tc.expectedError))
			}
		})
	}
}

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
			name: "valid_msg",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				return &msg
			},
		},
		{
			name: "valid_msg_with_nil_data",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.Data = nil
				return &msg
			},
		},
		{
			name: "invalid_id",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.ID = invalidNFTID
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_sender",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.Sender = invalidAccount
				return &msg
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
		{
			name: "invalid_classID",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.ClassID = "x"
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_uri",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.URI = string(make([]byte, 257))
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_uri_hash",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.URIHash = strings.Repeat("x", 129)
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_data_wrong_type",
			messageFunc: func() *types.MsgMint {
				dataValue, err := codectypes.NewAnyWithValue(&types.MsgIssueClass{})
				requireT.NoError(err)
				msg := validMessage
				msg.Data = dataValue
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "valid_msg_with_dynamic_data",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.Data = func() *codectypes.Any {
					dataDynamic := types.DataDynamic{
						Items: []types.DataDynamicItem{
							{
								// no editors
								Editors: []types.DataEditor{},
								Data:    bytes.Repeat([]byte{0x01}, 5),
							},
							{
								// admin
								Editors: []types.DataEditor{
									types.DataEditor_admin,
								},
								Data: nil,
							},
							{
								// owner
								Editors: []types.DataEditor{
									types.DataEditor_admin,
								},
								Data: bytes.Repeat([]byte{0x01}, 5),
							},
							{
								// admin and owner
								Editors: []types.DataEditor{
									types.DataEditor_admin,
									types.DataEditor_owner,
								},
								Data: bytes.Repeat([]byte{0x01}, 5),
							},
						},
					}

					dataBytes, err := dataDynamic.Marshal()
					requireT.NoError(err)

					return &codectypes.Any{
						TypeUrl: "/" + proto.MessageName((*types.DataDynamic)(nil)),
						Value:   dataBytes,
					}
				}()
				return &msg
			},
		},
		{
			name: "invalid_valid_msg_empty_item",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.Data = func() *codectypes.Any {
					dataDynamic := types.DataDynamic{}
					dataBytes, err := dataDynamic.Marshal()
					requireT.NoError(err)
					return &codectypes.Any{
						TypeUrl: "/" + proto.MessageName((*types.DataDynamic)(nil)),
						Value:   dataBytes,
					}
				}()
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_msg_duplicated_editor",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.Data = func() *codectypes.Any {
					dataDynamic := types.DataDynamic{
						Items: []types.DataDynamicItem{
							{
								Editors: []types.DataEditor{
									types.DataEditor_admin, types.DataEditor_admin,
								},
								Data: nil,
							},
						},
					}
					dataBytes, err := dataDynamic.Marshal()
					requireT.NoError(err)
					return &codectypes.Any{
						TypeUrl: "/" + proto.MessageName((*types.DataDynamic)(nil)),
						Value:   dataBytes,
					}
				}()
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_msg_not_existing_editor",
			messageFunc: func() *types.MsgMint {
				msg := validMessage
				msg.Data = func() *codectypes.Any {
					dataDynamic := types.DataDynamic{
						Items: []types.DataDynamicItem{
							{
								Editors: []types.DataEditor{
									types.DataEditor(12),
								},
								Data: nil,
							},
						},
					}
					dataBytes, err := dataDynamic.Marshal()
					requireT.NoError(err)
					return &codectypes.Any{
						TypeUrl: "/" + proto.MessageName((*types.DataDynamic)(nil)),
						Value:   dataBytes,
					}
				}()
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			requireT := require.New(t)
			err := tc.messageFunc().ValidateBasic()
			if tc.expectedError == nil {
				requireT.NoError(err)
			} else {
				requireT.True(sdkerrors.IsOf(err, tc.expectedError))
			}
		})
	}
}

func TestMsgUpdateData_ValidateBasic(t *testing.T) {
	sender := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	validMessage := types.MsgUpdateData{
		Sender:  sender.String(),
		ClassID: fmt.Sprintf("symbol-%s", sender.String()),
		ID:      "my-id",
		Items: []types.DataDynamicIndexedItem{
			{
				Index: 0,
				Data:  nil,
			},
			{
				Index: 1,
				Data:  nil,
			},
		},
	}
	testCases := []struct {
		name          string
		messageFunc   func() *types.MsgUpdateData
		expectedError error
	}{
		{
			name: "valid_msg",
			messageFunc: func() *types.MsgUpdateData {
				msg := validMessage
				return &msg
			},
		},
		{
			name: "invalid_id",
			messageFunc: func() *types.MsgUpdateData {
				msg := validMessage
				msg.ID = invalidNFTID
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_sender",
			messageFunc: func() *types.MsgUpdateData {
				msg := validMessage
				msg.Sender = invalidAccount
				return &msg
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
		{
			name: "invalid_classID",
			messageFunc: func() *types.MsgUpdateData {
				msg := validMessage
				msg.ClassID = "x"
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_empty_items",
			messageFunc: func() *types.MsgUpdateData {
				msg := validMessage
				msg.Items = nil
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
		{
			name: "invalid_duplicated_index",
			messageFunc: func() *types.MsgUpdateData {
				msg := validMessage
				msg.Items = []types.DataDynamicIndexedItem{
					{
						Index: 0,
						Data:  nil,
					},
					{
						Index: 0,
						Data:  nil,
					},
				}
				return &msg
			},
			expectedError: types.ErrInvalidInput,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			requireT := require.New(t)
			err := tc.messageFunc().ValidateBasic()
			if tc.expectedError == nil {
				requireT.NoError(err)
			} else {
				requireT.True(sdkerrors.IsOf(err, tc.expectedError))
			}
		})
	}
}

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
			expectedError: cosmoserrors.ErrInvalidAddress,
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
			requireT := require.New(t)
			err := tc.messageFunc().ValidateBasic()
			if tc.expectedError == nil {
				requireT.NoError(err)
			} else {
				requireT.True(sdkerrors.IsOf(err, tc.expectedError))
			}
		})
	}
}

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
			expectedError: cosmoserrors.ErrInvalidAddress,
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
			requireT := require.New(t)
			err := tc.messageFunc().ValidateBasic()
			if tc.expectedError == nil {
				requireT.NoError(err)
			} else {
				requireT.True(sdkerrors.IsOf(err, tc.expectedError))
			}
		})
	}
}

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
			expectedError: cosmoserrors.ErrInvalidAddress,
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
			requireT := require.New(t)
			err := tc.messageFunc().ValidateBasic()
			if tc.expectedError == nil {
				requireT.NoError(err)
			} else {
				requireT.True(sdkerrors.IsOf(err, tc.expectedError))
			}
		})
	}
}

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
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
		{
			name: "invalid account",
			messageFunc: func() *types.MsgAddToWhitelist {
				msg := validMessage
				msg.Account = "devcore172"
				return &msg
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
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
			requireT := require.New(t)
			err := tc.messageFunc().ValidateBasic()
			if tc.expectedError == nil {
				requireT.NoError(err)
			} else {
				requireT.True(sdkerrors.IsOf(err, tc.expectedError))
			}
		})
	}
}

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
			expectedError: cosmoserrors.ErrInvalidAddress,
		},
		{
			name: "invalid account",
			messageFunc: func() *types.MsgRemoveFromWhitelist {
				msg := validMessage
				msg.Account = "devcore172"
				return &msg
			},
			expectedError: cosmoserrors.ErrInvalidAddress,
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
			requireT := require.New(t)
			err := tc.messageFunc().ValidateBasic()
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
		msg           sdk.Msg
		wantAminoJSON string
	}{
		{
			name: sdk.MsgTypeURL(&types.MsgIssueClass{}),
			msg: &types.MsgIssueClass{
				Issuer: address,
				Symbol: "ABC",
			},
			wantAminoJSON: `{"type":"assetnft/MsgIssueClass","value":{"issuer":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5","royalty_rate":"0","symbol":"ABC"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgMint{}),
			msg: &types.MsgMint{
				Sender:  address,
				ClassID: "classID",
			},
			wantAminoJSON: `{"type":"assetnft/MsgMint","value":{"class_id":"classID","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgBurn{}),
			msg: &types.MsgBurn{
				Sender:  address,
				ClassID: "classID",
				ID:      "nftID",
			},
			wantAminoJSON: `{"type":"assetnft/MsgBurn","value":{"class_id":"classID","id":"nftID","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgFreeze{}),
			msg: &types.MsgFreeze{
				Sender:  address,
				ClassID: "classID",
				ID:      "nftID",
			},
			wantAminoJSON: `{"type":"assetnft/MsgFreeze","value":{"class_id":"classID","id":"nftID","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgUnfreeze{}),
			msg: &types.MsgUnfreeze{
				Sender:  address,
				ClassID: "classID",
				ID:      "nftID",
			},
			wantAminoJSON: `{"type":"assetnft/MsgUnfreeze","value":{"class_id":"classID","id":"nftID","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgAddToWhitelist{}),
			msg: &types.MsgAddToWhitelist{
				Sender:  address,
				ClassID: "classID",
				ID:      "nftID",
			},
			wantAminoJSON: `{"type":"assetnft/MsgAddToWhitelist","value":{"class_id":"classID","id":"nftID","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgRemoveFromWhitelist{}),
			msg: &types.MsgRemoveFromWhitelist{
				Sender:  address,
				ClassID: "classID",
				ID:      "nftID",
			},
			wantAminoJSON: `{"type":"assetnft/MsgRemoveFromWhitelist","value":{"class_id":"classID","id":"nftID","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
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
