//go:build integrationtests

package modules

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdksigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

var maxMemo = strings.Repeat("-", 256) // cosmos sdk is configured to accept maximum memo of 256 characters by default

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
		InitialAmount: sdk.NewInt(100_000_000_000_000_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_freezing, // enable the feature to make the computation more complicated
		},
	}

	numAccountsToFund := 1000 // 1700 was the max accepted
	iterationsToFund := 2

	inputItem := banktypes.Input{
		Address: issuer.String(),
		Coins:   sdk.NewCoins(),
	}
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)
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

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages:                    append([]sdk.Msg{issueMsg}, multiSendMsgs...),
		NondeterministicMessagesGas: 10_000_000, // to cover extra bytes because of the message size
		Amount:                      getIssueFee(ctx, t, chain.ClientContext).Amount,
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
			chain.TxFactory().WithSimulateAndExecute(true),
			msg,
		)
		requireT.NoError(err)
		t.Logf("Successfully sent batch MultiSend tx, hash: %s, gasUse:%d", res.TxHash, res.GasUsed)
	}
	t.Logf("It takes %s to fund %d accounts with MultiSend", time.Since(start), numAccountsToFund*iterationsToFund)

	assertBatchAccounts(ctx, chain, sdk.NewCoins(sdk.NewCoin(coinToFund.Denom, coinToFund.Amount.MulRaw(int64(iterationsToFund)))), fundedAccounts, denom, requireT)
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
		InitialAmount: sdk.NewInt(100_000_000_000_000_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_freezing, // enable the feature to make the computation more complicated
		},
	}

	numAccountsToFund := 400 // 600 was the max accepted
	iterationsToFund := 3

	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)
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
	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: fundMsgs,
		Amount:   getIssueFee(ctx, t, chain.ClientContext).Amount,
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
		t.Logf("Successfully sent batch BankSend tx, hash: %s, gasUse:%d", res.TxHash, res.GasUsed)
	}
	t.Logf("It takes %s to fund %d accounts with BankSend", time.Since(start), numAccountsToFund*iterationsToFund)

	assertBatchAccounts(ctx, chain, sdk.NewCoins(sdk.NewCoin(coinToFund.Denom, coinToFund.Amount.MulRaw(int64(iterationsToFund)))), fundedAccounts, denom, requireT)
}

// TestBankSendDeterministicGas checks that transfer takes the deterministic amount of gas.
func TestBankSendDeterministicGas(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	sender := chain.GenAccount()
	recipient := chain.GenAccount()

	amountToSend := sdk.NewInt(1000)
	chain.FundAccountsWithOptions(ctx, t, sender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   amountToSend,
	})

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(amountToSend)),
	}

	clientCtx := chain.ClientContext.WithFromAddress(sender)
	bankSendGas := chain.GasLimitByMsgs(&banktypes.MsgSend{})
	res, err := client.BroadcastTx(
		ctx,
		clientCtx,
		chain.TxFactory().
			WithMemo(maxMemo). // memo is set to max length here to charge as much gas as possible
			WithGas(bankSendGas),
		msg)
	require.NoError(t, err)
	require.Equal(t, bankSendGas, uint64(res.GasUsed))
}

// TestBankSendDeterministicGasTwoBankSends checks that transfer takes the deterministic amount of gas.
func TestBankSendDeterministicGasTwoBankSends(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	sender := chain.GenAccount()
	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	bankSend1 := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   recipient1.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdk.NewInt(1000))),
	}
	bankSend2 := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   recipient2.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdk.NewInt(1000))),
	}

	chain.FundAccountsWithOptions(ctx, t, sender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{bankSend1, bankSend2},
		Amount:   sdk.NewInt(2000),
	})

	gasExpected := chain.GasLimitByMultiSendMsgs(&banktypes.MsgSend{}, &banktypes.MsgSend{})
	clientCtx := chain.ChainContext.ClientContext.WithFromAddress(sender)
	txf := chain.ChainContext.TxFactory().WithGas(gasExpected)
	result, err := client.BroadcastTx(ctx, clientCtx, txf, bankSend1, bankSend2)
	require.NoError(t, err)
	require.EqualValues(t, gasExpected, uint64(result.GasUsed))
}

