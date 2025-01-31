//go:build integrationtests

package modules

import (
	"context"
	"encoding/json"
	"testing"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v5/integration-tests"
	moduleswasm "github.com/CoreumFoundation/coreum/v5/integration-tests/contracts/modules"
	"github.com/CoreumFoundation/coreum/v5/pkg/client"
	"github.com/CoreumFoundation/coreum/v5/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v5/x/deterministicgas"
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
		Mul(sdkmath.LegacyOneDec().Sub(feeModel.MaxDiscount))

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
	amountToSendFromMultisigAccount := int64(1_000_000)

	signersCount := 7
	multisigTreshold := 6
	multisigPublicKey, keyNamesSet, err := chain.GenMultisigAccount(signersCount, multisigTreshold)
	requireT.NoError(err)
	multisigAddress := sdk.AccAddress(multisigPublicKey.Address())

	chain.FundAccountWithOptions(ctx, t, multisigAddress, integration.BalancesOptions{
		// for gas estimation to work wee need account to exist on chain so we fund it with to be sent amount.
		Amount: sdkmath.NewInt(amountToSendFromMultisigAccount),
	})

	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	// prepare account to be funded from the multisig
	recipientAddr := recipient.String()
	coinsToSendToRecipient := sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(amountToSendFromMultisigAccount)))

	bankSendMsg := &banktypes.MsgSend{
		FromAddress: multisigAddress.String(),
		ToAddress:   recipientAddr,
		Amount:      coinsToSendToRecipient,
	}

	_, gasEstimation, err := client.CalculateGas(
		ctx,
		chain.ClientContext.WithFromAddress(multisigAddress),
		chain.TxFactory(),
		bankSendMsg,
	)
	requireT.NoError(err)

	// fund the multisig account
	chain.FundAccountWithOptions(ctx, t, multisigAddress, integration.BalancesOptions{
		// because of 6/7 multisig gas exceeds FixedGas, and we need to fund it to pay fees.
		Amount: sdkmath.NewInt(int64(gasEstimation)),
	})

	_, err = chain.SignAndBroadcastMultisigTx(
		ctx,
		chain.ClientContext.WithFromAddress(multisigAddress),
		// We intentionally use simulation instead of using `WithGas(chain.GasLimitByMsgs(bankSendMsg))`.
		// We do it to test simulation for multisig account.
		chain.TxFactoryAuto(),
		bankSendMsg,
		keyNamesSet[0])
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)
	t.Log("Partially signed tx executed with expected error")

	// sign and submit with the min threshold
	txRes, err := chain.SignAndBroadcastMultisigTx(
		ctx,
		chain.ClientContext.WithFromAddress(multisigAddress),
		chain.TxFactoryAuto(),
		bankSendMsg,
		keyNamesSet[:multisigTreshold]...)
	requireT.NoError(err)
	t.Logf("Fully signed tx executed, txHash:%s, gasUsed:%d, gasWanted:%d", txRes.TxHash, txRes.GasUsed, txRes.GasWanted)

	// Real gas used might be less that estimation for multisig account (especially when there are many signers)
	// because in ConsumeTxSizeGasDecorator (cosmos-sdk@v0.47.5/x/auth/ante/basic.go:99) amount of bytes is estimated
	// for the worst case.
	requireT.LessOrEqual(txRes.GasUsed, int64(gasEstimation))

	recipientBalances, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: recipientAddr,
	})
	requireT.NoError(err)
	requireT.Equal(coinsToSendToRecipient, recipientBalances.Balances)
}

