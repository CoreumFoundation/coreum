package keeper_test

import (
	"fmt"
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
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/testutil/simapp"
	testcontracts "github.com/CoreumFoundation/coreum/v6/x/asset/ft/keeper/test-contracts"
	"github.com/CoreumFoundation/coreum/v6/x/asset/ft/types"
	wibctransfertypes "github.com/CoreumFoundation/coreum/v6/x/wibctransfer/types"
)

var (
	AmountDisallowedTrigger               = sdkmath.NewInt(testcontracts.AmountDisallowedTrigger)
	AmountBurningTrigger                  = sdkmath.NewInt(testcontracts.AmountBurningTrigger)
	AmountMintingTrigger                  = sdkmath.NewInt(testcontracts.AmountMintingTrigger)
	AmountIgnoreBurnRateTrigger           = sdkmath.NewInt(testcontracts.AmountIgnoreBurnRateTrigger)
	AmountIgnoreSendCommissionRateTrigger = sdkmath.NewInt(testcontracts.AmountIgnoreSendCommissionRateTrigger)
	AmountDEXExpectToSpendTrigger         = sdkmath.NewInt(testcontracts.AmountDEXExpectToSpendTrigger)
	AmountDEXExpectToReceiveTrigger       = sdkmath.NewInt(testcontracts.AmountDEXExpectToReceiveTrigger)
)

func TestKeeper_Extension_Issue(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
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

	subunit := "extensionabc"
	attachedAmount := sdkmath.NewInt(500)
	issuerAmount := sdkmath.NewInt(277)
	denom := types.BuildDenom(subunit, issuer)
	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Description:   "ABC Desc",
		Subunit:       subunit,
		Precision:     8,
		InitialAmount: attachedAmount.Add(issuerAmount),
		Features:      []types.Feature{types.Feature_extension},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
			Funds:  sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(500))),
		},
	}

	_, err = ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	requireT.Equal(types.BuildDenom(settings.Subunit, settings.Issuer), denom)

	gotToken, err := ftKeeper.GetToken(ctx, denom)
	requireT.NoError(err)
	requireT.EqualValues(denom, gotToken.Denom)
	requireT.EqualValues(settings.Issuer.String(), gotToken.Issuer)
	requireT.EqualValues(settings.Symbol, gotToken.Symbol)
	requireT.EqualValues(settings.Description, gotToken.Description)
	requireT.EqualValues(strings.ToLower(settings.Subunit), gotToken.Subunit)
	requireT.EqualValues(settings.Precision, gotToken.Precision)
	requireT.EqualValues([]types.Feature{types.Feature_extension}, gotToken.Features)
	requireT.EqualValues(sdkmath.LegacyNewDec(0), gotToken.BurnRate)
	requireT.EqualValues(sdkmath.LegacyNewDec(0), gotToken.SendCommissionRate)
	requireT.EqualValues(types.CurrentTokenVersion, gotToken.Version)
	requireT.EqualValues(settings.URI, gotToken.URI)
	requireT.EqualValues(settings.URIHash, gotToken.URIHash)
	requireT.Len(gotToken.ExtensionCWAddress, 66)

	contractAddress, err := sdk.AccAddressFromBech32(gotToken.ExtensionCWAddress)
	requireT.NoError(err)
	// check the account state
	contractBalance := bankKeeper.GetBalance(ctx, contractAddress, denom)
	requireT.Equal(sdk.NewCoin(denom, attachedAmount).String(), contractBalance.String())

	issuedAssetBalance := bankKeeper.GetBalance(ctx, issuer, denom)
	requireT.Equal(sdk.NewCoin(denom, issuerAmount).String(), issuedAssetBalance.String())

	// send 2 coin will succeed
	receiver := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, settings.Issuer, receiver, sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(2))))
	requireT.NoError(err)
	balance := bankKeeper.GetBalance(ctx, receiver, denom)
	requireT.EqualValues("2", balance.Amount.String())

	// send 7 coin will fail.
	// the test contract is written as such that sending 7 will fail.
	err = bankKeeper.SendCoins(ctx, settings.Issuer, receiver, sdk.NewCoins(
		sdk.NewCoin(denom, AmountDisallowedTrigger)),
	)
	requireT.ErrorIs(err, types.ErrExtensionCallFailed)
	balance = bankKeeper.GetBalance(ctx, receiver, denom)
	requireT.EqualValues("2", balance.Amount.String())
}

