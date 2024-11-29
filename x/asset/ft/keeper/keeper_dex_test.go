package keeper_test

import (
	"fmt"
	"math"
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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v5/testutil/simapp"
	testcontracts "github.com/CoreumFoundation/coreum/v5/x/asset/ft/keeper/test-contracts"
	"github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	cwasmtypes "github.com/CoreumFoundation/coreum/v5/x/wasm/types"
)

func TestKeeper_ValidateSpendableNotFT(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false)

	ftKeeper := testApp.AssetFTKeeper

	acc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	coinToSend := sdk.NewInt64Coin(denom1, 1000)
	requireT.NoError(testApp.FundAccount(ctx, acc, sdk.NewCoins(coinToSend)))

	requireT.NoError(
		ftKeeper.DEXCheckOrderAmounts(
			ctx,
			types.DEXOrder{Creator: acc},
			coinToSend,
			sdk.NewInt64Coin(denom1, 0),
		),
	)

	// lock a half
	requireT.NoError(ftKeeper.DEXIncreaseLocked(ctx, acc, sdk.NewInt64Coin(denom1, 500)))

	err := ftKeeper.DEXCheckOrderAmounts(
		ctx,
		types.DEXOrder{Creator: acc},
		coinToSend,
		sdk.NewInt64Coin(denom1, 0),
	)
	// validate one more time
	requireT.ErrorIs(err, types.ErrDEXInsufficientSpendableBalance)
	requireT.ErrorContains(
		err,
		fmt.Sprintf(
			"%s is not available, available %s", coinToSend.String(),
			sdk.NewInt64Coin(denom1, 500).String(),
		),
	)
}

func TestKeeper_DEXExpectedToReceive(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	sender := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	noFeaturesIssuanceSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     1,
		Description:   "ABC Desc",
		InitialAmount: sdkmath.NewInt(666),
		Features:      []types.Feature{},
	}

	noFeaturesDenom, err := ftKeeper.Issue(ctx, noFeaturesIssuanceSettings)
	requireT.NoError(err)
	_, err = ftKeeper.GetToken(ctx, noFeaturesDenom)
	requireT.NoError(err)

	// function passed but nothing is reserved
	requireT.NoError(ftKeeper.DEXIncreaseExpectedToReceive(
		ctx, recipient, sdk.NewCoin(noFeaturesDenom, sdkmath.NewInt(1)),
	))
	requireT.True(ftKeeper.GetDEXExpectedToReceivedBalance(ctx, recipient, noFeaturesDenom).IsZero())

	// increase for not asset FT denom, passes but nothing is reserved
	notFTDenom := types.BuildDenom("nonexist", issuer)
	requireT.NoError(ftKeeper.DEXIncreaseExpectedToReceive(
		ctx, recipient, sdk.NewCoin(notFTDenom, sdkmath.NewInt(10)),
	))
	requireT.True(
		ftKeeper.GetDEXExpectedToReceivedBalance(ctx, recipient, "nonexist").IsZero(),
	)

	whitelistingIssuanceSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(666),
		Features:      []types.Feature{types.Feature_whitelisting},
	}

	whitelistingDenom, err := ftKeeper.Issue(ctx, whitelistingIssuanceSettings)
	requireT.NoError(err)

	// set whitelisted balance
	coinToSend := sdk.NewCoin(whitelistingDenom, sdkmath.NewInt(100))
	// whitelist sender and whitelistingDenom
	requireT.NoError(ftKeeper.SetWhitelistedBalance(ctx, issuer, sender, coinToSend))
	requireT.NoError(bankKeeper.SendCoins(ctx, issuer, sender, sdk.NewCoins(coinToSend)))
	// send without the expected to received balance
	requireT.NoError(ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, coinToSend))
	requireT.NoError(bankKeeper.SendCoins(ctx, sender, recipient, sdk.NewCoins(coinToSend)))
	// return coin
	requireT.NoError(bankKeeper.SendCoins(ctx, recipient, sender, sdk.NewCoins(coinToSend)))
	// increase expected to received balance
	coinToIncreaseExpectedToReceive := sdk.NewCoin(whitelistingDenom, sdkmath.NewInt(1))
	requireT.NoError(ftKeeper.DEXIncreaseExpectedToReceive(ctx, recipient, coinToIncreaseExpectedToReceive))
	requireT.Equal(
		coinToIncreaseExpectedToReceive.String(),
		ftKeeper.GetDEXExpectedToReceivedBalance(ctx, recipient, whitelistingDenom).String(),
	)
	// try to send with the increased part
	requireT.ErrorIs(
		bankKeeper.SendCoins(ctx, sender, recipient, sdk.NewCoins(coinToSend)),
		types.ErrWhitelistedLimitExceeded,
	)

	// try to decrease more that the balance
	requireT.ErrorIs(
		cosmoserrors.ErrInsufficientFunds,
		ftKeeper.DEXDecreaseExpectedToReceive(
			ctx, recipient, coinToIncreaseExpectedToReceive.Add(coinToIncreaseExpectedToReceive),
		),
	)

	requireT.NoError(ftKeeper.DEXDecreaseExpectedToReceive(ctx, recipient, coinToIncreaseExpectedToReceive))
	requireT.True(ftKeeper.GetDEXExpectedToReceivedBalance(ctx, recipient, whitelistingDenom).IsZero())
	// send without decreased amount
	requireT.NoError(bankKeeper.SendCoins(ctx, sender, recipient, sdk.NewCoins(coinToSend)))
}

