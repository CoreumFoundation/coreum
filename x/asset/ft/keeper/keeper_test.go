package keeper_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/testutil/event"
	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
	wbankkeeper "github.com/CoreumFoundation/coreum/x/wbank/keeper"
)

func TestKeeper_Issue(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	ftParams := types.Params{
		IssueFee: sdk.NewInt64Coin(constant.DenomDev, 10_000_000),
	}
	ftKeeper.SetParams(ctx, ftParams)

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	requireT.NoError(testApp.FundAccount(ctx, addr, sdk.NewCoins(ftParams.IssueFee)))

	settings := types.IssueSettings{
		Issuer:        addr,
		Symbol:        "ABC",
		Description:   "ABC Desc",
		Subunit:       "abc",
		Precision:     8,
		InitialAmount: sdk.NewInt(777),
		Features:      []types.Feature{types.Feature_freezing},
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	// verify issue fee was burnt

	burntStr, err := event.FindStringEventAttribute(ctx.EventManager().ABCIEvents(), banktypes.EventTypeCoinBurn, sdk.AttributeKeyAmount)
	requireT.NoError(err)
	requireT.Equal(ftParams.IssueFee.String(), burntStr)

	// check that balance is 0 meaning issue fee was taken

	balance := bankKeeper.GetBalance(ctx, addr, constant.DenomDev)
	requireT.Equal(sdk.ZeroInt().String(), balance.Amount.String())

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
		BurnRate:           sdk.NewDec(0),
		SendCommissionRate: sdk.NewDec(0),
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
	_, err = ftKeeper.Issue(ctx, st)
	requireT.True(errors.Is(types.ErrInvalidInput, err))

	// check duplicate symbol
	st = settings
	st.Subunit = "test-subunit"
	st.Symbol = "aBc"
	_, err = ftKeeper.Issue(ctx, st)
	requireT.True(errors.Is(types.ErrInvalidInput, err))
}

func TestKeeper_IssueValidateSymbol(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})
	ftKeeper := testApp.AssetFTKeeper

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
		settings := types.IssueSettings{
			Issuer:        addr,
			Symbol:        symbol,
			Subunit:       "subunit",
			Description:   "ABC Desc",
			InitialAmount: sdk.NewInt(777),
			Features:      []types.Feature{types.Feature_freezing},
		}

		_, err := ftKeeper.Issue(ctx, settings)
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

func TestKeeper_IssueValidateSubunit(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})
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
			Description:   "ABC Desc",
			InitialAmount: sdk.NewInt(777),
			Features:      []types.Feature{types.Feature_freezing},
		}

		_, err := ftKeeper.Issue(ctx, settings)
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

func TestKeeper_Issue_WithZeroIssueFee(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper

	ftParams := types.Params{
		IssueFee: sdk.NewCoin(constant.DenomDev, sdk.ZeroInt()),
	}
	ftKeeper.SetParams(ctx, ftParams)

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        addr,
		Symbol:        "ABC",
		Description:   "ABC Desc",
		Subunit:       "abc",
		Precision:     8,
		InitialAmount: sdk.NewInt(777),
		Features:      []types.Feature{types.Feature_freezing},
	}

	_, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)
}

func TestKeeper_Issue_WithNoFundsCoveringFee(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper

	ftParams := types.Params{
		IssueFee: sdk.NewInt64Coin(constant.DenomDev, 10_000_000),
	}
	ftKeeper.SetParams(ctx, ftParams)

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	settings := types.IssueSettings{
		Issuer:        addr,
		Symbol:        "ABC",
		Description:   "ABC Desc",
		Subunit:       "abc",
		Precision:     8,
		InitialAmount: sdk.NewInt(777),
		Features:      []types.Feature{types.Feature_freezing},
	}

	_, err := ftKeeper.Issue(ctx, settings)
	requireT.ErrorIs(err, sdkerrors.ErrInsufficientFunds)
}

