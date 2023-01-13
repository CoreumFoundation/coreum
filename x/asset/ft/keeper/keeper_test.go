package keeper_test

import (
	"errors"
	"fmt"
	"math/big"
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
	"github.com/CoreumFoundation/coreum/x/asset/ft/keeper"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
	wbankkeeper "github.com/CoreumFoundation/coreum/x/wbank/keeper"
)

func TestKeeper_ValidateSymbol(t *testing.T) {
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
			Features:      []types.TokenFeature{types.TokenFeature_freeze}, //nolint:nosnakecase
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

func TestKeeper_ValidateSubunit(t *testing.T) {
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
		settings := types.IssueSettings{
			Issuer:        addr,
			Symbol:        "symbol",
			Subunit:       subunit,
			Description:   "ABC Desc",
			InitialAmount: sdk.NewInt(777),
			Features:      []types.TokenFeature{types.TokenFeature_freeze}, //nolint:nosnakecase
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
		Features:      []types.TokenFeature{types.TokenFeature_freeze}, //nolint:nosnakecase
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
	requireT.Equal(types.FT{
		Denom:              denom,
		Issuer:             settings.Issuer.String(),
		Symbol:             settings.Symbol,
		Description:        settings.Description,
		Subunit:            strings.ToLower(settings.Subunit),
		Precision:          settings.Precision,
		Features:           []types.TokenFeature{types.TokenFeature_freeze}, //nolint:nosnakecase
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
		Features:      []types.TokenFeature{types.TokenFeature_freeze}, //nolint:nosnakecase
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
		Features:      []types.TokenFeature{types.TokenFeature_freeze}, //nolint:nosnakecase
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
		Features: []types.TokenFeature{
			types.TokenFeature_freeze, //nolint:nosnakecase
			types.TokenFeature_burn,   //nolint:nosnakecase
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
		Features: []types.TokenFeature{
			types.TokenFeature_mint, //nolint:nosnakecase
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
		Features: []types.TokenFeature{
			types.TokenFeature_freeze, //nolint:nosnakecase
			types.TokenFeature_mint,   //nolint:nosnakecase
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
		Features: []types.TokenFeature{
			types.TokenFeature_burn,   //nolint:nosnakecase
			types.TokenFeature_freeze, //nolint:nosnakecase
		},
	}

	burnableDenom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	// send to new recipient address
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(sdk.NewCoin(burnableDenom, sdk.NewInt(100))))
	requireT.NoError(err)

	// try to burn as non-issuer
	err = ftKeeper.Burn(ctx, recipient, sdk.NewCoin(burnableDenom, sdk.NewInt(100)))
	requireT.NoError(err)

	// burn tokens and check balance and total supply
	err = ftKeeper.Burn(ctx, issuer, sdk.NewCoin(burnableDenom, sdk.NewInt(100)))
	requireT.NoError(err)

	balance := bankKeeper.GetBalance(ctx, issuer, burnableDenom)
	requireT.EqualValues(sdk.NewCoin(burnableDenom, sdk.NewInt(577)), balance)

	totalSupply, err := bankKeeper.TotalSupply(sdk.WrapSDKContext(ctx), &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewInt(577), totalSupply.Supply.AmountOf(burnableDenom))

	// try to burn frozen amount
	err = ftKeeper.Freeze(ctx, issuer, issuer, sdk.NewCoin(burnableDenom, sdk.NewInt(600)))
	requireT.NoError(err)

	err = ftKeeper.Burn(ctx, issuer, sdk.NewCoin(burnableDenom, sdk.NewInt(100)))
	requireT.ErrorIs(sdkerrors.ErrInsufficientFunds, err)
}

//nolint:funlen // there are too many tests cases
func TestKeeperCalculateRateShare(t *testing.T) {
	genAccount := func() string {
		return sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
	}
	var accounts []string
	for i := 0; i < 11; i++ {
		accounts = append(accounts, genAccount())
	}
	issuer := genAccount()
	pow10 := func(ex int64) sdk.Int {
		return sdk.NewIntFromBigInt(big.NewInt(0).Exp(big.NewInt(10), big.NewInt(ex), nil))
	}
	testCases := []struct {
		rate      string
		senders   map[string]sdk.Int
		receivers map[string]sdk.Int
		shares    map[string]sdk.Int
	}{
		{
			rate:    "0.5",
			senders: map[string]sdk.Int{},
			shares:  map[string]sdk.Int{},
		},
		{
			rate: "0.5",
			senders: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(5),
				accounts[1]: sdk.NewInt(5),
			},
			receivers: map[string]sdk.Int{
				issuer: sdk.NewInt(10),
			},
			shares: map[string]sdk.Int{},
		},
		{
			rate: "0.5",
			senders: map[string]sdk.Int{
				issuer: sdk.NewInt(10),
			},
			receivers: map[string]sdk.Int{
				accounts[5]: sdk.NewInt(5),
				accounts[6]: sdk.NewInt(5),
			},
			shares: map[string]sdk.Int{},
		},
		{
			rate: "0.1",
			senders: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(400),
				accounts[1]: sdk.NewInt(600),
			},
			receivers: map[string]sdk.Int{
				accounts[10]: sdk.NewInt(1000),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(40),
				accounts[1]: sdk.NewInt(60),
			},
		},
		{
			rate: "0.1",
			senders: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(399),
				accounts[1]: sdk.NewInt(602),
			},
			receivers: map[string]sdk.Int{
				accounts[10]: sdk.NewInt(1001),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(40),
				accounts[1]: sdk.NewInt(61),
			},
		},
		{
			rate: "0.1",
			senders: map[string]sdk.Int{
				issuer:      sdk.NewInt(90),
				accounts[0]: sdk.NewInt(29),
				accounts[1]: sdk.NewInt(32),
			},
			receivers: map[string]sdk.Int{
				genAccount(): sdk.NewInt(90 + 29 + 32),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(3),
				accounts[1]: sdk.NewInt(4),
			},
		},
		{
			rate: "0.01",
			senders: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(30000),
				accounts[1]: sdk.NewInt(20000),
			},
			receivers: map[string]sdk.Int{
				issuer:       sdk.NewInt(30000),
				genAccount(): sdk.NewInt(20000),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(120),
				accounts[1]: sdk.NewInt(80),
			},
		},
		{
			rate: "0.01001",
			senders: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(30000),
				accounts[1]: sdk.NewInt(20000),
			},
			receivers: map[string]sdk.Int{
				issuer:       sdk.NewInt(30000),
				genAccount(): sdk.NewInt(20000),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(121),
				accounts[1]: sdk.NewInt(81),
			},
		},
		{
			rate: "0.1234",
			senders: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(80),
				accounts[1]: sdk.NewInt(17),
			},
			receivers: map[string]sdk.Int{
				genAccount(): sdk.NewInt(97),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(10),
				accounts[1]: sdk.NewInt(3),
			},
		},
		{
			rate: "0.1",
			senders: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(1),
				accounts[1]: sdk.NewInt(2),
				accounts[2]: sdk.NewInt(9),
			},
			receivers: map[string]sdk.Int{
				genAccount(): sdk.NewInt(12),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(1),
				accounts[1]: sdk.NewInt(1),
				accounts[2]: sdk.NewInt(1),
			},
		},
		{
			rate: "0.01",
			senders: map[string]sdk.Int{
				issuer:      sdk.NewInt(2100),
				accounts[0]: sdk.NewInt(1100),
				accounts[1]: sdk.NewInt(1700),
				accounts[2]: sdk.NewInt(1900),
			},
			receivers: map[string]sdk.Int{
				issuer:       sdk.NewInt(2100),
				genAccount(): sdk.NewInt(300),
				genAccount(): sdk.NewInt(1100),
				genAccount(): sdk.NewInt(3300),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(11),
				accounts[1]: sdk.NewInt(17),
				accounts[2]: sdk.NewInt(19),
			},
		},
		{
			rate: "0.01",
			senders: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(100).Mul(pow10(24)),
				accounts[1]: sdk.NewInt(300).Mul(pow10(25)),
				accounts[2]: sdk.NewInt(500).Mul(pow10(26)),
			},
			receivers: map[string]sdk.Int{
				genAccount(): sdk.NewInt(100).Mul(pow10(24)),
				genAccount(): sdk.NewInt(300).Mul(pow10(25)),
				genAccount(): sdk.NewInt(500).Mul(pow10(26)),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(100).Mul(pow10(22)),
				accounts[1]: sdk.NewInt(300).Mul(pow10(23)),
				accounts[2]: sdk.NewInt(500).Mul(pow10(24)),
			},
		},
		{
			rate: "0.99",
			senders: map[string]sdk.Int{
				issuer:      sdk.NewInt(2100),
				accounts[0]: sdk.NewInt(1100),
				accounts[1]: sdk.NewInt(1700),
				accounts[2]: sdk.NewInt(2728),
			},
			receivers: map[string]sdk.Int{
				genAccount(): sdk.NewInt(2100),
				genAccount(): sdk.NewInt(1000),
				genAccount(): sdk.NewInt(1800),
				genAccount(): sdk.NewInt(2728),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(1089),
				accounts[1]: sdk.NewInt(1683),
				accounts[2]: sdk.NewInt(2701),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		name := fmt.Sprintf("%+v", tc)
		t.Run(name, func(t *testing.T) {
			assertT := assert.New(t)
			shares := keeper.CalculateRateShares(sdk.MustNewDecFromStr(tc.rate), issuer, tc.senders, tc.receivers)
			for account, share := range shares {
				assertT.EqualValues(tc.shares[account].String(), share.String())
			}
		})
	}
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
			Features:      []types.TokenFeature{},
			BurnRate:      sdk.MustNewDecFromStr(fmt.Sprintf("0.%d", i+1)),
		}

		denom, err := assetKeeper.Issue(ctx, settings)
		requireT.NoError(err)
		denoms = append(denoms, denom)

		// create 2 recipient for every issuer to allow for complex test cases
		recipients = append(recipients, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()))
		recipients = append(recipients, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()))
	}

	var testCases = []struct {
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
		Features:      []types.TokenFeature{},
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
		Features:      []types.TokenFeature{},
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
		Features:           []types.TokenFeature{},
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
		Features:           []types.TokenFeature{},
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
		Features:           []types.TokenFeature{},
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