func TestKeeper_Extension_IBC(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	codeID, _, err := testApp.WasmPermissionedKeeper.Create(
		ctx, issuer, testcontracts.AssetExtensionWasm, &wasmtypes.AllowEverybody,
	)
	requireT.NoError(err)

	settingsWithoutIBC := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(666),
		Features: []types.Feature{
			types.Feature_whitelisting,
			types.Feature_extension,
		},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
		},
	}

	denomWithoutIBC, err := ftKeeper.Issue(ctx, settingsWithoutIBC)
	requireT.NoError(err)

	settingsWithIBC := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(666),
		Features: []types.Feature{
			types.Feature_whitelisting,
			types.Feature_extension,
			types.Feature_ibc,
		},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
		},
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

func TestKeeper_Extension_Whitelist(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

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

	coinToSend := sdk.NewCoin(denom, sdkmath.NewInt(100))
	coinsToSend := sdk.NewCoins(coinToSend)
	// send
	requireT.ErrorIs(bankKeeper.SendCoins(ctx, issuer, recipient, coinsToSend), types.ErrWhitelistedLimitExceeded)
	// multi-send
	requireT.ErrorIs(bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{Address: issuer.String(), Coins: coinsToSend},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}}), types.ErrWhitelistedLimitExceeded)

	// set whitelisted balance (for 2 sends)
	requireT.NoError(ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, coinToSend.Add(coinToSend)))

	// check that expected to fail extension call fails
	err = bankKeeper.SendCoins(ctx, settings.Issuer, recipient, sdk.NewCoins(
		sdk.NewCoin(denom, AmountDisallowedTrigger)),
	)
	requireT.ErrorIs(err, types.ErrExtensionCallFailed)
	requireT.ErrorContains(err, "7 is not allowed")

	requireT.NoError(bankKeeper.SendCoins(ctx, issuer, recipient, coinsToSend))
	// multi-send
	requireT.NoError(bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{Address: issuer.String(), Coins: coinsToSend},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}}))

	bankBalance := bankKeeper.GetBalance(ctx, recipient, denom)
	requireT.Equal(coinToSend.Add(coinToSend).String(), bankBalance.String())

	// try to send more
	coinsToSend = sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(1)))
	// send
	requireT.ErrorIs(bankKeeper.SendCoins(ctx, issuer, recipient, coinsToSend), types.ErrWhitelistedLimitExceeded)
	// multi-send
	requireT.ErrorIs(bankKeeper.InputOutputCoins(ctx,
		banktypes.Input{Address: issuer.String(), Coins: coinsToSend},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}}), types.ErrWhitelistedLimitExceeded)
}

func TestKeeper_Extension_FreezeUnfreeze(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	recipient1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

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

	// check that expected to fail extension call fails
	err = bankKeeper.SendCoins(ctx, settings.Issuer, recipient1, sdk.NewCoins(
		sdk.NewCoin(denom, AmountDisallowedTrigger)),
	)
	requireT.ErrorIs(err, types.ErrExtensionCallFailed)
	requireT.ErrorContains(err, "7 is not allowed")

	// send coins to recipient1
	err = bankKeeper.SendCoins(ctx, issuer, recipient1, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(100)),
	))
	requireT.NoError(err)

	// freeze
	requireT.NoError(ftKeeper.Freeze(ctx, issuer, recipient1, sdk.NewCoin(denom, sdkmath.NewInt(120))))

	// try to send more than available
	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(80)))
	// send
	requireT.ErrorIs(bankKeeper.SendCoins(ctx, recipient1, recipient2, coinsToSend), cosmoserrors.ErrInsufficientFunds)
	// multi-send
	requireT.ErrorIs(bankKeeper.InputOutputCoins(
		ctx,
		banktypes.Input{Address: recipient1.String(), Coins: coinsToSend},
		[]banktypes.Output{{Address: recipient2.String(), Coins: coinsToSend}}),
		cosmoserrors.ErrInsufficientFunds,
	)

	// unfreeze
	requireT.NoError(ftKeeper.Unfreeze(ctx, issuer, recipient1, sdk.NewCoin(denom, sdkmath.NewInt(120))))

	// send
	requireT.NoError(bankKeeper.SendCoins(ctx, recipient1, recipient2, coinsToSend))

	// freeze globally
	requireT.NoError(ftKeeper.SetGlobalFreeze(ctx, denom, true))

	// try to send more than available
	requireT.ErrorIs(
		bankKeeper.SendCoins(
			ctx, recipient1, recipient2, sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(1))),
		), types.ErrGloballyFrozen,
	)
}

