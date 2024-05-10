package keeper_test

import (
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	testcontracts "github.com/CoreumFoundation/coreum/v4/x/asset/ft/keeper/test-contracts"
	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

const (
	MagicAmountDisallowed = 7
	MagicAmountBurning    = 101
	MagicAmountMinting    = 105
)

func TestKeeper_Extension_Issue(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	codeID, _, err := testApp.WasmPermissionedKeeper.Create(
		ctx, issuer, testcontracts.AssetExtensionWasm, &wasmtypes.AllowEverybody,
	)
	requireT.NoError(err)

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Description:   "ABC Desc",
		Subunit:       "extensionabc",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(777),
		Features:      []types.Feature{types.Feature_extension},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
		},
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	requireT.Equal(types.BuildDenom(settings.Subunit, settings.Issuer), denom)

	gotToken, err := ftKeeper.GetToken(ctx, denom)
	requireT.NoError(err)
	requireT.EqualValues(gotToken.Denom, denom)
	requireT.EqualValues(gotToken.Issuer, settings.Issuer.String())
	requireT.EqualValues(gotToken.Symbol, settings.Symbol)
	requireT.EqualValues(gotToken.Description, settings.Description)
	requireT.EqualValues(gotToken.Subunit, strings.ToLower(settings.Subunit))
	requireT.EqualValues(gotToken.Precision, settings.Precision)
	requireT.EqualValues(gotToken.Features, []types.Feature{types.Feature_extension})
	requireT.EqualValues(gotToken.BurnRate, sdk.NewDec(0))
	requireT.EqualValues(gotToken.SendCommissionRate, sdk.NewDec(0))
	requireT.EqualValues(gotToken.Version, types.CurrentTokenVersion)
	requireT.EqualValues(gotToken.URI, settings.URI)
	requireT.EqualValues(gotToken.URIHash, settings.URIHash)
	requireT.EqualValues(66, len(gotToken.ExtensionCWAddress))

	// check the account state
	issuedAssetBalance := bankKeeper.GetBalance(ctx, issuer, denom)
	requireT.Equal(sdk.NewCoin(denom, settings.InitialAmount).String(), issuedAssetBalance.String())

	// send 1 coin will succeed
	receiver := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, settings.Issuer, receiver, sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(2))))
	requireT.NoError(err)
	balance := bankKeeper.GetBalance(ctx, receiver, denom)
	requireT.EqualValues("2", balance.Amount.String())

	// send 7 coin will fail.
	// the POC contract is written as such that sending 7 will fail.
	// TODO replace with more meningful checks.
	err = bankKeeper.SendCoins(ctx, settings.Issuer, receiver, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(MagicAmountDisallowed))),
	)
	requireT.ErrorIs(err, types.ErrExtensionCallFailed)
	balance = bankKeeper.GetBalance(ctx, receiver, denom)
	requireT.EqualValues("2", balance.Amount.String())
}

func TestKeeper_Extension_Whitelist(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	codeID, _, err := testApp.WasmPermissionedKeeper.Create(
		ctx, issuer, testcontracts.AssetExtensionWasm, &wasmtypes.AllowEverybody,
	)
	requireT.NoError(err)

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(666),
		Features: []types.Feature{
			types.Feature_whitelisting,
			types.Feature_extension,
		},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
		},
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
		Features: []types.Feature{
			types.Feature_extension,
		},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
		},
	}

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	unwhitelistableDenom, err := ftKeeper.Issue(ctx, unwhitelistableSettings)
	requireT.NoError(err)
	_, err = ftKeeper.GetToken(ctx, unwhitelistableDenom)
	requireT.NoError(err)

	// set whitelisted balance to 0
	requireT.NoError(ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, sdk.NewCoin(denom, sdkmath.NewInt(0))))
	whitelistedBalance := ftKeeper.GetWhitelistedBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(0)).String(), whitelistedBalance.String())

	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100)))
	// send
	err = bankKeeper.SendCoins(ctx, issuer, recipient, coinsToSend)
	requireT.ErrorContains(err, "Whitelisted limit exceeded.")
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: issuer.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}})
	requireT.ErrorContains(err, "Whitelisted limit exceeded.")

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
		[]banktypes.Input{{Address: issuer.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}})
	requireT.NoError(err)

	// try to send more
	coinsToSend = sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(1)))
	// send
	err = bankKeeper.SendCoins(ctx, issuer, recipient, coinsToSend)
	requireT.ErrorContains(err, "Whitelisted limit exceeded.")
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: issuer.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}})
	requireT.ErrorContains(err, "Whitelisted limit exceeded.")

	// reduce whitelisting limit below the current balance
	err = ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, sdk.NewCoin(denom, sdkmath.NewInt(80)))
	requireT.NoError(err)

	// TODO: add a special condition into the smart contract to ignore default whitelisting, so we will know that
	// we are able to overwrite the default behavior via smart contract.
}

