//go:build integrationtests

package ibc

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"
	"github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	integrationtests "github.com/CoreumFoundation/coreum/v6/integration-tests"
	"github.com/CoreumFoundation/coreum/v6/testutil/integration"
)

// TestICAController tests the ICA controller capabilities.
func TestICACoreumController(t *testing.T) {
	t.Parallel()
	ctx, chains := integrationtests.NewChainsTestingContext(t)
	testICAIntegration(ctx, t, chains.Coreum.Chain, chains.Gaia)
}

// TestICAController tests the ICA host capabilities.
func TestICACoreumHost(t *testing.T) {
	t.Parallel()
	ctx, chains := integrationtests.NewChainsTestingContext(t)
	testICAIntegration(ctx, t, chains.Gaia, chains.Coreum.Chain)
}

// TestICADeterministicGas tests the ICA deterministic gas messages.
func TestICADeterministicGas(t *testing.T) {
	t.Parallel()

	requireT := require.New(t)

	ctx, chains := integrationtests.NewChainsTestingContext(t)

	controllerCaller := chains.Coreum.GenAccount()
	chains.Coreum.Faucet.FundAccounts(ctx, t, integration.FundedAccount{
		Address: controllerCaller,
		Amount:  chains.Coreum.NewCoin(sdkmath.NewIntWithDecimal(1, 7)),
	})
	_, controllerToHostConnectionID := chains.Coreum.AwaitForIBCClientAndConnectionIDs(
		ctx, t, chains.Gaia.ChainSettings.ChainID,
	)
	msgRegisterInterchainAccountOnHost := icacontrollertypes.MsgRegisterInterchainAccount{
		ConnectionId: controllerToHostConnectionID,
		Owner:        chains.Coreum.MustConvertToBech32Address(controllerCaller),
		Ordering:     types.UNORDERED,
	}
	txRes, err := chains.Coreum.BroadcastTxWithSigner(
		ctx,
		chains.Coreum.TxFactory().WithGas(chains.Coreum.GasLimitByMsgs(&icacontrollertypes.MsgRegisterInterchainAccount{})),
		controllerCaller,
		&msgRegisterInterchainAccountOnHost,
	)
	requireT.NoError(err)
	requireT.EqualValues(txRes.GasUsed, chains.Coreum.GasLimitByMsgs(&icacontrollertypes.MsgRegisterInterchainAccount{}))
}

func testICAIntegration(
	ctx context.Context,
	t *testing.T,
	controllerChain integration.Chain,
	hostChain integration.Chain,
) {
	t.Helper()

	requireT := require.New(t)

	controllerAcc := controllerChain.GenAccount()
	controllerChain.Faucet.FundAccounts(ctx, t, integration.FundedAccount{
		Address: controllerAcc,
		Amount:  controllerChain.NewCoin(sdkmath.NewIntWithDecimal(1, 7)),
	})

	t.Logf(
		"Generated and funded an account on controller chain: %s",
		controllerChain.MustConvertToBech32Address(controllerAcc),
	)

	_, controllerToHostConnectionID := controllerChain.AwaitForIBCClientAndConnectionIDs(
		ctx,
		t,
		hostChain.ChainSettings.ChainID,
	)

	msgRegisterInterchainAccount := icacontrollertypes.MsgRegisterInterchainAccount{
		ConnectionId: controllerToHostConnectionID,
		Owner:        controllerChain.MustConvertToBech32Address(controllerAcc),
		Ordering:     types.UNORDERED,
	}
	_, err := controllerChain.BroadcastTxWithSigner(
		ctx,
		controllerChain.TxFactoryAuto(),
		controllerAcc,
		&msgRegisterInterchainAccount,
	)
	requireT.NoError(err)

	t.Logf(
		"Waiting for ICA account on controller chain (%s) to be created on host chain (%s), connectionID: %s.",
		controllerChain.ChainSettings.ChainID,
		hostChain.ChainSettings.ChainID,
		controllerToHostConnectionID,
	)

	controllerChainICAControllerClient := icacontrollertypes.NewQueryClient(controllerChain.ClientContext)

	var hostICAAcc sdk.AccAddress
	require.NoError(t, controllerChain.AwaitState(ctx, func(ctx context.Context) error {
		icaAccRes, err := controllerChainICAControllerClient.InterchainAccount(ctx,
			&icacontrollertypes.QueryInterchainAccountRequest{
				Owner:        controllerChain.MustConvertToBech32Address(controllerAcc),
				ConnectionId: controllerToHostConnectionID,
			})
		if err != nil {
			return retry.Retryable(errors.Errorf("ICA account is not ready yet, %s", err))
		}
		_, hostICAAcc, err = bech32.DecodeAndConvert(icaAccRes.Address)
		require.NoError(t, err)
		return nil
	}))

	t.Logf("Corresponding account is created on the host chain: %s", hostChain.MustConvertToBech32Address(hostICAAcc))

	// fund the ICA account on host
	hostChain.Faucet.FundAccounts(ctx, t, integration.FundedAccount{
		Address: hostICAAcc,
		Amount:  hostChain.NewCoin(sdkmath.NewIntWithDecimal(1, 8)),
	})

	// generate the host recipient, but don't fund it
	hostRecipient := hostChain.GenAccount()

	amtToSendOnHost := hostChain.NewCoin(sdkmath.NewIntWithDecimal(1, 6))
	msgBankSendOnHost := banktypes.MsgSend{
		FromAddress: hostChain.MustConvertToBech32Address(hostICAAcc),
		ToAddress:   hostChain.MustConvertToBech32Address(hostRecipient),
		Amount:      sdk.NewCoins(amtToSendOnHost),
	}

	icaMsgData, err := icatypes.SerializeCosmosTx(
		hostChain.ClientContext.Codec(),
		[]proto.Message{&msgBankSendOnHost},
		icatypes.EncodingProtobuf,
	)
	require.NoError(t, err)

	msgICAMsgBankSendOnHost := icacontrollertypes.MsgSendTx{
		Owner:        controllerChain.MustConvertToBech32Address(controllerAcc),
		ConnectionId: controllerToHostConnectionID,
		PacketData: icatypes.InterchainAccountPacketData{
			Type: icatypes.EXECUTE_TX,
			Data: icaMsgData,
		},
		// Relative timeout timestamp provided will be added to the current block time during transaction execution.
		RelativeTimeout: uint64(time.Hour),
	}

	_, err = controllerChain.BroadcastTxWithSigner(
		ctx,
		controllerChain.TxFactoryAuto(),
		controllerAcc,
		&msgICAMsgBankSendOnHost,
	)
	require.NoError(t, err)
	requireT.NoError(hostChain.AwaitForBalance(ctx, t, hostRecipient, amtToSendOnHost))
}