// TestBankSendDeterministicGasManyCoins checks that transfer takes the higher deterministic amount of gas when more coins are transferred.
func TestBankSendDeterministicGasManyCoins(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	const numOfTokens = 3

	sender := chain.GenAccount()
	recipient := chain.GenAccount()

	amountToSend := sdk.NewInt(1000)

	issueMsgs := make([]sdk.Msg, 0, numOfTokens)
	for i := 0; i < numOfTokens; i++ {
		issueMsgs = append(issueMsgs, &assetfttypes.MsgIssue{
			Issuer:        sender.String(),
			Symbol:        fmt.Sprintf("TOK%d", i),
			Subunit:       fmt.Sprintf("tok%d", i),
			Precision:     1,
			Description:   fmt.Sprintf("TOK%d Description", i),
			InitialAmount: amountToSend,
		})
	}

	chain.FundAccountsWithOptions(ctx, t, sender, integrationtests.BalancesOptions{
		Messages: append([]sdk.Msg{&banktypes.MsgSend{
			Amount: make(sdk.Coins, numOfTokens),
		}}, issueMsgs...),
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount.MulRaw(numOfTokens),
	})

	// Issue fungible tokens
	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsgs...)),
		issueMsgs...,
	)
	require.NoError(t, err)

	coinsToSend := sdk.NewCoins()

	tokenIssuedEvts, err := event.FindTypedEvents[*assetfttypes.EventIssued](res.Events)
	require.NoError(t, err)
	require.Equal(t, numOfTokens, len(tokenIssuedEvts))

	for _, e := range tokenIssuedEvts {
		coinsToSend = coinsToSend.Add(sdk.NewCoin(e.Denom, amountToSend))
	}

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   recipient.String(),
		Amount:      coinsToSend,
	}

	clientCtx := chain.ClientContext.WithFromAddress(sender)

	bankSendGas := chain.GasLimitByMsgs(msg)
	msgGas, ok := chain.DeterministicGasConfig.GasRequiredByMessage(msg)
	require.True(t, ok)
	require.Equal(t, chain.DeterministicGasConfig.FixedGas+msgGas, bankSendGas)

	res, err = client.BroadcastTx(
		ctx,
		clientCtx,
		chain.TxFactory().
			WithMemo(maxMemo). // memo is set to max length here to charge as much gas as possible
			WithGas(bankSendGas),
		msg)
	require.NoError(t, err)
	require.Equal(t, bankSendGas, uint64(res.GasUsed))
}

// TestBankSendFailsIfNotEnoughGasIsProvided checks that transfer fails if not enough gas is provided.
func TestBankSendFailsIfNotEnoughGasIsProvided(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	sender := chain.GenAccount()

	amountToSend := sdk.NewInt(1000)
	chain.FundAccountsWithOptions(ctx, t, sender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   amountToSend,
	})

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   sender.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(amountToSend)),
	}

	clientCtx := chain.ClientContext.WithFromAddress(sender)
	bankSendGas := chain.GasLimitByMsgs(&banktypes.MsgSend{})
	_, err := client.BroadcastTx(
		ctx,
		clientCtx,
		chain.TxFactory().
			WithGas(bankSendGas-1), // gas less than expected
		msg)

	require.True(t, cosmoserrors.ErrOutOfGas.Is(err))
}

// TestBankSendGasEstimation checks that gas is correctly estimated for send message.
func TestBankSendGasEstimation(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	sender := chain.GenAccount()

	amountToSend := sdk.NewInt(1000)
	chain.FundAccountsWithOptions(ctx, t, sender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   amountToSend,
	})

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   sender.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(amountToSend)),
	}

	clientCtx := chain.ClientContext.WithFromAddress(sender)
	bankSendGas := chain.GasLimitByMsgs(&banktypes.MsgSend{})
	_, estimatedGas, err := client.CalculateGas(
		ctx,
		clientCtx,
		chain.TxFactory().
			WithGas(bankSendGas),
		msg)
	require.NoError(t, err)
	assert.Equal(t, bankSendGas, estimatedGas)
}