func TestKeeper_Extension_FreezeUnfreeze(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	codeID, _, err := testApp.WasmPermissionedKeeper.Create(
		ctx, issuer, testcontracts.AssetExtensionWasm, &wasmtypes.AllowEverybody,
	)
	requireT.NoError(err)

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(666),
		Features: []types.Feature{
			types.Feature_freezing,
			types.Feature_extension,
		},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
		},
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	token, err := ftKeeper.GetToken(ctx, denom)
	requireT.NoError(err)

	extensionCWAddress, err := sdk.AccAddressFromBech32(token.ExtensionCWAddress)
	requireT.NoError(err)

	unfreezableSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     1,
		Description:   "ABC Desc",
		InitialAmount: sdkmath.NewInt(666),
		Features:      []types.Feature{types.Feature_extension},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
		},
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
	requireT.ErrorIs(err, types.ErrTokenNotFound)

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
	requireT.ErrorContains(err, "Requested transfer token is frozen.")
	// return attached fund of failed transaction
	err = bankKeeper.SendCoins(ctx, extensionCWAddress, recipient, coinsToSend)
	requireT.NoError(err)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: recipient.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: issuer.String(), Coins: coinsToSend}})
	requireT.ErrorContains(err, "Requested transfer token is frozen.")
	// return attached fund of failed transaction
	err = bankKeeper.SendCoins(ctx, extensionCWAddress, recipient, coinsToSend)
	requireT.NoError(err)

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
		[]banktypes.Input{{Address: recipient.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: recipient2.String(), Coins: coinsToSend}})
	requireT.NoError(err) // Requested transfer token is frozen.
	balance = bankKeeper.GetBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(80)), balance)
	balance = bankKeeper.GetBalance(ctx, recipient2, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(20)), balance)

	// try to unfreeze from non issuer address
	err = ftKeeper.Unfreeze(ctx, randomAddr, recipient, sdk.NewCoin(denom, sdkmath.NewInt(80)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

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
		[]banktypes.Input{{Address: recipient.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: recipient2.String(), Coins: coinsToSend}})
	requireT.NoError(err)
	balance = bankKeeper.GetBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(0)), balance)
	balance = bankKeeper.GetBalance(ctx, recipient2, denom)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(100)), balance)
}