func TestKeeper_Mint(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	// Issue an unmintable fungible token
	settings := types.IssueSettings{
		Issuer:        addr,
		Symbol:        "NotMintable",
		Subunit:       "notmintable",
		InitialAmount: sdk.NewInt(777),
		Features: []types.Feature{
			types.Feature_freezing,
			types.Feature_burning,
		},
	}

	unmintableDenom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)
	requireT.Equal(types.BuildDenom(settings.Symbol, settings.Issuer), unmintableDenom)

	// try to mint unmintable token
	err = ftKeeper.Mint(ctx, addr, sdk.NewCoin(unmintableDenom, sdk.NewInt(100)))
	requireT.ErrorIs(types.ErrFeatureDisabled, err)

	// Issue a mintable fungible token
	settings = types.IssueSettings{
		Issuer:        addr,
		Symbol:        "mintable",
		Subunit:       "mintable",
		InitialAmount: sdk.NewInt(777),
		Features: []types.Feature{
			types.Feature_minting,
		},
	}

	mintableDenom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	// try to mint as non-issuer
	randomAddr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	err = ftKeeper.Mint(ctx, randomAddr, sdk.NewCoin(mintableDenom, sdk.NewInt(100)))
	requireT.ErrorIs(sdkerrors.ErrUnauthorized, err)

	// mint tokens and check balance and total supply
	err = ftKeeper.Mint(ctx, addr, sdk.NewCoin(mintableDenom, sdk.NewInt(100)))
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

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	recipient := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	// Issue an unburnable fungible token
	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "NotBurnable",
		Subunit:       "notburnable",
		InitialAmount: sdk.NewInt(777),
		Features: []types.Feature{
			types.Feature_freezing,
			types.Feature_minting,
		},
	}

	unburnableDenom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)
	requireT.Equal(types.BuildDenom(settings.Symbol, settings.Issuer), unburnableDenom)

	// send to new recipient address
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(sdk.NewCoin(unburnableDenom, sdk.NewInt(100))))
	requireT.NoError(err)

	// try to burn unburnable token from the recipient account
	err = ftKeeper.Burn(ctx, recipient, sdk.NewCoin(unburnableDenom, sdk.NewInt(100)))
	requireT.ErrorIs(types.ErrFeatureDisabled, err)

	// try to burn unburnable token from the issuer account
	err = ftKeeper.Burn(ctx, issuer, sdk.NewCoin(unburnableDenom, sdk.NewInt(100)))
	requireT.NoError(err)

	// Issue a burnable fungible token
	settings = types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "burnable",
		Subunit:       "burnable",
		InitialAmount: sdk.NewInt(777),
		Features: []types.Feature{
			types.Feature_burning,
			types.Feature_freezing,
		},
	}

	burnableDenom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	// send to new recipient address
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(sdk.NewCoin(burnableDenom, sdk.NewInt(200))))
	requireT.NoError(err)

	// try to burn as non-issuer
	err = ftKeeper.Burn(ctx, recipient, sdk.NewCoin(burnableDenom, sdk.NewInt(100)))
	requireT.NoError(err)

	// burn tokens and check balance and total supply
	err = ftKeeper.Burn(ctx, issuer, sdk.NewCoin(burnableDenom, sdk.NewInt(100)))
	requireT.NoError(err)

	balance := bankKeeper.GetBalance(ctx, issuer, burnableDenom)
	requireT.EqualValues(sdk.NewCoin(burnableDenom, sdk.NewInt(477)), balance)

	totalSupply, err := bankKeeper.TotalSupply(sdk.WrapSDKContext(ctx), &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewInt(577), totalSupply.Supply.AmountOf(burnableDenom))

	// try to freeze the issuer (issuer can't be frozen)
	err = ftKeeper.Freeze(ctx, issuer, issuer, sdk.NewCoin(burnableDenom, sdk.NewInt(600)))
	requireT.ErrorIs(sdkerrors.ErrUnauthorized, err)

	// try to burn non-issuer frozen coins
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(burnableDenom, sdk.NewInt(100)))
	requireT.NoError(err)
	err = ftKeeper.Burn(ctx, recipient, sdk.NewCoin(burnableDenom, sdk.NewInt(100)))
	requireT.ErrorIs(sdkerrors.ErrInsufficientFunds, err)
}