// TestAuthUnexpectedSequenceNumber test verifies that we correctly handle error reporting invalid account
// sequence number used to sign transaction.
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

	singlesigAddress := chain.GenAccount()

	multisigPublicKey1, _, err := chain.GenMultisigAccount(3, 2)
	require.NoError(t, err)
	multisigAddress1 := sdk.AccAddress(multisigPublicKey1.Address())

	multisigPublicKey2, _, err := chain.GenMultisigAccount(7, 6)
	require.NoError(t, err)
	multisigAddress2 := sdk.AccAddress(multisigPublicKey2.Address())

	dgc := deterministicgas.DefaultConfig()
	authParams, err := authtypes.NewQueryClient(chain.ClientContext).Params(ctx, &authtypes.QueryParamsRequest{})
	require.NoError(t, err)

	// For accounts to exist on chain we need to fund them at least with min amount (1ucore).
	chain.FundAccountWithOptions(ctx, t, singlesigAddress, integration.BalancesOptions{Amount: sdkmath.NewInt(1)})
	chain.FundAccountWithOptions(ctx, t, multisigAddress1, integration.BalancesOptions{Amount: sdkmath.NewInt(1)})
	chain.FundAccountWithOptions(ctx, t, multisigAddress2, integration.BalancesOptions{Amount: sdkmath.NewInt(1)})

	// For deterministic messages we are able to assert that gas estimation is equal to exact number.
	testsDeterm := []struct {
		name             string
		fromAddress      sdk.AccAddress
		msgs             []sdk.Msg
		simulateUnsigned bool
		expectedGas      func(txBytes []byte) uint64
	}{
		{
			name:        "singlesig_bank_send",
			fromAddress: singlesigAddress,
			msgs: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: singlesigAddress.String(),
					ToAddress:   singlesigAddress.String(),
					Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1))),
				},
			},
			// single signature no extra bytes.
			expectedGas: func(txBytes []byte) uint64 {
				return dgc.FixedGas + deterministicgas.BankSendPerCoinGas
			},
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
			expectedGas: func(txBytes []byte) uint64 {
				return dgc.FixedGas + deterministicgas.BankSendPerCoinGas
			},
		},
		{
			name:        "multisig_6_7_bank_send",
			fromAddress: multisigAddress2,
			msgs: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: multisigAddress2.String(),
					ToAddress:   multisigAddress2.String(),
					Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1))),
				},
			},
			expectedGas: func(txBytes []byte) uint64 {
				signatureCost := uint64(3790)
				bytesCost := uint64(len(txBytes)) * authParams.Params.TxSizeCostPerByte
				expectedGas := dgc.FixedGas + deterministicgas.BankSendPerCoinGas + bytesCost + signatureCost
				return expectedGas
			},
		},
		{
			name:        "multisig_6_7_bank_send_unsigned",
			fromAddress: multisigAddress2,
			msgs: []sdk.Msg{
				&banktypes.MsgSend{
					FromAddress: multisigAddress2.String(),
					ToAddress:   multisigAddress2.String(),
					Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1))),
				},
			},
			simulateUnsigned: true,
			expectedGas: func(txBytes []byte) uint64 {
				signatureCostDifference := uint64(1610)
				bytesCost := uint64(len(txBytes)) * authParams.Params.TxSizeCostPerByte
				expectedGas := dgc.FixedGas + deterministicgas.BankSendPerCoinGas + bytesCost - signatureCostDifference
				return expectedGas
			},
		},
	}
	for _, tt := range testsDeterm {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			txBytes, err := client.BuildTxForSimulation(
				ctx,
				chain.ClientContext.WithFromAddress(tt.fromAddress).WithUnsignedSimulation(tt.simulateUnsigned),
				chain.TxFactory(),
				tt.msgs...,
			)
			require.NoError(t, err)

			_, estimatedGas, err := client.CalculateGas(
				ctx,
				chain.ClientContext.WithFromAddress(tt.fromAddress).WithUnsignedSimulation(tt.simulateUnsigned),
				chain.TxFactory(),
				tt.msgs...,
			)
			require.NoError(t, err)

			require.Equal(t, int(tt.expectedGas(txBytes)), int(estimatedGas))
		})
	}

	// For non-deterministic messages we need to deploy a contract.
	// Any address could be admin since we are not going to execute it but just estimate.
	admin := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{Amount: sdkmath.NewInt(1_000_000)})

	initialPayload, err := json.Marshal(moduleswasm.SimpleState{
		Count: 1337,
	})
	require.NoError(t, err)
	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		chain.TxFactoryAuto(),
		admin,
		moduleswasm.SimpleStateWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayload,
			Label:      "simple_state",
		},
	)
	require.NoError(t, err)

	wasmPayload, err := moduleswasm.MethodToEmptyBodyPayload(moduleswasm.SimpleIncrement)
	require.NoError(t, err)

	// For non-deterministic messages we are unable to know exact number, so we do just basic assertion.
	testsNonDeterm := []struct {
		name        string
		fromAddress sdk.AccAddress
		msgs        []sdk.Msg
	}{
		{
			name:        "singlesig_wasm_execute_contract",
			fromAddress: singlesigAddress,
			msgs: []sdk.Msg{
				&wasmtypes.MsgExecuteContract{
					Sender:   singlesigAddress.String(),
					Contract: contractAddr,
					Msg:      wasmtypes.RawContractMessage(wasmPayload),
					Funds:    sdk.Coins{},
				},
			},
		},
		{
			name:        "multisig_2_3_wasm_execute_contract",
			fromAddress: multisigAddress1,
			msgs: []sdk.Msg{
				&wasmtypes.MsgExecuteContract{
					Sender:   multisigAddress1.String(),
					Contract: contractAddr,
					Msg:      wasmtypes.RawContractMessage(wasmPayload),
					Funds:    sdk.Coins{},
				},
			},
		},
	}
	for _, tt := range testsNonDeterm {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, estimatedGas, err := client.CalculateGas(
				ctx,
				chain.ClientContext.WithFromAddress(tt.fromAddress),
				chain.TxFactory(),
				tt.msgs...,
			)
			require.NoError(t, err)
			require.Positive(t, int(estimatedGas))
		})
	}
}

