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

	multisigPublicKey, keyNamesSet, err := chain.GenMultisigAccount(4, 3)
	require.NoError(t, err)

	multisigAddress := sdk.AccAddress(multisigPublicKey.Address())
	_ = keyNamesSet
	//signer1KeyName := keyNamesSet[0]
	//signer2KeyName := keyNamesSet[1]

	chain.FundAccountWithOptions(ctx, t, sender, integration.BalancesOptions{Amount: sdkmath.NewInt(1)})
	chain.FundAccountWithOptions(ctx, t, multisigAddress, integration.BalancesOptions{Amount: sdkmath.NewInt(1)})

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
			name:        "multisig bank send",
			fromAddress: multisigAddress,
			msgs: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: multisigAddress.String(),
					ToAddress:   multisigAddress.String(),
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