//nolint:dupl // We don't care
func TestKeeper_BurnRate_BankSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

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
		InitialAmount: sdk.NewInt(600),
		Features:      []types.Feature{},
		BurnRate:      sdk.MustNewDecFromStr("1.01"),
	}

	_, err := assetKeeper.Issue(ctx, settings)
	requireT.ErrorIs(types.ErrInvalidInput, err)

	// issue token
	settings = types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     6,
		Description:   "DEF Desc",
		InitialAmount: sdk.NewInt(600),
		Features:      []types.Feature{},
		BurnRate:      sdk.MustNewDecFromStr("0.25"),
	}

	denom, err := assetKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// send from issuer to recipient (burn must not apply)
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(500)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient: 500,
		&issuer:    100,
	})

	// send from recipient1 to recipient2 (burn must apply)
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(100)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient:  375,
		&recipient2: 100,
		&issuer:     100,
	})

	// send from recipient to issuer account (burn must not apply)
	err = bankKeeper.SendCoins(ctx, recipient, issuer, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(375)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient2: 100,
		&issuer:     475,
	})
}

//nolint:funlen // This is a complex test scenario and breaking it down will make it harder to read
func TestKeeper_BurnRate_BankMultiSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

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
			Issuer:        issuers[i],
			Symbol:        fmt.Sprintf("DEF%d", i),
			Subunit:       fmt.Sprintf("def%d", i),
			Precision:     6,
			Description:   "DEF Desc",
			InitialAmount: sdk.NewInt(1000),
			Features:      []types.Feature{},
			BurnRate:      sdk.MustNewDecFromStr(fmt.Sprintf("0.%d", i+1)),
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
		inputs       []banktypes.Input
		outputs      []banktypes.Output
		distribution map[string]map[*sdk.AccAddress]int64
	}{
		{
			name: "send from issuer to other accounts",
			inputs: []banktypes.Input{
				{Address: issuers[0].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(200)))},
				{Address: issuers[1].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[1], sdk.NewInt(200)))},
			},
			outputs: []banktypes.Output{
				{Address: recipients[0].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdk.NewInt(100)),
					sdk.NewCoin(denoms[1], sdk.NewInt(100)),
				)},
				{Address: recipients[1].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdk.NewInt(100)),
					sdk.NewCoin(denoms[1], sdk.NewInt(100)),
				)},
			},
			distribution: map[string]map[*sdk.AccAddress]int64{
				denoms[0]: {
					&issuers[0]:    800,
					&recipients[0]: 100,
					&recipients[1]: 100,
				},
				denoms[1]: {
					&issuers[1]:    800,
					&recipients[0]: 100,
					&recipients[1]: 100,
				},
			},
		},
		{
			name: "include issuer in senders",
			inputs: []banktypes.Input{
				{Address: issuers[0].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(90)))},
				{Address: recipients[0].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(29)))},
				{Address: recipients[1].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(32)))},
			},
			outputs: []banktypes.Output{
				{Address: recipients[2].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdk.NewInt(89)),
				)},
				{Address: recipients[3].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdk.NewInt(62)),
				)},
			},
			distribution: map[string]map[*sdk.AccAddress]int64{
				denoms[0]: {
					&issuers[0]:    710,
					&recipients[0]: 68, // 100 - 29 - 3 (burn = roundup(29 * 10%))
					&recipients[1]: 64, // 100 - 32 - 4 (burn = roundup(32 * 10%))
					&recipients[2]: 89,
					&recipients[3]: 62,
				},
			},
		},
		{
			name: "include issuer in receivers",
			inputs: []banktypes.Input{
				{Address: recipients[0].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[1], sdk.NewInt(60)))},
				{Address: recipients[1].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[1], sdk.NewInt(40)))},
			},
			outputs: []banktypes.Output{
				{Address: issuers[1].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdk.NewInt(40)),
				)},
				{Address: recipients[2].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdk.NewInt(25)),
				)},
				{Address: recipients[3].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdk.NewInt(35)),
				)},
			},
			distribution: map[string]map[*sdk.AccAddress]int64{
				denoms[1]: {
					&issuers[1]:    840,
					&recipients[0]: 32, // 100 - 60 - 8 (burn = roundup(60 * (60/100) * 20%))
					&recipients[1]: 55, // 100 - 40 - 5 (burn = roundup(40 * (60/100) * 20%))
					&recipients[2]: 25,
					&recipients[3]: 35,
				},
			},
		},
		{
			name: "send all coins back to issuers",
			inputs: []banktypes.Input{
				// coin[0]
				{Address: recipients[0].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(68)))},
				{Address: recipients[1].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(64)))},
				{Address: recipients[2].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(89)))},
				{Address: recipients[3].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(62)))},
				// coin[1]
				{Address: recipients[0].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[1], sdk.NewInt(32)))},
				{Address: recipients[1].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[1], sdk.NewInt(55)))},
				{Address: recipients[2].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[1], sdk.NewInt(25)))},
				{Address: recipients[3].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[1], sdk.NewInt(35)))},
			},
			outputs: []banktypes.Output{
				{Address: issuers[0].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdk.NewInt(283)),
				)},
				{Address: issuers[1].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdk.NewInt(147)),
				)},
			},
			distribution: map[string]map[*sdk.AccAddress]int64{
				denoms[0]: {
					&issuers[0]: 993,
				},
				denoms[1]: {
					&issuers[1]: 987,
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

//nolint:dupl // We don't care
func TestKeeper_SendCommissionRate_BankSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

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
		InitialAmount:      sdk.NewInt(600),
		Features:           []types.Feature{},
		SendCommissionRate: sdk.MustNewDecFromStr("1.01"),
	}

	_, err := assetKeeper.Issue(ctx, settings)
	requireT.ErrorIs(types.ErrInvalidInput, err)

	// issue token
	settings = types.IssueSettings{
		Issuer:             issuer,
		Symbol:             "DEF",
		Subunit:            "def",
		Precision:          6,
		Description:        "DEF Desc",
		InitialAmount:      sdk.NewInt(600),
		Features:           []types.Feature{},
		SendCommissionRate: sdk.MustNewDecFromStr("0.25"),
	}

	denom, err := assetKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// send from issuer to recipient (send commission rate must not apply)
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(500)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient: 500,
		&issuer:    100,
	})

	// send from recipient1 to recipient2 (send commission rate must apply)
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(100)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient:  375,
		&recipient2: 100,
		&issuer:     125,
	})

	// send from recipient to issuer account (send commission rate must not apply)
	err = bankKeeper.SendCoins(ctx, recipient, issuer, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(375)),
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
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

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
		InitialAmount:      sdk.NewInt(600),
		Features:           []types.Feature{},
		BurnRate:           sdk.MustNewDecFromStr("0.5"),
		SendCommissionRate: sdk.MustNewDecFromStr("0.25"),
	}

	denom, err := assetKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// send from issuer to recipient (fees must not apply)
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(500)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient: 500,
		&issuer:    100,
	})

	// send from recipient1 to recipient2 (fees must apply)
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(100)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient:  325,
		&recipient2: 100,
		&issuer:     125,
	})

	// send from recipient to issuer account (fees must not apply)
	err = bankKeeper.SendCoins(ctx, recipient, issuer, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(325)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient2: 100,
		&issuer:     450,
	})
}

