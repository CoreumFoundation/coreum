//go:build integrationtests

package modules

import (
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdksigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
)

var maxMemo = strings.Repeat("-", 256) // cosmos sdk is configured to accept maximum memo of 256 characters by default

// TestSendDeterministicGas checks that transfer takes the deterministic amount of gas
func TestSendDeterministicGas(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	sender := chain.GenAccount()
	recipient := chain.GenAccount()

	amountToSend := sdk.NewInt(1000)
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   amountToSend,
	}))

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(amountToSend)),
	}

	clientCtx := chain.ClientContext.WithFromAddress(sender)
	bankSendGas := chain.GasLimitByMsgs(&banktypes.MsgSend{})
	res, err := tx.BroadcastTx(
		ctx,
		clientCtx,
		chain.TxFactory().
			WithMemo(maxMemo). // memo is set to max length here to charge as much gas as possible
			WithGas(bankSendGas),
		msg)
	require.NoError(t, err)
	require.Equal(t, bankSendGas, uint64(res.GasUsed))
}

// TestSendDeterministicGasTwoBankSends checks that transfer takes the deterministic amount of gas
func TestSendDeterministicGasTwoBankSends(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	sender := chain.GenAccount()
	receiver1 := chain.GenAccount()
	receiver2 := chain.GenAccount()

	bankSend1 := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   receiver1.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdk.NewInt(1000))),
	}
	bankSend2 := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   receiver2.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdk.NewInt(1000))),
	}

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{bankSend1, bankSend2},
		Amount:   sdk.NewInt(2000),
	}))

	gasExpected := chain.GasLimitByMultiSendMsgs(&banktypes.MsgSend{}, &banktypes.MsgSend{})
	clientCtx := chain.ChainContext.ClientContext.WithFromAddress(sender)
	txf := chain.ChainContext.TxFactory().WithGas(gasExpected)
	result, err := tx.BroadcastTx(ctx, clientCtx, txf, bankSend1, bankSend2)
	require.NoError(t, err)
	require.EqualValues(t, gasExpected, uint64(result.GasUsed))
}

// TestSendDeterministicGasManyCoins checks that transfer takes the higher deterministic amount of gas when more coins are transferred
func TestSendDeterministicGasManyCoins(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	const numOfTokens = 3

	sender := chain.GenAccount()
	recipient := chain.GenAccount()
	deterministicGasConfig := chain.DeterministicGas()

	amountToSend := sdk.NewInt(1000)

	issueMsgs := make([]sdk.Msg, 0, numOfTokens)
	for i := 0; i < numOfTokens; i++ {
		issueMsgs = append(issueMsgs, &assettypes.MsgIssueFungibleToken{
			Issuer:        sender.String(),
			Symbol:        fmt.Sprintf("TOK%d", i),
			Subunit:       fmt.Sprintf("tok%d", i),
			Description:   fmt.Sprintf("TOK%d Description", i),
			Recipient:     sender.String(),
			InitialAmount: amountToSend,
		})
	}

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, integrationtests.BalancesOptions{
		Messages: append([]sdk.Msg{&banktypes.MsgSend{
			Amount: make(sdk.Coins, numOfTokens),
		}}, issueMsgs...),
	}))

	// Issue fungible tokens
	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsgs...)),
		issueMsgs...,
	)
	require.NoError(t, err)

	coinsToSend := sdk.NewCoins()

	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assettypes.EventFungibleTokenIssued](res.Events)
	require.NoError(t, err)
	require.Equal(t, numOfTokens, len(fungibleTokenIssuedEvts))

	for _, e := range fungibleTokenIssuedEvts {
		coinsToSend = coinsToSend.Add(sdk.NewCoin(e.Denom, amountToSend))
	}

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   recipient.String(),
		Amount:      coinsToSend,
	}

	clientCtx := chain.ClientContext.WithFromAddress(sender)
	zeroBankSendGas := chain.GasLimitByMsgs(&banktypes.MsgSend{})
	require.Equal(t, deterministicGasConfig.FixedGas+deterministicGasConfig.BankSendPerEntry, zeroBankSendGas)

	bankSendGas := chain.GasLimitByMsgs(msg)
	require.Equal(t, deterministicGasConfig.FixedGas+numOfTokens*deterministicGasConfig.BankSendPerEntry, bankSendGas)

	res, err = tx.BroadcastTx(
		ctx,
		clientCtx,
		chain.TxFactory().
			WithMemo(maxMemo). // memo is set to max length here to charge as much gas as possible
			WithGas(bankSendGas),
		msg)
	require.NoError(t, err)
	require.Equal(t, bankSendGas, uint64(res.GasUsed))
}

