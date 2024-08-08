package keeper_test

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/v4/testutil/event"
	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	wbankkeeper "github.com/CoreumFoundation/coreum/v4/x/wbank/keeper"
	wibctransfertypes "github.com/CoreumFoundation/coreum/v4/x/wibctransfer/types"
)

func TestKeeper_Issue(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	ftParams := types.DefaultParams()
	ftParams.IssueFee = sdk.NewInt64Coin(constant.DenomDev, 10_000_000)
	requireT.NoError(ftKeeper.SetParams(ctx, ftParams))

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	requireT.NoError(testApp.FundAccount(
		ctx,
		addr,
		sdk.NewCoins(sdk.NewCoin(ftParams.IssueFee.Denom, ftParams.IssueFee.Amount.MulRaw(5)))),
	)

	settings := types.IssueSettings{
		Issuer:        addr,
		Symbol:        "ABC",
		Description:   "ABC Desc",
		Subunit:       "abc",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(777),
		Features:      []types.Feature{types.Feature_freezing},
		URI:           "https://my-class-meta.invalid/1",
		URIHash:       "content-hash",
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	// verify issue fee was burnt

	burntStr, err := event.FindStringEventAttribute(
		ctx.EventManager().ABCIEvents(),
		banktypes.EventTypeCoinBurn,
		sdk.AttributeKeyAmount,
	)
	requireT.NoError(err)
	requireT.Equal(ftParams.IssueFee.String(), burntStr)

	// check that balance is 0 meaning issue fee was taken

	balance := bankKeeper.GetBalance(ctx, addr, constant.DenomDev)
	requireT.Equal(ftParams.IssueFee.Amount.MulRaw(4).String(), balance.Amount.String())

	requireT.Equal(types.BuildDenom(settings.Subunit, settings.Issuer), denom)

	gotToken, err := ftKeeper.GetToken(ctx, denom)
	requireT.NoError(err)
	requireT.Equal(types.Token{
		Denom:              denom,
		Issuer:             settings.Issuer.String(),
		Symbol:             settings.Symbol,
		Description:        settings.Description,
		Subunit:            strings.ToLower(settings.Subunit),
		Precision:          settings.Precision,
		Features:           []types.Feature{types.Feature_freezing},
		BurnRate:           sdkmath.LegacyNewDec(0),
		SendCommissionRate: sdkmath.LegacyNewDec(0),
		Version:            types.CurrentTokenVersion,
		URI:                settings.URI,
		URIHash:            settings.URIHash,
		Admin:              settings.Issuer.String(),
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
				Denom:    denom,
				Exponent: 0,
			},
			{
				Denom:    settings.Symbol,
				Exponent: settings.Precision,
			},
		},
		Base:    denom,
		Display: settings.Symbol,
		URI:     settings.URI,
		URIHash: settings.URIHash,
	}, storedMetadata)

	// check the account state
	issuedAssetBalance := bankKeeper.GetBalance(ctx, addr, denom)
	requireT.Equal(sdk.NewCoin(denom, settings.InitialAmount).String(), issuedAssetBalance.String())

	// check duplicate subunit
	st := settings
	st.Symbol = "test-symbol"
	_, err = ftKeeper.Issue(ctx, st)
	requireT.ErrorIs(err, types.ErrInvalidInput)

	// check duplicate symbol
	requireT.NoError(testApp.FundAccount(ctx, addr, sdk.NewCoins(ftParams.IssueFee)))
	st = settings
	st.Subunit = "subunit"
	st.Symbol = "aBc"
	_, err = ftKeeper.Issue(ctx, st)
	requireT.ErrorIs(err, types.ErrInvalidInput)
	requireT.True(strings.Contains(err.Error(), "duplicate"))

	// try to create token containing non-existing feature
	settings.Symbol = "CDE"
	settings.Subunit = "subunit2"
	settings.Features = append(settings.Features, 10000)
	_, err = ftKeeper.Issue(ctx, settings)
	requireT.ErrorIs(err, types.ErrInvalidInput)

	// try to create token containing doubled feature
	settings.Symbol = "EFG"
	settings.Subunit = "subunit3"
	settings.Features = append(settings.Features, settings.Features[0])
	_, err = ftKeeper.Issue(ctx, settings)
	requireT.ErrorIs(err, types.ErrInvalidInput)
}

func TestKeeper_Issue_ZeroPrecision(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper
	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        addr,
		Symbol:        "ABC",
		Description:   "ABC Desc",
		Subunit:       "abc",
		Precision:     0,
		InitialAmount: sdkmath.NewInt(777),
		Features:      []types.Feature{types.Feature_freezing},
		URI:           "https://my-class-meta.invalid/1",
		URIHash:       "content-hash",
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	gotToken, err := ftKeeper.GetToken(ctx, denom)
	requireT.NoError(err)
	requireT.Equal(types.Token{
		Denom:              denom,
		Issuer:             settings.Issuer.String(),
		Symbol:             settings.Symbol,
		Description:        settings.Description,
		Subunit:            strings.ToLower(settings.Subunit),
		Precision:          settings.Precision,
		Features:           []types.Feature{types.Feature_freezing},
		BurnRate:           sdkmath.LegacyNewDec(0),
		SendCommissionRate: sdkmath.LegacyNewDec(0),
		Version:            types.CurrentTokenVersion,
		URI:                settings.URI,
		URIHash:            settings.URIHash,
		Admin:              settings.Issuer.String(),
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
				Denom:    denom,
				Exponent: 0,
			},
		},
		Base:    denom,
		Display: denom,
		URI:     settings.URI,
		URIHash: settings.URIHash,
	}, storedMetadata)
}

func TestKeeper_IssueEqualDisplayAndBaseDenom(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper

	ftParams := types.DefaultParams()
	ftParams.IssueFee = sdk.NewInt64Coin(constant.DenomDev, 10_000_000)
	requireT.NoError(ftKeeper.SetParams(ctx, ftParams))

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	requireT.NoError(testApp.FundAccount(ctx, addr, sdk.NewCoins(ftParams.IssueFee)))
	subunit := "abc"
	denom := types.BuildDenom(subunit, addr)

	settings := types.IssueSettings{
		Issuer:        addr,
		Symbol:        denom,
		Description:   "ABC Desc",
		Subunit:       subunit,
		Precision:     8,
		InitialAmount: sdkmath.NewInt(777),
		Features:      []types.Feature{types.Feature_freezing},
	}

	_, err := ftKeeper.Issue(ctx, settings)
	requireT.Error(err)
	requireT.True(strings.Contains(err.Error(), "duplicate denomination"))
}

func TestKeeper_IssueValidateSymbol(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})
	ftKeeper := testApp.AssetFTKeeper

	unacceptableSymbols := []string{
		"core",
		"ucore",
		"Core",
		"uCore",
		"CORE",
		"UCORE",
		"3abc",
		"3ABC",
		"C",
	}

	acceptableSymbols := []string{
		"btc-devcore1phjrez5j2wp5qzp0zvlqavasvw60mkp2zmfe6h",
		"BTC-devcore1phjrez5j2wp5qzp0zvlqavasvw60mkp2zmfe6h",
		"ABC-1",
		"ABC1",
		"ABC/1",
		"coreum",
		"ucoreum",
		"Coreum",
		"COREeum.",
		"uCoreum",
		"COREeum",
		"A1234567890123456789012345678901234567890123456789012345678901234567890",
		"AB1234567890123456789012345678901234567890123456789012345678901234567890",
	}

	assertValidSymbol := func(symbol string, isValid bool) {
		addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		settings := types.IssueSettings{
			Issuer:        addr,
			Symbol:        symbol,
			Subunit:       "subunit",
			Description:   "ABC Desc",
			Precision:     1,
			InitialAmount: sdkmath.NewInt(777),
			Features:      []types.Feature{types.Feature_freezing},
		}

		_, err := ftKeeper.Issue(ctx, settings)
		if !isValid {
			requireT.ErrorIs(err, types.ErrInvalidInput, "symbol:%s", symbol)
		} else {
			requireT.NoError(err, "symbol:%s", symbol)
		}
	}

	for _, symbol := range unacceptableSymbols {
		assertValidSymbol(symbol, false)
	}

	for _, symbol := range acceptableSymbols {
		assertValidSymbol(symbol, true)
	}
}

func TestKeeper_IssueValidateSubunit(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})
	ftKeeper := testApp.AssetFTKeeper

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
		"AB123456789012345678901234567890123456789012345678",
	}

	acceptableSubunits := []string{
		"abc1",
		"coreum",
		"ucoreum",
		"a1234567890123456789012345678901234567890123456789",
	}

	assertValidSubunit := func(subunit string, isValid bool) {
		addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		settings := types.IssueSettings{
			Issuer:        addr,
			Symbol:        "symbol",
			Subunit:       subunit,
			Precision:     1,
			Description:   "ABC Desc",
			InitialAmount: sdkmath.NewInt(777),
			Features:      []types.Feature{types.Feature_freezing},
		}

		_, err := ftKeeper.Issue(ctx, settings)
		if isValid {
			requireT.NoError(err)
		} else {
			requireT.ErrorIs(err, types.ErrInvalidInput, "subunit", subunit)
		}
	}

	for _, su := range unacceptableSubunits {
		assertValidSubunit(su, false)
	}

	for _, su := range acceptableSubunits {
		assertValidSubunit(su, true)
	}
}

func TestKeeper_Issue_WithZeroIssueFee(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper

	ftParams := types.DefaultParams()
	ftParams.IssueFee = sdk.NewCoin(constant.DenomDev, sdkmath.ZeroInt())
	requireT.NoError(ftKeeper.SetParams(ctx, ftParams))

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        addr,
		Symbol:        "ABC",
		Description:   "ABC Desc",
		Subunit:       "abc",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(777),
		Features:      []types.Feature{types.Feature_freezing},
	}

	_, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)
}

func TestKeeper_Issue_WithNoFundsCoveringFee(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper

	ftParams := types.DefaultParams()
	ftParams.IssueFee = sdk.NewInt64Coin(constant.DenomDev, 10_000_000)
	requireT.NoError(ftKeeper.SetParams(ctx, ftParams))

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	settings := types.IssueSettings{
		Issuer:        addr,
		Symbol:        "ABC",
		Description:   "ABC Desc",
		Subunit:       "abc",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(777),
		Features:      []types.Feature{types.Feature_freezing},
	}

	_, err := ftKeeper.Issue(ctx, settings)
	requireT.ErrorIs(err, cosmoserrors.ErrInsufficientFunds)
}