// TestBankMultiSendDeterministicGasManyCoins checks that transfer takes the higher deterministic amount of gas when more coins are transferred.
func TestBankMultiSendDeterministicGasManyCoins(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	const numOfTokens = 3

	sender := chain.GenAccount()
	recipient := chain.GenAccount()

	amountToSend := sdk.NewInt(1000)

	issueMsgs := make([]sdk.Msg, 0, numOfTokens)
	for i := 0; i < numOfTokens; i++ {
		issueMsgs = append(issueMsgs, &assetfttypes.MsgIssue{
			Issuer:        sender.String(),
			Symbol:        fmt.Sprintf("TOK%d", i),
			Subunit:       fmt.Sprintf("tok%d", i),
			Description:   fmt.Sprintf("TOK%d Description", i),
			Precision:     1,
			InitialAmount: amountToSend,
		})
	}

	chain.FundAccountsWithOptions(ctx, t, sender, integrationtests.BalancesOptions{
		Messages: append([]sdk.Msg{&banktypes.MsgMultiSend{
			Inputs: []banktypes.Input{
				{
					Coins: make(sdk.Coins, numOfTokens),
				},
			},
			Outputs: []banktypes.Output{
				{
					Coins: make(sdk.Coins, numOfTokens),
				},
			},
		}}, issueMsgs...),
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount.MulRaw(numOfTokens),
	})

	// Issue fungible tokens
	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsgs...)),
		issueMsgs...,
	)
	require.NoError(t, err)

	coinsToSend := sdk.NewCoins()

	tokenIssuedEvts, err := event.FindTypedEvents[*assetfttypes.EventIssued](res.Events)
	require.NoError(t, err)
	require.Equal(t, numOfTokens, len(tokenIssuedEvts))

	for _, e := range tokenIssuedEvts {
		coinsToSend = coinsToSend.Add(sdk.NewCoin(e.Denom, amountToSend))
	}

	msg := &banktypes.MsgMultiSend{
		Inputs: []banktypes.Input{
			{
				Address: sender.String(),
				Coins:   coinsToSend,
			},
		},
		Outputs: []banktypes.Output{
			{
				Address: recipient.String(),
				Coins:   coinsToSend,
			},
		},
	}

	clientCtx := chain.ClientContext.WithFromAddress(sender)
	bankMultiSendGas := chain.GasLimitByMsgs(msg)

	res, err = client.BroadcastTx(
		ctx,
		clientCtx,
		chain.TxFactory().
			WithMemo(maxMemo). // memo is set to max length here to charge as much gas as possible
			WithGas(bankMultiSendGas),
		msg)
	require.NoError(t, err)
	require.Equal(t, bankMultiSendGas, uint64(res.GasUsed))
}

