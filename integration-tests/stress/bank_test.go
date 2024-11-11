//go:build integrationtests

package stress

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v5/integration-tests"
	"github.com/CoreumFoundation/coreum/v5/pkg/client"
	"github.com/CoreumFoundation/coreum/v5/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
)

// TestBankMultiSendBatchOutputs tests MultiSend message with maximum amount of addresses.
func TestBankMultiSendBatchOutputs(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issuer := chain.GenAccount()
	requireT := require.New(t)

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "TOK1",
		Subunit:       "tok1",
		Precision:     1,
		Description:   "TOK1 Description",
		InitialAmount: sdkmath.NewInt(100_000_000_000_000_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_freezing, // enable the feature to make the computation more complicated
		},
	}

	numAccountsToFund := 20 // 1700 was the max accepted
	iterationsToFund := 2

	inputItem := banktypes.Input{
		Address: issuer.String(),
		Coins:   sdk.NewCoins(),
	}
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)
	outputItems := make([]banktypes.Output, 0, numAccountsToFund)
	fundedAccounts := make([]sdk.AccAddress, 0, numAccountsToFund)
	coinToFund := sdk.NewCoin(denom, sdkmath.NewInt(10_000_000_000))

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

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages:                    append([]sdk.Msg{issueMsg}, multiSendMsgs...),
		NondeterministicMessagesGas: 50_000_000, // to cover extra bytes because of the message size
		Amount:                      chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	// issue fungible tokens
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	// send coins in loop
	start := time.Now()
	for _, msg := range multiSendMsgs {
		res, err := client.BroadcastTx(
			ctx,
			chain.ClientContext.WithFromAddress(issuer),
			// we estimate here since the th size is grater then allowed for the deterministic fee
			chain.TxFactoryAuto(),
			msg,
		)
		requireT.NoError(err)
		t.Logf("Successfully sent batch MultiSend tx, hash: %s, gasUse:%d", res.TxHash, res.GasUsed)
	}
	t.Logf("It takes %s to fund %d accounts with MultiSend", time.Since(start), numAccountsToFund*iterationsToFund)

	assertBatchAccounts(
		ctx,
		chain,
		sdk.NewCoins(sdk.NewCoin(coinToFund.Denom, coinToFund.Amount.MulRaw(int64(iterationsToFund)))),
		fundedAccounts,
		denom,
		requireT,
	)
}

// TestBankSendBatchMsgs tests BankSend message with maximum amount of accounts.
func TestBankSendBatchMsgs(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issuer := chain.GenAccount()
	requireT := require.New(t)

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "TOK1",
		Subunit:       "tok1",
		Precision:     1,
		Description:   "TOK1 Description",
		InitialAmount: sdkmath.NewInt(100_000_000_000_000_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_freezing, // enable the feature to make the computation more complicated
		},
	}

	numAccountsToFund := 20 // 600 was the max accepted
	iterationsToFund := 3

	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)
	bankSendSendMsgs := make([]sdk.Msg, 0, numAccountsToFund)
	coinToFund := sdk.NewCoin(denom, sdkmath.NewInt(10_000_000_000))
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
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: fundMsgs,
		Amount:   chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	// issue fungible tokens
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	// send coins in loop
	start := time.Now()
	for i := 0; i < iterationsToFund; i++ {
		res, err := client.BroadcastTx(
			ctx,
			chain.ClientContext.WithFromAddress(issuer),
			chain.TxFactory().WithGas(chain.GasLimitByMsgs(bankSendSendMsgs...)),
			bankSendSendMsgs...)
		requireT.NoError(err)
		t.Logf("Successfully sent batch BankSend tx, hash: %s, gasUsed:%d", res.TxHash, res.GasUsed)
	}
	t.Logf("It takes %s to fund %d accounts with BankSend", time.Since(start), numAccountsToFund*iterationsToFund)

	assertBatchAccounts(
		ctx,
		chain,
		sdk.NewCoins(sdk.NewCoin(coinToFund.Denom, coinToFund.Amount.MulRaw(int64(iterationsToFund)))),
		fundedAccounts,
		denom,
		requireT,
	)
}

func assertBatchAccounts(
	ctx context.Context,
	chain integration.CoreumChain,
	expectedCoins sdk.Coins,
	fundedAccounts []sdk.AccAddress,
	denom string,
	requireT *require.Assertions,
) {
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