func TestKeeper_Mint(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	// Issue an unmintable fungible token
	settings := types.IssueSettings{
		Issuer:        addr,
		Symbol:        "NotMintable",
		Subunit:       "notmintable",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(777),
		Features: []types.Feature{
			types.Feature_freezing,
			types.Feature_burning,
		},
	}

	unmintableDenom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)
	requireT.Equal(types.BuildDenom(settings.Symbol, settings.Issuer), unmintableDenom)

	// try to mint unmintable token
	err = ftKeeper.Mint(ctx, addr, addr, sdk.NewCoin(unmintableDenom, sdkmath.NewInt(100)))
	requireT.ErrorIs(err, types.ErrFeatureDisabled)

	// Issue a mintable fungible token
	settings = types.IssueSettings{
		Issuer:        addr,
		Symbol:        "mintable",
		Subunit:       "mintable",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(777),
		Features: []types.Feature{
			types.Feature_minting,
		},
	}

	mintableDenom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	// try to mint as non-issuer
	randomAddr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	err = ftKeeper.Mint(ctx, randomAddr, randomAddr, sdk.NewCoin(mintableDenom, sdkmath.NewInt(100)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// mint tokens and check balance and total supply
	err = ftKeeper.Mint(ctx, addr, addr, sdk.NewCoin(mintableDenom, sdkmath.NewInt(100)))
	requireT.NoError(err)

	balance := bankKeeper.GetBalance(ctx, addr, mintableDenom)
	requireT.EqualValues(sdk.NewCoin(mintableDenom, sdkmath.NewInt(877)), balance)

	totalSupply, err := bankKeeper.TotalSupply(ctx, &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(877), totalSupply.Supply.AmountOf(mintableDenom))

	// mint to another account
	err = ftKeeper.Mint(ctx, addr, randomAddr, sdk.NewCoin(mintableDenom, sdkmath.NewInt(100)))
	requireT.NoError(err)

	balance = bankKeeper.GetBalance(ctx, randomAddr, mintableDenom)
	requireT.EqualValues(sdk.NewCoin(mintableDenom, sdkmath.NewInt(100)), balance)

	totalSupply, err = bankKeeper.TotalSupply(ctx, &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(977), totalSupply.Supply.AmountOf(mintableDenom))
}

func TestKeeper_Burn(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	recipient := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	// Issue an unburnable fungible token
	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "NotBurnable",
		Subunit:       "notburnable",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(777),
		Features: []types.Feature{
			types.Feature_freezing,
			types.Feature_minting,
		},
	}

	unburnableDenom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)
	requireT.Equal(types.BuildDenom(settings.Symbol, settings.Issuer), unburnableDenom)

	// send to new recipient address
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(sdk.NewCoin(unburnableDenom, sdkmath.NewInt(100))))
	requireT.NoError(err)

	// try to burn unburnable token from the recipient account
	err = ftKeeper.Burn(ctx, recipient, sdk.NewCoin(unburnableDenom, sdkmath.NewInt(100)))
	requireT.ErrorIs(err, types.ErrFeatureDisabled)

	// try to burn unburnable token from the issuer account
	err = ftKeeper.Burn(ctx, issuer, sdk.NewCoin(unburnableDenom, sdkmath.NewInt(100)))
	requireT.NoError(err)

	// Issue a burnable fungible token
	settings = types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "burnable",
		Subunit:       "burnable",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(777),
		Features: []types.Feature{
			types.Feature_burning,
			types.Feature_freezing,
		},
	}

	burnableDenom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	// send to new recipient address
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(sdk.NewCoin(burnableDenom, sdkmath.NewInt(200))))
	requireT.NoError(err)

	// try to burn as non-issuer
	err = ftKeeper.Burn(ctx, recipient, sdk.NewCoin(burnableDenom, sdkmath.NewInt(100)))
	requireT.NoError(err)

	// burn tokens and check balance and total supply
	err = ftKeeper.Burn(ctx, issuer, sdk.NewCoin(burnableDenom, sdkmath.NewInt(100)))
	requireT.NoError(err)

	balance := bankKeeper.GetBalance(ctx, issuer, burnableDenom)
	requireT.EqualValues(sdk.NewCoin(burnableDenom, sdkmath.NewInt(477)), balance)

	totalSupply, err := bankKeeper.TotalSupply(ctx, &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(577), totalSupply.Supply.AmountOf(burnableDenom))

	// try to freeze the issuer (issuer can't be frozen)
	err = ftKeeper.Freeze(ctx, issuer, issuer, sdk.NewCoin(burnableDenom, sdkmath.NewInt(600)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to burn non-issuer frozen coins
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(burnableDenom, sdkmath.NewInt(100)))
	requireT.NoError(err)
	err = ftKeeper.Burn(ctx, recipient, sdk.NewCoin(burnableDenom, sdkmath.NewInt(100)))
	requireT.ErrorIs(err, cosmoserrors.ErrInsufficientFunds)
	err = ftKeeper.Unfreeze(ctx, issuer, recipient, sdk.NewCoin(burnableDenom, sdkmath.NewInt(100)))
	requireT.NoError(err)

	// DEX lock coins and try to burn
	err = ftKeeper.DEXLock(ctx, recipient, sdk.NewCoin(burnableDenom, sdkmath.NewInt(100)))
	requireT.NoError(err)
	err = ftKeeper.Burn(ctx, recipient, sdk.NewCoin(burnableDenom, sdkmath.NewInt(100)))
	requireT.ErrorIs(err, cosmoserrors.ErrInsufficientFunds)
}

func TestKeeper_BurnRate_BankSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	assetKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper
	ba := newBankAsserter(ctx, t, bankKeeper)

	// issue with more than 1 burn rate
	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     6,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(600),
		Features:      []types.Feature{},
		BurnRate:      sdkmath.LegacyMustNewDecFromStr("1.01"),
	}

	_, err := assetKeeper.Issue(ctx, settings)
	requireT.ErrorIs(err, types.ErrInvalidInput)

	// issue token
	settings = types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     6,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(600),
		Features:      []types.Feature{},
		BurnRate:      sdkmath.LegacyMustNewDecFromStr("0.25"),
	}

	denom, err := assetKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// send from issuer to recipient (burn must not apply)
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(500)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient: 500,
		&issuer:    100,
	})

	// send from recipient1 to recipient2 (burn must apply)
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(100)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient:  375,
		&recipient2: 100,
		&issuer:     100,
	})

	// send from recipient to issuer account (burn must not apply)
	err = bankKeeper.SendCoins(ctx, recipient, issuer, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(375)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient2: 100,
		&issuer:     475,
	})
}

func TestKeeper_BurnRate_BankMultiSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	assetKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper
	ba := newBankAsserter(ctx, t, bankKeeper)

	// issue 2 tokens
	var recipients []sdk.AccAddress
	var issuers []sdk.AccAddress
	var denoms []string
	for i := 0; i < 2; i++ {
		issuers = append(issuers, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()))
		settings := types.IssueSettings{
			Issuer:             issuers[i],
			Symbol:             fmt.Sprintf("DEF%d", i),
			Subunit:            fmt.Sprintf("def%d", i),
			Precision:          6,
			Description:        "DEF Desc",
			InitialAmount:      sdkmath.NewInt(1000),
			Features:           []types.Feature{},
			BurnRate:           sdkmath.LegacyNewDec(int64(i + 1)).QuoInt64(10), // 10% and 20% respectively
			SendCommissionRate: sdkmath.LegacyNewDec(int64(i + 1)).QuoInt64(20), // 5% and 10% respectively
		}

		denom, err := assetKeeper.Issue(ctx, settings)
		requireT.NoError(err)
		denoms = append(denoms, denom)

		// create 2 recipient for every issuer to allow for complex test cases
		recipients = append(recipients, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()))
		recipients = append(recipients, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()))
	}

	testCases := []struct {
		name         string
		inputs       banktypes.Input
		outputs      []banktypes.Output
		distribution map[string]map[*sdk.AccAddress]int64
	}{
		{
			name: "send from issuer1 to other accounts",
			inputs: banktypes.Input{
				Address: issuers[1].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[1], sdkmath.NewInt(600))),
			},
			outputs: []banktypes.Output{
				{Address: recipients[0].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdkmath.NewInt(100)),
				)},
				{Address: recipients[1].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdkmath.NewInt(100)),
				)},
				{Address: issuers[0].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdkmath.NewInt(400)),
				)},
			},
			distribution: map[string]map[*sdk.AccAddress]int64{
				denoms[1]: {
					&issuers[1]:    400,
					&issuers[0]:    400,
					&recipients[0]: 100,
					&recipients[1]: 100,
				},
			},
		},
		{
			name: "send from issuer0 to other accounts",
			inputs: banktypes.Input{
				Address: issuers[0].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdkmath.NewInt(200)),
					sdk.NewCoin(denoms[1], sdkmath.NewInt(200)),
				),
			},
			outputs: []banktypes.Output{
				{Address: recipients[0].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdkmath.NewInt(100)),
					sdk.NewCoin(denoms[1], sdkmath.NewInt(100)),
				)},
				{Address: recipients[1].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdkmath.NewInt(100)),
					sdk.NewCoin(denoms[1], sdkmath.NewInt(100)),
				)},
			},
			distribution: map[string]map[*sdk.AccAddress]int64{
				denoms[0]: {
					&issuers[0]:    800,
					&recipients[0]: 100,
					&recipients[1]: 100,
				},
				denoms[1]: {
					&issuers[1]:    420, // (400 + 200*10%)
					&issuers[0]:    140, // (400 - 200 - 200*10%(commison) - 200*20% (burn))
					&recipients[0]: 200,
					&recipients[1]: 200,
				},
			},
		},
		{
			name: "include issuer in recipients",
			inputs: banktypes.Input{
				Address: recipients[0].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdkmath.NewInt(60)),
					sdk.NewCoin(denoms[1], sdkmath.NewInt(60)),
				),
			},
			outputs: []banktypes.Output{
				{Address: issuers[1].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdkmath.NewInt(25)),
				)},
				{Address: issuers[0].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdkmath.NewInt(15)),
				)},
				{Address: recipients[2].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdkmath.NewInt(11)),
				)},
				{Address: recipients[3].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdkmath.NewInt(9)),
				)},
				{Address: issuers[1].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdkmath.NewInt(25)),
				)},
				{Address: issuers[0].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdkmath.NewInt(15)),
				)},
				{Address: recipients[2].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdkmath.NewInt(11)),
				)},
				{Address: recipients[3].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdkmath.NewInt(9)),
				)},
			},
			distribution: map[string]map[*sdk.AccAddress]int64{
				denoms[0]: {
					&issuers[0]:    819, // 800 + 15 + 4 (5% commission)
					&issuers[1]:    25,
					&recipients[0]: 30, // 100 - 60 - 6 (10% burn) - 4(commission)
					&recipients[1]: 100,
					&recipients[2]: 11,
					&recipients[3]: 9,
				},
				denoms[1]: {
					&issuers[1]:    450, // 420 + 25 + 5 (10% commission)
					&issuers[0]:    155, // 140 + 15
					&recipients[0]: 127, // 200 - 60 - 8 (20% burn) - 5 (commission)
					&recipients[1]: 200,
					&recipients[2]: 11,
					&recipients[3]: 9,
				},
			},
		},
	}

	for counter, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%s case #%d", tc.name, counter), func(t *testing.T) {
			err := bankKeeper.InputOutputCoins(ctx, tc.inputs, tc.outputs)
			requireT.NoError(err)

			for denom, dist := range tc.distribution {
				ba.assertCoinDistribution(denom, dist)
			}
		})
	}
}

func TestKeeper_SendCommissionRate_BankSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	assetKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper
	ba := newBankAsserter(ctx, t, bankKeeper)

	// issue with more than 1 send commission rate
	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	settings := types.IssueSettings{
		Issuer:             issuer,
		Symbol:             "DEF",
		Subunit:            "def",
		Precision:          6,
		Description:        "DEF Desc",
		InitialAmount:      sdkmath.NewInt(600),
		Features:           []types.Feature{},
		SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("1.01"),
	}

	_, err := assetKeeper.Issue(ctx, settings)
	requireT.ErrorIs(err, types.ErrInvalidInput)

	// issue token
	settings = types.IssueSettings{
		Issuer:             issuer,
		Symbol:             "DEF",
		Subunit:            "def",
		Precision:          6,
		Description:        "DEF Desc",
		InitialAmount:      sdkmath.NewInt(600),
		Features:           []types.Feature{},
		SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("0.25"),
	}

	denom, err := assetKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// send from issuer to recipient (send commission rate must not apply)
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(500)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient: 500,
		&issuer:    100,
	})

	// send from recipient1 to recipient2 (send commission rate must apply)
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(100)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient:  375,
		&recipient2: 100,
		&issuer:     125,
	})

	// send from recipient to issuer account (send commission rate must not apply)
	err = bankKeeper.SendCoins(ctx, recipient, issuer, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(375)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient2: 100,
		&issuer:     500,
	})
}

