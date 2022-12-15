package keeper_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestMain(m *testing.M) {
	n, err := config.NetworkByChainID(constant.ChainIDDev)
	if err != nil {
		panic(err)
	}
	n.SetSDKConfig()
	m.Run()
}

func TestKeeper_ValidateSymbol(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})
	assetKeeper := testApp.AssetKeeper

	unacceptableSymbols := []string{
		"ABC/1",
		"core",
		"ucore",
		"Core",
		"uCore",
		"CORE",
		"UCORE",
		"3abc",
		"3ABC",
	}

	acceptableSymbols := []string{
		"btc-devcore1phjrez5j2wp5qzp0zvlqavasvw60mkp2zmfe6h",
		"BTC-devcore1phjrez5j2wp5qzp0zvlqavasvw60mkp2zmfe6h",
		"ABC-1",
		"ABC1",
		"coreum",
		"ucoreum",
		"Coreum",
		"uCoreum",
		"COREeum",
		"A1234567890123456789012345678901234567890123456789012345678901234567890",
		"AB1234567890123456789012345678901234567890123456789012345678901234567890",
	}

	assertValidSymbol := func(symbol string, isValid bool) {
		addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		settings := types.IssueFungibleTokenSettings{
			Issuer:        addr,
			Symbol:        symbol,
			Subunit:       "subunit",
			Description:   "ABC Desc",
			InitialAmount: sdk.NewInt(777),
			Features:      []types.FungibleTokenFeature{types.FungibleTokenFeature_freeze}, //nolint:nosnakecase
		}

		_, err := assetKeeper.IssueFungibleToken(ctx, settings)
		if types.ErrInvalidInput.Is(err) == isValid {
			requireT.Equal(types.ErrInvalidInput.Is(err), !isValid)
		}
	}

	for _, symbol := range unacceptableSymbols {
		assertValidSymbol(symbol, false)
	}

	for _, symbol := range acceptableSymbols {
		assertValidSymbol(symbol, true)
	}
}

func TestKeeper_ValidateSubunit(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})
	assetKeeper := testApp.AssetKeeper

	unacceptableSubunits := []string{
		"ABC-1",
		"ABC/1",
		"btc-devcore1phjrez5j2wp5qzp0zvlqavasvw60mkp2zmfe6h",
		"BTC-devcore1phjrez5j2wp5qzp0zvlqavasvw60mkp2zmfe6h",
		"core",
		"ucore",
		"Core",
		"uCore",
		"CORE",
		"UCORE",
		"3abc",
		"3ABC",
		"uCoreum",
		"Coreum",
		"COREeum",
		"AB1234567890123456789012345678901234567890123456789012345678901234567890",
	}

	acceptableSubunits := []string{
		"abc1",
		"coreum",
		"ucoreum",
		"a1234567890123456789012345678901234567890123456789012345678901234567890",
	}

	assertValidSubunit := func(subunit string, isValid bool) {
		addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		settings := types.IssueFungibleTokenSettings{
			Issuer:        addr,
			Symbol:        "symbol",
			Subunit:       subunit,
			Description:   "ABC Desc",
			InitialAmount: sdk.NewInt(777),
			Features:      []types.FungibleTokenFeature{types.FungibleTokenFeature_freeze}, //nolint:nosnakecase
		}

		_, err := assetKeeper.IssueFungibleToken(ctx, settings)
		if isValid {
			requireT.NoError(err)
		} else {
			requireT.ErrorIs(types.ErrInvalidInput, err, "subunit", subunit)
		}
	}

	for _, su := range unacceptableSubunits {
		assertValidSubunit(su, false)
	}

	for _, su := range acceptableSubunits {
		assertValidSubunit(su, true)
	}
}

