package keeper_test

import (
	"strings"
	"testing"

	codetypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

func TestKeeper_IssueClass(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	nftKeeper := testApp.AssetNFTKeeper

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	dataString := "metadata"
	dataValue, err := codetypes.NewAnyWithValue(&gogotypes.BytesValue{Value: []byte(dataString)})
	requireT.NoError(err)
	settings := types.IssueClassSettings{
		Issuer:      addr,
		Name:        "name",
		Symbol:      "Symbol",
		Description: "description",
		URI:         "https://my-class-meta.invalid/1",
		URIHash:     "content-hash",
		Data:        dataValue,
		Features: []types.ClassFeature{
			types.ClassFeature_burn, //nolint:nosnakecase // generated variable
		},
	}

	classID, err := nftKeeper.IssueClass(ctx, settings)
	requireT.NoError(err)
	requireT.EqualValues(strings.ToLower(settings.Symbol)+"-"+addr.String(), classID)

	nativeNFTClass, found := testApp.NFTKeeper.GetClass(ctx, classID)
	requireT.True(found)
	// we check line by line because of the data field
	requireT.Equal(settings.Name, nativeNFTClass.Name)
	requireT.Equal(settings.Symbol, nativeNFTClass.Symbol)
	requireT.Equal(settings.Description, nativeNFTClass.Description)
	requireT.Equal(settings.URI, nativeNFTClass.Uri)
	requireT.Equal(settings.URIHash, nativeNFTClass.UriHash)
	requireT.Equal(string(settings.Data.Value), string(nativeNFTClass.Data.Value))

	nftClass, err := nftKeeper.GetClass(ctx, classID)
	requireT.NoError(err)

	// we check line by line because of the data field
	requireT.Equal(settings.Name, nftClass.Name)
	requireT.Equal(settings.Symbol, nftClass.Symbol)
	requireT.Equal(settings.Description, nftClass.Description)
	requireT.Equal(settings.URI, nftClass.URI)
	requireT.Equal(settings.URIHash, nftClass.URIHash)
	requireT.Equal(string(settings.Data.Value), string(nftClass.Data.Value))
	requireT.Equal([]types.ClassFeature{types.ClassFeature_burn}, nftClass.Features) //nolint:nosnakecase // generated variable

	// try to duplicate
	settings.Symbol = "SYMBOL"
	_, err = nftKeeper.IssueClass(ctx, settings)
	requireT.True(types.ErrInvalidInput.Is(err))

	// try to get non-existing class
	_, err = nftKeeper.GetClass(ctx, "invalid")
	requireT.ErrorIs(types.ErrClassNotFound, err)
}

func TestKeeper_Mint(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	nftKeeper := testApp.AssetNFTKeeper

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	classSettings := types.IssueClassSettings{
		Issuer: addr,
		Symbol: "symbol",
	}

	classID, err := nftKeeper.IssueClass(ctx, classSettings)
	requireT.NoError(err)
	requireT.EqualValues(classSettings.Symbol+"-"+addr.String(), classID)

	dataString := "metadata"
	dataValue, err := codetypes.NewAnyWithValue(&gogotypes.BytesValue{Value: []byte(dataString)})
	requireT.NoError(err)
	settings := types.MintSettings{
		Sender:  addr,
		ClassID: classID,
		ID:      "my-id",
		URI:     "https://my-nft-meta.invalid/1",
		URIHash: "content-hash",
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
	requireT.True(types.ErrInvalidInput.Is(err))

	// try to min from not issuer account
	settings.Sender = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	err = nftKeeper.Mint(ctx, settings)
	requireT.True(sdkerrors.ErrUnauthorized.Is(err))
}

func TestKeeper_Burn(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	assetNFTKeeper := testApp.AssetNFTKeeper
	nftKeeper := testApp.NFTKeeper

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	recipient := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	classSettings := types.IssueClassSettings{
		Issuer: issuer,
		Symbol: "symbol",
		Features: []types.ClassFeature{
			types.ClassFeature_burn, //nolint:nosnakecase // generated variable
		},
	}

	classID, err := assetNFTKeeper.IssueClass(ctx, classSettings)
	requireT.NoError(err)

	nftID := "my-id"
	settings := types.MintSettings{
		Sender:  issuer,
		ClassID: classID,
		ID:      nftID,
	}

	// mint NFT
	err = assetNFTKeeper.Mint(ctx, settings)
	requireT.NoError(err)

	// try to burn non-existing nft
	err = assetNFTKeeper.Burn(ctx, issuer, classID, "invalid")
	requireT.ErrorIs(types.ErrNFTNotFound, err)

	// try to burn from not owner account
	err = assetNFTKeeper.Burn(ctx, recipient, classID, nftID)
	requireT.ErrorIs(sdkerrors.ErrUnauthorized, err)

	// burn the nft
	err = assetNFTKeeper.Burn(ctx, issuer, classID, nftID)
	requireT.NoError(err)

	// try to burn the nft one more time
	err = assetNFTKeeper.Burn(ctx, issuer, classID, nftID)
	requireT.ErrorIs(types.ErrNFTNotFound, err)

	// issue class without burning feature
	classSettings = types.IssueClassSettings{
		Issuer: issuer,
		Symbol: "symbol2",
	}

	classID, err = assetNFTKeeper.IssueClass(ctx, classSettings)
	requireT.NoError(err)

	settings = types.MintSettings{
		Sender:  issuer,
		ClassID: classID,
		ID:      nftID,
	}

	// mint NFT
	err = assetNFTKeeper.Mint(ctx, settings)
	requireT.NoError(err)

	// try burn the nft with the disabled feature from the issuer account
	err = assetNFTKeeper.Burn(ctx, issuer, classID, nftID)
	requireT.NoError(err)

	// mint the nft one more time
	err = assetNFTKeeper.Mint(ctx, settings)
	requireT.NoError(err)

	err = nftKeeper.Transfer(ctx, settings.ClassID, settings.ID, recipient)
	requireT.NoError(err)

	// try burn the nft with the disabled feature from the recipient account
	err = assetNFTKeeper.Burn(ctx, recipient, classID, nftID)
	requireT.ErrorIs(sdkerrors.ErrUnauthorized, err)
}