//nolint:funlen // this is complex test scenario and breaking it down is not helpful
func TestKeeper_FreezeUnfreeze(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Description:   "DEF Desc",
		InitialAmount: sdk.NewInt(666),
		Features:      []types.Feature{types.Feature_freezing},
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	unfreezableSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Subunit:       "abc",
		Description:   "ABC Desc",
		InitialAmount: sdk.NewInt(666),
		Features:      []types.Feature{},
	}

	unfreezableDenom, err := ftKeeper.Issue(ctx, unfreezableSettings)
	requireT.NoError(err)
	_, err = ftKeeper.GetToken(ctx, unfreezableDenom)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(100)),
		sdk.NewCoin(unfreezableDenom, sdk.NewInt(100)),
	))
	requireT.NoError(err)

	// try to freeze non-existent denom
	nonExistentDenom := types.BuildDenom("nonexist", issuer)
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(nonExistentDenom, sdk.NewInt(10)))
	assertT.True(sdkerrors.IsOf(err, types.ErrTokenNotFound))

	// try to freeze unfreezable Token
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(unfreezableDenom, sdk.NewInt(10)))
	assertT.ErrorIs(types.ErrFeatureDisabled, err)

	// try to freeze from non issuer address
	randomAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = ftKeeper.Freeze(ctx, randomAddr, recipient, sdk.NewCoin(denom, sdk.NewInt(10)))
	assertT.ErrorIs(sdkerrors.ErrUnauthorized, err)

	// try to freeze 0 balance
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(0)))
	requireT.ErrorIs(sdkerrors.ErrInvalidCoins, err)

	// try to unfreeze 0 balance
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(0)))
	requireT.ErrorIs(sdkerrors.ErrInvalidCoins, err)

	// try to freeze more than balance
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(110)))
	requireT.NoError(err)
	frozenBalance := ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	assertT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(110)), frozenBalance)

	// try to unfreeze more than frozen balance
	err = ftKeeper.Unfreeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(130)))
	requireT.ErrorIs(sdkerrors.ErrInsufficientFunds, err)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	assertT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(110)), frozenBalance)

	// set frozen balance back to zero
	err = ftKeeper.Unfreeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(110)))
	requireT.NoError(err)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	assertT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(0)).String(), frozenBalance.String())

	// freeze, query frozen
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(40)))
	requireT.NoError(err)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(40)).String(), frozenBalance.String())

	// test query all frozen
	allBalances, pageRes, err := ftKeeper.GetAccountsFrozenBalances(ctx, &query.PageRequest{})
	assertT.NoError(err)
	assertT.Len(allBalances, 1)
	assertT.EqualValues(1, pageRes.GetTotal())
	assertT.EqualValues(recipient.String(), allBalances[0].Address)
	requireT.Equal(sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(40))).String(), allBalances[0].Coins.String())

	// increase frozen and query
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(40)))
	requireT.NoError(err)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(80)), frozenBalance)

	// try to send more than available
	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(80)))
	// send
	err = bankKeeper.SendCoins(ctx, recipient, issuer, coinsToSend)
	assertT.True(sdkerrors.IsOf(err, sdkerrors.ErrInsufficientFunds))
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: recipient.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: issuer.String(), Coins: coinsToSend}})
	assertT.True(sdkerrors.IsOf(err, sdkerrors.ErrInsufficientFunds))

	// try to send unfrozen balance
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	coinsToSend = sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(10)))
	// send
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, coinsToSend)
	requireT.NoError(err)
	balance := bankKeeper.GetBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(90)), balance)
	balance = bankKeeper.GetBalance(ctx, recipient2, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(10)), balance)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: recipient.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: recipient2.String(), Coins: coinsToSend}})
	requireT.NoError(err)
	balance = bankKeeper.GetBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(80)), balance)
	balance = bankKeeper.GetBalance(ctx, recipient2, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(20)), balance)

	// try to unfreeze from non issuer address
	err = ftKeeper.Unfreeze(ctx, randomAddr, recipient, sdk.NewCoin(denom, sdk.NewInt(80)))
	assertT.True(sdkerrors.IsOf(err, sdkerrors.ErrUnauthorized))

	// unfreeze, query frozen, and try to send
	err = ftKeeper.Unfreeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(80)))
	requireT.NoError(err)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(0)), frozenBalance)
	coinsToSend = sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(40)))
	// send
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, coinsToSend)
	requireT.NoError(err)
	balance = bankKeeper.GetBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(40)), balance)
	balance = bankKeeper.GetBalance(ctx, recipient2, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(60)), balance)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: recipient.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: recipient2.String(), Coins: coinsToSend}})
	requireT.NoError(err)
	balance = bankKeeper.GetBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(0)), balance)
	balance = bankKeeper.GetBalance(ctx, recipient2, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(100)), balance)
}