// TestAuthSignModeDirectAux tests SignModeDirectAux signing mode.
func TestAuthSignModeDirectAux(t *testing.T) {
	t.Parallel()
	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	// Tipper does not pay any tips yet because TipDecorator is not integrated yet (it is still in beta).
	tipper := chain.GenAccount()
	feePayer := chain.GenAccount()
	recipient := chain.GenAccount()

	amountToSend := sdkmath.NewIntFromUint64(1000)
	chain.FundAccountWithOptions(ctx, t, feePayer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})
	chain.Faucet.FundAccounts(ctx, t, integration.FundedAccount{
		Address: tipper,
		Amount:  chain.NewCoin(amountToSend),
	})

	msg := &banktypes.MsgSend{
		FromAddress: tipper.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(amountToSend)),
	}

	tipperKey, err := chain.ClientContext.Keyring().KeyByAddress(tipper)
	requireT.NoError(err)
	tipperPubKey, err := tipperKey.GetPubKey()
	requireT.NoError(err)

	feePayerKey, err := chain.ClientContext.Keyring().KeyByAddress(feePayer)
	requireT.NoError(err)

	tipperAccountInfo, err := client.GetAccountInfo(ctx, chain.ClientContext, tipper)
	requireT.NoError(err)

	feePayerAccountInfo, err := client.GetAccountInfo(ctx, chain.ClientContext, feePayer)
	requireT.NoError(err)

	builder := clienttx.NewAuxTxBuilder()
	builder.SetChainID(chain.ClientContext.ChainID())
	requireT.NoError(builder.SetMsgs(msg))
	builder.SetAddress(tipper.String())
	requireT.NoError(builder.SetPubKey(tipperPubKey))
	builder.SetAccountNumber(tipperAccountInfo.GetAccountNumber())
	builder.SetSequence(tipperAccountInfo.GetSequence())
	requireT.NoError(builder.SetSignMode(signing.SignMode_SIGN_MODE_DIRECT_AUX))

	signBytes, err := builder.GetSignBytes()
	requireT.NoError(err)

	tipperSignature, _, err := chain.ClientContext.Keyring().SignByAddress(
		tipper, signBytes, signing.SignMode_SIGN_MODE_DIRECT_AUX,
	)
	requireT.NoError(err)

	builder.SetSignature(tipperSignature)
	tipperSignerData, err := builder.GetAuxSignerData()
	requireT.NoError(err)

	gas := chain.GasLimitByMsgs(msg)
	txBuilder := chain.ClientContext.TxConfig().NewTxBuilder()
	requireT.NoError(txBuilder.AddAuxSignerData(tipperSignerData))
	txBuilder.SetFeePayer(feePayer)
	txBuilder.SetFeeAmount(sdk.NewCoins(
		chain.NewCoin(
			chain.ChainSettings.GasPrice.
				Mul(sdkmath.LegacyNewDecFromInt(sdkmath.NewIntFromUint64(gas))).
				Ceil().
				RoundInt(),
		),
	))
	txBuilder.SetGasLimit(gas)

	requireT.NoError(clienttx.Sign(
		ctx,
		chain.TxFactory().
			WithAccountNumber(feePayerAccountInfo.GetAccountNumber()).
			WithSequence(feePayerAccountInfo.GetSequence()),
		feePayerKey.Name,
		txBuilder,
		false,
	))
	txBytes, err := chain.ClientContext.TxConfig().TxEncoder()(txBuilder.GetTx())
	requireT.NoError(err)

	_, err = client.BroadcastRawTx(ctx, chain.ClientContext, txBytes)
	requireT.NoError(err)

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	tipperBalanceResp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: tipper.String(),
		Denom:   chain.ChainSettings.Denom,
	})
	requireT.NoError(err)

	feePayerBalanceResp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: feePayer.String(),
		Denom:   chain.ChainSettings.Denom,
	})
	requireT.NoError(err)

	recipientBalanceResp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   chain.ChainSettings.Denom,
	})
	requireT.NoError(err)

	requireT.Equal(chain.NewCoin(sdkmath.ZeroInt()).String(), tipperBalanceResp.Balance.String())
	requireT.Equal(chain.NewCoin(sdkmath.ZeroInt()).String(), feePayerBalanceResp.Balance.String())
	requireT.Equal(chain.NewCoin(amountToSend).String(), recipientBalanceResp.Balance.String())
}