func TestKeeper_DEXLocked(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

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
	requireT.ErrorIs(ftKeeper.DEXIncreaseLocked(ctx, acc, coinToSend), types.ErrDEXInsufficientSpendableBalance)
	requireT.NoError(bankKeeper.SendCoins(ctx, issuer, acc, sdk.NewCoins(coinToSend)))

	// try to send full balance with the vesting locked coins
	requireT.ErrorIs(
		bankKeeper.SendCoins(ctx, acc, acc, sdk.NewCoins(coinToSend.Add(vestingCoin))),
		cosmoserrors.ErrInsufficientFunds,
	)
	requireT.ErrorIs(
		ftKeeper.DEXCheckOrderAmounts(
			ctx,
			types.DEXOrder{Creator: acc},
			coinToSend.Add(vestingCoin),
			sdk.NewInt64Coin(denom1, 0),
		),
		types.ErrDEXInsufficientSpendableBalance,
	)
	// send max allowed amount
	requireT.NoError(bankKeeper.SendCoins(ctx, acc, acc, sdk.NewCoins(coinToSend)))

	// lock full allowed amount (but without the amount locked by vesting)
	requireT.NoError(ftKeeper.DEXIncreaseLocked(ctx, acc, coinToSend))
	// try to send at least one coin
	requireT.ErrorIs(
		bankKeeper.SendCoins(ctx, acc, acc, sdk.NewCoins(sdk.NewInt64Coin(denom, 1))),
		cosmoserrors.ErrInsufficientFunds,
	)
	requireT.ErrorIs(
		ftKeeper.DEXCheckOrderAmounts(
			ctx,
			types.DEXOrder{Creator: acc},
			sdk.NewInt64Coin(denom, 1),
			sdk.NewInt64Coin(denom1, 0),
		),
		types.ErrDEXInsufficientSpendableBalance,
	)
	// DEX decrease locked full balance
	requireT.NoError(ftKeeper.DEXDecreaseLocked(ctx, acc, coinToSend))
	// DEX lock one more time
	requireT.NoError(ftKeeper.DEXIncreaseLocked(ctx, acc, coinToSend))

	balance := bankKeeper.GetBalance(ctx, acc, denom)
	requireT.Equal(coinToSend.Add(vestingCoin).String(), balance.String())

	// try to DEX lock coins which are locked by the vesting
	requireT.ErrorIs(ftKeeper.DEXIncreaseLocked(ctx, acc, vestingCoin), types.ErrDEXInsufficientSpendableBalance)

	// try lock decrease locked full balance
	requireT.ErrorIs(ftKeeper.DEXDecreaseLocked(ctx, acc, balance), cosmoserrors.ErrInsufficientFunds)
	requireT.ErrorIs(
		ftKeeper.DEXDecreaseLocked(ctx, acc, balance),
		cosmoserrors.ErrInsufficientFunds,
	)

	// decrease locked part
	requireT.NoError(ftKeeper.DEXDecreaseLocked(ctx, acc, sdk.NewInt64Coin(denom, 400)))
	requireT.Equal(sdk.NewInt64Coin(denom, 600).String(), ftKeeper.GetDEXLockedBalance(ctx, acc, denom).String())
	spendableBalance, err := ftKeeper.GetSpendableBalance(ctx, acc, denom)
	requireT.NoError(err)
	requireT.Equal(sdk.NewInt64Coin(denom, 400).String(), spendableBalance.String())

	// freeze locked balance
	requireT.NoError(ftKeeper.Freeze(ctx, issuer, acc, coinToSend))
	// 1050 - total, 600 locked by dex, 50 locked by bank, 1000 frozen
	spendableBalance, err = ftKeeper.GetSpendableBalance(ctx, acc, denom)
	requireT.NoError(err)
	requireT.Equal(sdk.NewInt64Coin(denom, 50).String(), spendableBalance.String())

	// decrease locked 2d part, even when it's frozen we allow it
	requireT.NoError(ftKeeper.DEXDecreaseLocked(ctx, acc, sdk.NewInt64Coin(denom, 600)))
	requireT.Equal(sdkmath.ZeroInt().String(), ftKeeper.GetDEXLockedBalance(ctx, acc, denom).Amount.String())

	// check order amounts are spendable with frozen coins
	requireT.ErrorIs(
		ftKeeper.DEXCheckOrderAmounts(
			ctx,
			types.DEXOrder{Creator: acc},
			coinToSend,
			sdk.NewInt64Coin(denom1, 0),
		),
		types.ErrDEXInsufficientSpendableBalance,
	)

	// unfreeze part
	requireT.NoError(ftKeeper.Unfreeze(ctx, issuer, acc, sdk.NewInt64Coin(denom, 300)))
	frozenBalance, err := ftKeeper.GetFrozenBalance(ctx, acc, denom)
	requireT.NoError(err)
	requireT.Equal(sdk.NewInt64Coin(denom, 700).String(), frozenBalance.String())

	// now 700 frozen, 50 locked by vesting, 1050 balance
	// try to use more than allowed
	err = ftKeeper.DEXCheckOrderAmounts(
		ctx,
		types.DEXOrder{Creator: acc},
		sdk.NewInt64Coin(denom, 351),
		sdk.NewInt64Coin(denom1, 0),
	)
	requireT.ErrorIs(err, types.ErrDEXInsufficientSpendableBalance)
	requireT.ErrorContains(err, "available 350")

	// try to send more than allowed
	err = bankKeeper.SendCoins(ctx, acc, acc, sdk.NewCoins(sdk.NewInt64Coin(denom, 351)))
	requireT.ErrorIs(err, cosmoserrors.ErrInsufficientFunds)
	requireT.ErrorContains(err, "available 350")

	// try to use with global freezing
	requireT.NoError(ftKeeper.GloballyFreeze(ctx, issuer, denom))
	requireT.ErrorContains(
		ftKeeper.DEXCheckOrderAmounts(
			ctx,
			types.DEXOrder{Creator: acc},
			sdk.NewInt64Coin(denom, 350),
			sdk.NewInt64Coin(denom1, 0),
		),
		fmt.Sprintf("usage of %s for DEX is blocked because the token is globally frozen", denom),
	)
	spendableBalance, err = ftKeeper.GetSpendableBalance(ctx, acc, denom)
	requireT.NoError(err)
	requireT.True(spendableBalance.IsZero())
	// globally unfreeze now and check that we can use the previously locked amount
	requireT.NoError(ftKeeper.GloballyUnfreeze(ctx, issuer, denom))
	requireT.NoError(
		ftKeeper.DEXCheckOrderAmounts(
			ctx,
			types.DEXOrder{Creator: acc},
			sdk.NewInt64Coin(denom, 350),
			sdk.NewInt64Coin(denom1, 0),
		),
	)
	requireT.NoError(ftKeeper.DEXIncreaseLocked(ctx, acc, sdk.NewInt64Coin(denom, 350)))
	// freeze more than balance
	requireT.NoError(ftKeeper.Freeze(ctx, issuer, acc, sdk.NewInt64Coin(denom, 1_000_000)))
}

