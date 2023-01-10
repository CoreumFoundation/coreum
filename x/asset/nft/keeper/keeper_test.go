package keeper_test

import (
	"strings"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/testutil/event"
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
	dataValue, err := codectypes.NewAnyWithValue(&types.DataBytes{Data: []byte(dataString)})
	requireT.NoError(err)
	settings := types.IssueClassSettings{
		Issuer:      addr,
		Name:        "name",
		Symbol:      "Symbol",
		Description: "description",
		URI:         "https://my-class-meta.invalid/1",
		URIHash:     "content-hash",
		Data:        dataValue,
	}

	classID, err := nftKeeper.IssueClass(ctx, settings)
	requireT.NoError(err)
	requireT.EqualValues(strings.ToLower(settings.Symbol)+"-"+addr.String(), classID)

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
	settings.Symbol = "SYMBOL"
	_, err = nftKeeper.IssueClass(ctx, settings)
	requireT.True(types.ErrInvalidInput.Is(err))
}

func TestKeeper_Mint(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	nftKeeper := testApp.AssetNFTKeeper
	bankKeeper := testApp.BankKeeper

	nftParams := types.Params{
		MintFee: sdk.NewInt64Coin(constant.DenomDev, 10_000_000),
	}
	nftKeeper.SetParams(ctx, nftParams)

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	requireT.NoError(testApp.FundAccount(ctx, addr, sdk.NewCoins(nftParams.MintFee)))
	classSettings := types.IssueClassSettings{
		Issuer: addr,
		Symbol: "symbol",
	}

	classID, err := nftKeeper.IssueClass(ctx, classSettings)
	requireT.NoError(err)
	requireT.EqualValues(classSettings.Symbol+"-"+addr.String(), classID)

	dataString := "metadata"
	dataValue, err := codectypes.NewAnyWithValue(&types.DataBytes{Data: []byte(dataString)})
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

	// verify issue fee was burnt

	burntStr, err := event.FindStringEventAttribute(ctx.EventManager().ABCIEvents(), banktypes.EventTypeCoinBurn, sdk.AttributeKeyAmount)
	requireT.NoError(err)
	requireT.Equal(nftParams.MintFee.String(), burntStr)

	// check that balance is 0 meaning issue fee was taken

	balance := bankKeeper.GetBalance(ctx, addr, constant.DenomDev)
	requireT.Equal(sdk.ZeroInt().String(), balance.Amount.String())

	// mint second NFT with the same ID
	err = nftKeeper.Mint(ctx, settings)
	requireT.True(types.ErrInvalidInput.Is(err))

	// try to mint from not issuer account
	settings.Sender = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	err = nftKeeper.Mint(ctx, settings)
	requireT.True(sdkerrors.ErrUnauthorized.Is(err))
}

func TestKeeper_Mint_WithZeroMintFee(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	nftKeeper := testApp.AssetNFTKeeper

	nftParams := types.Params{
		MintFee: sdk.NewCoin(constant.DenomDev, sdk.ZeroInt()),
	}
	nftKeeper.SetParams(ctx, nftParams)

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	classSettings := types.IssueClassSettings{
		Issuer: addr,
		Symbol: "symbol",
	}

	classID, err := nftKeeper.IssueClass(ctx, classSettings)
	requireT.NoError(err)
	requireT.EqualValues(classSettings.Symbol+"-"+addr.String(), classID)

	requireT.NoError(err)
	settings := types.MintSettings{
		Sender:  addr,
		ClassID: classID,
		ID:      "my-id",
		URI:     "https://my-nft-meta.invalid/1",
		URIHash: "content-hash",
	}

	// mint NFT
	err = nftKeeper.Mint(ctx, settings)
	requireT.NoError(err)
}

func TestKeeper_Mint_WithNoFundsCoveringFee(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	nftKeeper := testApp.AssetNFTKeeper

	nftParams := types.Params{
		MintFee: sdk.NewInt64Coin(constant.DenomDev, 10_000_000),
	}
	nftKeeper.SetParams(ctx, nftParams)

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	classSettings := types.IssueClassSettings{
		Issuer: addr,
		Symbol: "symbol",
	}

	classID, err := nftKeeper.IssueClass(ctx, classSettings)
	requireT.NoError(err)
	requireT.EqualValues(classSettings.Symbol+"-"+addr.String(), classID)

	requireT.NoError(err)
	settings := types.MintSettings{
		Sender:  addr,
		ClassID: classID,
		ID:      "my-id",
		URI:     "https://my-nft-meta.invalid/1",
		URIHash: "content-hash",
	}

	// mint NFT
	requireT.ErrorIs(nftKeeper.Mint(ctx, settings), sdkerrors.ErrInsufficientFunds)
}