func TestKeeper_IssueFungibleToken(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	assetKeeper := testApp.AssetKeeper
	bankKeeper := testApp.BankKeeper

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	settings := types.IssueFungibleTokenSettings{
		Issuer:        addr,
		Symbol:        "ABC",
		Description:   "ABC Desc",
		Subunit:       "abc",
		Precision:     8,
		InitialAmount: sdk.NewInt(777),
		Features:      []types.FungibleTokenFeature{types.FungibleTokenFeature_freeze}, //nolint:nosnakecase
	}

	denom, err := assetKeeper.IssueFungibleToken(ctx, settings)
	requireT.NoError(err)
	requireT.Equal(types.BuildFungibleTokenDenom(settings.Subunit, settings.Issuer), denom)

	gotToken, err := assetKeeper.GetFungibleToken(ctx, denom)
	requireT.NoError(err)
	requireT.Equal(types.FungibleToken{
		Denom:       denom,
		Issuer:      settings.Issuer.String(),
		Symbol:      settings.Symbol,
		Description: settings.Description,
		Subunit:     strings.ToLower(settings.Subunit),
		Precision:   settings.Precision,
		Features:    []types.FungibleTokenFeature{types.FungibleTokenFeature_freeze}, //nolint:nosnakecase
		BurnRate:    sdk.NewDec(0),
	}, gotToken)

	// check the metadata
	storedMetadata, found := bankKeeper.GetDenomMetaData(ctx, denom)
	requireT.True(found)
	requireT.Equal(banktypes.Metadata{
		Name:        settings.Symbol,
		Symbol:      settings.Symbol,
		Description: settings.Description,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    settings.Symbol,
				Exponent: settings.Precision,
			},
			{
				Denom:    denom,
				Exponent: 0,
			},
		},
		Base:    denom,
		Display: settings.Symbol,
	}, storedMetadata)

	// check the account state
	issuedAssetBalance := bankKeeper.GetBalance(ctx, addr, denom)
	requireT.Equal(sdk.NewCoin(denom, settings.InitialAmount).String(), issuedAssetBalance.String())

	// check duplicate subunit
	st := settings
	st.Symbol = "test-symbol"
	_, err = assetKeeper.IssueFungibleToken(ctx, st)
	requireT.True(errors.Is(types.ErrInvalidInput, err))

	// check duplicate symbol
	st = settings
	st.Subunit = "test-subunit"
	st.Symbol = "aBc"
	_, err = assetKeeper.IssueFungibleToken(ctx, st)
	requireT.True(errors.Is(types.ErrInvalidInput, err))
}

func TestKeeper_Mint(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	assetKeeper := testApp.AssetKeeper
	bankKeeper := testApp.BankKeeper

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	// Issue an unmintable fungible token
	settings := types.IssueFungibleTokenSettings{
		Issuer:        addr,
		Symbol:        "NotMintable",
		Subunit:       "notmintable",
		InitialAmount: sdk.NewInt(777),
		Features: []types.FungibleTokenFeature{
			types.FungibleTokenFeature_freeze, //nolint:nosnakecase
			types.FungibleTokenFeature_burn,   //nolint:nosnakecase
		},
	}

	unmintableDenom, err := assetKeeper.IssueFungibleToken(ctx, settings)
	requireT.NoError(err)
	requireT.Equal(types.BuildFungibleTokenDenom(settings.Symbol, settings.Issuer), unmintableDenom)

	// try to mint unmintable token
	err = assetKeeper.MintFungibleToken(ctx, addr, sdk.NewCoin(unmintableDenom, sdk.NewInt(100)))
	requireT.Error(err)
	requireT.True(types.ErrFeatureNotActive.Is(err))

	// Issue a mintable fungible token
	settings = types.IssueFungibleTokenSettings{
		Issuer:        addr,
		Symbol:        "mintable",
		Subunit:       "mintable",
		InitialAmount: sdk.NewInt(777),
		Features: []types.FungibleTokenFeature{
			types.FungibleTokenFeature_mint, //nolint:nosnakecase
		},
	}

	mintableDenom, err := assetKeeper.IssueFungibleToken(ctx, settings)
	requireT.NoError(err)

	// try to mint as non-issuer
	randomAddr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	err = assetKeeper.MintFungibleToken(ctx, randomAddr, sdk.NewCoin(mintableDenom, sdk.NewInt(100)))
	requireT.Error(err)
	requireT.True(sdkerrors.ErrUnauthorized.Is(err))

	// mint tokens and check balance and total supply
	err = assetKeeper.MintFungibleToken(ctx, addr, sdk.NewCoin(mintableDenom, sdk.NewInt(100)))
	requireT.NoError(err)

	balance := bankKeeper.GetBalance(ctx, addr, mintableDenom)
	requireT.EqualValues(sdk.NewCoin(mintableDenom, sdk.NewInt(877)), balance)

	totalSupply, err := bankKeeper.TotalSupply(sdk.WrapSDKContext(ctx), &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewInt(877), totalSupply.Supply.AmountOf(mintableDenom))
}

