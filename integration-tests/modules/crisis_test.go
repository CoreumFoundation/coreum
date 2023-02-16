//go:build integrationtests

package modules

import (
	"testing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
)

// TestVerifyInvariantMessageIsDenied verifies that transactions containing crisis.MsgVerifyInvariant message are rejected.
// We do it because that message does not work and Cosmos SDK team decided to not fix the bug.
func TestVerifyInvariantMessageIsDenied(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	sender := chain.GenAccount()

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, integrationtests.BalancesOptions{
		Amount: chain.NetworkConfig.CrisisConfig.Amount,
	}))

	// the gas price is too low
	_, err := client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().WithSimulateAndExecute(true),
		&crisistypes.MsgVerifyInvariant{
			Sender:              sender.String(),
			InvariantModuleName: banktypes.ModuleName,
			InvariantRoute:      "total-supply",
		})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unauthorized")
}
