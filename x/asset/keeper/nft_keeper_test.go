package keeper_test

import (
	"testing"

	codetypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestNonFungibleTokenKeeper_CreateNonFungibleTokenClass(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	nftKeeper := testApp.AssetNonFungibleTokenKeeper

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	dataString := "metadata"
	dataValue, err := codetypes.NewAnyWithValue(&gogotypes.BytesValue{Value: []byte(dataString)})
	requireT.NoError(err)
	settings := types.CreateNonFungibleTokenClassSettings{
		Creator:     addr,
		Name:        "name",
		Symbol:      "symbol",
		Description: "description",
		URI:         "https://my-class-meta.int/1",
		URIHash:     "35b326a2b3b605270c26185c38d2581e937b2eae0418b4964ef521efe79cdf34", // sha256
		Data:        dataValue,
	}

	classID, err := nftKeeper.CreateClass(ctx, settings)
	requireT.NoError(err)
	requireT.EqualValues(settings.Symbol+"-"+addr.String(), classID)

	class, found := testApp.NFTKeeper.GetClass(ctx, classID)
	requireT.True(found)
	// we check line by line because of the data field
	requireT.Equal(settings.Name, class.Name)
	requireT.Equal(settings.Symbol, class.Symbol)
	requireT.Equal(settings.Description, class.Description)
	requireT.Equal(settings.URI, class.Uri)
	requireT.Equal(settings.URIHash, class.UriHash)
	requireT.Equal(string(settings.Data.Value), string(class.Data.Value))

	// try to duplicate
	_, err = nftKeeper.CreateClass(ctx, settings)
	requireT.True(types.ErrInvalidNonFungibleTokenClass.Is(err))
}

func TestNonFungibleTokenKeeper_MintNonFungibleToken(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	nftKeeper := testApp.AssetNonFungibleTokenKeeper

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	classSettings := types.CreateNonFungibleTokenClassSettings{
		Creator: addr,
		Symbol:  "symbol",
	}

	classID, err := nftKeeper.CreateClass(ctx, classSettings)
	requireT.NoError(err)
	requireT.EqualValues(classSettings.Symbol+"-"+addr.String(), classID)

	dataString := "metadata"
	dataValue, err := codetypes.NewAnyWithValue(&gogotypes.BytesValue{Value: []byte(dataString)})
	requireT.NoError(err)
	settings := types.MintNonFungibleTokenSettings{
		Sender:  addr,
		ClassID: classID,
		ID:      "my-id",
		URI:     "https://my-nft-meta.int/1",
		URIHash: "9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9", // sha256
		Data:    dataValue,
	}

	// mint first NFT
	err = nftKeeper.Mint(ctx, settings)
	requireT.NoError(err)

	nft, found := testApp.NFTKeeper.GetNFT(ctx, classID, settings.ID)
	requireT.True(found)
	// we check line by line because of the data field
	requireT.Equal(settings.ClassID, nft.ClassId)
	requireT.Equal(settings.ID, nft.Id)
	requireT.Equal(settings.URI, nft.Uri)
	requireT.Equal(settings.URIHash, nft.UriHash)
	requireT.Equal(string(settings.Data.Value), string(nft.Data.Value))

	nftOwner := testApp.NFTKeeper.GetOwner(ctx, classID, settings.ID)
	requireT.Equal(addr, nftOwner)

	// mint second NFT with the same ID
	err = nftKeeper.Mint(ctx, settings)
	requireT.True(types.ErrInvalidNonFungibleToken.Is(err))

	// try to min from not creator account
	settings.Sender = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	err = nftKeeper.Mint(ctx, settings)
	requireT.True(sdkerrors.ErrUnauthorized.Is(err))
}