func TestKeeper_BurnRateAndSendCommissionRate_BankSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	assetKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper
	ba := newBankAsserter(ctx, t, bankKeeper)

	// issue token
	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	settings := types.IssueSettings{
		Issuer:             issuer,
		Symbol:             "DEF",
		Subunit:            "def",
		Precision:          6,
		Description:        "DEF Desc",
		InitialAmount:      sdkmath.NewInt(600),
		Features:           []types.Feature{},
		BurnRate:           sdkmath.LegacyMustNewDecFromStr("0.5"),
		SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("0.25"),
	}

	denom, err := assetKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// send from issuer to recipient (fees must not apply)
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(500)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient: 500,
		&issuer:    100,
	})

	// send from recipient1 to recipient2 (fees must apply)
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(100)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient:  325,
		&recipient2: 100,
		&issuer:     125,
	})

	// send from recipient to issuer account (fees must not apply)
	err = bankKeeper.SendCoins(ctx, recipient, issuer, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(325)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient2: 100,
		&issuer:     450,
	})
}

func TestKeeper_FreezeUnfreeze(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(666),
		Features:      []types.Feature{types.Feature_freezing},
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	unfreezableSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     1,
		Description:   "ABC Desc",
		InitialAmount: sdkmath.NewInt(666),
		Features:      []types.Feature{},
	}

	unfreezableDenom, err := ftKeeper.Issue(ctx, unfreezableSettings)
	requireT.NoError(err)
	_, err = ftKeeper.GetToken(ctx, unfreezableDenom)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(100)),
		sdk.NewCoin(unfreezableDenom, sdkmath.NewInt(100)),
	))
	requireT.NoError(err)

	// try to freeze non-existent denom
	nonExistentDenom := types.BuildDenom("nonexist", issuer)
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(nonExistentDenom, sdkmath.NewInt(10)))
	assertT.True(sdkerrors.IsOf(err, types.ErrTokenNotFound))

	// try to freeze unfreezable Token
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(unfreezableDenom, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, types.ErrFeatureDisabled)

	// try to freeze from non issuer address
	randomAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = ftKeeper.Freeze(ctx, randomAddr, recipient, sdk.NewCoin(denom, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to freeze 0 balance
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdkmath.NewInt(0)))
	requireT.ErrorIs(err, cosmoserrors.ErrInvalidCoins)

	// try to unfreeze 0 balance
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdkmath.NewInt(0)))
	requireT.ErrorIs(err, cosmoserrors.ErrInvalidCoins)

	// try to freeze more than balance
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdkmath.NewInt(110)))
	requireT.NoError(err)
	frozenBalance := ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	assertT.EqualValues(sdk.NewCoin(denom, sdkmath.NewInt(110)), frozenBalance)

	// try to unfreeze more than frozen balance
	err = ftKeeper.Unfreeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdkmath.NewInt(130)))
	requireT.ErrorIs(err, cosmoserrors.ErrInsufficientFunds)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	assertT.EqualValues(sdk.NewCoin(denom, sdkmath.NewInt(110)), frozenBalance)

	// set frozen balance back to zero
	err = ftKeeper.Unfreeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdkmath.NewInt(110)))
	requireT.NoError(err)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	assertT.EqualValues(sdk.NewCoin(denom, sdkmath.NewInt(0)).String(), frozenBalance.String())

	// freeze, query frozen
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdkmath.NewInt(40)))
	requireT.NoError(err)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(40)).String(), frozenBalance.String())

	// test query all frozen
	allBalances, pageRes, err := ftKeeper.GetAccountsFrozenBalances(ctx, &query.PageRequest{})
	requireT.NoError(err)
	assertT.Len(allBalances, 1)
	assertT.EqualValues(1, pageRes.GetTotal())
	assertT.EqualValues(recipient.String(), allBalances[0].Address)
	requireT.Equal(sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(40))).String(), allBalances[0].Coins.String())

	// increase frozen and query
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdkmath.NewInt(40)))
	requireT.NoError(err)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(80)), frozenBalance)

	// try to send more than available
	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(80)))
	// send
	err = bankKeeper.SendCoins(ctx, recipient, issuer, coinsToSend)
	assertT.True(sdkerrors.IsOf(err, cosmoserrors.ErrInsufficientFunds))
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{Address: recipient.String(), Coins: coinsToSend},
		[]banktypes.Output{{Address: issuer.String(), Coins: coinsToSend}})
	assertT.True(sdkerrors.IsOf(err, cosmoserrors.ErrInsufficientFunds))

	// try to send unfrozen balance
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	coinsToSend = sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(10)))
	// send
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, coinsToSend)
	requireT.NoError(err)
	balance := bankKeeper.GetBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(90)), balance)
	balance = bankKeeper.GetBalance(ctx, recipient2, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(10)), balance)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{Address: recipient.String(), Coins: coinsToSend},
		[]banktypes.Output{{Address: recipient2.String(), Coins: coinsToSend}})
	requireT.NoError(err)
	balance = bankKeeper.GetBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(80)), balance)
	balance = bankKeeper.GetBalance(ctx, recipient2, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(20)), balance)

	// try to unfreeze from non issuer address
	err = ftKeeper.Unfreeze(ctx, randomAddr, recipient, sdk.NewCoin(denom, sdkmath.NewInt(80)))
	assertT.True(sdkerrors.IsOf(err, cosmoserrors.ErrUnauthorized))

	// set absolute frozen amount
	err = ftKeeper.SetFrozen(ctx, issuer, recipient, sdk.NewCoin(denom, sdkmath.NewInt(100)))
	requireT.NoError(err)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(100)), frozenBalance)

	// unfreeze, query frozen, and try to send
	err = ftKeeper.Unfreeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdkmath.NewInt(100)))
	requireT.NoError(err)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(0)), frozenBalance)
	coinsToSend = sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(40)))
	// send
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, coinsToSend)
	requireT.NoError(err)
	balance = bankKeeper.GetBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(40)), balance)
	balance = bankKeeper.GetBalance(ctx, recipient2, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(60)), balance)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{Address: recipient.String(), Coins: coinsToSend},
		[]banktypes.Output{{Address: recipient2.String(), Coins: coinsToSend}})
	requireT.NoError(err)
	balance = bankKeeper.GetBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(0)), balance)
	balance = bankKeeper.GetBalance(ctx, recipient2, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(100)), balance)
}

func TestKeeper_GlobalFreezeUnfreeze(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	freezableSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "FREEZE",
		Subunit:       "freeze",
		Precision:     6,
		Description:   "FREEZE Desc",
		InitialAmount: sdkmath.NewInt(777),
		Features:      []types.Feature{types.Feature_freezing},
	}

	freezableDenom, err := ftKeeper.Issue(ctx, freezableSettings)
	requireT.NoError(err)

	unfreezableSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "NOFREEZE",
		Subunit:       "nofreeze",
		Precision:     6,
		Description:   "NOFREEZE Desc",
		InitialAmount: sdkmath.NewInt(777),
		Features:      []types.Feature{},
	}

	unfreezableDenom, err := ftKeeper.Issue(ctx, unfreezableSettings)
	requireT.NoError(err)
	_, err = ftKeeper.GetToken(ctx, unfreezableDenom)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(
		sdk.NewCoin(freezableDenom, sdkmath.NewInt(100)),
		sdk.NewCoin(unfreezableDenom, sdkmath.NewInt(100)),
	))
	requireT.NoError(err)

	// try to global-freeze non-existent
	nonExistentDenom := types.BuildDenom("nonexist", issuer)
	err = ftKeeper.GloballyFreeze(ctx, issuer, nonExistentDenom)
	assertT.True(sdkerrors.IsOf(err, types.ErrTokenNotFound))

	// try to global-freeze unfreezable Token
	err = ftKeeper.GloballyFreeze(ctx, issuer, unfreezableDenom)
	requireT.ErrorIs(err, types.ErrFeatureDisabled)

	// try to global-freeze from non issuer address
	randomAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = ftKeeper.GloballyFreeze(ctx, randomAddr, freezableDenom)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// freeze twice to check global-freeze idempotence
	err = ftKeeper.GloballyFreeze(ctx, issuer, freezableDenom)
	requireT.NoError(err)
	err = ftKeeper.GloballyFreeze(ctx, issuer, freezableDenom)
	requireT.NoError(err)
	frozenToken, err := ftKeeper.GetToken(ctx, freezableDenom)
	requireT.NoError(err)
	assertT.True(frozenToken.GloballyFrozen)

	// try to global-unfreeze from non issuer address
	err = ftKeeper.GloballyUnfreeze(ctx, randomAddr, freezableDenom)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// unfreeze twice to check global-unfreeze idempotence
	err = ftKeeper.GloballyUnfreeze(ctx, issuer, freezableDenom)
	requireT.NoError(err)
	err = ftKeeper.GloballyUnfreeze(ctx, issuer, freezableDenom)
	requireT.NoError(err)
	unfrozenToken, err := ftKeeper.GetToken(ctx, freezableDenom)
	requireT.NoError(err)
	assertT.False(unfrozenToken.GloballyFrozen)

	// freeze, try to send & verify balance
	err = ftKeeper.GloballyFreeze(ctx, issuer, freezableDenom)
	requireT.NoError(err)
	coinsToSend := sdk.NewCoins(sdk.NewCoin(freezableDenom, sdkmath.NewInt(10)))
	// send
	err = bankKeeper.SendCoins(ctx, recipient, randomAddr, coinsToSend)
	requireT.ErrorIs(err, types.ErrGloballyFrozen)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{Address: recipient.String(), Coins: coinsToSend},
		[]banktypes.Output{{Address: randomAddr.String(), Coins: coinsToSend}})
	requireT.ErrorIs(err, types.ErrGloballyFrozen)

	// unfreeze, try to send & verify balance
	err = ftKeeper.GloballyUnfreeze(ctx, issuer, freezableDenom)
	requireT.NoError(err)
	coinsToSend = sdk.NewCoins(sdk.NewCoin(freezableDenom, sdkmath.NewInt(6)))
	// send
	err = bankKeeper.SendCoins(ctx, recipient, randomAddr, coinsToSend)
	requireT.NoError(err)
	balance := bankKeeper.GetBalance(ctx, randomAddr, freezableDenom)
	requireT.Equal(sdk.NewCoin(freezableDenom, sdkmath.NewInt(6)), balance)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{Address: recipient.String(), Coins: coinsToSend},
		[]banktypes.Output{{Address: randomAddr.String(), Coins: coinsToSend}})
	requireT.NoError(err)
	balance = bankKeeper.GetBalance(ctx, randomAddr, freezableDenom)
	requireT.Equal(sdk.NewCoin(freezableDenom, sdkmath.NewInt(12)), balance)
}

