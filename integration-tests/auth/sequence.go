package auth

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// TestUnexpectedSequenceNumber test verifies that we correctly handle error reporting invalid account sequence number
// used to sign transaction
func TestUnexpectedSequenceNumber(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := testing.RandomWallet()

	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(
			sender,
			chain.NewCoin(testing.ComputeNeededBalance(
				chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
				chain.GasLimitByMsgs(&banktypes.MsgSend{}),
				1,
				sdk.NewInt(10),
			)),
		),
	))

	coredClient := chain.Client

	accNum, accSeq, err := coredClient.GetNumberSequence(ctx, sender.Key.Address())
	require.NoError(t, err)

	sender.AccountNumber = accNum
	sender.AccountSequence = accSeq + 1 // Intentionally set incorrect sequence number

	// Broadcast a transaction using incorrect sequence number
	txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
		Base: tx.BaseInput{
			Signer:   sender,
			GasLimit: chain.GasLimitByMsgs(&banktypes.MsgSend{}),
			GasPrice: chain.NewDecCoin(chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice),
		},
		Sender:   sender,
		Receiver: sender,
		Amount:   chain.NewCoin(sdk.NewInt(1)),
	})
	require.NoError(t, err)
	_, err = coredClient.Broadcast(ctx, txBytes)
	require.Error(t, err) // We expect error

	// We expect that we get an error saying what the correct sequence number should be
	expectedSeq, ok, err2 := client.ExpectedSequenceFromError(err)
	require.NoError(t, err2)
	if !ok {
		require.Fail(t, "Unexpected error", err.Error())
	}
	require.Equal(t, accSeq, expectedSeq)
}