func TestKeeper_Extension_Burn(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	recipient := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	codeID, _, err := testApp.WasmPermissionedKeeper.Create(
		ctx, issuer, testcontracts.AssetExtensionWasm, &wasmtypes.AllowEverybody,
	)
	requireT.NoError(err)

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
			types.Feature_extension,
		},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
		},
	}

	unburnableDenom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)
	requireT.Equal(types.BuildDenom(settings.Symbol, settings.Issuer), unburnableDenom)

	token, err := ftKeeper.GetToken(ctx, unburnableDenom)
	requireT.NoError(err)

	unburnableDenomExtensionCWAddress, err := sdk.AccAddressFromBech32(token.ExtensionCWAddress)
	requireT.NoError(err)

	// send to new recipient address
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(sdk.NewCoin(unburnableDenom, sdkmath.NewInt(102))))
	requireT.NoError(err)

	coinsToBurn := sdk.NewCoins(sdk.NewCoin(unburnableDenom, sdkmath.NewInt(MagicAmountBurning)))

	// try to burn unburnable token from the recipient account
	err = bankKeeper.SendCoins(ctx, recipient, issuer, coinsToBurn)
	requireT.ErrorContains(err, "Feature disabled.")
	// return attached fund of failed transaction
	err = bankKeeper.SendCoins(ctx, unburnableDenomExtensionCWAddress, recipient, coinsToBurn)
	requireT.NoError(err)

	issuerBalanceBefore := bankKeeper.GetBalance(ctx, issuer, unburnableDenom)
	cwExtensionBalanceBefore := bankKeeper.GetBalance(ctx, unburnableDenomExtensionCWAddress, unburnableDenom)

	// try to burn unburnable token from the issuer account
	err = bankKeeper.SendCoins(ctx, issuer, issuer, coinsToBurn)
	requireT.NoError(err)

	issuerBalanceAfter := bankKeeper.GetBalance(ctx, issuer, unburnableDenom)
	cwExtensionBalanceAfter := bankKeeper.GetBalance(ctx, unburnableDenomExtensionCWAddress, unburnableDenom)

	// the amount should be burnt
	requireT.Equal(
		issuerBalanceBefore.String(),
		issuerBalanceAfter.Add(sdk.NewCoin(unburnableDenom, sdkmath.NewInt(MagicAmountBurning))).String(),
	)
	requireT.Equal(cwExtensionBalanceBefore.String(), cwExtensionBalanceAfter.String())

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
			types.Feature_extension,
		},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
		},
	}

	burnableDenom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	token, err = ftKeeper.GetToken(ctx, burnableDenom)
	requireT.NoError(err)

	extensionCWAddress, err := sdk.AccAddressFromBech32(token.ExtensionCWAddress)
	requireT.NoError(err)

	// send to new recipient address
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(sdk.NewCoin(burnableDenom, sdkmath.NewInt(202))))
	requireT.NoError(err)

	recipientBalanceBefore := bankKeeper.GetBalance(ctx, recipient, burnableDenom)
	cwExtensionBalanceBefore = bankKeeper.GetBalance(ctx, extensionCWAddress, burnableDenom)

	// try to burn as non-issuer
	err = bankKeeper.SendCoins(ctx, recipient, issuer, sdk.NewCoins(
		sdk.NewCoin(burnableDenom, sdkmath.NewInt(MagicAmountBurning))),
	)
	requireT.NoError(err)

	recipientBalanceAfter := bankKeeper.GetBalance(ctx, recipient, burnableDenom)
	cwExtensionBalanceAfter = bankKeeper.GetBalance(ctx, extensionCWAddress, burnableDenom)

	// the amount should be burnt
	requireT.Equal(
		recipientBalanceBefore.String(),
		recipientBalanceAfter.Add(sdk.NewCoin(burnableDenom, sdkmath.NewInt(MagicAmountBurning))).String(),
	)
	requireT.Equal(cwExtensionBalanceBefore.String(), cwExtensionBalanceAfter.String())

	issuerBalanceBefore = bankKeeper.GetBalance(ctx, issuer, burnableDenom)
	cwExtensionBalanceBefore = bankKeeper.GetBalance(ctx, extensionCWAddress, burnableDenom)

	// burn tokens and check balance and total supply
	err = bankKeeper.SendCoins(ctx, issuer, issuer, sdk.NewCoins(
		sdk.NewCoin(burnableDenom, sdkmath.NewInt(MagicAmountBurning))),
	)
	requireT.NoError(err)

	issuerBalanceAfter = bankKeeper.GetBalance(ctx, issuer, burnableDenom)
	cwExtensionBalanceAfter = bankKeeper.GetBalance(ctx, extensionCWAddress, burnableDenom)

	// the amount should be burnt
	requireT.Equal(
		issuerBalanceBefore.String(),
		issuerBalanceAfter.Add(sdk.NewCoin(burnableDenom, sdkmath.NewInt(MagicAmountBurning))).String(),
	)
	requireT.Equal(cwExtensionBalanceBefore.String(), cwExtensionBalanceAfter.String())

	balance := bankKeeper.GetBalance(ctx, issuer, burnableDenom)
	requireT.EqualValues(sdk.NewCoin(burnableDenom, sdkmath.NewInt(474)), balance)

	totalSupply, err := bankKeeper.TotalSupply(sdk.WrapSDKContext(ctx), &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(575), totalSupply.Supply.AmountOf(burnableDenom))

	// try to freeze the issuer (issuer can't be frozen)
	err = ftKeeper.Freeze(ctx, issuer, issuer, sdk.NewCoin(burnableDenom, sdkmath.NewInt(600)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to burn non-issuer frozen coins
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(burnableDenom, sdkmath.NewInt(MagicAmountBurning)))
	requireT.NoError(err)
	err = bankKeeper.SendCoins(ctx, recipient, issuer, sdk.NewCoins(
		sdk.NewCoin(burnableDenom, sdkmath.NewInt(MagicAmountBurning))),
	)
	requireT.ErrorContains(err, "Requested transfer token is frozen.")
}