// TestTxWithMultipleSignatures verifies that transaction with multiple signatures is executed correctly.
// For more details check: func signTxWithMultipleSignatures.
func TestTxWithMultipleSignatures(t *testing.T) {
	t.Parallel()
	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	sender1 := chain.GenAccount()
	sender2 := chain.GenAccount()
	receiver := chain.GenAccount()

	sendAmount1 := chain.NewCoin(sdkmath.NewInt(100))
	sendAmount2 := chain.NewCoin(sdkmath.NewInt(50))

	msgs := []sdk.Msg{
		&banktypes.MsgSend{
			FromAddress: sender1.String(),
			ToAddress:   receiver.String(),
			Amount:      sdk.NewCoins(sendAmount1),
		},
		&banktypes.MsgSend{
			FromAddress: sender2.String(),
			ToAddress:   receiver.String(),
			Amount:      sdk.NewCoins(sendAmount2),
		},
	}

	chain.FundAccountWithOptions(ctx, t, sender1, integration.BalancesOptions{
		Amount:   sendAmount1.Amount,
		Messages: msgs, // note that first signer pays fees for the whole tx.
	})
	chain.FundAccountWithOptions(ctx, t, sender2, integration.BalancesOptions{
		Amount: sendAmount2.Amount,
	})

	tx := signTxWithMultipleSignatures(ctx, t, chain, msgs, []sdk.AccAddress{sender1, sender2})

	txBytes, err := chain.ClientContext.TxConfig().TxEncoder()(tx)
	requireT.NoError(err)

	_, err = client.BroadcastRawTx(ctx, chain.ClientContext, txBytes)
	requireT.NoError(err)

	balanceResp, err := banktypes.NewQueryClient(chain.ClientContext).Balance(
		ctx,
		&banktypes.QueryBalanceRequest{
			Address: receiver.String(),
			Denom:   chain.ChainSettings.Denom,
		},
	)
	requireT.NoError(err)
	requireT.Equal(sendAmount1.Amount.Add(sendAmount2.Amount).String(), balanceResp.Balance.Amount.String())
}