func TestKeeper_DEXBlockSmartContracts(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEFBLK",
		Subunit:       "defblk",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 10),
		Features: []types.Feature{
			types.Feature_block_smart_contracts,
		},
	}
	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)
	blockSmartContractCoin := sdk.NewInt64Coin(denom, 50)
	requireT.NoError(bankKeeper.SendCoins(ctx, issuer, acc, sdk.NewCoins(blockSmartContractCoin)))
	// triggered from native call
	requireT.NoError(
		ftKeeper.DEXCheckOrderAmounts(
			ctx,
			types.DEXOrder{Creator: acc},
			blockSmartContractCoin,
			sdk.NewInt64Coin(denom1, 1),
		),
	)

	ctxFromSmartContract := cwasmtypes.WithSmartContractSender(ctx, acc.String())
	blockingErr := fmt.Sprintf("usage of %s is not supported for DEX in smart contract", denom)
	testApp.MintAndSendCoin(t, ctxFromSmartContract, acc, sdk.NewCoins(sdk.NewInt64Coin(denom1, 1)))
	requireT.ErrorContains(
		ftKeeper.DEXCheckOrderAmounts(
			ctxFromSmartContract,
			types.DEXOrder{Creator: acc},
			blockSmartContractCoin,
			sdk.NewInt64Coin(denom1, 1),
		),
		blockingErr,
	)
	requireT.ErrorContains(
		ftKeeper.DEXCheckOrderAmounts(
			ctxFromSmartContract,
			types.DEXOrder{Creator: acc},
			sdk.NewInt64Coin(denom1, 1),
			blockSmartContractCoin,
		),
		blockingErr,
	)

	// but still allowed to lock by admin
	testApp.MintAndSendCoin(t, ctxFromSmartContract, issuer, sdk.NewCoins(sdk.NewInt64Coin(denom1, 1)))
	requireT.NoError(
		ftKeeper.DEXCheckOrderAmounts(
			ctxFromSmartContract,
			types.DEXOrder{Creator: issuer},
			blockSmartContractCoin,
			sdk.NewInt64Coin(denom1, 1),
		),
	)
	requireT.NoError(
		ftKeeper.DEXCheckOrderAmounts(
			ctxFromSmartContract,
			types.DEXOrder{Creator: issuer},
			sdk.NewInt64Coin(denom1, 1),
			blockSmartContractCoin,
		),
	)
}

