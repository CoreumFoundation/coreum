package auth

import (
	"context"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
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
		testing.FundedAccount{
			Wallet: sender,
			Amount: testing.MustNewCoin(t, testing.ComputeNeededBalance(
				chain.NetworkConfig.Fee.FeeModel.InitialGasPrice,
				chain.NetworkConfig.Fee.DeterministicGas.BankSend,
				1,
				sdk.NewInt(10),
			), chain.NetworkConfig.TokenSymbol),
		},
	))

	privateKey := secp256k1.PrivKey{Key: sender.Key}
	senderAddress := sdk.AccAddress(privateKey.PubKey().Address())
	info, err := chain.ClientCtx.AccountRetriever.GetAccount(chain.ClientCtx, senderAddress)
	require.NoError(t, err)

	var accInfo tx.AccountInfo
	accInfo.Number = info.GetAccountNumber()
	accInfo.Sequence = info.GetSequence() + 1 // Intentionally set incorrect sequence number

	msg := &banktypes.MsgSend{
		FromAddress: senderAddress.String(),
		ToAddress:   senderAddress.String(),
		Amount: []sdk.Coin{
			{Denom: chain.NetworkConfig.TokenSymbol, Amount: sdk.NewInt(1)},
		},
	}
	signInput := tx.SignInput{
		PrivateKey:  privateKey,
		GasLimit:    chain.NetworkConfig.Fee.DeterministicGas.BankSend,
		GasPrice:    sdk.Coin{Amount: chain.NetworkConfig.Fee.FeeModel.InitialGasPrice, Denom: chain.NetworkConfig.TokenSymbol},
		AccountInfo: accInfo,
	}

	// Broadcast a transaction using incorrect sequence number
	_, err = tx.BroadcastAsync(ctx, chain.ClientCtx, signInput, msg)
	require.Error(t, err) // We expect error

	// We expect that we get an error saying what the correct sequence number should be
	expectedSeq, ok, err2 := client.ExpectedSequenceFromError(err)
	require.NoError(t, err2)
	if !ok {
		require.Fail(t, "Unexpected error", err.Error())
	}
	require.Equal(t, info.GetSequence(), expectedSeq)
}