// TestSendFailsIfNotEnoughGasIsProvided checks that transfer fails if not enough gas is provided
func TestSendFailsIfNotEnoughGasIsProvided(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	sender := chain.GenAccount()

	amountToSend := sdk.NewInt(1000)
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   amountToSend,
	}))

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   sender.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(amountToSend)),
	}

	clientCtx := chain.ClientContext.WithFromAddress(sender)
	bankSendGas := chain.GasLimitByMsgs(&banktypes.MsgSend{})
	_, err := tx.BroadcastTx(
		ctx,
		clientCtx,
		chain.TxFactory().
			WithGas(bankSendGas-1), // gas less than expected
		msg)

	require.True(t, cosmoserrors.ErrOutOfGas.Is(err))
}

// TestSendGasEstimation checks that gas is correctly estimated for send message
func TestSendGasEstimation(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	sender := chain.GenAccount()

	amountToSend := sdk.NewInt(1000)
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   amountToSend,
	}))

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   sender.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(amountToSend)),
	}

	clientCtx := chain.ClientContext.WithFromAddress(sender)
	bankSendGas := chain.GasLimitByMsgs(&banktypes.MsgSend{})
	_, estimatedGas, err := tx.CalculateGas(
		ctx,
		clientCtx,
		chain.TxFactory().
			WithGas(bankSendGas),
		msg)
	require.NoError(t, err)
	assert.Equal(t, bankSendGas, estimatedGas)
}

// TestMultiSendDeterministicGasManyCoins checks that transfer takes the higher deterministic amount of gas when more coins are transferred
func TestMultiSendDeterministicGasManyCoins(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	const numOfTokens = 3

	sender := chain.GenAccount()
	recipient := chain.GenAccount()
	deterministicGasConfig := chain.DeterministicGas()

	amountToSend := sdk.NewInt(1000)

	issueMsgs := make([]sdk.Msg, 0, numOfTokens)
	for i := 0; i < numOfTokens; i++ {
		issueMsgs = append(issueMsgs, &assettypes.MsgIssueFungibleToken{
			Issuer:        sender.String(),
			Symbol:        fmt.Sprintf("TOK%d", i),
			Subunit:       fmt.Sprintf("tok%d", i),
			Description:   fmt.Sprintf("TOK%d Description", i),
			Recipient:     sender.String(),
			InitialAmount: amountToSend,
		})
	}

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, integrationtests.BalancesOptions{
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
	}))

	// Issue fungible tokens
	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsgs...)),
		issueMsgs...,
	)
	require.NoError(t, err)

	coinsToSend := sdk.NewCoins()

	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assettypes.EventFungibleTokenIssued](res.Events)
	require.NoError(t, err)
	require.Equal(t, numOfTokens, len(fungibleTokenIssuedEvts))

	for _, e := range fungibleTokenIssuedEvts {
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

	zeroBankMultiSendGas := chain.GasLimitByMsgs(&banktypes.MsgMultiSend{})
	require.Equal(t, deterministicGasConfig.FixedGas+deterministicGasConfig.BankMultiSendPerEntry, zeroBankMultiSendGas)

	bankMultiSendGas := chain.GasLimitByMsgs(msg)
	require.Equal(t, deterministicGasConfig.FixedGas+numOfTokens*deterministicGasConfig.BankMultiSendPerEntry, bankMultiSendGas)

	res, err = tx.BroadcastTx(
		ctx,
		clientCtx,
		chain.TxFactory().
			WithMemo(maxMemo). // memo is set to max length here to charge as much gas as possible
			WithGas(bankMultiSendGas),
		msg)
	require.NoError(t, err)
	require.Equal(t, bankMultiSendGas, uint64(res.GasUsed))
}