// TestBankMultiSend tests MultiSend message.
//
//nolint:funlen // there are many tests
func TestBankMultiSend(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	sender := chain.GenAccount()
	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	amount := sdk.NewInt(1000)

	issueMsgs := []sdk.Msg{
		&assetfttypes.MsgIssue{
			Issuer:        sender.String(),
			Symbol:        "TOK1",
			Subunit:       "tok1",
			Precision:     1,
			Description:   "TOK1 Description",
			InitialAmount: amount,
		},
		&assetfttypes.MsgIssue{
			Issuer:        sender.String(),
			Symbol:        "TOK2",
			Subunit:       "tok2",
			Precision:     1,
			Description:   "TOK2 Description",
			InitialAmount: amount,
		},
	}

	chain.FundAccountsWithOptions(ctx, t, sender, integrationtests.BalancesOptions{
		Messages: append([]sdk.Msg{&banktypes.MsgMultiSend{
			Inputs: []banktypes.Input{
				{Coins: make(sdk.Coins, 2)},
			},
			Outputs: []banktypes.Output{
				{Coins: make(sdk.Coins, 2)},
				{Coins: make(sdk.Coins, 2)},
			},
		}}, issueMsgs...),
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount.MulRaw(int64(len(issueMsgs))),
	})

	// Issue fungible tokens
	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsgs...)),
		issueMsgs...,
	)
	require.NoError(t, err)

	tokenIssuedEvts, err := event.FindTypedEvents[*assetfttypes.EventIssued](res.Events)
	require.NoError(t, err)
	require.Equal(t, len(issueMsgs), len(tokenIssuedEvts))

	denom1 := tokenIssuedEvts[0].Denom
	denom2 := tokenIssuedEvts[1].Denom

	msg := &banktypes.MsgMultiSend{
		Inputs: []banktypes.Input{
			{
				Address: sender.String(),
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin(denom1, 1000),
					sdk.NewInt64Coin(denom2, 1000),
				),
			},
		},
		Outputs: []banktypes.Output{
			{
				Address: recipient1.String(),
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin(denom1, 600),
					sdk.NewInt64Coin(denom2, 400),
				),
			},
			{
				Address: recipient2.String(),
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin(denom1, 400),
					sdk.NewInt64Coin(denom2, 600),
				),
			},
		},
	}

	clientCtx := chain.ClientContext.WithFromAddress(sender)
	bankMultiSendGas := chain.GasLimitByMsgs(msg)
	res, err = client.BroadcastTx(
		ctx,
		clientCtx,
		chain.TxFactory().
			WithMemo(maxMemo). // memo is set to max length here to charge as much gas as possible
			WithGas(bankMultiSendGas),
		msg)
	require.NoError(t, err)
	require.Equal(t, bankMultiSendGas, uint64(res.GasUsed))

	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	qres, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: sender.String()})
	require.NoError(t, err)
	require.Empty(t, qres.Balances)

	recipient1AllBalancesRes, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: recipient1.String()})
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(denom1, 600), sdk.NewInt64Coin(denom2, 400)), recipient1AllBalancesRes.Balances)

	recipient2AllBalancesRes, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: recipient2.String()})
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(denom1, 400), sdk.NewInt64Coin(denom2, 600)), recipient2AllBalancesRes.Balances)
}