func TestKeeper_DEXSettings_BlockDEX(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	ft1Settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 10),
		Features: []types.Feature{
			types.Feature_freezing,
			types.Feature_dex_block,
		},
	}

	invalidFT1Settings := ft1Settings
	invalidFT1Settings.DEXSettings = &types.DEXSettings{
		WhitelistedDenoms: []string{denom1},
	}
	trialCtx := simapp.CopyContextWithMultiStore(ctx)
	_, err := ftKeeper.Issue(trialCtx, invalidFT1Settings)
	requireT.ErrorIs(err, types.ErrFeatureDisabled)

	ft1Denom, err := ftKeeper.Issue(ctx, ft1Settings)
	requireT.NoError(err)
	dexBlockedCoin := sdk.NewInt64Coin(ft1Denom, 50)
	requireT.NoError(bankKeeper.SendCoins(ctx, issuer, acc, sdk.NewCoins(dexBlockedCoin)))

	errStr := fmt.Sprintf("usage of %s is not supported for DEX, the token has dex_block", ft1Denom)
	requireT.ErrorContains(ftKeeper.DEXCheckOrderAmounts(
		ctx,
		types.DEXOrder{Creator: acc},
		dexBlockedCoin,
		sdk.NewInt64Coin(denom1, 0),
	), errStr)
	requireT.ErrorContains(ftKeeper.DEXCheckOrderAmounts(
		ctx,
		types.DEXOrder{Creator: acc},
		sdk.NewInt64Coin(denom1, 0),
		dexBlockedCoin,
	), errStr)
}

func TestKeeper_DEXSettings_WhitelistedDenom(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	ft1Settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 10),
		Features: []types.Feature{
			types.Feature_dex_whitelisted_denoms,
		},
		DEXSettings: &types.DEXSettings{
			WhitelistedDenoms: []string{
				denom1,
			},
		},
	}
	ft1Denom, err := ftKeeper.Issue(ctx, ft1Settings)
	requireT.NoError(err)

	ft2Settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF2",
		Subunit:       "def2",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 10),
		Features: []types.Feature{
			types.Feature_dex_whitelisted_denoms,
		},
		DEXSettings: &types.DEXSettings{
			WhitelistedDenoms: []string{
				ft1Denom,
			},
		},
	}
	ft2Denom, err := ftKeeper.Issue(ctx, ft2Settings)
	requireT.NoError(err)

	ft1CoinToLock := sdk.NewInt64Coin(ft1Denom, 10)
	requireT.NoError(bankKeeper.SendCoins(ctx, issuer, acc, sdk.NewCoins(ft1CoinToLock)))
	errStr := fmt.Sprintf("denom %s not whitelisted for %s", denom2, ft1Denom)
	requireT.ErrorContains(
		ftKeeper.DEXCheckOrderAmounts(
			ctx,
			types.DEXOrder{Creator: acc},
			ft1CoinToLock,
			sdk.NewInt64Coin(denom2, 1),
		),
		errStr,
	)

	requireT.NoError(ftKeeper.DEXCheckOrderAmounts(
		ctx,
		types.DEXOrder{Creator: acc},
		ft1CoinToLock,
		sdk.NewInt64Coin(denom1, 1),
	))

	denom2CoinToLock := sdk.NewInt64Coin(denom2, 10)
	testApp.MintAndSendCoin(t, ctx, acc, sdk.NewCoins(denom2CoinToLock))
	// can't lock the receive denom
	errStr = fmt.Sprintf("denom %s not whitelisted for %s", denom2, ft1Denom)
	requireT.ErrorContains(
		ftKeeper.DEXCheckOrderAmounts(
			ctx,
			types.DEXOrder{Creator: acc},
			denom2CoinToLock,
			sdk.NewInt64Coin(ft1Denom, 1),
		),
		errStr,
	)

	// both not ft
	requireT.NoError(
		ftKeeper.DEXCheckOrderAmounts(
			ctx,
			types.DEXOrder{Creator: acc},
			denom2CoinToLock,
			sdk.NewInt64Coin(denom1, 1),
		),
	)

	// try to lock both not ft coins
	ft2CoinToLock := sdk.NewInt64Coin(ft2Denom, 10)
	requireT.NoError(bankKeeper.SendCoins(ctx, issuer, acc, sdk.NewCoins(ft2CoinToLock)))
	errStr = fmt.Sprintf("denom %s not whitelisted for %s", ft2Denom, ft1Denom)
	requireT.ErrorContains(
		ftKeeper.DEXCheckOrderAmounts(
			ctx,
			types.DEXOrder{Creator: acc},
			ft2CoinToLock,
			sdk.NewInt64Coin(ft1Denom, 1),
		),
		errStr,
	)
	requireT.NoError(ftKeeper.UpdateDEXWhitelistedDenoms(ctx, issuer, ft1Denom, []string{ft2Denom}))
	// now we can lock
	requireT.NoError(
		ftKeeper.DEXCheckOrderAmounts(
			ctx,
			types.DEXOrder{Creator: acc},
			ft2CoinToLock,
			sdk.NewInt64Coin(ft1Denom, 1),
		),
	)
	//
	// lock not ft denoms without settings
	denom3CoinToLock := sdk.NewInt64Coin(denom3, 10)
	testApp.MintAndSendCoin(t, ctx, acc, sdk.NewCoins(denom3CoinToLock))
	requireT.NoError(
		ftKeeper.DEXCheckOrderAmounts(
			ctx,
			types.DEXOrder{Creator: acc},
			denom3CoinToLock,
			sdk.NewInt64Coin(denom4, 1),
		),
	)
}