func TestKeeper_Clawback(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper
	accountKeeper := testApp.AccountKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(666),
		Features: []types.Feature{
			types.Feature_clawback,
			types.Feature_freezing,
		},
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	clawbackDisabledSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     1,
		Description:   "ABC Desc",
		InitialAmount: sdkmath.NewInt(666),
		Features:      []types.Feature{},
	}

	clawbackDisabledDenom, err := ftKeeper.Issue(ctx, clawbackDisabledSettings)
	requireT.NoError(err)
	_, err = ftKeeper.GetToken(ctx, clawbackDisabledDenom)
	requireT.NoError(err)

	from := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, issuer, from, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(100)),
		sdk.NewCoin(clawbackDisabledDenom, sdkmath.NewInt(100)),
	))
	requireT.NoError(err)

	// try to clawback non-existent denom
	nonExistentDenom := types.BuildDenom("nonexist", issuer)
	err = ftKeeper.Clawback(ctx, issuer, from, sdk.NewCoin(nonExistentDenom, sdkmath.NewInt(10)))
	assertT.True(sdkerrors.IsOf(err, types.ErrTokenNotFound))

	// try to clawback clawbackDisabled Token
	err = ftKeeper.Clawback(ctx, issuer, from, sdk.NewCoin(clawbackDisabledDenom, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, types.ErrFeatureDisabled)

	// try to clawback by non issuer address
	randomAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = ftKeeper.Clawback(ctx, randomAddr, from, sdk.NewCoin(denom, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to clawback from issuer address
	randomAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = ftKeeper.Clawback(ctx, randomAddr, issuer, sdk.NewCoin(denom, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to clawback from module address
	moduleAddr := accountKeeper.GetModuleAccount(ctx, types.ModuleName).GetAddress()
	err = ftKeeper.Clawback(ctx, issuer, moduleAddr, sdk.NewCoin(denom, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to clawback 0 balance
	err = ftKeeper.Clawback(ctx, issuer, from, sdk.NewCoin(denom, sdkmath.NewInt(0)))
	requireT.ErrorIs(err, cosmoserrors.ErrInvalidCoins)

	// try to clawback more than balance
	err = ftKeeper.Clawback(ctx, issuer, from, sdk.NewCoin(denom, sdkmath.NewInt(110)))
	requireT.ErrorIs(err, cosmoserrors.ErrInsufficientFunds)

	// try to clawback locked balance
	err = ftKeeper.DEXLock(ctx, from, sdk.NewCoin(denom, sdkmath.NewInt(100)))
	requireT.NoError(err)
	err = ftKeeper.Clawback(ctx, issuer, from, sdk.NewCoin(denom, sdkmath.NewInt(40)))
	requireT.ErrorIs(err, cosmoserrors.ErrInsufficientFunds)
	err = ftKeeper.DEXUnlock(ctx, from, sdk.NewCoin(denom, sdkmath.NewInt(100)))
	requireT.NoError(err)

	// clawback, query balance
	issuerBalanceBefore := bankKeeper.GetBalance(ctx, issuer, denom)
	accountBalanceBefore := bankKeeper.GetBalance(ctx, from, denom)
	err = ftKeeper.Clawback(ctx, issuer, from, sdk.NewCoin(denom, sdkmath.NewInt(40)))
	requireT.NoError(err)
	issuerBalanceAfter := bankKeeper.GetBalance(ctx, issuer, denom)
	accountBalanceAfter := bankKeeper.GetBalance(ctx, from, denom)
	requireT.Equal(issuerBalanceBefore.Add(sdk.NewCoin(denom, sdkmath.NewInt(40))), issuerBalanceAfter)
	requireT.Equal(accountBalanceBefore.Sub(sdk.NewCoin(denom, sdkmath.NewInt(40))), accountBalanceAfter)

	// clawback frozen token, query balance
	err = ftKeeper.Freeze(ctx, issuer, from, sdk.NewCoin(denom, sdkmath.NewInt(60)))
	requireT.NoError(err)
	err = ftKeeper.Clawback(ctx, issuer, from, sdk.NewCoin(denom, sdkmath.NewInt(60)))
	requireT.NoError(err)
	accountBalance := bankKeeper.GetBalance(ctx, from, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(0)), accountBalance)
}

func TestKeeper_Whitelist(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(666),
		Features:      []types.Feature{types.Feature_whitelisting},
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	unwhitelistableSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     1,
		Description:   "ABC Desc",
		InitialAmount: sdkmath.NewInt(666),
		Features:      []types.Feature{},
	}

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	unwhitelistableDenom, err := ftKeeper.Issue(ctx, unwhitelistableSettings)
	requireT.NoError(err)
	_, err = ftKeeper.GetToken(ctx, unwhitelistableDenom)
	requireT.NoError(err)

	// whitelisting fails on unwhitelistable token
	err = ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, sdk.NewCoin(unwhitelistableDenom, sdkmath.NewInt(1)))
	requireT.ErrorIs(err, types.ErrFeatureDisabled)

	// try to whitelist non-existent denom
	nonExistentDenom := types.BuildDenom("nonexist", issuer)
	err = ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, sdk.NewCoin(nonExistentDenom, sdkmath.NewInt(10)))
	assertT.True(sdkerrors.IsOf(err, types.ErrTokenNotFound))

	// try to whitelist from non issuer address
	randomAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = ftKeeper.SetWhitelistedBalance(ctx, randomAddr, recipient, sdk.NewCoin(denom, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to whitelist the issuer (issuer can't be whitelisted)
	err = ftKeeper.SetWhitelistedBalance(ctx, issuer, issuer, sdk.NewCoin(denom, sdkmath.NewInt(1)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// set whitelisted balance to 0
	requireT.NoError(ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, sdk.NewCoin(denom, sdkmath.NewInt(0))))
	whitelistedBalance := ftKeeper.GetWhitelistedBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(0)).String(), whitelistedBalance.String())

	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100)))
	// send
	err = bankKeeper.SendCoins(ctx, issuer, recipient, coinsToSend)
	requireT.ErrorIs(err, types.ErrWhitelistedLimitExceeded)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{Address: issuer.String(), Coins: coinsToSend},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}})
	requireT.True(types.ErrWhitelistedLimitExceeded.Is(err))

	// set whitelisted balance to 100
	requireT.NoError(ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, sdk.NewCoin(denom, sdkmath.NewInt(100))))
	whitelistedBalance = ftKeeper.GetWhitelistedBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(100)).String(), whitelistedBalance.String())

	// test query all whitelisted balances
	allBalances, pageRes, err := ftKeeper.GetAccountsWhitelistedBalances(ctx, &query.PageRequest{})
	requireT.NoError(err)
	assertT.Len(allBalances, 1)
	assertT.EqualValues(1, pageRes.GetTotal())
	assertT.EqualValues(recipient.String(), allBalances[0].Address)
	requireT.Equal(sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100))).String(), allBalances[0].Coins.String())

	coinsToSend = sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(50)),
		sdk.NewCoin(unwhitelistableDenom, sdkmath.NewInt(50)),
	)
	// send
	err = bankKeeper.SendCoins(ctx, issuer, recipient, coinsToSend)
	requireT.NoError(err)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{Address: issuer.String(), Coins: coinsToSend},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}})
	requireT.NoError(err)

	// try to send more
	coinsToSend = sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(1)))
	// send
	err = bankKeeper.SendCoins(ctx, issuer, recipient, coinsToSend)
	requireT.ErrorIs(err, types.ErrWhitelistedLimitExceeded)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{Address: issuer.String(), Coins: coinsToSend},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}})
	requireT.ErrorIs(err, types.ErrWhitelistedLimitExceeded)

	// try to whitelist from non issuer address
	err = ftKeeper.SetWhitelistedBalance(ctx, randomAddr, recipient, sdk.NewCoin(denom, sdkmath.NewInt(80)))
	assertT.True(sdkerrors.IsOf(err, cosmoserrors.ErrUnauthorized))

	// reduce whitelisting limit below the current balance
	err = ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, sdk.NewCoin(denom, sdkmath.NewInt(80)))
	requireT.NoError(err)
}

func TestKeeper_FreezeWhitelistMultiSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings1 := types.IssueSettings{
		Issuer:        issuer1,
		Symbol:        "DEF1",
		Subunit:       "def1",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(1000),
		Features:      []types.Feature{types.Feature_freezing},
	}

	settings2 := types.IssueSettings{
		Issuer:        issuer1,
		Symbol:        "DEF2",
		Subunit:       "def2",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(2000),
		Features:      []types.Feature{types.Feature_whitelisting},
	}

	bondDenom, err := testApp.StakingKeeper.BondDenom(ctx)
	requireT.NoError(err)
	// fund with the native coin
	err = testApp.FundAccount(ctx, issuer1, sdk.NewCoins(sdk.NewCoin(bondDenom, sdkmath.NewInt(1000))))
	requireT.NoError(err)

	denom1, err := ftKeeper.Issue(ctx, settings1)
	requireT.NoError(err)

	denom2, err := ftKeeper.Issue(ctx, settings2)
	requireT.NoError(err)

	recipient1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// freeze denom1 partially on the recipient1
	err = ftKeeper.Freeze(ctx, issuer1, recipient1, sdk.NewCoin(denom1, sdkmath.NewInt(10)))
	requireT.NoError(err)

	// whitelist denom2 partially on the recipient2
	err = ftKeeper.SetWhitelistedBalance(ctx, issuer1, recipient2, sdk.NewCoin(denom2, sdkmath.NewInt(10)))
	requireT.NoError(err)

	// multi-send valid amount
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{
			Address: issuer1.String(), Coins: sdk.NewCoins(
				sdk.NewCoin(denom1, sdkmath.NewInt(15)),
				sdk.NewCoin(denom2, sdkmath.NewInt(10)),
				sdk.NewCoin(bondDenom, sdkmath.NewInt(20)),
			),
		},
		[]banktypes.Output{
			// the recipient1 has frozen balance so that amount can be received
			{Address: recipient1.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom1, sdkmath.NewInt(15)))},
			// the recipient2 has whitelisted balance so that is the max amount recipient2 can receive
			{Address: recipient2.String(), Coins: sdk.NewCoins(
				sdk.NewCoin(denom2, sdkmath.NewInt(10)),
				sdk.NewCoin(bondDenom, sdkmath.NewInt(20)),
			)},
		})
	requireT.NoError(err)

	balance := bankKeeper.GetBalance(ctx, recipient1, denom1)
	requireT.Equal(sdk.NewCoin(denom1, sdkmath.NewInt(15)).String(), balance.String())
	balance = bankKeeper.GetBalance(ctx, recipient2, denom2)
	requireT.Equal(sdk.NewCoin(denom2, sdkmath.NewInt(10)).String(), balance.String())
	balance = bankKeeper.GetBalance(ctx, recipient2, bondDenom)
	requireT.Equal(sdk.NewCoin(bondDenom, sdkmath.NewInt(20)).String(), balance.String())

	// multi-send invalid frozen amount
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{
			// we can't return 15 coins since 10 are frozen
			Address: recipient1.String(), Coins: sdk.NewCoins(
				sdk.NewCoin(denom1, sdkmath.NewInt(15)),
				sdk.NewCoin(denom2, sdkmath.NewInt(10)),
			),
		},
		[]banktypes.Output{
			{Address: issuer1.String(), Coins: sdk.NewCoins(
				sdk.NewCoin(denom1, sdkmath.NewInt(15)),
				sdk.NewCoin(denom2, sdkmath.NewInt(10)),
			)},
		})
	requireT.ErrorIs(err, cosmoserrors.ErrInsufficientFunds)

	// multi-send invalid whitelisted amount
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{
			Address: issuer1.String(), Coins: sdk.NewCoins(
				sdk.NewCoin(denom1, sdkmath.NewInt(15)),
				sdk.NewCoin(denom2, sdkmath.NewInt(15)),
			),
		},
		[]banktypes.Output{
			{Address: recipient1.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom1, sdkmath.NewInt(15)))},
			// the recipient2 has whitelisted 10 so can't receive 15
			{Address: recipient2.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom2, sdkmath.NewInt(15)))},
		})
	requireT.ErrorIs(err, types.ErrWhitelistedLimitExceeded)
}

