package v1_test

import (
	"encoding/base64"
	"os"
	"testing"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	v1 "github.com/CoreumFoundation/coreum/x/asset/nft/legacy/v1"
	assetnfttypes "github.com/CoreumFoundation/coreum/x/asset/nft/types"
	"github.com/CoreumFoundation/coreum/x/nft"
	"github.com/cosmos/cosmos-sdk/simapp"
)

func TestMigrateWasmCreatedNFTData(t *testing.T) {
	requireT := require.New(t)
	simapp.FlagEnabledValue = true
	_, db, dir, logger, _, err := simapp.SetupSimulation("goleveldb-app-sim", "Simulation")
	requireT.NoError(err, "simulation setup failed")

	t.Cleanup(func() {
		db.Close()
		err = os.RemoveAll(dir)
		requireT.NoError(err)
	})

	encoding := config.NewEncodingConfig(app.ModuleBasics)
	network, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	if err != nil {
		panic(err)
	}
	network.SetSDKConfig()

	app.ChosenNetwork = network
	simApp := app.New(
		logger,
		db,
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		0,
		encoding,
		simapp.EmptyAppOptions{},
	)

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
			issuer: wasmkeeper.BuildContractAddressClassic(1, 12),
			data:   []byte("some data"),
			symbol: "smartsymbol",
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
			name:   "issuer is normal user",
			issuer: sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address()),
			symbol: "normalsymbol",
			nftID:  "nft1",
			data:   []byte("some data"),
			assertion: func(t *testing.T, class nft.Class, nft nft.NFT, data []byte) {
				requireT := require.New(t)
				// we write the encoded data, after migration the same encoded data should be returned
				// because issuer is not smart contract
				encodedBytes := base64.StdEncoding.EncodeToString(data)
				dataBytes := &assetnfttypes.DataBytes{Data: []byte(encodedBytes)}
				dataAny, err := codectypes.NewAnyWithValue(dataBytes)
				requireT.NoError(err)
				// check class
				requireT.EqualValues(class.Data.Value, dataAny.Value)
				// check nft
				requireT.EqualValues(dataAny.Value, nft.Data.Value)
			},
		},
	}

	for _, tc := range testCases {
		encodedData := base64.StdEncoding.EncodeToString(tc.data)
		encodedDataBytes := &assetnfttypes.DataBytes{Data: []byte(encodedData)}
		encodedDataAny, err := codectypes.NewAnyWithValue(encodedDataBytes)
		requireT.NoError(err)
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
	err = v1.MigrateWasmCreatedNFTData(ctx, simApp.NFTKeeper.Keeper, simApp.AssetNFTKeeper)
	requireT.NoError(err)

	// run assertions
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			classID := assetnfttypes.BuildClassID(tc.symbol, tc.issuer)
			class, found := simApp.NFTKeeper.GetClass(ctx, classID)
			requireT.True(found)

			nft, found := simApp.NFTKeeper.GetNFT(ctx, classID, tc.nftID)
			requireT.True(found)
			tc.assertion(t, class, nft, tc.data)
		})
	}
}
