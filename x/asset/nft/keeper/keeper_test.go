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
		Features: []types.ClassFeature{
			types.ClassFeature_burning, //nolint:nosnakecase // generated variable
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
	requireT.Equal([]types.ClassFeature{types.ClassFeature_burning}, nftClass.Features) //nolint:nosnakecase // generated variable

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
			types.ClassFeature_burning, //nolint:nosnakecase // generated variable
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

	// mint NFT
	err = assetNFTKeeper.Mint(ctx, settings)
	requireT.NoError(err)

	err = nftKeeper.Transfer(ctx, settings.ClassID, settings.ID, recipient)
	requireT.NoError(err)

	// try burn the nft with the enabled feature from the recipient account
	err = assetNFTKeeper.Burn(ctx, recipient, classID, nftID)
	requireT.NoError(err)

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
	requireT.ErrorIs(types.ErrFeatureDisabled, err)
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

func TestKeeper_Freeze(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	assetNFTKeeper := testApp.AssetNFTKeeper
	nftKeeper := testApp.NFTKeeper

	nftParams := types.Params{
		MintFee: sdk.NewInt64Coin(constant.DenomDev, 0),
	}
	assetNFTKeeper.SetParams(ctx, nftParams)

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	classSettings := types.IssueClassSettings{
		Issuer: issuer,
		Symbol: "symbol",
		Features: []types.ClassFeature{
			types.ClassFeature_freezing, //nolint:nosnakecase
		},
	}

	classID, err := assetNFTKeeper.IssueClass(ctx, classSettings)
	requireT.NoError(err)

	requireT.NoError(err)
	settings := types.MintSettings{
		Sender:  issuer,
		ClassID: classID,
		ID:      "my-id",
		URI:     "https://my-nft-meta.invalid/1",
		URIHash: "content-hash",
	}

	// mint NFT
	requireT.NoError(assetNFTKeeper.Mint(ctx, settings))

	// freeze NFT
	nftID := settings.ID
	requireT.NoError(assetNFTKeeper.Freeze(ctx, issuer, classID, nftID))
	isFrozen, err := assetNFTKeeper.IsFrozen(ctx, classID, nftID)
	requireT.NoError(err)
	requireT.True(isFrozen)

	// transfer from issuer (although it is frozen, the issuer can send)
	recipient := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	err = nftKeeper.Transfer(ctx, classID, nftID, recipient)
	requireT.NoError(err)

	// transfer from non-issuer (must fail)
	recipient2 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	err = nftKeeper.Transfer(ctx, classID, nftID, recipient2)
	requireT.Error(err)
	requireT.True(sdkerrors.ErrUnauthorized.Is(err))

	// unfreeze
	requireT.NoError(assetNFTKeeper.Unfreeze(ctx, issuer, classID, nftID))
	isFrozen, err = assetNFTKeeper.IsFrozen(ctx, classID, nftID)
	requireT.NoError(err)
	requireT.False(isFrozen)

	// transfer from non-issuer (must succeed)
	err = nftKeeper.Transfer(ctx, classID, nftID, recipient2)
	requireT.NoError(err)
}

func TestKeeper_Freeze_Unfreezable(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	assetNFTKeeper := testApp.AssetNFTKeeper

	nftParams := types.Params{
		MintFee: sdk.NewInt64Coin(constant.DenomDev, 0),
	}
	assetNFTKeeper.SetParams(ctx, nftParams)

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	classSettings := types.IssueClassSettings{
		Issuer:   issuer,
		Symbol:   "symbol",
		Features: []types.ClassFeature{},
	}

	classID, err := assetNFTKeeper.IssueClass(ctx, classSettings)
	requireT.NoError(err)

	requireT.NoError(err)
	settings := types.MintSettings{
		Sender:  issuer,
		ClassID: classID,
		ID:      "my-id",
		URI:     "https://my-nft-meta.invalid/1",
		URIHash: "content-hash",
	}

	// mint NFT
	requireT.NoError(assetNFTKeeper.Mint(ctx, settings))

	// freeze NFT
	nftID := settings.ID
	err = assetNFTKeeper.Freeze(ctx, issuer, classID, nftID)
	requireT.Error(err)
	requireT.True(types.ErrFeatureDisabled.Is(err))
}

func TestKeeper_Freeze_Nonexistent(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	assetNFTKeeper := testApp.AssetNFTKeeper
	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	nftParams := types.Params{
		MintFee: sdk.NewInt64Coin(constant.DenomDev, 0),
	}
	assetNFTKeeper.SetParams(ctx, nftParams)

	// try to freeze NFT when Class does not exists
	err := assetNFTKeeper.Freeze(ctx, issuer, types.BuildClassID("symbol", issuer), "random-id")
	requireT.Error(err)
	requireT.True(types.ErrClassNotFound.Is(err))

	// issue class
	classSettings := types.IssueClassSettings{
		Issuer: issuer,
		Symbol: "symbol",
		Features: []types.ClassFeature{
			types.ClassFeature_freezing, //nolint:nosnakecase
		},
	}

	classID, err := assetNFTKeeper.IssueClass(ctx, classSettings)
	requireT.NoError(err)

	// try to freeze when NFT does not exists
	err = assetNFTKeeper.Freeze(ctx, issuer, classID, "random-id")
	requireT.Error(err)
	requireT.True(types.ErrNFTNotFound.Is(err))
}