func TestKeeper_Extension_Burn(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
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

	coinsToBurn := sdk.NewCoins(sdk.NewCoin(unburnableDenom, AmountBurningTrigger))

	// try to burn unburnable token from the recipient account and make sure that extension can do it
	err = bankKeeper.SendCoins(ctx, recipient, issuer, coinsToBurn)
	requireT.NoError(err)

	issuerBalanceBefore := bankKeeper.GetBalance(ctx, issuer, unburnableDenom)
	cwExtensionBalanceBefore := bankKeeper.GetBalance(ctx, unburnableDenomExtensionCWAddress, unburnableDenom)
	totalSupplyBefore, err := bankKeeper.TotalSupply(ctx, &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(676), totalSupplyBefore.Supply.AmountOf(unburnableDenom))

	// try to burn unburnable token from the issuer account
	err = bankKeeper.SendCoins(ctx, issuer, issuer, coinsToBurn)
	requireT.NoError(err)

	issuerBalanceAfter := bankKeeper.GetBalance(ctx, issuer, unburnableDenom)
	cwExtensionBalanceAfter := bankKeeper.GetBalance(ctx, unburnableDenomExtensionCWAddress, unburnableDenom)
	totalSupplyAfter, err := bankKeeper.TotalSupply(ctx, &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(575), totalSupplyAfter.Supply.AmountOf(unburnableDenom))

	// the amount should be burnt
	requireT.Equal(
		issuerBalanceBefore.String(),
		issuerBalanceAfter.Add(sdk.NewCoin(unburnableDenom, AmountBurningTrigger)).String(),
	)
	requireT.Equal(cwExtensionBalanceBefore.String(), cwExtensionBalanceAfter.String())
	requireT.Equal(
		totalSupplyBefore.Supply.String(),
		totalSupplyAfter.Supply.Add(sdk.NewCoin(unburnableDenom, AmountBurningTrigger)).String(),
	)

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
	totalSupplyBefore, err = bankKeeper.TotalSupply(ctx, &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(777), totalSupplyBefore.Supply.AmountOf(burnableDenom))

	// try to burn as non-issuer
	err = bankKeeper.SendCoins(ctx, recipient, issuer, sdk.NewCoins(
		sdk.NewCoin(burnableDenom, AmountBurningTrigger)),
	)
	requireT.NoError(err)

	recipientBalanceAfter := bankKeeper.GetBalance(ctx, recipient, burnableDenom)
	cwExtensionBalanceAfter = bankKeeper.GetBalance(ctx, extensionCWAddress, burnableDenom)
	totalSupplyAfter, err = bankKeeper.TotalSupply(ctx, &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(676), totalSupplyAfter.Supply.AmountOf(burnableDenom))

	// the amount should be burnt
	requireT.Equal(
		recipientBalanceBefore.String(),
		recipientBalanceAfter.Add(sdk.NewCoin(burnableDenom, AmountBurningTrigger)).String(),
	)
	requireT.Equal(cwExtensionBalanceBefore.String(), cwExtensionBalanceAfter.String())
	requireT.Equal(
		totalSupplyBefore.Supply.String(),
		totalSupplyAfter.Supply.Add(sdk.NewCoin(burnableDenom, AmountBurningTrigger)).String(),
	)

	issuerBalanceBefore = bankKeeper.GetBalance(ctx, issuer, burnableDenom)
	cwExtensionBalanceBefore = bankKeeper.GetBalance(ctx, extensionCWAddress, burnableDenom)
	totalSupplyBefore, err = bankKeeper.TotalSupply(ctx, &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(676), totalSupplyBefore.Supply.AmountOf(burnableDenom))

	// burn tokens and check balance and total supply
	err = bankKeeper.SendCoins(ctx, issuer, issuer, sdk.NewCoins(
		sdk.NewCoin(burnableDenom, AmountBurningTrigger)),
	)
	requireT.NoError(err)

	issuerBalanceAfter = bankKeeper.GetBalance(ctx, issuer, burnableDenom)
	cwExtensionBalanceAfter = bankKeeper.GetBalance(ctx, extensionCWAddress, burnableDenom)
	totalSupplyAfter, err = bankKeeper.TotalSupply(ctx, &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(575), totalSupplyAfter.Supply.AmountOf(burnableDenom))

	// the amount should be burnt
	requireT.Equal(
		issuerBalanceBefore.String(),
		issuerBalanceAfter.Add(sdk.NewCoin(burnableDenom, AmountBurningTrigger)).String(),
	)
	requireT.Equal(cwExtensionBalanceBefore.String(), cwExtensionBalanceAfter.String())
	requireT.Equal(
		totalSupplyBefore.Supply.String(),
		totalSupplyAfter.Supply.Add(sdk.NewCoin(burnableDenom, AmountBurningTrigger)).String(),
	)

	balance := bankKeeper.GetBalance(ctx, issuer, burnableDenom)
	requireT.EqualValues(sdk.NewCoin(burnableDenom, sdkmath.NewInt(474)), balance)

	totalSupply, err := bankKeeper.TotalSupply(ctx, &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(575), totalSupply.Supply.AmountOf(burnableDenom))

	// try to freeze the issuer (issuer can't be frozen)
	err = ftKeeper.Freeze(ctx, issuer, issuer, sdk.NewCoin(burnableDenom, sdkmath.NewInt(600)))
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// try to burn non-issuer frozen coins
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(burnableDenom, AmountBurningTrigger))
	requireT.NoError(err)
	err = bankKeeper.SendCoins(ctx, recipient, issuer, sdk.NewCoins(
		sdk.NewCoin(burnableDenom, AmountBurningTrigger)),
	)
	requireT.ErrorIs(err, cosmoserrors.ErrInsufficientFunds)
}