func TestKeeper_DEXLimitsWithGlobalFreeze(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false)

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	ft1Settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEFONE",
		Subunit:       "defone",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 10),
		Features: []types.Feature{
			types.Feature_freezing,
		},
	}
	ft1Denom, err := ftKeeper.Issue(ctx, ft1Settings)
	requireT.NoError(err)

	ft2Settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEFTOW",
		Subunit:       "deftwo",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 10),
		Features: []types.Feature{
			types.Feature_freezing,
		},
	}
	ft2Denom, err := ftKeeper.Issue(ctx, ft2Settings)
	requireT.NoError(err)

	// fund acc
	ft1CoinToSend := sdk.NewInt64Coin(ft1Denom, 100)
	ft2CoinToSend := sdk.NewInt64Coin(ft2Denom, 100)
	requireT.NoError(bankKeeper.SendCoins(ctx, issuer, acc, sdk.NewCoins(ft1CoinToSend)))
	requireT.NoError(bankKeeper.SendCoins(ctx, issuer, acc, sdk.NewCoins(ft2CoinToSend)))

	// check that it's allowed to increase and decrease the limits
	requireT.NoError(
		ftKeeper.DEXCheckOrderAmounts(
			ctx,
			types.DEXOrder{Creator: acc},
			ft1CoinToSend,
			ft2CoinToSend,
		),
	)

	// globally freeze
	requireT.NoError(ftKeeper.SetGlobalFreeze(ctx, ft1CoinToSend.Denom, true))
	requireT.ErrorContains(
		ftKeeper.DEXCheckOrderAmounts(
			simapp.CopyContextWithMultiStore(ctx),
			types.DEXOrder{Creator: acc},
			ft1CoinToSend,
			ft2CoinToSend,
		),
		fmt.Sprintf("usage of %s for DEX is blocked because the token is globally frozen", ft1CoinToSend.Denom),
	)

	requireT.NoError(ftKeeper.SetGlobalFreeze(ctx, ft1CoinToSend.Denom, false))
	requireT.NoError(ftKeeper.SetGlobalFreeze(ctx, ft2CoinToSend.Denom, true))
	requireT.ErrorContains(
		ftKeeper.DEXCheckOrderAmounts(
			simapp.CopyContextWithMultiStore(ctx),
			types.DEXOrder{Creator: acc},
			ft1CoinToSend,
			ft2CoinToSend,
		),
		fmt.Sprintf("usage of %s for DEX is blocked because the token is globally frozen", ft2CoinToSend.Denom),
	)

	// admin still can increase the limits
	requireT.NoError(
		ftKeeper.DEXCheckOrderAmounts(
			ctx,
			types.DEXOrder{Creator: issuer},
			ft1CoinToSend,
			ft2CoinToSend,
		),
	)
}

func TestKeeper_LockedNotFT(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false)

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	faucet := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	requireT.NoError(testApp.FundAccount(ctx, faucet, sdk.NewCoins(sdk.NewCoin(denom1, sdkmath.NewIntWithDecimal(1, 10)))))

	// create acc with permanently locked coins (native)
	vestingCoin := sdk.NewInt64Coin(denom1, 50)
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
	requireT.Equal(vestingCoin.Amount.String(), bankKeeper.LockedCoins(ctx, acc).AmountOf(denom1).String())

	coinToSend := sdk.NewInt64Coin(denom1, 1000)
	// try to lock more than balance
	requireT.ErrorIs(ftKeeper.DEXIncreaseLocked(ctx, acc, coinToSend), types.ErrDEXInsufficientSpendableBalance)
	requireT.NoError(bankKeeper.SendCoins(ctx, faucet, acc, sdk.NewCoins(coinToSend)))

	// try to send full balance with the vesting locked coins
	requireT.ErrorIs(
		bankKeeper.SendCoins(ctx, acc, acc, sdk.NewCoins(coinToSend.Add(vestingCoin))),
		cosmoserrors.ErrInsufficientFunds,
	)

	// lock full allowed amount (but without the amount locked by vesting)
	requireT.NoError(ftKeeper.DEXIncreaseLocked(ctx, acc, coinToSend))

	// try to send at least one coin
	requireT.ErrorIs(
		bankKeeper.SendCoins(ctx, acc, acc, sdk.NewCoins(sdk.NewInt64Coin(denom1, 1))),
		cosmoserrors.ErrInsufficientFunds,
	)

	balance := bankKeeper.GetBalance(ctx, acc, denom1)
	requireT.Equal(coinToSend.Add(vestingCoin).String(), balance.String())

	// try lock coins which are locked by the vesting
	requireT.ErrorIs(ftKeeper.DEXIncreaseLocked(ctx, acc, vestingCoin), types.ErrDEXInsufficientSpendableBalance)

	// try decrease locked full balance
	requireT.ErrorIs(ftKeeper.DEXDecreaseLocked(ctx, acc, balance), cosmoserrors.ErrInsufficientFunds)

	// decrease locked part
	requireT.NoError(ftKeeper.DEXDecreaseLocked(ctx, acc, sdk.NewInt64Coin(denom1, 400)))
	requireT.Equal(sdk.NewInt64Coin(denom1, 600).String(), ftKeeper.GetDEXLockedBalance(ctx, acc, denom1).String())
}

