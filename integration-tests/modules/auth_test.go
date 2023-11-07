//go:build integrationtests

package modules

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	"github.com/CoreumFoundation/coreum/v3/pkg/client"
	"github.com/CoreumFoundation/coreum/v3/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v3/x/deterministicgas"
)

// TestAuthFeeLimits verifies that invalid message gas won't be accepted.
func TestAuthFeeLimits(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	sender := chain.GenAccount()

	feeModel := getFeemodelParams(ctx, t, chain.ClientContext)
	maxBlockGas := feeModel.MaxBlockGas
	chain.FundAccountWithOptions(ctx, t, sender, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&assetfttypes.MsgIssue{},
		},
		NondeterministicMessagesGas: uint64(maxBlockGas) + 100,
		Amount:                      chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   sender.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1))),
	}

	gasPriceWithMaxDiscount := feeModel.InitialGasPrice.
		Mul(sdk.OneDec().Sub(feeModel.MaxDiscount))

	// the gas price is too low
	_, err := client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(chain.NewDecCoin(gasPriceWithMaxDiscount.QuoInt64(2)).String()),
		msg)
	require.True(t, cosmoserrors.ErrInsufficientFee.Is(err))

	// no gas price
	_, err = client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(""),
		msg)
	require.True(t, cosmoserrors.ErrInsufficientFee.Is(err))

	// more gas than MaxBlockGas
	_, err = client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(uint64(maxBlockGas+1)),
		msg)
	require.Error(t, err)

	// gas equal MaxBlockGas, the tx should pass
	_, err = client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(uint64(maxBlockGas)),
		msg)
	require.NoError(t, err)

	// fee paid in another coin is rejected
	const subunit = "uzzz" // uzzz is intentionally selected to put it on second position, after ucore, in sorted coins
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        sender.String(),
		Symbol:        "ZZZ",
		Subunit:       subunit,
		Precision:     6,
		Description:   "ZZZ Description",
		InitialAmount: sdkmath.NewInt(1000),
		Features:      []assetfttypes.Feature{},
	}
	denom := assetfttypes.BuildDenom(subunit, sender)
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	require.NoError(t, err)

	_, err = client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(sdk.NewInt64Coin(denom, 1).String()),
		msg)
	require.Error(t, err)
	require.True(t, cosmoserrors.ErrInvalidCoins.Is(err))

	// fee paid both in core and another coin is rejected
	_, err = client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(chain.TxFactory().GasPrices().Add(sdk.NewInt64DecCoin(denom, 1)).Sort().String()),
		msg)
	require.Error(t, err)
	require.True(t, cosmoserrors.ErrInvalidCoins.Is(err))
}

// TestAuthMultisig tests the cosmos-sdk multisig accounts and API.
func TestAuthMultisig(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	recipient := chain.GenAccount()
	amountToSendFromMultisigAccount := int64(1000)

	signersCount := 7
	multisigTreshold := 6
	multisigPublicKey, keyNamesSet, err := chain.GenMultisigAccount(signersCount, multisigTreshold)
	requireT.NoError(err)
	multisigAddress := sdk.AccAddress(multisigPublicKey.Address())

	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	// fund the multisig account
	chain.FundAccountWithOptions(ctx, t, multisigAddress, integration.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   sdkmath.NewInt(amountToSendFromMultisigAccount),
	})

	// prepare account to be funded from the multisig
	recipientAddr := recipient.String()
	coinsToSendToRecipient := sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(amountToSendFromMultisigAccount)))

	bankSendMsg := &banktypes.MsgSend{
		FromAddress: multisigAddress.String(),
		ToAddress:   recipientAddr,
		Amount:      coinsToSendToRecipient,
	}
	_, err = chain.SignAndBroadcastMultisigTx(
		ctx,
		chain.ClientContext.WithFromAddress(multisigAddress),
		// We intentionally use simulation instead of using `WithGas(chain.GasLimitByMsgs(bankSendMsg))`.
		// We do it to test simulation for multisig account.
		chain.TxFactory().WithSimulateAndExecute(true),
		bankSendMsg,
		keyNamesSet[0])
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)
	t.Log("Partially signed tx executed with expected error")

	_, estimatedGas, err := client.CalculateGas(
		ctx,
		chain.ClientContext.WithFromAddress(multisigAddress),
		chain.TxFactory().WithGasAdjustment(1.0),
		bankSendMsg,
	)
	requireT.NoError(err)

	// sign and submit with the min threshold
	txRes, err := chain.SignAndBroadcastMultisigTx(
		ctx,
		chain.ClientContext.WithFromAddress(multisigAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(bankSendMsg)),
		bankSendMsg,
		keyNamesSet[:multisigTreshold]...)
	requireT.NoError(err)
	t.Logf("Fully signed tx executed, txHash:%s, gas:%d", txRes.TxHash, txRes.GasUsed)

	//requireT.Equal(txRes.GasUsed, txRes.GasWanted) // another option to reproduce is to use chain.TxFactory().WithSimulateAndExecute(true) & this assertion.

	requireT.Equal(txRes.GasUsed, int64(estimatedGas)) // this shouldn't fail.

	recipientBalances, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: recipientAddr,
	})
	requireT.NoError(err)
	requireT.Equal(coinsToSendToRecipient, recipientBalances.Balances)
}