func TestKeeper_Extension_Mint(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
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
		sdk.NewCoin(unmintableDenom, AmountMintingTrigger)),
	)
	requireT.ErrorContains(err, "feature minting is disabled")

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

	coinsToMint := sdk.NewCoins(sdk.NewCoin(mintableDenom, AmountMintingTrigger))

	randomAddr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	err = bankKeeper.SendCoins(ctx, addr, randomAddr, sdk.NewCoins(sdk.NewCoin(mintableDenom, sdkmath.NewInt(125))))
	requireT.NoError(err)

	// try to mint as non-issuer, which should succeed if the extension permits
	err = bankKeeper.SendCoins(ctx, randomAddr, randomAddr, coinsToMint)
	requireT.NoError(err)

	// mint tokens and check balance and total supply
	err = bankKeeper.SendCoins(ctx, addr, addr, coinsToMint)
	requireT.NoError(err)

	balance := bankKeeper.GetBalance(ctx, addr, mintableDenom)
	requireT.EqualValues(sdk.NewCoin(mintableDenom, sdkmath.NewInt(757)), balance)

	totalSupply, err := bankKeeper.TotalSupply(ctx, &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(987), totalSupply.Supply.AmountOf(mintableDenom))

	// mint to another account
	err = bankKeeper.SendCoins(ctx, addr, randomAddr, coinsToMint)
	requireT.NoError(err)

	balance = bankKeeper.GetBalance(ctx, randomAddr, mintableDenom)
	requireT.EqualValues(sdk.NewCoin(mintableDenom, sdkmath.NewInt(335)), balance)

	totalSupply, err = bankKeeper.TotalSupply(ctx, &banktypes.QueryTotalSupplyRequest{})
	requireT.NoError(err)
	requireT.EqualValues(sdkmath.NewInt(1092), totalSupply.Supply.AmountOf(mintableDenom))
}