func TestKeeper_UpdateDEXUnifiedRefAmount(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	ft1Settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(777),
		DEXSettings: &types.DEXSettings{
			UnifiedRefAmount: lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("0.01")),
		},
	}

	// try to issue without the feature enabled, but with the settings
	_, err := ftKeeper.Issue(simapp.CopyContextWithMultiStore(ctx), ft1Settings)
	requireT.ErrorIs(err, types.ErrFeatureDisabled)

	ft1Settings.Features = []types.Feature{
		types.Feature_dex_unified_ref_amount_change,
	}

	ft1Denom, err := ftKeeper.Issue(ctx, ft1Settings)
	requireT.NoError(err)

	gotToken, err := ftKeeper.GetToken(ctx, ft1Denom)
	requireT.NoError(err)
	expectToken := types.Token{
		Denom:              ft1Denom,
		Issuer:             ft1Settings.Issuer.String(),
		Symbol:             ft1Settings.Symbol,
		Subunit:            strings.ToLower(ft1Settings.Subunit),
		Precision:          ft1Settings.Precision,
		BurnRate:           sdkmath.LegacyNewDec(0),
		SendCommissionRate: sdkmath.LegacyNewDec(0),
		Version:            types.CurrentTokenVersion,
		Admin:              ft1Settings.Issuer.String(),
		Features:           ft1Settings.Features,
		DEXSettings:        ft1Settings.DEXSettings,
	}
	requireT.Equal(expectToken, gotToken)

	// try to update with the invalid settings
	unifiedRefAmount := sdkmath.LegacyMustNewDecFromStr("-0.01")
	requireT.ErrorIs(
		ftKeeper.UpdateDEXUnifiedRefAmount(ctx, issuer, ft1Denom, unifiedRefAmount), types.ErrInvalidInput,
	)

	// try to update from not issuer
	unifiedRefAmount = sdkmath.LegacyMustNewDecFromStr("0.01")
	randomAddr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	requireT.ErrorIs(ftKeeper.UpdateDEXUnifiedRefAmount(
		ctx, randomAddr, ft1Denom, unifiedRefAmount), cosmoserrors.ErrUnauthorized,
	)

	// update the settings
	requireT.NoError(ftKeeper.UpdateDEXUnifiedRefAmount(ctx, issuer, ft1Denom, unifiedRefAmount))

	gotToken, err = ftKeeper.GetToken(ctx, ft1Denom)
	requireT.NoError(err)
	expectToken.DEXSettings = &types.DEXSettings{
		UnifiedRefAmount: &unifiedRefAmount,
	}
	requireT.Equal(expectToken, gotToken)

	// update the settings one more time but with the gov acc
	unifiedRefAmount = sdkmath.LegacyMustNewDecFromStr("999")
	requireT.NoError(ftKeeper.UpdateDEXUnifiedRefAmount(
		ctx, authtypes.NewModuleAddress(govtypes.ModuleName), ft1Denom, unifiedRefAmount),
	)

	gotToken, err = ftKeeper.GetToken(ctx, ft1Denom)
	requireT.NoError(err)
	expectToken.DEXSettings = &types.DEXSettings{
		UnifiedRefAmount: &unifiedRefAmount,
	}
	requireT.Equal(expectToken, gotToken)

	// update the different setting to check that we don't affect other
	whitelistedDenoms := []string{denom1}
	requireT.NoError(ftKeeper.UpdateDEXWhitelistedDenoms(
		ctx, authtypes.NewModuleAddress(govtypes.ModuleName), ft1Denom, whitelistedDenoms,
	))
	unifiedRefAmount = sdkmath.LegacyMustNewDecFromStr("777")
	requireT.NoError(ftKeeper.UpdateDEXUnifiedRefAmount(
		ctx, authtypes.NewModuleAddress(govtypes.ModuleName), ft1Denom, unifiedRefAmount),
	)
	gotToken, err = ftKeeper.GetToken(ctx, ft1Denom)
	requireT.NoError(err)
	expectToken.DEXSettings = &types.DEXSettings{
		UnifiedRefAmount:  &unifiedRefAmount,
		WhitelistedDenoms: whitelistedDenoms,
	}
	requireT.Equal(expectToken, gotToken)

	// try to update settings for the not FT denom from not gov
	requireT.ErrorIs(
		ftKeeper.UpdateDEXUnifiedRefAmount(ctx, issuer, denom1, unifiedRefAmount), cosmoserrors.ErrUnauthorized,
	)
	requireT.NoError(
		ftKeeper.UpdateDEXUnifiedRefAmount(
			ctx, authtypes.NewModuleAddress(govtypes.ModuleName), denom1, unifiedRefAmount,
		),
	)

	dexSettings, err := ftKeeper.GetDEXSettings(ctx, denom1)
	requireT.NoError(err)

	requireT.Equal(types.DEXSettings{
		UnifiedRefAmount: &unifiedRefAmount,
	}, dexSettings)
}