func TestKeeper_DEXLockAndUnlock(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false)

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 10),
		Features:      []types.Feature{types.Feature_freezing},
	}
	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	acc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	// create acc with permanently vesting locked coins
	vestingCoin := sdk.NewInt64Coin(denom, 50)
	baseVestingAccount, err := vestingtypes.NewDelayedVestingAccount(
		authtypes.NewBaseAccountWithAddress(acc),
		sdk.NewCoins(vestingCoin),
		math.MaxInt64,
	)
	requireT.NoError(err)
	account := testApp.App.AccountKeeper.NewAccount(ctx, baseVestingAccount)
	testApp.AccountKeeper.SetAccount(ctx, account)
	requireT.NoError(bankKeeper.SendCoins(ctx, issuer, acc, sdk.NewCoins(vestingCoin)))
	// check vesting locked amount
	requireT.Equal(vestingCoin.Amount.String(), bankKeeper.LockedCoins(ctx, acc).AmountOf(denom).String())

	coinToSend := sdk.NewInt64Coin(denom, 1000)
	// try to DEX lock more than balance
	requireT.ErrorIs(ftKeeper.DEXLock(ctx, acc, coinToSend), cosmoserrors.ErrInsufficientFunds)
	requireT.NoError(bankKeeper.SendCoins(ctx, issuer, acc, sdk.NewCoins(coinToSend)))

	// try to send full balance with the vesting locked coins
	requireT.ErrorIs(
		bankKeeper.SendCoins(ctx, acc, acc, sdk.NewCoins(coinToSend.Add(vestingCoin))),
		cosmoserrors.ErrInsufficientFunds,
	)
	// send max allowed amount
	requireT.NoError(bankKeeper.SendCoins(ctx, acc, acc, sdk.NewCoins(coinToSend)))

	// lock full allowed amount (but without the amount locked by vesting)
	requireT.NoError(ftKeeper.DEXLock(ctx, acc, coinToSend))
	// try to send at least one coin
	requireT.ErrorIs(
		bankKeeper.SendCoins(ctx, acc, acc, sdk.NewCoins(sdk.NewInt64Coin(denom, 1))),
		cosmoserrors.ErrInsufficientFunds,
	)
	// DEX unlock full balance
	requireT.NoError(ftKeeper.DEXUnlock(ctx, acc, coinToSend))
	// DEX lock one more time
	requireT.NoError(ftKeeper.DEXLock(ctx, acc, coinToSend))

	balance := bankKeeper.GetBalance(ctx, acc, denom)
	requireT.Equal(coinToSend.Add(vestingCoin).String(), balance.String())

	// try to DEX lock coins which are locked by the vesting
	requireT.ErrorIs(ftKeeper.DEXLock(ctx, acc, vestingCoin), cosmoserrors.ErrInsufficientFunds)

	// try lock unlock full balance
	requireT.ErrorIs(ftKeeper.DEXUnlock(ctx, acc, balance), cosmoserrors.ErrInsufficientFunds)

	// unlock part
	requireT.NoError(ftKeeper.DEXUnlock(ctx, acc, sdk.NewInt64Coin(denom, 400)))
	requireT.Equal(sdk.NewInt64Coin(denom, 600).String(), ftKeeper.GetDEXLockedBalance(ctx, acc, denom).String())

	// freeze locked balance
	requireT.NoError(ftKeeper.Freeze(ctx, issuer, acc, coinToSend))

	// unlock 2d part, even when it's frozen we allow it
	requireT.NoError(ftKeeper.DEXUnlock(ctx, acc, sdk.NewInt64Coin(denom, 600)))
	requireT.Equal(sdkmath.ZeroInt().String(), ftKeeper.GetDEXLockedBalance(ctx, acc, denom).Amount.String())

	// try to lock now with the frozen coins
	requireT.ErrorIs(ftKeeper.DEXLock(ctx, acc, coinToSend), cosmoserrors.ErrInsufficientFunds)

	// unfreeze part
	requireT.NoError(ftKeeper.Unfreeze(ctx, issuer, acc, sdk.NewInt64Coin(denom, 300)))
	requireT.Equal(sdk.NewInt64Coin(denom, 700).String(), ftKeeper.GetFrozenBalance(ctx, acc, denom).String())

	// now 700 frozen, 50 locked by vesting, 1050 balance
	// try to lock more than allowed
	err = ftKeeper.DEXLock(ctx, acc, sdk.NewInt64Coin(denom, 351))
	requireT.ErrorIs(err, cosmoserrors.ErrInsufficientFunds)
	requireT.ErrorContains(err, "available 350")

	// try to send more than allowed
	err = bankKeeper.SendCoins(ctx, acc, acc, sdk.NewCoins(sdk.NewInt64Coin(denom, 351)))
	requireT.ErrorIs(err, cosmoserrors.ErrInsufficientFunds)
	requireT.ErrorContains(err, "available 350")

	// try to lock with global freezing
	requireT.NoError(ftKeeper.GloballyFreeze(ctx, issuer, denom))
	requireT.ErrorIs(ftKeeper.DEXLock(ctx, acc, sdk.NewInt64Coin(denom, 350)), cosmoserrors.ErrInsufficientFunds)
	// globally unfreeze now and check that we can lock max allowed
	requireT.NoError(ftKeeper.GloballyUnfreeze(ctx, issuer, denom))
	requireT.NoError(ftKeeper.DEXLock(ctx, acc, sdk.NewInt64Coin(denom, 350)))
	// freeze more than balance
	requireT.NoError(ftKeeper.Freeze(ctx, issuer, acc, sdk.NewInt64Coin(denom, 1_000_000)))
}

func TestKeeper_LockAndUnlockNotFT(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false)

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	faucet := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	denom := "denom1"
	requireT.NoError(testApp.FundAccount(ctx, faucet, sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewIntWithDecimal(1, 10)))))

	// create acc with permanently locked coins (native)
	vestingCoin := sdk.NewInt64Coin(denom, 50)
	baseVestingAccount, err := vestingtypes.NewDelayedVestingAccount(
		authtypes.NewBaseAccountWithAddress(acc),
		sdk.NewCoins(vestingCoin),
		math.MaxInt64,
	)
	requireT.NoError(err)
	account := testApp.App.AccountKeeper.NewAccount(ctx, baseVestingAccount)
	testApp.AccountKeeper.SetAccount(ctx, account)
	requireT.NoError(bankKeeper.SendCoins(ctx, faucet, acc, sdk.NewCoins(vestingCoin)))
	// check bank locked amount
	requireT.Equal(vestingCoin.Amount.String(), bankKeeper.LockedCoins(ctx, acc).AmountOf(denom).String())

	coinToSend := sdk.NewInt64Coin(denom, 1000)
	// try to lock more than balance
	requireT.ErrorIs(ftKeeper.DEXLock(ctx, acc, coinToSend), cosmoserrors.ErrInsufficientFunds)
	requireT.NoError(bankKeeper.SendCoins(ctx, faucet, acc, sdk.NewCoins(coinToSend)))

	// try to send full balance with the vesting locked coins
	requireT.ErrorIs(
		bankKeeper.SendCoins(ctx, acc, acc, sdk.NewCoins(coinToSend.Add(vestingCoin))),
		cosmoserrors.ErrInsufficientFunds,
	)

	// lock full allowed amount (but without the amount locked by vesting)
	requireT.NoError(ftKeeper.DEXLock(ctx, acc, coinToSend))

	// try to send at least one coin
	requireT.ErrorIs(
		bankKeeper.SendCoins(ctx, acc, acc, sdk.NewCoins(sdk.NewInt64Coin(denom, 1))),
		cosmoserrors.ErrInsufficientFunds,
	)

	balance := bankKeeper.GetBalance(ctx, acc, denom)
	requireT.Equal(coinToSend.Add(vestingCoin).String(), balance.String())

	// try lock coins which are locked by the vesting
	requireT.ErrorIs(ftKeeper.DEXLock(ctx, acc, vestingCoin), cosmoserrors.ErrInsufficientFunds)

	// try lock unlock full balance
	requireT.ErrorIs(ftKeeper.DEXUnlock(ctx, acc, balance), cosmoserrors.ErrInsufficientFunds)

	// unlock part
	requireT.NoError(ftKeeper.DEXUnlock(ctx, acc, sdk.NewInt64Coin(denom, 400)))
	requireT.Equal(sdk.NewInt64Coin(denom, 600).String(), ftKeeper.GetDEXLockedBalance(ctx, acc, denom).String())
}

func TestKeeper_IBC(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settingsWithoutIBC := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(666),
	}

	denomWithoutIBC, err := ftKeeper.Issue(ctx, settingsWithoutIBC)
	requireT.NoError(err)

	settingsWithIBC := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     1,
		Description:   "ABC Desc",
		InitialAmount: sdkmath.NewInt(666),
		Features:      []types.Feature{types.Feature_ibc},
	}

	denomWithIBC, err := ftKeeper.Issue(ctx, settingsWithIBC)
	requireT.NoError(err)

	// Trick the ctx to look like an outgoing IBC,
	// so we may use regular bank send to test the logic.
	ctx = sdk.UnwrapSDKContext(wibctransfertypes.WithPurpose(ctx, wibctransfertypes.PurposeOut))

	// transferring denom with disabled IBC should fail
	err = bankKeeper.SendCoins(
		ctx,
		issuer,
		recipient,
		sdk.NewCoins(sdk.NewCoin(denomWithoutIBC, sdkmath.NewInt(100))),
	)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// transferring denom with enabled IBC should succeed
	err = bankKeeper.SendCoins(
		ctx,
		issuer,
		recipient,
		sdk.NewCoins(sdk.NewCoin(denomWithIBC, sdkmath.NewInt(100))),
	)
	requireT.NoError(err)
}

// TestKeeper_AllInOne tests send and multi send with tokens that have all features enabled
// and applied.
func TestKeeper_AllInOne(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(1000),
		Features: []types.Feature{
			types.Feature_freezing,
			types.Feature_burning,
			types.Feature_minting,
			types.Feature_whitelisting,
		},
		BurnRate:           sdkmath.LegacyMustNewDecFromStr("0.1"),
		SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("0.05"),
	}

	bondDenom, err := testApp.StakingKeeper.BondDenom(ctx)
	requireT.NoError(err)
	// fund with the native coin
	err = testApp.FundAccount(ctx, issuer, sdk.NewCoins(sdk.NewCoin(bondDenom, sdkmath.NewInt(1000))))
	requireT.NoError(err)

	denom1, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	recipient1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// freeze denom1 partially on the recipient1
	err = ftKeeper.Freeze(ctx, issuer, recipient1, sdk.NewCoin(denom1, sdkmath.NewInt(10)))
	requireT.NoError(err)

	// whitelist recipients
	requireT.NoError(ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient1, sdk.NewCoin(denom1, sdkmath.NewInt(10))))
	requireT.NoError(ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient2, sdk.NewCoin(denom1, sdkmath.NewInt(10))))

	// multi-send valid amount
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{
			Address: issuer.String(), Coins: sdk.NewCoins(
				sdk.NewCoin(denom1, sdkmath.NewInt(20)),
				sdk.NewCoin(bondDenom, sdkmath.NewInt(40)),
			),
		},
		[]banktypes.Output{
			// the recipient1 has frozen balance so that amount can be received
			{Address: recipient1.String(), Coins: sdk.NewCoins(
				sdk.NewCoin(denom1, sdkmath.NewInt(10)),
				sdk.NewCoin(bondDenom, sdkmath.NewInt(20)),
			)},
			// the recipient2 has whitelisted balance so that is the max amount recipient2 can receive
			{Address: recipient2.String(), Coins: sdk.NewCoins(
				sdk.NewCoin(denom1, sdkmath.NewInt(10)),
				sdk.NewCoin(bondDenom, sdkmath.NewInt(20)),
			)},
		})
	requireT.NoError(err)
}