// signTxWithMultipleSignatures signs a transaction with multiple signatures.
// Reference: cosmos-sdk/testutil/sims/tx_helpers.go (GenSignedMockTx)
// Note the difference between multisig account transaction and multiple signer account tx.
//
// multisig account tx signature sample:
// "signatures": [
//
//	  "CkAnDHXdaoGxCtO97cMJOxAAg2r5M286FnvZ1Dm2lOiHGhnFesLrNHmdmEFJH8yzaMuBGpMgLs2NsjrP3aD4J..."
//	]
//
// multiple signer account tx:
// "signatures": [
//
//	  "80/z4w/4JaNoxSOBRt1J5bOXyZN27V5Jn9Ssfp/FQ9l5wn/z5jcHMpXTIt7EIcW5vU9nFaoztL+SwYG8FTzC9Q==",
//	  "CBeIHV6NTWPfOcxn/bTKUI/OMOT0SQk3jstEvGgmbhpQJJPDSpC2mQmm8f9AOHBI78FxJ4li2AuCRhFBZEm0Zw=="
//	]
//
// Multisig account tx contains single string in array where multiple signatures are combined.
// While multiple signer account tx contains each signature as a separate element in array.
func signTxWithMultipleSignatures(
	ctx context.Context,
	t *testing.T,
	chain integration.CoreumChain,
	msgs []sdk.Msg,
	signers []sdk.AccAddress,
) sdk.Tx {
	requireT := require.New(t)

	txConfig := chain.ClientContext.TxConfig()
	signMod, err := authsign.APISignModeToInternal(txConfig.SignModeHandler().DefaultMode())
	requireT.NoError(err)

	signerAccInfos := make([]sdk.AccountI, len(signers))
	// Fetch account info for all signers.
	for i, signer := range signers {
		accInfo, err := client.GetAccountInfo(ctx, chain.ClientContext, signer)
		requireT.NoError(err)
		signerAccInfos[i] = accInfo
	}

	sigs := make([]signing.SignatureV2, len(signers))

	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos. This is needed for GetSignBytes to return all bytes
	// which should be signed by each signer.
	// Check cosmos-sdk@v0.47.5/x/auth/tx/direct.go DirectSignBytes for more details.
	for i, signer := range signers {
		k, err := chain.TxFactory().Keybase().KeyByAddress(signer)
		requireT.NoError(err)

		pubKey, err := k.GetPubKey()
		requireT.NoError(err)

		sigs[i] = signing.SignatureV2{
			PubKey: pubKey,
			Data: &signing.SingleSignatureData{
				SignMode: signMod,
			},
			Sequence: signerAccInfos[i].GetSequence(),
		}
	}

	txBuilder, err := chain.TxFactory().
		WithGas(chain.GasLimitByMsgs(msgs...)).
		BuildUnsignedTx(msgs...)
	requireT.NoError(err)
	requireT.NoError(txBuilder.SetSignatures(sigs...))

	// 2nd round: sign and set real signatures.
	for i, signer := range signers {
		signerData := authsign.SignerData{
			Address:       signer.String(),
			ChainID:       chain.ChainContext.ChainSettings.ChainID,
			AccountNumber: signerAccInfos[i].GetAccountNumber(),
			Sequence:      signerAccInfos[i].GetSequence(),
			PubKey:        sigs[i].PubKey,
		}
		signBytes, err := authsign.GetSignBytesAdapter(
			ctx,
			txConfig.SignModeHandler(),
			signMod,
			signerData,
			txBuilder.GetTx(),
		)
		requireT.NoError(err)
		sig, _, err := chain.TxFactory().Keybase().SignByAddress(signer, signBytes, signMod)
		requireT.NoError(err)

		sigs[i].Data.(*signing.SingleSignatureData).Signature = sig
	}
	requireT.NoError(txBuilder.SetSignatures(sigs...))

	return txBuilder.GetTx()
}