//nolint:funlen // this is complex tests scenario and breaking it down is not beneficial
func TestKeeper_GlobalFreezeUnfreeze(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	freezableSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "FREEZE",
		Subunit:       "freeze",
		Precision:     6,
		Description:   "FREEZE Desc",
		InitialAmount: sdk.NewInt(777),
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
		InitialAmount: sdk.NewInt(777),
		Features:      []types.Feature{},
	}

	unfreezableDenom, err := ftKeeper.Issue(ctx, unfreezableSettings)
	requireT.NoError(err)
	_, err = ftKeeper.GetToken(ctx, unfreezableDenom)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(
		sdk.NewCoin(freezableDenom, sdk.NewInt(100)),
		sdk.NewCoin(unfreezableDenom, sdk.NewInt(100)),
	))
	requireT.NoError(err)

	// try to global-freeze non-existent
	nonExistentDenom := types.BuildDenom("nonexist", issuer)
	err = ftKeeper.GloballyFreeze(ctx, issuer, nonExistentDenom)
	assertT.True(sdkerrors.IsOf(err, types.ErrTokenNotFound))

	// try to global-freeze unfreezable Token
	err = ftKeeper.GloballyFreeze(ctx, issuer, unfreezableDenom)
	assertT.ErrorIs(types.ErrFeatureDisabled, err)

	// try to global-freeze from non issuer address
	randomAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = ftKeeper.GloballyFreeze(ctx, randomAddr, freezableDenom)
	assertT.ErrorIs(sdkerrors.ErrUnauthorized, err)

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
	assertT.ErrorIs(sdkerrors.ErrUnauthorized, err)

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
	coinsToSend := sdk.NewCoins(sdk.NewCoin(freezableDenom, sdk.NewInt(10)))
	// send
	err = bankKeeper.SendCoins(ctx, recipient, randomAddr, coinsToSend)
	assertT.ErrorIs(types.ErrGloballyFrozen, err)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: recipient.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: randomAddr.String(), Coins: coinsToSend}})
	assertT.ErrorIs(types.ErrGloballyFrozen, err)

	// unfreeze, try to send & verify balance
	err = ftKeeper.GloballyUnfreeze(ctx, issuer, freezableDenom)
	requireT.NoError(err)
	coinsToSend = sdk.NewCoins(sdk.NewCoin(freezableDenom, sdk.NewInt(6)))
	// send
	err = bankKeeper.SendCoins(ctx, recipient, randomAddr, coinsToSend)
	requireT.NoError(err)
	balance := bankKeeper.GetBalance(ctx, randomAddr, freezableDenom)
	requireT.Equal(sdk.NewCoin(freezableDenom, sdk.NewInt(6)), balance)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: recipient.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: randomAddr.String(), Coins: coinsToSend}})
	requireT.NoError(err)
	balance = bankKeeper.GetBalance(ctx, randomAddr, freezableDenom)
	requireT.Equal(sdk.NewCoin(freezableDenom, sdk.NewInt(12)), balance)
}

