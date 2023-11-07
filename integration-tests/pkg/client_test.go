//go:build integrationtests

package pkg

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	"github.com/CoreumFoundation/coreum/v3/pkg/client"
	"github.com/CoreumFoundation/coreum/v3/testutil/integration"
)

func TestCalculateGas(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	sender := chain.GenAccount()

	multisigPublicKey1, keyNamesSet, err := chain.GenMultisigAccount(4, 3)
	require.NoError(t, err)
	multisigAddress1 := sdk.AccAddress(multisigPublicKey1.Address())
	_ = keyNamesSet

	multisigPublicKey2, keyNamesSet, err := chain.GenMultisigAccount(5, 4)
	require.NoError(t, err)
	multisigAddress2 := sdk.AccAddress(multisigPublicKey2.Address())
	_ = keyNamesSet

	chain.FundAccountWithOptions(ctx, t, sender, integration.BalancesOptions{Amount: sdkmath.NewInt(1)})
	chain.FundAccountWithOptions(ctx, t, multisigAddress1, integration.BalancesOptions{Amount: sdkmath.NewInt(1)})
	chain.FundAccountWithOptions(ctx, t, multisigAddress2, integration.BalancesOptions{Amount: sdkmath.NewInt(1)})

	tests := []struct {
		name        string
		fromAddress sdk.AccAddress
		msgs        []sdk.Msg
		expectedGas int
	}{
		{
			name:        "single address send",
			fromAddress: sender,
			msgs: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: sender.String(),
					ToAddress:   sender.String(),
					Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1))),
				},
			},
			expectedGas: 65_000 + 1*50_000 + 0*10 + (1-1)*1000,
		},
		{
			name:        "multisig 3/4 bank send",
			fromAddress: multisigAddress1,
			msgs: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: multisigAddress1.String(),
					ToAddress:   multisigAddress1.String(),
					Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1))),
				},
			},
			expectedGas: 65_000 + 1*50_000 + 0*10 + (3-1)*1000, // extra 2 signatures
		},
		{
			name:        "multisig 4/5 bank send",
			fromAddress: multisigAddress1,
			msgs: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: multisigAddress1.String(),
					ToAddress:   multisigAddress1.String(),
					Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1))),
				},
			},
			expectedGas: 65_000 + 1*50_000 + 0*10 + (3-1)*1000, // extra 2 signatures
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, estimatedGas, err := client.CalculateGas(
				ctx,
				chain.ClientContext.WithFromAddress(test.fromAddress),
				chain.TxFactory(),
				test.msgs...,
			)
			chainEstimatedGas := chain.GasLimitByMsgs(test.msgs...)

			fmt.Printf("estimatedGas: %v\n", estimatedGas)
			fmt.Printf("chainEstimatedGas: %v\n", chainEstimatedGas)
			require.NoError(t, err)
			require.EqualValues(t, test.expectedGas, int(estimatedGas))
		})
	}
}
