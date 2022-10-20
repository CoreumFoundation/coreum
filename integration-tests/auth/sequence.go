package auth

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// TestUnexpectedSequenceNumber test verifies that we correctly handle error reporting invalid account sequence number
// used to sign transaction
func TestUnexpectedSequenceNumber(ctx context.Context, t testing.T, chain testing.Chain) {
	sender, err := chain.GenFundedAccount(ctx)
	require.NoError(t, err)

	clientCtx := chain.ClientContext
	accInfo, err := tx.GetAccountInfo(ctx, clientCtx, sender)
	require.NoError(t, err)

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   sender.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdk.NewInt(1))),
	}

	gasPrice, err := tx.GetGasPrice(ctx, clientCtx)
	require.NoError(t, err)

	_, err = tx.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithSequence(accInfo.GetSequence()+1). // incorrect sequence
			WithAccountNumber(accInfo.GetAccountNumber()).
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(gasPrice.String()),
		msg)
	require.ErrorIs(t, cosmoserrors.ErrWrongSequence, err)
}