func TestKeeper_GetIssuerTokens(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})
	ftKeeper := testApp.AssetFTKeeper

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	numberOfTokens := 5
	for i := 0; i < numberOfTokens; i++ {
		settings := types.IssueSettings{
			Issuer:        addr,
			Symbol:        "ABC" + uuid.NewString()[:4],
			Description:   "ABC Desc",
			Subunit:       "abc" + uuid.NewString()[:4],
			Precision:     8,
			InitialAmount: sdkmath.NewInt(10),
		}
		denom, err := ftKeeper.Issue(ctx, settings)
		requireT.NoError(err)
		requireT.Equal(types.BuildDenom(settings.Subunit, settings.Issuer), denom)
	}

	tokens, _, err := ftKeeper.GetIssuerTokens(ctx, addr, &query.PageRequest{
		Limit: 3,
	})
	requireT.NoError(err)
	requireT.Len(tokens, 3)

	tokens, _, err = ftKeeper.GetIssuerTokens(ctx, addr, &query.PageRequest{
		Limit: uint64(numberOfTokens + 1),
	})
	requireT.NoError(err)
	requireT.Len(tokens, numberOfTokens)
}

type bankAssertion struct {
	t   require.TestingT
	bk  wbankkeeper.BaseKeeperWrapper
	ctx sdk.Context
}

func (ba bankAssertion) assertCoinDistribution(denom string, dist map[*sdk.AccAddress]int64) {
	requireT := require.New(ba.t)
	total := int64(0)
	for acc, expectedBalance := range dist {
		total += expectedBalance
		getBalance := ba.bk.GetBalance(ba.ctx, *acc, denom)
		requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(expectedBalance)).String(), getBalance.String())
	}

	totalSupply := ba.bk.GetSupply(ba.ctx, denom)
	requireT.Equal(totalSupply.String(), sdk.NewCoin(denom, sdkmath.NewInt(total)).String())
}

func newBankAsserter(
	ctx sdk.Context,
	t require.TestingT,
	bk wbankkeeper.BaseKeeperWrapper,
) bankAssertion {
	return bankAssertion{
		t:   t,
		bk:  bk,
		ctx: ctx,
	}
}

func TestKeeper_TransferAdmin(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(666),
		Features:      []types.Feature{},
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	newAdmin := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// try to transfer admin of non-existent denom
	nonExistentDenom := types.BuildDenom("nonexist", issuer)
	err = ftKeeper.TransferAdmin(ctx, issuer, newAdmin, nonExistentDenom)
	requireT.ErrorIs(err, types.ErrTokenNotFound)

	// try to transfer admin from non admin address
	randomAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = ftKeeper.TransferAdmin(ctx, randomAddr, newAdmin, denom)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// transfer admin, query admin of definition
	err = ftKeeper.TransferAdmin(ctx, issuer, newAdmin, denom)
	requireT.NoError(err)
	def, err := ftKeeper.GetDefinition(ctx, denom)
	requireT.NoError(err)
	requireT.Equal(newAdmin.String(), def.Admin)

	// try to transfer from issuer which is not admin anymore
	err = ftKeeper.TransferAdmin(ctx, issuer, randomAddr, denom)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to transfer from admin back to the original issuer
	err = ftKeeper.TransferAdmin(ctx, newAdmin, issuer, denom)
	requireT.NoError(err)
}

func TestKeeper_TransferAdmin_Mint(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	originalIssuerAddr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	adminAddr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	// Issue an unmintable fungible token
	settings := types.IssueSettings{
		Issuer:        originalIssuerAddr,
		Symbol:        "NotMintable",
		Subunit:       "notmintable",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(777),
		Features: []types.Feature{
			types.Feature_freezing,
			types.Feature_burning,
		},
	}

	unmintableDenom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)
	requireT.Equal(types.BuildDenom(settings.Symbol, settings.Issuer), unmintableDenom)

	// transfer the unmintableDenom to new admin
	err = ftKeeper.TransferAdmin(ctx, originalIssuerAddr, adminAddr, unmintableDenom)
	requireT.NoError(err)

	// try to mint unmintable token
	err = ftKeeper.Mint(ctx, adminAddr, adminAddr, sdk.NewCoin(unmintableDenom, sdkmath.NewInt(100)))
	requireT.ErrorIs(err, types.ErrFeatureDisabled)

	// Issue a mintable fungible token
	settings = types.IssueSettings{
		Issuer:        originalIssuerAddr,
		Symbol:        "mintable",
		Subunit:       "mintable",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(777),
		Features: []types.Feature{
			types.Feature_minting,
		},
	}

	mintableDenom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	// transfer the mintableDenom to new admin
	err = ftKeeper.TransferAdmin(ctx, originalIssuerAddr, adminAddr, mintableDenom)
	requireT.NoError(err)

	// try to mint as non-issuer
	randomAddr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	err = ftKeeper.Mint(ctx, randomAddr, randomAddr, sdk.NewCoin(mintableDenom, sdkmath.NewInt(100)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to mint as original issuer
	err = ftKeeper.Mint(ctx, originalIssuerAddr, originalIssuerAddr, sdk.NewCoin(mintableDenom, sdkmath.NewInt(100)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// mint tokens and check balance and total supply
	err = ftKeeper.Mint(ctx, adminAddr, adminAddr, sdk.NewCoin(mintableDenom, sdkmath.NewInt(100)))
	requireT.NoError(err)

	balance := bankKeeper.GetBalance(ctx, adminAddr, mintableDenom)
	requireT.EqualValues(sdk.NewCoin(mintableDenom, sdkmath.NewInt(100)).String(), balance.String())

	totalSupply, err := bankKeeper.TotalSupply(ctx, &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(877), totalSupply.Supply.AmountOf(mintableDenom))

	// mint to another account as original issuer
	err = ftKeeper.Mint(ctx, originalIssuerAddr, randomAddr, sdk.NewCoin(mintableDenom, sdkmath.NewInt(100)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// mint to another account
	err = ftKeeper.Mint(ctx, adminAddr, randomAddr, sdk.NewCoin(mintableDenom, sdkmath.NewInt(100)))
	requireT.NoError(err)

	balance = bankKeeper.GetBalance(ctx, randomAddr, mintableDenom)
	requireT.EqualValues(sdk.NewCoin(mintableDenom, sdkmath.NewInt(100)), balance)

	totalSupply, err = bankKeeper.TotalSupply(ctx, &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(977), totalSupply.Supply.AmountOf(mintableDenom))
}

func TestKeeper_TransferAdmin_Burn(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	originalIssuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	admin := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	recipient := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	// Issue an unburnable fungible token
	settings := types.IssueSettings{
		Issuer:        originalIssuer,
		Symbol:        "NotBurnable",
		Subunit:       "notburnable",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(877),
		Features: []types.Feature{
			types.Feature_freezing,
			types.Feature_minting,
		},
	}

	unburnableDenom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)
	requireT.Equal(types.BuildDenom(settings.Symbol, settings.Issuer), unburnableDenom)

	// transfer the unburnableDenom to new admin
	err = ftKeeper.TransferAdmin(ctx, originalIssuer, admin, unburnableDenom)
	requireT.NoError(err)

	// send to admin address as original issuer
	err = bankKeeper.SendCoins(ctx, originalIssuer, admin, sdk.NewCoins(sdk.NewCoin(unburnableDenom, sdkmath.NewInt(777))))
	requireT.NoError(err)

	// send to new recipient address as admin
	err = bankKeeper.SendCoins(ctx, admin, recipient, sdk.NewCoins(sdk.NewCoin(unburnableDenom, sdkmath.NewInt(100))))
	requireT.NoError(err)

	// try to burn unburnable token from the recipient account
	err = ftKeeper.Burn(ctx, recipient, sdk.NewCoin(unburnableDenom, sdkmath.NewInt(100)))
	requireT.ErrorIs(err, types.ErrFeatureDisabled)

	// try to burn unburnable token from the original issuer account
	err = ftKeeper.Burn(ctx, originalIssuer, sdk.NewCoin(unburnableDenom, sdkmath.NewInt(100)))
	requireT.ErrorIs(err, types.ErrFeatureDisabled)

	// try to burn unburnable token from the admin account
	err = ftKeeper.Burn(ctx, admin, sdk.NewCoin(unburnableDenom, sdkmath.NewInt(100)))
	requireT.NoError(err)

	// Issue a burnable fungible token
	settings = types.IssueSettings{
		Issuer:        originalIssuer,
		Symbol:        "burnable",
		Subunit:       "burnable",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(877),
		Features: []types.Feature{
			types.Feature_burning,
			types.Feature_freezing,
		},
	}

	burnableDenom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	// transfer the burnableDenom to new admin
	err = ftKeeper.TransferAdmin(ctx, originalIssuer, admin, burnableDenom)
	requireT.NoError(err)

	// send to admin address as original issuer
	err = bankKeeper.SendCoins(ctx, originalIssuer, admin, sdk.NewCoins(sdk.NewCoin(burnableDenom, sdkmath.NewInt(777))))
	requireT.NoError(err)

	// send to new recipient address
	err = bankKeeper.SendCoins(ctx, admin, recipient, sdk.NewCoins(sdk.NewCoin(burnableDenom, sdkmath.NewInt(200))))
	requireT.NoError(err)

	// try to burn as non-admin
	err = ftKeeper.Burn(ctx, recipient, sdk.NewCoin(burnableDenom, sdkmath.NewInt(50)))
	requireT.NoError(err)

	// try to burn as original issuer
	err = ftKeeper.Burn(ctx, originalIssuer, sdk.NewCoin(burnableDenom, sdkmath.NewInt(50)))
	requireT.NoError(err)

	// burn tokens and check balance and total supply
	err = ftKeeper.Burn(ctx, admin, sdk.NewCoin(burnableDenom, sdkmath.NewInt(100)))
	requireT.NoError(err)

	balance := bankKeeper.GetBalance(ctx, admin, burnableDenom)
	requireT.EqualValues(sdk.NewCoin(burnableDenom, sdkmath.NewInt(477)), balance)

	totalSupply, err := bankKeeper.TotalSupply(ctx, &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(677), totalSupply.Supply.AmountOf(burnableDenom))

	// try to freeze the original issuer
	err = ftKeeper.Freeze(ctx, originalIssuer, admin, sdk.NewCoin(burnableDenom, sdkmath.NewInt(600)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to freeze the admin (admin can't be frozen)
	err = ftKeeper.Freeze(ctx, admin, admin, sdk.NewCoin(burnableDenom, sdkmath.NewInt(600)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to burn non-admin frozen coins
	err = ftKeeper.Freeze(ctx, admin, recipient, sdk.NewCoin(burnableDenom, sdkmath.NewInt(100)))
	requireT.NoError(err)
	err = ftKeeper.Burn(ctx, recipient, sdk.NewCoin(burnableDenom, sdkmath.NewInt(100)))
	requireT.ErrorIs(err, cosmoserrors.ErrInsufficientFunds)
}

func TestKeeper_TransferAdmin_FreezeUnfreeze(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	admin := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(666),
		Features:      []types.Feature{types.Feature_freezing},
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100))))
	requireT.NoError(err)

	err = ftKeeper.TransferAdmin(ctx, issuer, admin, denom)
	requireT.NoError(err)

	// try to freeze from issuer address which is non admin anymore
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// freeze, query frozen
	err = ftKeeper.Freeze(ctx, admin, recipient, sdk.NewCoin(denom, sdkmath.NewInt(40)))
	requireT.NoError(err)
	frozenBalance := ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(40)).String(), frozenBalance.String())

	// test query all frozen
	allBalances, pageRes, err := ftKeeper.GetAccountsFrozenBalances(ctx, &query.PageRequest{})
	requireT.NoError(err)
	assertT.Len(allBalances, 1)
	assertT.EqualValues(1, pageRes.GetTotal())
	assertT.EqualValues(recipient.String(), allBalances[0].Address)
	requireT.Equal(sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(40))).String(), allBalances[0].Coins.String())

	// try to unfreeze from issuer address which is non admin anymore
	err = ftKeeper.Unfreeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdkmath.NewInt(80)))
	assertT.True(sdkerrors.IsOf(err, cosmoserrors.ErrUnauthorized))

	// set absolute frozen amount
	err = ftKeeper.SetFrozen(ctx, admin, recipient, sdk.NewCoin(denom, sdkmath.NewInt(100)))
	requireT.NoError(err)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(100)), frozenBalance)

	// unfreeze, query frozen
	err = ftKeeper.Unfreeze(ctx, admin, recipient, sdk.NewCoin(denom, sdkmath.NewInt(100)))
	requireT.NoError(err)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(0)), frozenBalance)
}

