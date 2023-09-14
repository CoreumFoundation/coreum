package v1_test

import (
	"encoding/base64"
	"testing"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v3/testutil/simapp"
	v1 "github.com/CoreumFoundation/coreum/v3/x/asset/nft/migrations/v1"
	assetnfttypes "github.com/CoreumFoundation/coreum/v3/x/asset/nft/types"
	"github.com/CoreumFoundation/coreum/v3/x/nft"
)

func TestMigrateWasmCreatedNFTData(t *testing.T) {
	requireT := require.New(t)

	simApp := simapp.New()

	ctx := simApp.NewContext(true, tmproto.Header{})

	testCases := []struct {
		name      string
		issuer    sdk.AccAddress
		symbol    string
		nftID     string
		data      []byte
		assertion func(*testing.T, nft.Class, nft.NFT, []byte)
	}{
		{
			name:   "issuer is smart contract",
			issuer: wasmkeeper.BuildContractAddressClassic(1, 1),
			data:   []byte("some data"),
			symbol: "symbol1",
			nftID:  "nft1",
			assertion: func(t *testing.T, class nft.Class, nft nft.NFT, data []byte) {
				requireT := require.New(t)
				// we write the encoded data, after migration the raw data (not encoded) should be returned
				dataBytes := &assetnfttypes.DataBytes{Data: data}
				dataAny, err := codectypes.NewAnyWithValue(dataBytes)
				requireT.NoError(err)
				// check class
				requireT.EqualValues(class.Data.Value, dataAny.Value)
				// check nft
				requireT.EqualValues(dataAny.Value, nft.Data.Value)
			},
		},
		{
			name:   "nil data",
			issuer: wasmkeeper.BuildContractAddressClassic(1, 2),
			data:   nil,
			symbol: "symbol2",
			nftID:  "nft1",
			assertion: func(t *testing.T, class nft.Class, nft nft.NFT, data []byte) {
				requireT := require.New(t)
				requireT.Nil(class.Data)
				requireT.Nil(nft.Data)
			},
		},
		{
			name:   "issuer is normal user",
			issuer: sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address()),
			symbol: "symbol3",
			nftID:  "nft1",
			data:   []byte("some data"),
			assertion: func(t *testing.T, class nft.Class, nft nft.NFT, data []byte) {
				requireT := require.New(t)
				// we write the encoded data, after migration the same encoded data should be returned
				// because issuer is not smart contract
				dataAny := encodeDataToAny(t, data)
				// check class
				requireT.EqualValues(dataAny.Value, class.Data.Value)
				// check nft
				requireT.EqualValues(dataAny.Value, nft.Data.Value)
			},
		},
	}

	for _, tc := range testCases {
		encodedDataAny := encodeDataToAny(t, tc.data)
		issueMsg := assetnfttypes.IssueClassSettings{
			Issuer:      tc.issuer,
			Symbol:      tc.symbol,
			Name:        "sample name",
			Description: "some desc",
			Data:        encodedDataAny,
		}

		classID, err := simApp.AssetNFTKeeper.IssueClass(ctx, issueMsg)
		requireT.NoError(err)

		// mint nft
		mintNFT := nft.NFT{
			ClassId: classID,
			Id:      tc.nftID,
			Data:    encodedDataAny,
		}
		err = simApp.NFTKeeper.Mint(ctx, mintNFT, tc.issuer)
		requireT.NoError(err)
	}

	// migrate data
	err := v1.MigrateWasmCreatedNFTData(ctx, simApp.NFTKeeper.Keeper, simApp.AssetNFTKeeper, mockWasmKeeper{})
	requireT.NoError(err)

	// run assertions
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			requireT := require.New(t)
			classID := assetnfttypes.BuildClassID(tc.symbol, tc.issuer)
			class, found := simApp.NFTKeeper.GetClass(ctx, classID)
			requireT.True(found)

			nft, found := simApp.NFTKeeper.GetNFT(ctx, classID, tc.nftID)
			requireT.True(found)
			tc.assertion(t, class, nft, tc.data)
		})
	}
}

func encodeDataToAny(t *testing.T, data []byte) *codectypes.Any {
	if data == nil {
		return nil
	}
	encodedData := base64.StdEncoding.EncodeToString(data)
	encodedDataBytes := &assetnfttypes.DataBytes{Data: []byte(encodedData)}
	encodedDataAny, err := codectypes.NewAnyWithValue(encodedDataBytes)
	require.NoError(t, err)
	return encodedDataAny
}

type mockWasmKeeper struct{}

func (m mockWasmKeeper) HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool {
	return isSmartContractAddress(contractAddress)
}

func isSmartContractAddress(address sdk.AccAddress) bool {
	return len(address) == wasmtypes.ContractAddrLen
}