//nolint:funlen // this is complex test scenario and breaking it down is not helpful
func TestKeeper_Whitelist(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Description:   "DEF Desc",
		InitialAmount: sdk.NewInt(666),
		Features:      []types.Feature{types.Feature_whitelisting},
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	unwhitelistableSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Subunit:       "abc",
		Description:   "ABC Desc",
		InitialAmount: sdk.NewInt(666),
		Features:      []types.Feature{},
	}

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	unwhitelistableDenom, err := ftKeeper.Issue(ctx, unwhitelistableSettings)
	requireT.NoError(err)
	_, err = ftKeeper.GetToken(ctx, unwhitelistableDenom)
	requireT.NoError(err)

	// whitelisting fails on unwhitelistable token
	err = ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, sdk.NewCoin(unwhitelistableDenom, sdk.NewInt(1)))
	requireT.ErrorIs(types.ErrFeatureDisabled, err)

	// try to whitelist non-existent denom
	nonExistentDenom := types.BuildDenom("nonexist", issuer)
	err = ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, sdk.NewCoin(nonExistentDenom, sdk.NewInt(10)))
	assertT.True(sdkerrors.IsOf(err, types.ErrTokenNotFound))

	// try to whitelist from non issuer address
	randomAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = ftKeeper.SetWhitelistedBalance(ctx, randomAddr, recipient, sdk.NewCoin(denom, sdk.NewInt(10)))
	assertT.ErrorIs(sdkerrors.ErrUnauthorized, err)

	// try to whitelist the issuer (issuer can't be whitelisted)
	err = ftKeeper.SetWhitelistedBalance(ctx, issuer, issuer, sdk.NewCoin(denom, sdk.NewInt(1)))
	requireT.ErrorIs(sdkerrors.ErrUnauthorized, err)

	// set whitelisted balance to 0
	requireT.NoError(ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(0))))
	whitelistedBalance := ftKeeper.GetWhitelistedBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(0)).String(), whitelistedBalance.String())

	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100)))
	// send
	err = bankKeeper.SendCoins(ctx, issuer, recipient, coinsToSend)
	requireT.ErrorIs(types.ErrWhitelistedLimitExceeded, err)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: issuer.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}})
	requireT.True(types.ErrWhitelistedLimitExceeded.Is(err))

	// set whitelisted balance to 100
	requireT.NoError(ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(100))))
	whitelistedBalance = ftKeeper.GetWhitelistedBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(100)).String(), whitelistedBalance.String())

	// test query all whitelisted balances
	allBalances, pageRes, err := ftKeeper.GetAccountsWhitelistedBalances(ctx, &query.PageRequest{})
	assertT.NoError(err)
	assertT.Len(allBalances, 1)
	assertT.EqualValues(1, pageRes.GetTotal())
	assertT.EqualValues(recipient.String(), allBalances[0].Address)
	requireT.Equal(sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100))).String(), allBalances[0].Coins.String())

	coinsToSend = sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(50)),
		sdk.NewCoin(unwhitelistableDenom, sdk.NewInt(50)),
	)
	// send
	err = bankKeeper.SendCoins(ctx, issuer, recipient, coinsToSend)
	requireT.NoError(err)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: issuer.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}})
	requireT.NoError(err)

	// try to send more
	coinsToSend = sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1)))
	// send
	err = bankKeeper.SendCoins(ctx, issuer, recipient, coinsToSend)
	requireT.ErrorIs(types.ErrWhitelistedLimitExceeded, err)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: issuer.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}})
	requireT.ErrorIs(types.ErrWhitelistedLimitExceeded, err)

	// try to whitelist from non issuer address
	err = ftKeeper.SetWhitelistedBalance(ctx, randomAddr, recipient, sdk.NewCoin(denom, sdk.NewInt(80)))
	assertT.True(sdkerrors.IsOf(err, sdkerrors.ErrUnauthorized))
}

