//go:build integrationtests

package ibc

import (
	"context"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
)

func TestIBCTransfer(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	sender := chain.GenAccount()
	recipient, err := integrationtests.GenRandomAddress(integrationtests.GaiaAccountPrefix)
	require.NoError(t, err)

	sendCoin := chain.NewCoin(sdk.NewInt(1000))
	channelID := chain.GaiaContext.ChannelID
	// transfer tokens over ibc
	ibcSend := ibctransfertypes.MsgTransfer{
		SourcePort:    ibctransfertypes.PortID,
		SourceChannel: channelID,
		Token:         sendCoin,
		Sender:        sender.String(),
		Receiver:      recipient,
		TimeoutHeight: ibcclienttypes.Height{
			RevisionNumber: 1,
			RevisionHeight: 10000000, // TODO: calculate this value
		},
	}

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibcSend},
		Amount:   sendCoin.Amount,
	}))
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender).WithBroadcastMode(flags.BroadcastBlock),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(&ibcSend)),
		&ibcSend,
	)
	require.NoError(t, err)

	// Query other chain for balance
	gaiaBankClient := banktypes.NewQueryClient(chain.GaiaContext.ClientContext)

	retryCtx, retryCancel := context.WithTimeout(ctx, 20*time.Second)
	defer retryCancel()
	var balancesRecipient *banktypes.QueryAllBalancesResponse
	retry.Do(retryCtx, time.Second, func() error {
		balancesRecipient, err = gaiaBankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
			Address: recipient,
		})
		if err != nil {
			return err
		}

		if len(balancesRecipient.Balances) == 0 {
			return retry.Retryable(errors.New("balances is still empty"))
		}
		return nil
	})
	require.NoError(t, err)
	require.Len(t, balancesRecipient.Balances, 1)

	ibcAmount := ibctransfertypes.GetTransferCoin(ibctransfertypes.PortID, channelID, sendCoin.Denom, sendCoin.Amount)
	require.NoError(t, err)
	assert.EqualValues(t, ibcAmount.String(), balancesRecipient.Balances[0].String())
}