func TestKeeper_UpdateDEXWhitelistedDenoms(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	ft1Settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(777),
		Features: []types.Feature{
			types.Feature_dex_whitelisted_denoms,
		},
	}

	ft1Denom, err := ftKeeper.Issue(ctx, ft1Settings)
	requireT.NoError(err)

	gotToken, err := ftKeeper.GetToken(ctx, ft1Denom)
	requireT.NoError(err)
	expectToken := types.Token{
		Denom:              ft1Denom,
		Issuer:             ft1Settings.Issuer.String(),
		Symbol:             ft1Settings.Symbol,
		Subunit:            strings.ToLower(ft1Settings.Subunit),
		Precision:          ft1Settings.Precision,
		BurnRate:           sdkmath.LegacyNewDec(0),
		SendCommissionRate: sdkmath.LegacyNewDec(0),
		Version:            types.CurrentTokenVersion,
		Admin:              ft1Settings.Issuer.String(),
		DEXSettings:        ft1Settings.DEXSettings,
		Features: []types.Feature{
			types.Feature_dex_whitelisted_denoms,
		},
	}
	requireT.Equal(expectToken, gotToken)

	// try to update with the invalid whitelisted denoms
	whitelistedDenoms := []string{"1denom1"}
	requireT.ErrorIs(ftKeeper.UpdateDEXWhitelistedDenoms(ctx, issuer, ft1Denom, whitelistedDenoms), types.ErrInvalidInput)

	// try to update from not issuer
	whitelistedDenoms = []string{denom1}
	randomAddr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	requireT.ErrorIs(
		ftKeeper.UpdateDEXWhitelistedDenoms(ctx, randomAddr, ft1Denom, whitelistedDenoms), cosmoserrors.ErrUnauthorized,
	)

	requireT.NoError(ftKeeper.UpdateDEXWhitelistedDenoms(ctx, issuer, ft1Denom, whitelistedDenoms))

	gotToken, err = ftKeeper.GetToken(ctx, ft1Denom)
	requireT.NoError(err)
	expectToken.DEXSettings = &types.DEXSettings{
		WhitelistedDenoms: whitelistedDenoms,
	}
	requireT.Equal(expectToken, gotToken)

	// update the to empty list to allow all denoms
	whitelistedDenoms = make([]string, 0)
	requireT.NoError(ftKeeper.UpdateDEXWhitelistedDenoms(ctx, issuer, ft1Denom, whitelistedDenoms))

	gotToken, err = ftKeeper.GetToken(ctx, ft1Denom)
	requireT.NoError(err)
	expectToken.DEXSettings = &types.DEXSettings{
		WhitelistedDenoms: nil,
	}
	requireT.Equal(expectToken, gotToken)

	whitelistedDenoms = []string{denom1}

	// try to update settings for the not FT denom from not gov
	requireT.ErrorIs(
		ftKeeper.UpdateDEXWhitelistedDenoms(ctx, issuer, denom1, whitelistedDenoms), cosmoserrors.ErrUnauthorized,
	)
	// update from gov
	requireT.NoError(
		ftKeeper.UpdateDEXWhitelistedDenoms(
			ctx, authtypes.NewModuleAddress(govtypes.ModuleName), denom1, whitelistedDenoms,
		),
	)

	dexSettings, err := ftKeeper.GetDEXSettings(ctx, denom1)
	requireT.NoError(err)

	requireT.Equal(types.DEXSettings{
		WhitelistedDenoms: whitelistedDenoms,
	}, dexSettings)

	ft2Settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC2",
		Subunit:       "abc2",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(777),
		// no features
	}

	ft2Denom, err := ftKeeper.Issue(ctx, ft2Settings)
	requireT.NoError(err)

	whitelistedDenoms = []string{denom2}

	// try to update settings from issuer
	requireT.ErrorIs(
		ftKeeper.UpdateDEXWhitelistedDenoms(ctx, issuer, ft2Denom, whitelistedDenoms), types.ErrFeatureDisabled,
	)
	// update from gov
	requireT.NoError(
		ftKeeper.UpdateDEXWhitelistedDenoms(
			ctx, authtypes.NewModuleAddress(govtypes.ModuleName), ft2Denom, whitelistedDenoms,
		),
	)

	dexSettings, err = ftKeeper.GetDEXSettings(ctx, ft2Denom)
	requireT.NoError(err)

	requireT.Equal(types.DEXSettings{
		WhitelistedDenoms: whitelistedDenoms,
	}, dexSettings)
}

