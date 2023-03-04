package nft_test

import (
	"fmt"
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/nft"
	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
	rawnft "github.com/CoreumFoundation/coreum/x/nft"
)

//nolint:funlen
func TestInitAndExportGenesis(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	testApp := simapp.New()

	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})
	nftKeeper := testApp.AssetNFTKeeper

	// prepare the genesis data

	rawGenState := &rawnft.GenesisState{}

	// class definitions
	var classDefinitions []types.ClassDefinition
	for i := 0; i < 5; i++ {
		classDefinition := types.ClassDefinition{
			ID: fmt.Sprintf("class-id-%d", i),
			Features: []types.ClassFeature{
				types.ClassFeature_freezing,
				types.ClassFeature_whitelisting,
			},
			RoyaltyRate: sdk.MustNewDecFromStr(fmt.Sprintf("0.%d", (i+1)%10)),
		}

		rawGenState.Classes = append(rawGenState.Classes, &rawnft.Class{
			Id:     classDefinition.ID,
			Name:   fmt.Sprintf("name-%d", i),
			Symbol: fmt.Sprintf("symbol-%d", i),
		})
		classDefinitions = append(classDefinitions, classDefinition)
	}

	for i := 0; i < 5; i++ {
		rawGenState.Entries = append(rawGenState.Entries, &rawnft.Entry{
			Owner: sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
			Nfts: []*rawnft.NFT{
				{
					ClassId: fmt.Sprintf("class-id-%d", i),
					Id:      fmt.Sprintf("nft-id-1-%d", i),
				},
				{
					ClassId: fmt.Sprintf("class-id-%d", i),
					Id:      fmt.Sprintf("nft-id-2-%d", i),
				},
			},
		})
	}

	// Frozen NFTs
	var frozen []types.FrozenNFT
	for i := 0; i < 5; i++ {
		frozen = append(frozen, types.FrozenNFT{
			ClassID: fmt.Sprintf("class-id-%d", i),
			NftIDs: []string{
				fmt.Sprintf("nft-id-1-%d", i),
				fmt.Sprintf("nft-id-2-%d", i),
			},
		})
	}

	// Whitelisting
	var whitelisted []types.WhitelistedNFTAccounts
	for i := 0; i < 5; i++ {
		whitelisted = append(whitelisted, types.WhitelistedNFTAccounts{
			ClassID: fmt.Sprintf("class-id-%d", i),
			NftID:   fmt.Sprintf("nft-id-1-%d", i),
			Accounts: []string{
				sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
				sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
				sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
			},
		},
			types.WhitelistedNFTAccounts{
				ClassID: fmt.Sprintf("class-id-%d", i),
				NftID:   fmt.Sprintf("nft-id-2-%d", i),
				Accounts: []string{
					sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
					sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
					sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
				},
			})
	}

	// Burnt NFTs
	var burnt []types.BurntNFT
	for i := 0; i < 5; i++ {
		burnt = append(burnt, types.BurntNFT{
			ClassID: fmt.Sprintf("class-id-%d", i),
			NftIDs: []string{
				fmt.Sprintf("burnt-nft-id-1-%d", i),
				fmt.Sprintf("burnt-nft-id-2-%d", i),
			},
		})
	}

	genState := types.GenesisState{
		Params:                 types.DefaultParams(),
		ClassDefinitions:       classDefinitions,
		FrozenNFTs:             frozen,
		WhitelistedNFTAccounts: whitelisted,
		BurntNFTs:              burnt,
	}

	// init the keeper
	testApp.NFTKeeper.InitGenesis(ctx, rawGenState)
	nft.InitGenesis(ctx, nftKeeper, genState)

	// assert the keeper state

	// class definitions
	for _, definition := range classDefinitions {
		storedDefinition, err := nftKeeper.GetClassDefinition(ctx, definition.ID)
		requireT.NoError(err)
		assertT.EqualValues(definition, storedDefinition)
	}

	// params
	params := nftKeeper.GetParams(ctx)
	assertT.EqualValues(types.DefaultParams(), params)

	// check that export is equal import
	exportedGenState := nft.ExportGenesis(ctx, nftKeeper)
	assertT.ElementsMatch(genState.ClassDefinitions, exportedGenState.ClassDefinitions)
	assertT.ElementsMatch(genState.FrozenNFTs, exportedGenState.FrozenNFTs)

	for _, st := range genState.WhitelistedNFTAccounts {
		sort.Strings(st.Accounts)
	}
	for _, st := range exportedGenState.WhitelistedNFTAccounts {
		sort.Strings(st.Accounts)
	}
	assertT.ElementsMatch(genState.WhitelistedNFTAccounts, exportedGenState.WhitelistedNFTAccounts)
	assertT.ElementsMatch(genState.BurntNFTs, exportedGenState.BurntNFTs)
}
