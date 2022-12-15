package bank

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
)

// TestMultiSendBatchOutputs tests MultiSend message with maximum amount of addresses.
func TestMultiSendBatchOutputs(ctx context.Context, t testing.T, chain testing.Chain) {
	issuer := chain.GenAccount()
	requireT := require.New(t)

	issueMsg := &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "TOK1",
		Subunit:       "tok1",
		Description:   "TOK1 Description",
		InitialAmount: sdk.NewInt(100_000_000_000_000_000),
		Features: []assettypes.FungibleTokenFeature{
			assettypes.FungibleTokenFeature_freeze, //nolint:nosnakecase // enable the feature to make the computation more complicated
		},
	}

	numAccountsToFund := 1000 // 1700 was the max accepted
	iterationsToFund := 2

	inputItem := banktypes.Input{
		Address: issuer.String(),
		Coins:   sdk.NewCoins(),
	}
	denom := assettypes.BuildFungibleTokenDenom(issueMsg.Subunit, issuer)
	outputItems := make([]banktypes.Output, 0, numAccountsToFund)
	fundedAccounts := make([]sdk.AccAddress, 0, numAccountsToFund)
	coinToFund := sdk.NewCoin(denom, sdk.NewInt(10_000_000_000))

	for i := 0; i < numAccountsToFund; i++ {
		inputItem.Coins = inputItem.Coins.Add(coinToFund)
		recipient := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		fundedAccounts = append(fundedAccounts, recipient)
		outputItems = append(outputItems, banktypes.Output{
			Address: recipient.String(),
			Coins:   sdk.NewCoins(coinToFund),
		})
	}
	// prepare MultiSend messages
	multiSendMsgs := make([]sdk.Msg, 0, iterationsToFund)
	for i := 0; i < iterationsToFund; i++ {
		multiSendMsgs = append(multiSendMsgs, &banktypes.MsgMultiSend{
			Inputs: []banktypes.Input{
				inputItem,
			},
			Outputs: outputItems,
		})
	}

	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, issuer, testing.BalancesOptions{
		Messages: append([]sdk.Msg{issueMsg}, multiSendMsgs...),
		Amount:   sdk.NewInt(10_000_000), // add more coins to cover extra bytes because of the message size
	}))

	// issue fungible tokens
	_, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	// send coins in loop
	start := time.Now()
	for _, msg := range multiSendMsgs {
		res, err := tx.BroadcastTx(
			ctx,
			chain.ClientContext.WithFromAddress(issuer),
			// we estimate here since the th size is grater then allowed for the deterministic fee
			chain.TxFactory().WithSimulateAndExecute(true),
			msg,
		)
		requireT.NoError(err)
		logger.Get(ctx).Info(fmt.Sprintf("Successfully sent batch MultiSend tx, hash: %s, gasUse:%d", res.TxHash, res.GasUsed))
	}
	logger.Get(ctx).Info(fmt.Sprintf("It takes %s to fund %d accounts with MultiSend", time.Since(start), numAccountsToFund*iterationsToFund))

	assertBatchAccounts(ctx, chain, sdk.NewCoins(sdk.NewCoin(coinToFund.Denom, coinToFund.Amount.MulRaw(int64(iterationsToFund)))), fundedAccounts, denom, requireT)
}

// TestSendBatchMsgs tests BankSend message with maximum amount of accounts.
func TestSendBatchMsgs(ctx context.Context, t testing.T, chain testing.Chain) {
	issuer := chain.GenAccount()
	requireT := require.New(t)

	issueMsg := &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "TOK1",
		Subunit:       "tok1",
		Description:   "TOK1 Description",
		InitialAmount: sdk.NewInt(100_000_000_000_000_000),
		Features: []assettypes.FungibleTokenFeature{
			assettypes.FungibleTokenFeature_freeze, //nolint:nosnakecase // enable the feature to make the computation more complicated
		},
	}

	numAccountsToFund := 400 // 600 was the max accepted
	iterationsToFund := 3

	denom := assettypes.BuildFungibleTokenDenom(issueMsg.Subunit, issuer)
	bankSendSendMsgs := make([]sdk.Msg, 0, numAccountsToFund)
	coinToFund := sdk.NewCoin(denom, sdk.NewInt(10_000_000_000))
	fundedAccounts := make([]sdk.AccAddress, 0, numAccountsToFund)
	for i := 0; i < numAccountsToFund; i++ {
		recipient := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		fundedAccounts = append(fundedAccounts, recipient)
		bankSendSendMsgs = append(bankSendSendMsgs, &banktypes.MsgSend{
			FromAddress: issuer.String(),
			ToAddress:   recipient.String(),
			Amount:      sdk.NewCoins(coinToFund),
		})
	}

	fundMsgs := make([]sdk.Msg, 0)
	fundMsgs = append(fundMsgs, issueMsg)
	for i := 0; i < iterationsToFund; i++ {
		fundMsgs = append(fundMsgs, bankSendSendMsgs...)
	}
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, issuer, testing.BalancesOptions{
		Messages: fundMsgs,
	}))

	// issue fungible tokens
	_, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	// send coins in loop
	start := time.Now()
	for i := 0; i < iterationsToFund; i++ {
		res, err := tx.BroadcastTx(
			ctx,
			chain.ClientContext.WithFromAddress(issuer),
			chain.TxFactory().WithGas(chain.GasLimitByMsgs(bankSendSendMsgs...)),
			bankSendSendMsgs...)
		requireT.NoError(err)
		logger.Get(ctx).Info(fmt.Sprintf("Successfully sent batch BankSend tx, hash: %s, gasUse:%d", res.TxHash, res.GasUsed))
	}
	logger.Get(ctx).Info(fmt.Sprintf("It takes %s to fund %d accounts with BankSend", time.Since(start), numAccountsToFund*iterationsToFund))

	assertBatchAccounts(ctx, chain, sdk.NewCoins(sdk.NewCoin(coinToFund.Denom, coinToFund.Amount.MulRaw(int64(iterationsToFund)))), fundedAccounts, denom, requireT)
}

func assertBatchAccounts(
	ctx context.Context,
	chain testing.Chain,
	expectedCoins sdk.Coins,
	fundedAccounts []sdk.AccAddress,
	denom string,
	requireT *require.Assertions) {
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	for _, acc := range fundedAccounts {
		res, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
			Address: acc.String(),
			Denom:   denom,
		})
		requireT.NoError(err)
		requireT.Equal(expectedCoins.String(), res.Balance.String())
	}
}