// TestAuthUnexpectedSequenceNumber test verifies that we correctly handle error reporting invalid account sequence number
// used to sign transaction.
func TestAuthUnexpectedSequenceNumber(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	sender := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, sender, integration.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   sdkmath.NewInt(10),
	})

	clientCtx := chain.ClientContext
	accInfo, err := client.GetAccountInfo(ctx, clientCtx, sender)
	require.NoError(t, err)

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   sender.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1))),
	}

	_, err = client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithSequence(accInfo.GetSequence()+1). // incorrect sequence
			WithAccountNumber(accInfo.GetAccountNumber()).
			WithGas(chain.GasLimitByMsgs(msg)),
		msg)
	require.True(t, cosmoserrors.ErrWrongSequence.Is(err))
}

func TestGasEstimation(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	sender := chain.GenAccount()

	multisigPublicKey1, _, err := chain.GenMultisigAccount(3, 2)
	require.NoError(t, err)
	multisigAddress1 := sdk.AccAddress(multisigPublicKey1.Address())

	multisigPublicKey2, _, err := chain.GenMultisigAccount(7, 6)
	require.NoError(t, err)
	multisigAddress2 := sdk.AccAddress(multisigPublicKey2.Address())

	dgc := deterministicgas.DefaultConfig()

	// For accounts to exist on chain we need to fund them at least with min amount (1ucore).
	chain.FundAccountWithOptions(ctx, t, sender, integration.BalancesOptions{Amount: sdkmath.NewInt(1)})
	chain.FundAccountWithOptions(ctx, t, multisigAddress1, integration.BalancesOptions{Amount: sdkmath.NewInt(1)})
	chain.FundAccountWithOptions(ctx, t, multisigAddress2, integration.BalancesOptions{Amount: sdkmath.NewInt(1)})

	//initialPayload, err := json.Marshal(moduleswasm.SimpleState{
	//	Count: 1337,
	//})
	//requireT.NoError(err)
	//contractAddr, codeID, err := chain.Wasm.DeployAndInstantiateWASMContract(
	//	ctx,
	//	chain.TxFactory().WithSimulateAndExecute(true),
	//	admin,
	//	moduleswasm.SimpleStateWASM,
	//	integration.InstantiateConfig{
	//		AccessType: wasmtypes.AccessTypeUnspecified,
	//		Payload:    initialPayload,
	//		Label:      "simple_state",
	//	},
	//)
	//requireT.NoError(err)
	//chain.Wasm.ExecuteWASMContract()

	tests := []struct {
		name        string
		fromAddress sdk.AccAddress
		msgs        []sdk.Msg
		expectedGas uint64
	}{
		{
			name:        "singlesig_bank_send",
			fromAddress: sender,
			msgs: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: sender.String(),
					ToAddress:   sender.String(),
					Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1))),
				},
			},
			// single signature no extra bytes.
			expectedGas: dgc.FixedGas + 1*deterministicgas.BankSendPerCoinGas,
		},
		{
			name:        "multisig_2_3_bank_send",
			fromAddress: multisigAddress1,
			msgs: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: multisigAddress1.String(),
					ToAddress:   multisigAddress1.String(),
					Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1))),
				},
			},
			// single signature no extra bytes.
			// Note that multisig account and multiple signatures in a single tx are different.
			// Multisig tx still has single signature which is combination of multiple signatures so gas is charged for single sig.
			// Tx containing multiple signatures is a different case and gas is charged for each standalone signature.
			expectedGas: dgc.FixedGas + 1*deterministicgas.BankSendPerCoinGas,
		},
		// FIXME: This test fails. Probably because of bug.
		//{
		//	name:        "multisig_6_7_bank_send",
		//	fromAddress: multisigAddress2,
		//	msgs: []sdk.Msg{
		//		&banktypes.MsgSend{
		//			FromAddress: multisigAddress2.String(),
		//			ToAddress:   multisigAddress2.String(),
		//			Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1))),
		//		},
		//	},
		//	expectedGas:             dgc.FixedGas + 1*deterministicgas.BankSendPerCoinGas,
		//	expectedGasAllowedDelta: 0,
		//},
		{
			name:        "singlesig_auth_exec_and_bank_send",
			fromAddress: sender,
			msgs: []sdk.Msg{
				lo.ToPtr(
					authztypes.NewMsgExec(sender, []sdk.Msg{
						&banktypes.MsgSend{
							FromAddress: sender.String(),
							ToAddress:   sender.String(),
							Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1))),
						},
					})),
				&banktypes.MsgSend{
					FromAddress: sender.String(),
					ToAddress:   sender.String(),
					Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1))),
				},
			},
			// single signature no extra bytes.
			expectedGas: dgc.FixedGas + 1*deterministicgas.BankSendPerCoinGas + (1*deterministicgas.AuthzExecOverhead + 1*deterministicgas.BankSendPerCoinGas),
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
			require.NoError(t, err)
			require.Equal(t, test.expectedGas, estimatedGas)
		})
	}
}