func TestKeeper_Extension_BurnRate_BankSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	assetKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper
	ba := newBankAsserter(ctx, t, bankKeeper)

	admin := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	codeID, _, err := testApp.WasmPermissionedKeeper.Create(
		ctx, admin, testcontracts.AssetExtensionWasm, &wasmtypes.AllowEverybody,
	)
	requireT.NoError(err)

	// issue token
	settings := types.IssueSettings{
		Issuer:        admin,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     6,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(1100),
		Features:      []types.Feature{types.Feature_extension},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
		},
		BurnRate: sdkmath.LegacyMustNewDecFromStr("0.25"),
	}

	denom, err := assetKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// send from admin to recipient (burn must apply if the extension decides)
	err = bankKeeper.SendCoins(ctx, admin, recipient, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(500)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient: 500,
		&admin:     475, // 1100 - 500 - 125 (25% burn)
	})

	// send trigger amount from recipient1 to recipient2 (burn must not apply if the extension decides)
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, sdk.NewCoins(
		sdk.NewCoin(denom, AmountIgnoreBurnRateTrigger),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient:  392, // 500 - 108 (AmountIgnoreBurnRateTrigger)
		&recipient2: 108, // 108 (AmountIgnoreBurnRateTrigger)
		&admin:      475,
	})

	// send from recipient1 to recipient2 (burn must apply)
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(100)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient:  267, // 392 - 100 - 25 (25% burn)
		&recipient2: 208, // 108 + 100
		&admin:      475,
	})

	// send from recipient to admin account (burn must apply if the extension decides)
	err = bankKeeper.SendCoins(ctx, recipient, admin, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(213)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient2: 208,
		&admin:      688, // 474 + 213
	})
}

