package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v6/x/asset/ft/types"
)

func TestBaseKeeperWrapper_SpendableBalances(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false)

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	totalTokens := 10
	amountToSend := sdkmath.NewInt(100)
	denoms := make([]string, 0, totalTokens)
	for i := range totalTokens {
		settings := types.IssueSettings{
			Issuer:        issuer,
			Symbol:        fmt.Sprintf("DEF%d", i),
			Subunit:       fmt.Sprintf("def%d", i),
			Precision:     1,
			InitialAmount: sdkmath.NewInt(666),
			Features:      []types.Feature{types.Feature_freezing},
		}
		denom, err := ftKeeper.Issue(ctx, settings)
		requireT.NoError(err)
		denoms = append(denoms, denom)

		coinToSend := sdk.NewCoin(denom, amountToSend)
		err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(
			coinToSend,
		))
		requireT.NoError(err)
	}

	balances := bankKeeper.GetAllBalances(ctx, recipient)
	spendableBalancesRes, err := bankKeeper.SpendableBalances(ctx, &banktypes.QuerySpendableBalancesRequest{
		Address: recipient.String(),
	})
	requireT.NoError(err)
	requireT.Equal(balances.String(), spendableBalancesRes.Balances.String())

	denom := denoms[5]
	// freeze tokens
	coinToFreeze := sdk.NewCoin(denom, sdkmath.NewInt(10))
	err = ftKeeper.Freeze(ctx, issuer, recipient, coinToFreeze)
	requireT.NoError(err)

	// check that after the freezing the spendable balance is different
	spendableBalancesRes, err = bankKeeper.SpendableBalances(ctx, &banktypes.QuerySpendableBalancesRequest{
		Address: recipient.String(),
	})
	requireT.NoError(err)
	requireT.Equal(
		balances.AmountOf(denom).Sub(coinToFreeze.Amount).String(),
		spendableBalancesRes.Balances.AmountOf(denom).String(),
	)

	// check with global freeze
	err = ftKeeper.GloballyFreeze(ctx, issuer, denom)
	requireT.NoError(err)
	spendableBalancesRes, err = bankKeeper.SpendableBalances(ctx, &banktypes.QuerySpendableBalancesRequest{
		Address: recipient.String(),
	})
	requireT.NoError(err)
	requireT.Equal(
		sdkmath.ZeroInt().String(),
		spendableBalancesRes.Balances.AmountOf(denom).String(),
	)
}

func TestBaseKeeperWrapper_SpendableBalanceByDenom(t *testing.T) {
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
		InitialAmount: sdkmath.NewInt(666),
		Features:      []types.Feature{types.Feature_freezing},
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	coinToSend := sdk.NewCoin(denom, sdkmath.NewInt(100))
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(
		coinToSend,
	))
	requireT.NoError(err)

	// check that before the freezing the balance is correct
	balance := bankKeeper.GetBalance(ctx, recipient, denom)
	requireT.Equal(coinToSend, balance)
	spendableBalanceRes, err := bankKeeper.SpendableBalanceByDenom(ctx, &banktypes.QuerySpendableBalanceByDenomRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)

	requireT.Equal(balance.String(), spendableBalanceRes.Balance.String())

	// freeze tokens
	coinToFreeze := sdk.NewCoin(denom, sdkmath.NewInt(10))
	err = ftKeeper.Freeze(ctx, issuer, recipient, coinToFreeze)
	requireT.NoError(err)

	// check that after the freezing the balance is the same
	balance = bankKeeper.GetBalance(ctx, recipient, denom)
	requireT.Equal(coinToSend, balance)

	// check that after the freezing the spendable balance is different
	spendableBalanceRes, err = bankKeeper.SpendableBalanceByDenom(ctx, &banktypes.QuerySpendableBalanceByDenomRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(balance.Sub(coinToFreeze).String(), spendableBalanceRes.Balance.String())

	// check that after the locking the spendable balance is different
	coinToLock := sdk.NewCoin(denom, sdkmath.NewInt(10))
	err = ftKeeper.DEXIncreaseLocked(ctx, recipient, coinToLock)
	requireT.NoError(err)
	spendableBalanceRes, err = bankKeeper.SpendableBalanceByDenom(ctx, &banktypes.QuerySpendableBalanceByDenomRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(balance.Sub(coinToFreeze).String(), spendableBalanceRes.Balance.String())

	// freeze globally
	err = ftKeeper.GloballyFreeze(ctx, issuer, denom)
	requireT.NoError(err)
	// check that it is fully frozen now
	spendableBalanceRes, err = bankKeeper.SpendableBalanceByDenom(ctx, &banktypes.QuerySpendableBalanceByDenomRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.ZeroInt().String(), spendableBalanceRes.Balance.Amount.String())

	// query for the non-existing denom
	spendableBalanceRes, err = bankKeeper.SpendableBalanceByDenom(ctx, &banktypes.QuerySpendableBalanceByDenomRequest{
		Address: recipient.String(),
		Denom:   "nondenom",
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.ZeroInt().String(), spendableBalanceRes.Balance.Amount.String())

	// tests native denom
	nativeDenom := "ucore"
	coinToMindAndSend := sdk.NewCoin(nativeDenom, sdkmath.NewInt(100))
	err = bankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(coinToMindAndSend))
	requireT.NoError(err)
	err = bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, recipient, sdk.NewCoins(
		coinToMindAndSend,
	))
	requireT.NoError(err)
	balance = bankKeeper.GetBalance(ctx, recipient, nativeDenom)
	requireT.Equal(coinToMindAndSend, balance)
}