//nolint:funlen // this is complex test scenario and breaking it down is not helpful
func TestKeeper_FreezeWhitelistMultiSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	issuer2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings1 := types.IssueSettings{
		Issuer:        issuer1,
		Symbol:        "DEF",
		Subunit:       "def",
		Description:   "DEF Desc",
		InitialAmount: sdk.NewInt(1000),
		Features:      []types.Feature{types.Feature_freezing},
	}

	settings2 := types.IssueSettings{
		Issuer:        issuer2,
		Symbol:        "DEF",
		Subunit:       "def",
		Description:   "DEF Desc",
		InitialAmount: sdk.NewInt(2000),
		Features:      []types.Feature{types.Feature_whitelisting},
	}

	bondDenom := testApp.StakingKeeper.BondDenom(ctx)
	// fund with the native coin
	err := testApp.FundAccount(ctx, issuer1, sdk.NewCoins(sdk.NewCoin(bondDenom, sdk.NewInt(1000))))
	requireT.NoError(err)

	denom1, err := ftKeeper.Issue(ctx, settings1)
	requireT.NoError(err)

	denom2, err := ftKeeper.Issue(ctx, settings2)
	requireT.NoError(err)

	recipient1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// freeze denom1 partially on the recipient1
	err = ftKeeper.Freeze(ctx, issuer1, recipient1, sdk.NewCoin(denom1, sdk.NewInt(10)))
	requireT.NoError(err)

	// whitelist denom2 partially on the recipient2
	err = ftKeeper.SetWhitelistedBalance(ctx, issuer2, recipient2, sdk.NewCoin(denom2, sdk.NewInt(10)))
	requireT.NoError(err)

	// multi-send valid amount
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{
			{Address: issuer1.String(), Coins: sdk.NewCoins(
				sdk.NewCoin(denom1, sdk.NewInt(15)),
				sdk.NewCoin(bondDenom, sdk.NewInt(20)),
			)},
			{Address: issuer2.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom2, sdk.NewInt(10)))},
		},
		[]banktypes.Output{
			// the recipient1 has frozen balance so that amount can be received
			{Address: recipient1.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom1, sdk.NewInt(15)))},
			// the recipient2 has whitelisted balance so that is the max amount recipient2 can receive
			{Address: recipient2.String(), Coins: sdk.NewCoins(
				sdk.NewCoin(denom2, sdk.NewInt(10)),
				sdk.NewCoin(bondDenom, sdk.NewInt(20)),
			)},
		})
	requireT.NoError(err)

	balance := bankKeeper.GetBalance(ctx, recipient1, denom1)
	requireT.Equal(sdk.NewCoin(denom1, sdk.NewInt(15)).String(), balance.String())
	balance = bankKeeper.GetBalance(ctx, recipient2, denom2)
	requireT.Equal(sdk.NewCoin(denom2, sdk.NewInt(10)).String(), balance.String())
	balance = bankKeeper.GetBalance(ctx, recipient2, bondDenom)
	requireT.Equal(sdk.NewCoin(bondDenom, sdk.NewInt(20)).String(), balance.String())

	// multi-send invalid frozen amount
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{
			// we can't return 15 coins since 10 are frozen
			{Address: recipient1.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom1, sdk.NewInt(15)))},
			{Address: recipient2.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom2, sdk.NewInt(10)))},
		},
		[]banktypes.Output{
			{Address: issuer1.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom1, sdk.NewInt(15)))},
			{Address: issuer2.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom2, sdk.NewInt(10)))},
		})
	requireT.ErrorIs(sdkerrors.ErrInsufficientFunds, err)

	// multi-send invalid whitelisted amount
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{
			{Address: issuer1.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom1, sdk.NewInt(15)))},
			{Address: issuer2.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom2, sdk.NewInt(15)))},
		},
		[]banktypes.Output{
			{Address: recipient1.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom1, sdk.NewInt(15)))},
			// the recipient2 has whitelisted 10 so can't receive 15
			{Address: recipient2.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom2, sdk.NewInt(15)))},
		})
	requireT.ErrorIs(types.ErrWhitelistedLimitExceeded, err)
}

func TestKeeper_GetIssuerTokens(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})
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
			InitialAmount: sdk.NewInt(10),
		}
		denom, err := ftKeeper.Issue(ctx, settings)
		requireT.NoError(err)
		requireT.Equal(types.BuildDenom(settings.Subunit, settings.Issuer), denom)
	}

	tokens, _, err := ftKeeper.GetIssuerTokens(ctx, addr, &query.PageRequest{
		Limit: 3,
	})
	requireT.NoError(err)
	requireT.Equal(3, len(tokens))

	tokens, _, err = ftKeeper.GetIssuerTokens(ctx, addr, &query.PageRequest{
		Limit: uint64(numberOfTokens + 1),
	})
	requireT.NoError(err)
	requireT.Equal(numberOfTokens, len(tokens))
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
		requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(expectedBalance)).String(), getBalance.String())
	}

	totalSupply := ba.bk.GetSupply(ba.ctx, denom)
	requireT.Equal(totalSupply.String(), sdk.NewCoin(denom, sdk.NewInt(total)).String())
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