// TestBankMultiSendFromMultipleAccounts tests MultiSend message form multiple accounts.
//
//nolint:funlen // there are many tests
func TestBankMultiSendFromMultipleAccounts(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	sender1 := chain.GenAccount()
	sender1KeyInfo, err := chain.ClientContext.Keyring().KeyByAddress(sender1)
	requireT.NoError(err)

	sender2 := chain.GenAccount()
	sender2KeyInfo, err := chain.ClientContext.Keyring().KeyByAddress(sender2)
	requireT.NoError(err)

	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()
	recipient3 := chain.GenAccount()

	assetAmount := sdk.NewInt(1000)
	issue1Msg := &assetfttypes.MsgIssue{
		Issuer:        sender1.String(),
		Symbol:        "TOK1",
		Subunit:       "tok1",
		Precision:     1,
		Description:   "TOK1 Description",
		InitialAmount: assetAmount,
	}
	issue2Msg := &assetfttypes.MsgIssue{
		Issuer:        sender2.String(),
		Symbol:        "TOK2",
		Subunit:       "tok2",
		Precision:     1,
		Description:   "TOK2 Description",
		InitialAmount: assetAmount,
	}

	denom1 := assetfttypes.BuildDenom(issue1Msg.Subunit, sender1)
	denom2 := assetfttypes.BuildDenom(issue2Msg.Subunit, sender2)

	nativeAmountToSend := chain.NewCoin(sdk.NewInt(100))

	// define the message to send from multiple accounts to multiple
	multiSendMsg := &banktypes.MsgMultiSend{
		Inputs: []banktypes.Input{
			{
				Address: sender1.String(),
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin(denom1, 1000),
					chain.NewCoin(sdk.NewInt(100)),
				),
			},
			{
				Address: sender2.String(),
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin(denom2, 1000),
				),
			},
		},
		Outputs: []banktypes.Output{
			{
				Address: recipient1.String(),
				Coins: sdk.NewCoins(
					chain.NewCoin(sdk.NewInt(30)),
					sdk.NewInt64Coin(denom1, 600),
					sdk.NewInt64Coin(denom2, 400),
				),
			},
			{
				Address: recipient2.String(),
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin(denom1, 400),
					sdk.NewInt64Coin(denom2, 600),
				),
			},
			{
				Address: recipient3.String(),
				Coins: sdk.NewCoins(
					chain.NewCoin(sdk.NewInt(70)),
				),
			},
		},
	}

	issueFee := getIssueFee(ctx, t, chain.ClientContext).Amount

	// fund accounts
	chain.FundAccountsWithOptions(ctx, t, sender1, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			multiSendMsg,
			issue1Msg,
		},
		Amount: issueFee.Add(nativeAmountToSend.Amount),
	})
	chain.FundAccountsWithOptions(ctx, t, sender2, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{issue2Msg},
		Amount:   issueFee,
	})

	// issue first fungible token
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issue1Msg)),
		issue1Msg,
	)
	requireT.NoError(err)
	// issue second fungible token
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender2),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issue2Msg)),
		issue2Msg,
	)
	requireT.NoError(err)

	// create MultiSend tx message and sign it from 2 accounts
	sender1AccInfo, err := client.GetAccountInfo(ctx, chain.ClientContext, sender1)
	requireT.NoError(err)

	// set sender1 params for the signature
	txF := chain.TxFactory().
		WithAccountNumber(sender1AccInfo.GetAccountNumber()).
		WithSequence(sender1AccInfo.GetSequence()).
		WithGas(chain.GasLimitByMsgs(multiSendMsg)).
		WithSignMode(sdksigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)

	txBuilder, err := txF.BuildUnsignedTx(multiSendMsg)
	requireT.NoError(err)

	// sign from sender1
	err = client.Sign(txF, sender1KeyInfo.GetName(), txBuilder, false)
	requireT.NoError(err)

	sender2AccInfo, err := client.GetAccountInfo(ctx, chain.ClientContext, sender2)
	requireT.NoError(err)

	// set sender2 params for the signature
	txF = chain.TxFactory().
		WithAccountNumber(sender2AccInfo.GetAccountNumber()).
		WithSequence(sender2AccInfo.GetSequence()).
		WithGas(chain.GasLimitByMsgs(multiSendMsg)).
		WithSignMode(sdksigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)

	// sign from sender2
	err = client.Sign(txF, sender2KeyInfo.GetName(), txBuilder, false)
	requireT.NoError(err)

	// encode tx and broadcast
	encodedMultiSendTx, err := chain.ClientContext.TxConfig().TxEncoder()(txBuilder.GetTx())
	requireT.NoError(err)
	_, err = client.BroadcastRawTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender1),
		encodedMultiSendTx)
	requireT.NoError(err)

	// check the received balances
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	for _, output := range multiSendMsg.Outputs {
		res, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: output.Address})
		requireT.NoError(err)
		requireT.Equal(output.Coins, res.Balances)
	}
}

// FIXME (wojtek): add test verifying that transfer fails if sender is out of balance.

// TestBankCoreSend checks that core is transferred correctly between wallets.
func TestBankCoreSend(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	sender := chain.GenAccount()
	recipient := chain.GenAccount()

	senderInitialAmount := sdk.NewInt(100)
	recipientInitialAmount := sdk.NewInt(10)
	chain.FundAccountsWithOptions(ctx, t, sender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   senderInitialAmount,
	})
	chain.FundAccountsWithOptions(ctx, t, recipient, integrationtests.BalancesOptions{
		Amount: recipientInitialAmount,
	})

	// transfer tokens from sender to recipient
	amountToSend := sdk.NewInt(10)
	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(amountToSend)),
	}

	result, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)
	require.NoError(t, err)

	t.Logf("Transfer executed, txHash:%s", result.TxHash)

	// Query wallets for current balance
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	balancesSender, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: sender.String(),
		Denom:   chain.ChainSettings.Denom,
	})
	require.NoError(t, err)

	balancesRecipient, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   chain.ChainSettings.Denom,
	})
	require.NoError(t, err)

	assert.Equal(t, senderInitialAmount.Sub(amountToSend).String(), balancesSender.Balance.Amount.String())
	assert.Equal(t, recipientInitialAmount.Add(amountToSend).String(), balancesRecipient.Balance.Amount.String())
}

func assertBatchAccounts(
	ctx context.Context,
	chain integrationtests.CoreumChain,
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
