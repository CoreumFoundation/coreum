package bank

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
)

var maxMemo = strings.Repeat("-", 256) // cosmos sdk is configured to accept maximum memo of 256 characters by default

// TestSendDeterministicGas checks that transfer takes the deterministic amount of gas
func TestSendDeterministicGas(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := chain.GenAccount()
	recipient := chain.GenAccount()

	amountToSend := sdk.NewInt(1000)
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, testing.BalancesOptions{
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
func TestSendDeterministicGasTwoBankSends(ctx context.Context, t testing.T, chain testing.Chain) {
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

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, testing.BalancesOptions{
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

// TestSendDeterministicGasManyCoins checks that transfer takes the minimum deterministic amount of gas up to the limit of number of coins transferred
func TestSendDeterministicGasManyCoins(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := chain.GenAccount()
	recipient := chain.GenAccount()
	deterministicGasConfig := chain.DeterministicGas()

	amountToSend := sdk.NewInt(1000)

	issueMsgs := make([]sdk.Msg, 0, deterministicGasConfig.BankSendTokenNumberLimit)
	for i := 0; i < deterministicGasConfig.BankSendTokenNumberLimit; i++ {
		issueMsgs = append(issueMsgs, &assettypes.MsgIssueFungibleToken{
			Issuer:        sender.String(),
			Symbol:        fmt.Sprintf("TOK%d", i),
			Description:   fmt.Sprintf("TOK%d Description", i),
			Recipient:     sender.String(),
			InitialAmount: amountToSend,
		})
	}

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, testing.BalancesOptions{
		Messages: append([]sdk.Msg{&banktypes.MsgSend{}}, issueMsgs...),
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
	require.Equal(t, deterministicGasConfig.BankSendTokenNumberLimit, len(fungibleTokenIssuedEvts))

	for _, e := range fungibleTokenIssuedEvts {
		coinsToSend = coinsToSend.Add(sdk.NewCoin(e.Denom, amountToSend))
	}

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   recipient.String(),
		Amount:      coinsToSend,
	}

	clientCtx := chain.ClientContext.WithFromAddress(sender)
	bankSendGas := chain.GasLimitByMsgs(&banktypes.MsgSend{})
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

// TestSendDeterministicGasMoreCoins checks that transfer takes the higher deterministic amount of gas above the limit of number of coins transferred
func TestSendDeterministicGasMoreCoins(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := chain.GenAccount()
	recipient := chain.GenAccount()
	deterministicGasConfig := chain.DeterministicGas()

	amountToSend := sdk.NewInt(1000)

	numOfTokens := deterministicGasConfig.BankSendTokenNumberLimit + 3
	issueMsgs := make([]sdk.Msg, 0, numOfTokens)
	for i := 0; i < numOfTokens; i++ {
		issueMsgs = append(issueMsgs, &assettypes.MsgIssueFungibleToken{
			Issuer:        sender.String(),
			Symbol:        fmt.Sprintf("TOK%d", i),
			Description:   fmt.Sprintf("TOK%d Description", i),
			Recipient:     sender.String(),
			InitialAmount: amountToSend,
		})
	}

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, testing.BalancesOptions{
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
	minBankSendGas := chain.GasLimitByMsgs(&banktypes.MsgSend{})
	bankSendGas := chain.GasLimitByMsgs(msg)
	res, err = tx.BroadcastTx(
		ctx,
		clientCtx,
		chain.TxFactory().
			WithMemo(maxMemo). // memo is set to max length here to charge as much gas as possible
			WithGas(bankSendGas),
		msg)
	require.NoError(t, err)
	require.Equal(t, bankSendGas, uint64(res.GasUsed))
	require.Equal(t, minBankSendGas+3*deterministicGasConfig.BankSendAdditionalTransfer, bankSendGas)
}

// TestSendFailsIfNotEnoughGasIsProvided checks that transfer fails if not enough gas is provided
func TestSendFailsIfNotEnoughGasIsProvided(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := chain.GenAccount()

	amountToSend := sdk.NewInt(1000)
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, testing.BalancesOptions{
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
func TestSendGasEstimation(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := chain.GenAccount()

	amountToSend := sdk.NewInt(1000)
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, testing.BalancesOptions{
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

// TestMultiSendDeterministicGasManyCoins checks that transfer takes the minimum deterministic amount of gas up to the limit of number of coins transferred
func TestMultiSendDeterministicGasManyCoins(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := chain.GenAccount()
	recipient := chain.GenAccount()
	deterministicGasConfig := chain.DeterministicGas()

	amountToSend := sdk.NewInt(1000)

	issueMsgs := make([]sdk.Msg, 0, deterministicGasConfig.BankSendTokenNumberLimit)
	for i := 0; i < deterministicGasConfig.BankSendTokenNumberLimit; i++ {
		issueMsgs = append(issueMsgs, &assettypes.MsgIssueFungibleToken{
			Issuer:        sender.String(),
			Symbol:        fmt.Sprintf("TOK%d", i),
			Description:   fmt.Sprintf("TOK%d Description", i),
			Recipient:     sender.String(),
			InitialAmount: amountToSend,
		})
	}

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, testing.BalancesOptions{
		Messages: append([]sdk.Msg{&banktypes.MsgMultiSend{}}, issueMsgs...),
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
	require.Equal(t, deterministicGasConfig.BankSendTokenNumberLimit, len(fungibleTokenIssuedEvts))

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
	bankMultiSend := chain.GasLimitByMsgs(&banktypes.MsgMultiSend{})
	res, err = tx.BroadcastTx(
		ctx,
		clientCtx,
		chain.TxFactory().
			WithMemo(maxMemo). // memo is set to max length here to charge as much gas as possible
			WithGas(bankMultiSend),
		msg)
	require.NoError(t, err)
	require.Equal(t, bankMultiSend, uint64(res.GasUsed))
}

// TestMultiSendDeterministicGasMoreCoins checks that transfer takes the higher deterministic amount of gas above the limit of number of coins transferred
func TestMultiSendDeterministicGasMoreCoins(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := chain.GenAccount()
	recipient := chain.GenAccount()
	deterministicGasConfig := chain.DeterministicGas()

	amountToSend := sdk.NewInt(1000)

	numOfTokens := deterministicGasConfig.BankSendTokenNumberLimit + 3
	issueMsgs := make([]sdk.Msg, 0, numOfTokens)
	for i := 0; i < numOfTokens; i++ {
		issueMsgs = append(issueMsgs, &assettypes.MsgIssueFungibleToken{
			Issuer:        sender.String(),
			Symbol:        fmt.Sprintf("TOK%d", i),
			Description:   fmt.Sprintf("TOK%d Description", i),
			Recipient:     sender.String(),
			InitialAmount: amountToSend,
		})
	}

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, testing.BalancesOptions{
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
	minBankMultiSendGas := chain.GasLimitByMsgs(&banktypes.MsgMultiSend{})
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
	require.Equal(t, minBankMultiSendGas+3*deterministicGasConfig.BankSendAdditionalTransfer, bankMultiSendGas)
}