func TestKeeper_Extension_BurnRate_BankMultiSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	assetKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper
	ba := newBankAsserter(ctx, t, bankKeeper)

	// issue 2 tokens
	var admins []sdk.AccAddress
	var denoms []string
	admins = append(admins, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()))
	settings1 := types.IssueSettings{
		Issuer:             admins[0],
		Symbol:             "DEF0",
		Subunit:            "def0",
		Precision:          6,
		Description:        "DEF Desc",
		InitialAmount:      sdkmath.NewInt(1000),
		Features:           []types.Feature{},
		BurnRate:           sdkmath.LegacyNewDec(1).QuoInt64(10), // 10%
		SendCommissionRate: sdkmath.LegacyZeroDec(),
	}

	denom1, err := assetKeeper.Issue(ctx, settings1)
	requireT.NoError(err)
	denoms = append(denoms, denom1)

	// create 4 recipient for every admin to allow for complex test cases
	recipients := lo.RepeatBy(4, func(index int) sdk.AccAddress {
		return sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	})

	admins = append(admins, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()))

	codeID, _, err := testApp.WasmPermissionedKeeper.Create(
		ctx, admins[1], testcontracts.AssetExtensionWasm, &wasmtypes.AllowEverybody,
	)
	requireT.NoError(err)

	settings2 := types.IssueSettings{
		Issuer:        admins[1],
		Symbol:        "DEF1",
		Subunit:       "def1",
		Precision:     6,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(1000),
		Features:      []types.Feature{types.Feature_extension},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
		},
		BurnRate:           sdkmath.LegacyNewDec(1).QuoInt64(10), // 10%
		SendCommissionRate: sdkmath.LegacyZeroDec(),
	}

	denom2, err := assetKeeper.Issue(ctx, settings2)
	requireT.NoError(err)
	denoms = append(denoms, denom2)

	testCases := []struct {
		name         string
		input        banktypes.Input
		outputs      []banktypes.Output
		distribution map[string]map[*sdk.AccAddress]int64
	}{
		{
			name: "send from admin1 to other accounts",
			input: banktypes.Input{
				Address: admins[1].String(),
				Coins:   sdk.NewCoins(sdk.NewCoin(denoms[1], sdkmath.NewInt(600))),
			},
			outputs: []banktypes.Output{
				{Address: recipients[0].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdkmath.NewInt(100)),
				)},
				{Address: recipients[1].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdkmath.NewInt(100)),
				)},
				{Address: admins[0].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdkmath.NewInt(400)),
				)},
			},
			distribution: map[string]map[*sdk.AccAddress]int64{
				denoms[1]: {
					&admins[1]:     340, // 1000 - 600 - 60 (10% burn)
					&admins[0]:     400,
					&recipients[0]: 100,
					&recipients[1]: 100,
				},
			},
		},
		{
			name: "send from admin0 to other accounts",
			input: banktypes.Input{
				Address: admins[0].String(), Coins: sdk.NewCoins(
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
					&admins[0]:     800, // 1000 - 200
					&recipients[0]: 100,
					&recipients[1]: 100,
				},
				denoms[1]: {
					&admins[1]:     340,
					&admins[0]:     180, // 400 - 200 - 20 (10% burn)
					&recipients[0]: 200, // 100 + 100
					&recipients[1]: 200, // 100 + 100
				},
			},
		},
		{
			name: "include admin in recipients",
			input: banktypes.Input{
				Address: recipients[0].String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdkmath.NewInt(60)),
					sdk.NewCoin(denoms[1], sdkmath.NewInt(60)),
				),
			},
			outputs: []banktypes.Output{
				{Address: admins[1].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdkmath.NewInt(25)),
				)},
				{Address: admins[0].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdkmath.NewInt(15)),
				)},
				{Address: recipients[2].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdkmath.NewInt(11)),
				)},
				{Address: recipients[3].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdkmath.NewInt(9)),
				)},
				{Address: admins[1].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdkmath.NewInt(25)),
				)},
				{Address: admins[0].String(), Coins: sdk.NewCoins(
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
					&admins[0]:     815, // 800 + 15
					&admins[1]:     25,
					&recipients[0]: 34, // 100 - 60 - 6 (10% burn)
					&recipients[1]: 100,
					&recipients[2]: 11,
					&recipients[3]: 9,
				},
				denoms[1]: {
					&admins[1]:     365, // 340 + 25
					&admins[0]:     195, // 180 + 15
					&recipients[0]: 132, // 200 - 60 - (3+2+2+1) (10% burn of 25, 15, 11 and 9)
					&recipients[1]: 200,
					&recipients[2]: 11,
					&recipients[3]: 9,
				},
			},
		},
	}

	for counter, tc := range testCases {
		t.Run(fmt.Sprintf("%s case #%d", tc.name, counter), func(t *testing.T) {
			err := bankKeeper.InputOutputCoins(ctx, tc.input, tc.outputs)
			requireT.NoError(err)

			for denom, dist := range tc.distribution {
				ba.assertCoinDistribution(denom, dist)
			}
		})
	}
}

