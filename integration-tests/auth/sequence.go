package auth

import (
	"context"

	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// TestUnexpectedSequenceNumber test verifies that we correctly handle error reporting invalid account sequence number
// used to sign transaction
func TestUnexpectedSequenceNumber(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
	sender := testing.RandomWallet()

	return func(ctx context.Context) error {
			initialBalance, err := types.NewCoin2(types.NewInt(180000010), chain.Network.TokenSymbol())
			if err != nil {
				return err
			}
			return chain.Network.FundAccount(sender.Key.PubKey(), initialBalance.String())
		},
		func(ctx context.Context, t testing.T) {
			coredClient := chain.Client

			accNum, accSeq, err := coredClient.GetNumberSequence(ctx, sender.Key.Address())
			require.NoError(t, err)

			sender.AccountNumber = accNum
			sender.AccountSequence = accSeq + 1 // Intentionally set incorrect sequence number

			gasPrice, err := types.NewCoin2(chain.Network.InitialGasPrice2(), chain.Network.TokenSymbol())
			require.NoError(t, err)

			amount, err := types.NewCoin2(types.NewInt(1), chain.Network.TokenSymbol())
			require.NoError(t, err)

			// Broadcast a transaction using incorrect sequence number
			txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
				Base: tx.BaseInput{
					Signer:   sender,
					GasLimit: chain.Network.DeterministicGas().BankSend,
					GasPrice: gasPrice,
				},
				Sender:   sender,
				Receiver: sender,
				Amount:   amount,
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
}
