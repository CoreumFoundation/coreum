//go:build integrationtests

package modules

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	"github.com/CoreumFoundation/coreum/v3/pkg/client"
)

// TestVerifyInvariantMessageIsDenied verifies that transactions containing crisis.MsgVerifyInvariant message are rejected.
// We do it because that message does not work and Cosmos SDK team decided to not fix the bug.
func TestVerifyInvariantMessageIsDenied(t *testing.T) {
	t.Parallel()

	// This fee must correspond to the one set in genesis. Crisis module does not allow
	// to query it, and we don't want to store it in network config either.
	const invariantFee = 500_000_000_000

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	sender := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, sender, integrationtests.BalancesOptions{
		Amount: sdkmath.NewIntFromUint64(invariantFee),
	})

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