func TestKeeper_Extension_SendCommissionRate_BankSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	assetKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper
	ba := newBankAsserter(ctx, t, bankKeeper)

	admin := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	codeID, _, err := testApp.WasmPermissionedKeeper.Create(
		ctx, admin, testcontracts.AssetExtensionWasm, &wasmtypes.AllowEverybody,
	)
	requireT.NoError(err)

	// issue token
	settings := types.IssueSettings{
		Issuer:        admin,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     6,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(625),
		Features:      []types.Feature{types.Feature_extension},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
		},
		SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("0.25"),
	}

	denom, err := assetKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	token, err := assetKeeper.GetToken(ctx, denom)
	requireT.NoError(err)

	extension, err := sdk.AccAddressFromBech32(token.ExtensionCWAddress)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// send from admin to recipient (send commission rate must apply if the extension decides)
	err = bankKeeper.SendCoins(ctx, admin, recipient, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(500)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient: 500,
		&admin:     62, // 625 - 500 - 125 (25% commission from sender) + 62 (50% of the commission to the admin)
		&extension: 63, // 63 (50% of the commission to the extension)
	})

	// send trigger amount from recipient1 to recipient2 (send commission rate must not apply)
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, sdk.NewCoins(
		sdk.NewCoin(denom, AmountIgnoreSendCommissionRateTrigger),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient:  391, // 500 - 109 (AmountIgnoreSendCommissionRateTrigger)
		&recipient2: 109, // AmountIgnoreSendCommissionRateTrigger
		&admin:      62,
		&extension:  63,
	})

	// send from recipient1 to recipient2 (send commission rate must apply)
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(100)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient:  266, // 391 - 100 - 25 (25% commission rate from the sender)
		&recipient2: 209, // 109 + 100
		&admin:      74,  // 62 + 12 (50% of the commission to the admin)
		&extension:  76,  // 63 + 13 (50% of the commission to the extension)
	})

	// send from recipient to admin account (send commission rate must apply if the extension decides)
	err = bankKeeper.SendCoins(ctx, recipient, admin, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(100)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient:  141, // 266 - 100 - 25 (25% commission rate from the sender)
		&recipient2: 209,
		&admin:      186, // 74 + 100 + 12 (50% of the commission to the admin)
		&extension:  89,  // 76 + 13 (50% of the commission to the extension)
	})

	// clear admin, query admin of definition
	err = assetKeeper.ClearAdmin(ctx, admin, denom)
	requireT.NoError(err)
	def, err := assetKeeper.GetDefinition(ctx, denom)
	requireT.NoError(err)
	requireT.Empty(def.Admin)

	// send from recipient1 to recipient2 (send commission rate must apply, but all of it should go to the extension)
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, sdk.NewCoins(
		sdk.NewCoin(denom, sdkmath.NewInt(112)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient:  1, // 141 - 112 - 28 (25% commission rate from the sender)
		&recipient2: 321,
		&admin:      186, // previous admin does not receive anything
		&extension:  117, // 89 + 28 (100% of the commission to the extension, since there is no admin)
	})
}

func TestKeeper_Extension_ClearAdmin(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	bankKeeper := testApp.BankKeeper
	ftKeeper := testApp.AssetFTKeeper

	admin := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	sender := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	codeID, _, err := testApp.WasmPermissionedKeeper.Create(
		ctx, admin, testcontracts.AssetExtensionWasm, &wasmtypes.AllowEverybody,
	)
	requireT.NoError(err)

	settings := types.IssueSettings{
		Issuer:        admin,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(666),
		Features:      []types.Feature{types.Feature_extension},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
		},
		SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("0.1"),
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	token, err := ftKeeper.GetToken(ctx, denom)
	requireT.NoError(err)

	extensionCWAddress, err := sdk.AccAddressFromBech32(token.ExtensionCWAddress)
	requireT.NoError(err)

	// send some amount to an account
	err = bankKeeper.SendCoins(ctx, admin, sender, sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(200))))
	requireT.NoError(err)

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

	extensionBalanceBefore, err := bankKeeper.Balance(ctx, banktypes.NewQueryBalanceRequest(extensionCWAddress, denom))
	requireT.NoError(err)

	// send some amount between two accounts
	err = bankKeeper.SendCoins(ctx, sender, recipient, sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100))))
	requireT.NoError(err)

	extensionBalanceAfter, err := bankKeeper.Balance(ctx, banktypes.NewQueryBalanceRequest(extensionCWAddress, denom))
	requireT.NoError(err)

	requireT.Equal(
		"10",
		extensionBalanceAfter.Balance.Amount.Sub(extensionBalanceBefore.Balance.Amount).String(),
	)
}