func TestKeeper_TransferAdmin_GlobalFreezeUnfreeze(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	admin := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	freezableSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "FREEZE",
		Subunit:       "freeze",
		Precision:     6,
		Description:   "FREEZE Desc",
		InitialAmount: sdkmath.NewInt(877),
		Features:      []types.Feature{types.Feature_freezing},
	}

	freezableDenom, err := ftKeeper.Issue(ctx, freezableSettings)
	requireT.NoError(err)

	_, err = ftKeeper.GetToken(ctx, freezableDenom)
	requireT.NoError(err)

	err = ftKeeper.TransferAdmin(ctx, issuer, admin, freezableDenom)
	requireT.NoError(err)

	token, err := ftKeeper.GetToken(ctx, freezableDenom)
	requireT.NoError(err)
	assertT.Equal(admin.String(), token.Admin)

	unfreezableSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "NOFREEZE",
		Subunit:       "nofreeze",
		Precision:     6,
		Description:   "NOFREEZE Desc",
		InitialAmount: sdkmath.NewInt(877),
		Features:      []types.Feature{},
	}

	unfreezableDenom, err := ftKeeper.Issue(ctx, unfreezableSettings)
	requireT.NoError(err)
	_, err = ftKeeper.GetToken(ctx, unfreezableDenom)
	requireT.NoError(err)

	err = ftKeeper.TransferAdmin(ctx, issuer, admin, unfreezableDenom)
	requireT.NoError(err)

	token, err = ftKeeper.GetToken(ctx, unfreezableDenom)
	requireT.NoError(err)
	assertT.Equal(admin.String(), token.Admin)

	err = bankKeeper.SendCoins(ctx, issuer, admin, sdk.NewCoins(
		sdk.NewCoin(freezableDenom, sdkmath.NewInt(777)),
		sdk.NewCoin(unfreezableDenom, sdkmath.NewInt(777)),
	))
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, admin, recipient, sdk.NewCoins(
		sdk.NewCoin(freezableDenom, sdkmath.NewInt(100)),
		sdk.NewCoin(unfreezableDenom, sdkmath.NewInt(100)),
	))
	requireT.NoError(err)

	// try to global-freeze non-existent
	nonExistentDenom := types.BuildDenom("nonexist", admin)
	err = ftKeeper.GloballyFreeze(ctx, admin, nonExistentDenom)
	assertT.True(sdkerrors.IsOf(err, types.ErrTokenNotFound))

	// try to global-freeze unfreezable Token
	err = ftKeeper.GloballyFreeze(ctx, admin, unfreezableDenom)
	requireT.ErrorIs(err, types.ErrFeatureDisabled)

	// try to global-freeze from original issuer address which is not admin anymore
	err = ftKeeper.GloballyFreeze(ctx, issuer, freezableDenom)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to global-freeze from non admin address
	randomAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = ftKeeper.GloballyFreeze(ctx, randomAddr, freezableDenom)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// freeze twice to check global-freeze idempotence
	err = ftKeeper.GloballyFreeze(ctx, admin, freezableDenom)
	requireT.NoError(err)
	err = ftKeeper.GloballyFreeze(ctx, admin, freezableDenom)
	requireT.NoError(err)
	frozenToken, err := ftKeeper.GetToken(ctx, freezableDenom)
	requireT.NoError(err)
	assertT.True(frozenToken.GloballyFrozen)

	// try to global-unfreeze from original issuer address which is not admin anymore
	err = ftKeeper.GloballyUnfreeze(ctx, issuer, freezableDenom)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to global-unfreeze from non admin address
	err = ftKeeper.GloballyUnfreeze(ctx, randomAddr, freezableDenom)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// unfreeze twice to check global-unfreeze idempotence
	err = ftKeeper.GloballyUnfreeze(ctx, admin, freezableDenom)
	requireT.NoError(err)
	err = ftKeeper.GloballyUnfreeze(ctx, admin, freezableDenom)
	requireT.NoError(err)
	unfrozenToken, err := ftKeeper.GetToken(ctx, freezableDenom)
	requireT.NoError(err)
	assertT.False(unfrozenToken.GloballyFrozen)

	// freeze, try to send & verify balance
	err = ftKeeper.GloballyFreeze(ctx, admin, freezableDenom)
	requireT.NoError(err)
	coinsToSend := sdk.NewCoins(sdk.NewCoin(freezableDenom, sdkmath.NewInt(10)))
	// send
	err = bankKeeper.SendCoins(ctx, recipient, randomAddr, coinsToSend)
	requireT.ErrorIs(err, types.ErrGloballyFrozen)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{Address: recipient.String(), Coins: coinsToSend},
		[]banktypes.Output{{Address: randomAddr.String(), Coins: coinsToSend}})
	requireT.ErrorIs(err, types.ErrGloballyFrozen)

	// unfreeze, try to send & verify balance
	err = ftKeeper.GloballyUnfreeze(ctx, admin, freezableDenom)
	requireT.NoError(err)
	coinsToSend = sdk.NewCoins(sdk.NewCoin(freezableDenom, sdkmath.NewInt(6)))
	// send
	err = bankKeeper.SendCoins(ctx, recipient, randomAddr, coinsToSend)
	requireT.NoError(err)
	balance := bankKeeper.GetBalance(ctx, randomAddr, freezableDenom)
	requireT.Equal(sdk.NewCoin(freezableDenom, sdkmath.NewInt(6)), balance)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{Address: recipient.String(), Coins: coinsToSend},
		[]banktypes.Output{{Address: randomAddr.String(), Coins: coinsToSend}})
	requireT.NoError(err)
	balance = bankKeeper.GetBalance(ctx, randomAddr, freezableDenom)
	requireT.Equal(sdk.NewCoin(freezableDenom, sdkmath.NewInt(12)), balance)
}