func TestKeeper_DEXExtensions(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// extension
	codeID, _, err := testApp.WasmPermissionedKeeper.Create(
		ctx, issuer, testcontracts.AssetExtensionWasm, &wasmtypes.AllowEverybody,
	)
	requireT.NoError(err)
	settingsWithExtension := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEFEXT",
		Subunit:       "defext",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 10),
		Features: []types.Feature{
			types.Feature_freezing,
			types.Feature_extension,
			types.Feature_whitelisting,
		},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
		},
	}
	extensionDenom, err := ftKeeper.Issue(ctx, settingsWithExtension)
	requireT.NoError(err)
	coinWithExtension := sdk.NewCoin(extensionDenom, sdkmath.NewInt(200))
	requireT.NoError(ftKeeper.SetWhitelistedBalance(ctx, issuer, acc, coinWithExtension))
	coinWithoutExtension := sdk.NewInt64Coin(denom1, 123)
	testApp.MintAndSendCoin(t, ctx, acc, sdk.NewCoins(coinWithoutExtension))

	// the extension controls all features, but to locked, amount
	requireT.ErrorIs(
		ftKeeper.DEXCheckOrderAmounts(
			simapp.CopyContextWithMultiStore(ctx),
			types.DEXOrder{Creator: acc},
			coinWithExtension,
			coinWithoutExtension,
		),
		types.ErrDEXInsufficientSpendableBalance,
	)
	// send coins now to lock
	requireT.NoError(bankKeeper.SendCoins(ctx, issuer, acc, sdk.NewCoins(coinWithExtension)))

	// freeze locked balance, to check that we check freezing for extension as well
	requireT.NoError(ftKeeper.Freeze(ctx, issuer, acc, coinWithExtension))
	requireT.ErrorContains(
		ftKeeper.DEXCheckOrderAmounts(
			simapp.CopyContextWithMultiStore(ctx),
			types.DEXOrder{Creator: acc},
			coinWithExtension,
			coinWithoutExtension,
		),
		fmt.Sprintf("%s is not available, available 0%s", coinWithExtension.String(), coinWithExtension.Denom),
	)
	requireT.NoError(ftKeeper.Unfreeze(ctx, issuer, acc, coinWithExtension))

	// try with locked balance
	requireT.NoError(ftKeeper.DEXIncreaseLocked(ctx, acc, coinWithExtension))
	requireT.ErrorIs(
		ftKeeper.DEXCheckOrderAmounts(
			simapp.CopyContextWithMultiStore(ctx),
			types.DEXOrder{Creator: acc},
			coinWithExtension,
			coinWithoutExtension,
		),
		types.ErrDEXInsufficientSpendableBalance,
	)

	// check expected to send and receive with prohibited amounts
	requireT.NoError(ftKeeper.DEXDecreaseLocked(ctx, acc, coinWithExtension))

	// try with prohibited balances
	coinWithExtensionProhibitedToSpend := sdk.NewCoin(extensionDenom, AmountDEXExpectToSpendTrigger)
	requireT.ErrorContains(
		ftKeeper.DEXCheckOrderAmounts(
			simapp.CopyContextWithMultiStore(ctx),
			types.DEXOrder{Creator: acc},
			coinWithExtensionProhibitedToSpend,
			coinWithoutExtension,
		),
		"wasm error: DEX order placement is failed",
	)

	coinWithExtensionProhibitedToReceive := sdk.NewCoin(extensionDenom, AmountDEXExpectToReceiveTrigger)
	// check that we respect whitelisted balance
	requireT.ErrorIs(
		ftKeeper.DEXCheckOrderAmounts(
			simapp.CopyContextWithMultiStore(ctx),
			types.DEXOrder{Creator: acc},
			coinWithoutExtension,
			coinWithExtensionProhibitedToReceive,
		),
		types.ErrWhitelistedLimitExceeded,
	)

	whitelistedDenomBalance := bankKeeper.GetBalance(ctx, acc, extensionDenom)
	requireT.NoError(
		ftKeeper.SetWhitelistedBalance(
			ctx, issuer, acc, whitelistedDenomBalance.Add(coinWithExtensionProhibitedToReceive),
		),
	)
	requireT.ErrorContains(
		ftKeeper.DEXCheckOrderAmounts(
			simapp.CopyContextWithMultiStore(ctx),
			types.DEXOrder{Creator: acc},
			coinWithoutExtension,
			coinWithExtensionProhibitedToReceive,
		),
		"wasm error: DEX order placement is failed",
	)

	// increase expected to receive to check that we respect whitelisted balance
	require.NoError(t, ftKeeper.DEXIncreaseExpectedToReceive(ctx, acc, sdk.NewInt64Coin(extensionDenom, 1)))
	requireT.ErrorIs(
		ftKeeper.DEXCheckOrderAmounts(
			simapp.CopyContextWithMultiStore(ctx),
			types.DEXOrder{Creator: acc},
			coinWithoutExtension,
			coinWithExtensionProhibitedToReceive,
		),
		types.ErrWhitelistedLimitExceeded,
	)
}