func TestKeeper_Extension_Mint(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	codeID, _, err := testApp.WasmPermissionedKeeper.Create(
		ctx, addr, testcontracts.AssetExtensionWasm, &wasmtypes.AllowEverybody,
	)
	requireT.NoError(err)

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
			types.Feature_extension,
		},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
		},
	}

	unmintableDenom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)
	requireT.Equal(types.BuildDenom(settings.Symbol, settings.Issuer), unmintableDenom)

	// try to mint unmintable token
	err = bankKeeper.SendCoins(ctx, addr, addr, sdk.NewCoins(
		sdk.NewCoin(unmintableDenom, sdkmath.NewInt(MagicAmountMinting))),
	)
	requireT.ErrorContains(err, "Feature disabled.")

	// Issue a mintable fungible token
	settings = types.IssueSettings{
		Issuer:        addr,
		Symbol:        "mintable",
		Subunit:       "mintable",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(777),
		Features: []types.Feature{
			types.Feature_minting,
			types.Feature_extension,
		},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
		},
	}

	mintableDenom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	token, err := ftKeeper.GetToken(ctx, mintableDenom)
	requireT.NoError(err)

	extensionCWAddress, err := sdk.AccAddressFromBech32(token.ExtensionCWAddress)
	requireT.NoError(err)

	coinsToMint := sdk.NewCoins(sdk.NewCoin(mintableDenom, sdkmath.NewInt(MagicAmountMinting)))

	randomAddr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	err = bankKeeper.SendCoins(ctx, addr, randomAddr, sdk.NewCoins(sdk.NewCoin(mintableDenom, sdkmath.NewInt(125))))
	requireT.NoError(err)

	// try to mint as non-issuer
	err = bankKeeper.SendCoins(ctx, randomAddr, randomAddr, coinsToMint)
	requireT.ErrorContains(err, "Unauthorized.")

	// return attached fund of failed transaction
	err = bankKeeper.SendCoins(ctx, extensionCWAddress, randomAddr, coinsToMint)
	requireT.NoError(err)

	// mint tokens and check balance and total supply
	err = bankKeeper.SendCoins(ctx, addr, addr, coinsToMint)
	requireT.NoError(err)

	balance := bankKeeper.GetBalance(ctx, addr, mintableDenom)
	requireT.EqualValues(sdk.NewCoin(mintableDenom, sdkmath.NewInt(757)), balance)

	totalSupply, err := bankKeeper.TotalSupply(sdk.WrapSDKContext(ctx), &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(882), totalSupply.Supply.AmountOf(mintableDenom))

	// mint to another account
	err = bankKeeper.SendCoins(ctx, addr, randomAddr, coinsToMint)
	requireT.NoError(err)

	balance = bankKeeper.GetBalance(ctx, randomAddr, mintableDenom)
	requireT.EqualValues(sdk.NewCoin(mintableDenom, sdkmath.NewInt(230)), balance)

	totalSupply, err = bankKeeper.TotalSupply(sdk.WrapSDKContext(ctx), &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(987), totalSupply.Supply.AmountOf(mintableDenom))
}