// TestMultiSend tests MultiSend message
func TestMultiSend(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	sender := chain.GenAccount()
	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	amount := sdk.NewInt(1000)

	issueMsgs := []sdk.Msg{
		&assettypes.MsgIssueFungibleToken{
			Issuer:        sender.String(),
			Symbol:        "TOK1",
			Subunit:       "tok1",
			Description:   "TOK1 Description",
			Recipient:     sender.String(),
			InitialAmount: amount,
		},
		&assettypes.MsgIssueFungibleToken{
			Issuer:        sender.String(),
			Symbol:        "TOK2",
			Subunit:       "tok2",
			Description:   "TOK2 Description",
			Recipient:     sender.String(),
			InitialAmount: amount,
		},
	}

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, integrationtests.BalancesOptions{
		Messages: append([]sdk.Msg{&banktypes.MsgMultiSend{Outputs: []banktypes.Output{
			{Coins: make(sdk.Coins, 2)},
			{Coins: make(sdk.Coins, 2)},
		}}}, issueMsgs...),
	}))

	// Issue fungible tokens
	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsgs...)),
		issueMsgs...,
	)
	require.NoError(t, err)

	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assettypes.EventFungibleTokenIssued](res.Events)
	require.NoError(t, err)
	require.Equal(t, len(issueMsgs), len(fungibleTokenIssuedEvts))

	denom1 := fungibleTokenIssuedEvts[0].Denom
	denom2 := fungibleTokenIssuedEvts[1].Denom

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
	res, err = tx.BroadcastTx(
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

// TestMultiSendFromMultipleAccounts tests MultiSend message form multiple accounts.
func TestMultiSendFromMultipleAccounts(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

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
	issue1Msg := &assettypes.MsgIssueFungibleToken{
		Issuer:        sender1.String(),
		Symbol:        "TOK1",
		Subunit:       "tok1",
		Description:   "TOK1 Description",
		Recipient:     sender1.String(),
		InitialAmount: assetAmount,
	}
	issue2Msg := &assettypes.MsgIssueFungibleToken{
		Issuer:        sender2.String(),
		Symbol:        "TOK2",
		Subunit:       "tok2",
		Description:   "TOK2 Description",
		Recipient:     sender2.String(),
		InitialAmount: assetAmount,
	}

	denom1 := assettypes.BuildFungibleTokenDenom(issue1Msg.Subunit, sender1)
	denom2 := assettypes.BuildFungibleTokenDenom(issue2Msg.Subunit, sender2)

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

	// fund accounts
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, sender1, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			multiSendMsg,
			issue1Msg,
		},
		Amount: nativeAmountToSend.Amount,
	}))
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, sender2, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{issue2Msg},
	}))

	// issue first fungible token
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issue1Msg)),
		issue1Msg,
	)
	requireT.NoError(err)
	// issue second fungible token
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender2),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issue2Msg)),
		issue2Msg,
	)
	requireT.NoError(err)

	// create MultiSend tx message and sign it from 2 accounts
	sender1AccInfo, err := tx.GetAccountInfo(ctx, chain.ClientContext, sender1)
	requireT.NoError(err)

	// set sender1 params for the signature
	txF := chain.TxFactory().
		WithAccountNumber(sender1AccInfo.GetAccountNumber()).
		WithSequence(sender1AccInfo.GetSequence()).
		WithGas(chain.GasLimitByMsgs(multiSendMsg)).
		WithSignMode(sdksigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON) //nolint:nosnakecase // the sdk constant

	txBuilder, err := txF.BuildUnsignedTx(multiSendMsg)
	requireT.NoError(err)

	// sign from sender1
	err = tx.Sign(txF, sender1KeyInfo.GetName(), txBuilder, false)
	requireT.NoError(err)

	sender2AccInfo, err := tx.GetAccountInfo(ctx, chain.ClientContext, sender2)
	requireT.NoError(err)

	// set sender2 params for the signature
	txF = chain.TxFactory().
		WithAccountNumber(sender2AccInfo.GetAccountNumber()).
		WithSequence(sender2AccInfo.GetSequence()).
		WithGas(chain.GasLimitByMsgs(multiSendMsg)).
		WithSignMode(sdksigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON) //nolint:nosnakecase // the sdk constant

	// sign from sender2
	err = tx.Sign(txF, sender2KeyInfo.GetName(), txBuilder, false)
	requireT.NoError(err)

	// encode tx and broadcast
	encodedMultiSendTx, err := chain.ClientContext.TxConfig().TxEncoder()(txBuilder.GetTx())
	requireT.NoError(err)
	_, err = tx.BroadcastRawTx(
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

// TestCoreSend checks that core is transferred correctly between wallets
func TestCoreSend(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	sender := chain.GenAccount()
	recipient := chain.GenAccount()

	senderInitialAmount := sdk.NewInt(100)
	recipientInitialAmount := sdk.NewInt(10)
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   senderInitialAmount,
	}))
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, recipient, integrationtests.BalancesOptions{
		Amount: recipientInitialAmount,
	}))

	// transfer tokens from sender to recipient
	amountToSend := sdk.NewInt(10)
	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(amountToSend)),
	}

	result, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)
	require.NoError(t, err)

	logger.Get(ctx).Info("Transfer executed", zap.String("txHash", result.TxHash))

	// Query wallets for current balance
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	balancesSender, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: sender.String(),
		Denom:   chain.NetworkConfig.Denom,
	})
	require.NoError(t, err)

	balancesRecipient, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   chain.NetworkConfig.Denom,
	})
	require.NoError(t, err)

	assert.Equal(t, senderInitialAmount.Sub(amountToSend).String(), balancesSender.Balance.Amount.String())
	assert.Equal(t, recipientInitialAmount.Add(amountToSend).String(), balancesRecipient.Balance.Amount.String())
}