func TestKeeper_TransferAdmin_Clawback(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper
	accountKeeper := testApp.AccountKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	admin := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(1666),
		Features: []types.Feature{
			types.Feature_clawback,
			types.Feature_freezing,
		},
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	err = bankKeeper.SendCoins(ctx, issuer, admin, sdk.Coins{sdk.NewCoin(denom, sdkmath.NewInt(666))})
	requireT.NoError(err)

	// transfer the denom to new admin
	err = ftKeeper.TransferAdmin(ctx, issuer, admin, denom)
	requireT.NoError(err)

	clawbackDisabledSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     1,
		Description:   "ABC Desc",
		InitialAmount: sdkmath.NewInt(1666),
		Features:      []types.Feature{},
	}

	clawbackDisabledDenom, err := ftKeeper.Issue(ctx, clawbackDisabledSettings)
	requireT.NoError(err)
	_, err = ftKeeper.GetToken(ctx, clawbackDisabledDenom)
	requireT.NoError(err)

	err = bankKeeper.SendCoins(ctx, issuer, admin, sdk.Coins{sdk.NewCoin(clawbackDisabledDenom, sdkmath.NewInt(666))})
	requireT.NoError(err)

	// transfer the clawbackDisabledDenom to new admin
	err = ftKeeper.TransferAdmin(ctx, issuer, admin, clawbackDisabledDenom)
	requireT.NoError(err)

	account := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, admin, account, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(100)),
		sdk.NewCoin(clawbackDisabledDenom, sdkmath.NewInt(100)),
	))
	requireT.NoError(err)

	// try to clawback non-existent denom
	nonExistentDenom := types.BuildDenom("nonexist", admin)
	err = ftKeeper.Clawback(ctx, admin, account, sdk.NewCoin(nonExistentDenom, sdkmath.NewInt(10)))
	assertT.True(sdkerrors.IsOf(err, types.ErrTokenNotFound))

	// try to clawback clawbackDisabled Token
	err = ftKeeper.Clawback(ctx, admin, account, sdk.NewCoin(clawbackDisabledDenom, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, types.ErrFeatureDisabled)

	// try to clawback by non admin address
	randomAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = ftKeeper.Clawback(ctx, randomAddr, account, sdk.NewCoin(denom, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to clawback from admin address
	randomAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = ftKeeper.Clawback(ctx, randomAddr, admin, sdk.NewCoin(denom, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to clawback from module address
	moduleAddr := accountKeeper.GetModuleAccount(ctx, types.ModuleName).GetAddress()
	err = ftKeeper.Clawback(ctx, admin, moduleAddr, sdk.NewCoin(denom, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to clawback from the original issuer address which is not admin anymore
	err = ftKeeper.Clawback(ctx, issuer, account, sdk.NewCoin(denom, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to clawback 0 balance
	err = ftKeeper.Clawback(ctx, admin, account, sdk.NewCoin(denom, sdkmath.NewInt(0)))
	requireT.ErrorIs(err, cosmoserrors.ErrInvalidCoins)

	// try to clawback more than balance
	err = ftKeeper.Clawback(ctx, admin, account, sdk.NewCoin(denom, sdkmath.NewInt(110)))
	requireT.ErrorIs(err, cosmoserrors.ErrInsufficientFunds)

	// clawback, query balance
	issuerBalanceBefore := bankKeeper.GetBalance(ctx, admin, denom)
	accountBalanceBefore := bankKeeper.GetBalance(ctx, account, denom)
	err = ftKeeper.Clawback(ctx, admin, account, sdk.NewCoin(denom, sdkmath.NewInt(40)))
	requireT.NoError(err)
	issuerBalanceAfter := bankKeeper.GetBalance(ctx, admin, denom)
	accountBalanceAfter := bankKeeper.GetBalance(ctx, account, denom)
	requireT.Equal(issuerBalanceBefore.Add(sdk.NewCoin(denom, sdkmath.NewInt(40))), issuerBalanceAfter)
	requireT.Equal(accountBalanceBefore.Sub(sdk.NewCoin(denom, sdkmath.NewInt(40))), accountBalanceAfter)

	// clawback frozen token, query balance
	err = ftKeeper.Freeze(ctx, admin, account, sdk.NewCoin(denom, sdkmath.NewInt(60)))
	requireT.NoError(err)
	err = ftKeeper.Clawback(ctx, admin, account, sdk.NewCoin(denom, sdkmath.NewInt(60)))
	requireT.NoError(err)
	accountBalance := bankKeeper.GetBalance(ctx, account, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(0)), accountBalance)
}

func TestKeeper_TransferAdmin_Whitelist(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	admin := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(766),
		Features:      []types.Feature{types.Feature_whitelisting},
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	err = ftKeeper.TransferAdmin(ctx, issuer, admin, denom)
	requireT.NoError(err)

	unwhitelistableSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     1,
		Description:   "ABC Desc",
		InitialAmount: sdkmath.NewInt(766),
		Features:      []types.Feature{},
	}

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	unwhitelistableDenom, err := ftKeeper.Issue(ctx, unwhitelistableSettings)
	requireT.NoError(err)
	_, err = ftKeeper.GetToken(ctx, unwhitelistableDenom)
	requireT.NoError(err)

	err = ftKeeper.TransferAdmin(ctx, issuer, admin, unwhitelistableDenom)
	requireT.NoError(err)

	err = bankKeeper.SendCoins(ctx, issuer, admin, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(666)),
		sdk.NewCoin(unwhitelistableDenom, sdkmath.NewInt(666)),
	))
	requireT.NoError(err)

	// whitelisting fails on unwhitelistable token
	err = ftKeeper.SetWhitelistedBalance(ctx, admin, recipient, sdk.NewCoin(unwhitelistableDenom, sdkmath.NewInt(1)))
	requireT.ErrorIs(err, types.ErrFeatureDisabled)

	// try to whitelist non-existent denom
	nonExistentDenom := types.BuildDenom("nonexist", admin)
	err = ftKeeper.SetWhitelistedBalance(ctx, admin, recipient, sdk.NewCoin(nonExistentDenom, sdkmath.NewInt(10)))
	assertT.True(sdkerrors.IsOf(err, types.ErrTokenNotFound))

	// try to whitelist from original issuer which is not admin anymore
	err = ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, sdk.NewCoin(denom, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to whitelist from non admin address
	randomAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = ftKeeper.SetWhitelistedBalance(ctx, randomAddr, recipient, sdk.NewCoin(denom, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to whitelist the admin (admin can't be whitelisted)
	err = ftKeeper.SetWhitelistedBalance(ctx, admin, admin, sdk.NewCoin(denom, sdkmath.NewInt(1)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to whitelist the original issuer which is not admin anymore
	err = ftKeeper.SetWhitelistedBalance(ctx, admin, issuer, sdk.NewCoin(denom, sdkmath.NewInt(1)))
	requireT.NoError(err)

	// set whitelisted balance to 0
	requireT.NoError(ftKeeper.SetWhitelistedBalance(ctx, admin, recipient, sdk.NewCoin(denom, sdkmath.NewInt(0))))
	whitelistedBalance := ftKeeper.GetWhitelistedBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(0)).String(), whitelistedBalance.String())

	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100)))
	// send
	err = bankKeeper.SendCoins(ctx, admin, recipient, coinsToSend)
	requireT.ErrorIs(err, types.ErrWhitelistedLimitExceeded)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{Address: admin.String(), Coins: coinsToSend},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}})
	requireT.True(types.ErrWhitelistedLimitExceeded.Is(err))

	// set whitelisted balance to 100
	requireT.NoError(ftKeeper.SetWhitelistedBalance(ctx, admin, recipient, sdk.NewCoin(denom, sdkmath.NewInt(100))))
	whitelistedBalance = ftKeeper.GetWhitelistedBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(100)).String(), whitelistedBalance.String())

	// test query all whitelisted balances
	allBalances, pageRes, err := ftKeeper.GetAccountsWhitelistedBalances(ctx, &query.PageRequest{})
	requireT.NoError(err)
	assertT.Len(allBalances, 2)
	assertT.EqualValues(2, pageRes.GetTotal())

	if recipient.String() != allBalances[0].Address {
		slices.Reverse(allBalances)
	}

	assertT.EqualValues(recipient.String(), allBalances[0].Address)
	assertT.EqualValues(issuer.String(), allBalances[1].Address)
	requireT.Equal(sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100))).String(), allBalances[0].Coins.String())

	coinsToSend = sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(50)),
		sdk.NewCoin(unwhitelistableDenom, sdkmath.NewInt(50)),
	)
	// send
	err = bankKeeper.SendCoins(ctx, admin, recipient, coinsToSend)
	requireT.NoError(err)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{Address: admin.String(), Coins: coinsToSend},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}})
	requireT.NoError(err)

	// try to send more
	coinsToSend = sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(1)))
	// send
	err = bankKeeper.SendCoins(ctx, admin, recipient, coinsToSend)
	requireT.ErrorIs(err, types.ErrWhitelistedLimitExceeded)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{Address: admin.String(), Coins: coinsToSend},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}})
	requireT.ErrorIs(err, types.ErrWhitelistedLimitExceeded)

	// try to whitelist from non admin address
	err = ftKeeper.SetWhitelistedBalance(ctx, randomAddr, recipient, sdk.NewCoin(denom, sdkmath.NewInt(80)))
	assertT.True(sdkerrors.IsOf(err, cosmoserrors.ErrUnauthorized))

	// reduce whitelisting limit below the current balance
	err = ftKeeper.SetWhitelistedBalance(ctx, admin, recipient, sdk.NewCoin(denom, sdkmath.NewInt(80)))
	requireT.NoError(err)
}

func TestKeeper_TransferAdmin_IBC(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	admin := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settingsWithoutIBC := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(766),
	}

	denomWithoutIBC, err := ftKeeper.Issue(ctx, settingsWithoutIBC)
	requireT.NoError(err)

	err = ftKeeper.TransferAdmin(ctx, issuer, admin, denomWithoutIBC)
	requireT.NoError(err)

	settingsWithIBC := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     1,
		Description:   "ABC Desc",
		InitialAmount: sdkmath.NewInt(766),
		Features:      []types.Feature{types.Feature_ibc},
	}

	denomWithIBC, err := ftKeeper.Issue(ctx, settingsWithIBC)
	requireT.NoError(err)

	err = ftKeeper.TransferAdmin(ctx, issuer, admin, denomWithIBC)
	requireT.NoError(err)

	err = bankKeeper.SendCoins(ctx, issuer, admin, sdk.NewCoins(
		sdk.NewCoin(denomWithoutIBC, sdkmath.NewInt(666)),
		sdk.NewCoin(denomWithIBC, sdkmath.NewInt(666)),
	))
	requireT.NoError(err)

	// Trick the ctx to look like an outgoing IBC,
	// so we may use regular bank send to test the logic.
	ctx = sdk.UnwrapSDKContext(wibctransfertypes.WithPurpose(ctx, wibctransfertypes.PurposeOut))

	// transferring denom with disabled IBC should fail
	err = bankKeeper.SendCoins(
		ctx,
		admin,
		recipient,
		sdk.NewCoins(sdk.NewCoin(denomWithoutIBC, sdkmath.NewInt(100))),
	)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// transferring denom with enabled IBC should succeed
	err = bankKeeper.SendCoins(
		ctx,
		admin,
		recipient,
		sdk.NewCoins(sdk.NewCoin(denomWithIBC, sdkmath.NewInt(100))),
	)
	requireT.NoError(err)

	// transferring denom with enabled IBC should succeed
	err = bankKeeper.SendCoins(
		ctx,
		issuer,
		recipient,
		sdk.NewCoins(sdk.NewCoin(denomWithIBC, sdkmath.NewInt(100))),
	)
	requireT.NoError(err)
}

// TestKeeper_AllInOne tests send and multi send with tokens that have all features enabled
// and applied while the admin is transferred from issuer.
func TestKeeper_TransferAdmin_AllInOne(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	admin := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(1100),
		Features: []types.Feature{
			types.Feature_freezing,
			types.Feature_burning,
			types.Feature_minting,
			types.Feature_whitelisting,
		},
		BurnRate:           sdkmath.LegacyMustNewDecFromStr("0.1"),
		SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("0.05"),
	}

	bondDenom, err := testApp.StakingKeeper.BondDenom(ctx)
	requireT.NoError(err)
	// fund with the native coin
	err = testApp.FundAccount(ctx, issuer, sdk.NewCoins(sdk.NewCoin(bondDenom, sdkmath.NewInt(1000))))
	requireT.NoError(err)

	denom1, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	err = ftKeeper.TransferAdmin(ctx, issuer, admin, denom1)
	requireT.NoError(err)

	err = bankKeeper.SendCoins(ctx, issuer, admin, sdk.NewCoins(
		sdk.NewCoin(bondDenom, sdkmath.NewInt(1000)),
		sdk.NewCoin(denom1, sdkmath.NewInt(1000)),
	))
	requireT.NoError(err)

	recipient1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// freeze denom1 partially on the recipient1 as original issuer which is not admin anymore
	err = ftKeeper.Freeze(ctx, issuer, recipient1, sdk.NewCoin(denom1, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// freeze denom1 partially on the recipient1
	err = ftKeeper.Freeze(ctx, admin, recipient1, sdk.NewCoin(denom1, sdkmath.NewInt(10)))
	requireT.NoError(err)

	// whitelist recipients as original issuer which is not admin anymore
	err = ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient1, sdk.NewCoin(denom1, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)
	err = ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient2, sdk.NewCoin(denom1, sdkmath.NewInt(10)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// whitelist recipients
	requireT.NoError(ftKeeper.SetWhitelistedBalance(ctx, admin, recipient1, sdk.NewCoin(denom1, sdkmath.NewInt(10))))
	requireT.NoError(ftKeeper.SetWhitelistedBalance(ctx, admin, recipient2, sdk.NewCoin(denom1, sdkmath.NewInt(10))))

	// multi-send valid amount
	err = bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{
			Address: admin.String(), Coins: sdk.NewCoins(
				sdk.NewCoin(denom1, sdkmath.NewInt(20)),
				sdk.NewCoin(bondDenom, sdkmath.NewInt(40)),
			),
		},
		[]banktypes.Output{
			// the recipient1 has frozen balance so that amount can be received
			{Address: recipient1.String(), Coins: sdk.NewCoins(
				sdk.NewCoin(denom1, sdkmath.NewInt(10)),
				sdk.NewCoin(bondDenom, sdkmath.NewInt(20)),
			)},
			// the recipient2 has whitelisted balance so that is the max amount recipient2 can receive
			{Address: recipient2.String(), Coins: sdk.NewCoins(
				sdk.NewCoin(denom1, sdkmath.NewInt(10)),
				sdk.NewCoin(bondDenom, sdkmath.NewInt(20)),
			)},
		})
	requireT.NoError(err)
}

func TestKeeper_ClearAdmin(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	bankKeeper := testApp.BankKeeper
	ftKeeper := testApp.AssetFTKeeper

	admin := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	sender := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:             admin,
		Symbol:             "DEF",
		Subunit:            "def",
		Precision:          1,
		Description:        "DEF Desc",
		InitialAmount:      sdkmath.NewInt(666),
		Features:           []types.Feature{},
		SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("0.1"),
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	// send some amount to an account
	err = bankKeeper.SendCoins(ctx, admin, sender, sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100))))
	requireT.NoError(err)

	// try to clear admin of non-existent denom
	nonExistentDenom := types.BuildDenom("nonexist", admin)
	err = ftKeeper.ClearAdmin(ctx, admin, nonExistentDenom)
	requireT.ErrorIs(err, types.ErrTokenNotFound)

	// try to clear admin from non admin address
	randomAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = ftKeeper.ClearAdmin(ctx, randomAddr, denom)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// clear admin, query admin of definition
	err = ftKeeper.ClearAdmin(ctx, admin, denom)
	requireT.NoError(err)
	def, err := ftKeeper.GetDefinition(ctx, denom)
	requireT.NoError(err)
	requireT.Empty(def.Admin)

	// try to clear from the previous admin which is not admin anymore
	err = ftKeeper.ClearAdmin(ctx, admin, denom)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// send some amount between two accounts
	err = bankKeeper.SendCoins(ctx, sender, recipient, sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100))))
	requireT.NoError(err)
}