func TestKeeper_Burn(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	assetKeeper := testApp.AssetKeeper
	bankKeeper := testApp.BankKeeper

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	// Issue an unburnable fungible token
	settings := types.IssueFungibleTokenSettings{
		Issuer:        addr,
		Symbol:        "NotBurnable",
		Subunit:       "notburnable",
		InitialAmount: sdk.NewInt(777),
		Features: []types.FungibleTokenFeature{
			types.FungibleTokenFeature_freeze, //nolint:nosnakecase
			types.FungibleTokenFeature_mint,   //nolint:nosnakecase
		},
	}

	unburnableDenom, err := assetKeeper.IssueFungibleToken(ctx, settings)
	requireT.NoError(err)
	requireT.Equal(types.BuildFungibleTokenDenom(settings.Symbol, settings.Issuer), unburnableDenom)

	// try to burn unburnable token
	err = assetKeeper.BurnFungibleToken(ctx, addr, sdk.NewCoin(unburnableDenom, sdk.NewInt(100)))
	requireT.Error(err)
	requireT.True(types.ErrFeatureNotActive.Is(err))

	// Issue a burnable fungible token
	settings = types.IssueFungibleTokenSettings{
		Issuer:        addr,
		Symbol:        "burnable",
		Subunit:       "burnable",
		InitialAmount: sdk.NewInt(777),
		Features: []types.FungibleTokenFeature{
			types.FungibleTokenFeature_burn,   //nolint:nosnakecase
			types.FungibleTokenFeature_freeze, //nolint:nosnakecase
		},
	}

	burnableDenom, err := assetKeeper.IssueFungibleToken(ctx, settings)
	requireT.NoError(err)

	// try to burn as non-issuer
	randomAddr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	err = assetKeeper.BurnFungibleToken(ctx, randomAddr, sdk.NewCoin(burnableDenom, sdk.NewInt(100)))
	requireT.Error(err)
	requireT.True(sdkerrors.ErrUnauthorized.Is(err))

	// burn tokens and check balance and total supply
	err = assetKeeper.BurnFungibleToken(ctx, addr, sdk.NewCoin(burnableDenom, sdk.NewInt(100)))
	requireT.NoError(err)

	balance := bankKeeper.GetBalance(ctx, addr, burnableDenom)
	requireT.EqualValues(sdk.NewCoin(burnableDenom, sdk.NewInt(677)), balance)

	totalSupply, err := bankKeeper.TotalSupply(sdk.WrapSDKContext(ctx), &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewInt(677), totalSupply.Supply.AmountOf(burnableDenom))

	// try to burn frozen amount
	err = assetKeeper.FreezeFungibleToken(ctx, addr, addr, sdk.NewCoin(burnableDenom, sdk.NewInt(600)))
	requireT.NoError(err)

	err = assetKeeper.BurnFungibleToken(ctx, addr, sdk.NewCoin(burnableDenom, sdk.NewInt(100)))
	requireT.True(sdkerrors.ErrInsufficientFunds.Is(err))
}
