//go:build integrationtests

package ibc

import (
	"context"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v6/integration-tests"
	"github.com/CoreumFoundation/coreum/v6/pkg/client"
	"github.com/CoreumFoundation/coreum/v6/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v6/x/asset/ft/types"
	dextypes "github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

func TestIBCDexLimitOrdersMatching(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia
	assetFTClient := assetfttypes.NewQueryClient(coreumChain.ClientContext)
	dexClient := dextypes.NewQueryClient(coreumChain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(
		ctx, t, ibctransfertypes.PortID, coreumChain.ChainContext,
	)

	coreumIssuer := coreumChain.GenAccount()
	coreumSender := coreumChain.GenAccount()
	gaiaRecipient := gaiaChain.GenAccount()

	gaiaChain.Faucet.FundAccounts(ctx, t, integration.FundedAccount{
		Address: gaiaRecipient,
		Amount:  gaiaChain.NewCoin(sdkmath.NewInt(1_000_000)), // coin for the fees
	})

	issueFee := coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount

	coreumChain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: coreumIssuer,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&assetfttypes.MsgIssue{},
					&banktypes.MsgSend{},
					&banktypes.MsgSend{},
				},
				Amount: issueFee,
			},
		}, {
			Acc: coreumSender,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&ibctransfertypes.MsgTransfer{},
					&ibctransfertypes.MsgTransfer{},
				},
				Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(2).AddRaw(200_000),
			},
		},
	})

	denom1 := issueFT(ctx, t, coreumChain, coreumIssuer, sdkmath.NewIntWithDecimal(1, 6), assetfttypes.Feature_ibc)
	denom2 := issueFT(ctx, t, coreumChain, coreumIssuer, sdkmath.NewIntWithDecimal(1, 6))

	sendCoin := sdk.NewCoin(denom1, sdkmath.NewInt(100_000))
	halfCoin := sdk.NewCoin(denom1, sdkmath.NewInt(50_000))
	msgSend := &banktypes.MsgSend{
		FromAddress: coreumIssuer.String(),
		ToAddress:   coreumSender.String(),
		Amount:      sdk.NewCoins(sendCoin),
	}
	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.NoError(err)

	// ibc transfer half the amount
	ibcCoin := sdk.NewCoin(ConvertToIBCDenom(gaiaToCoreumChannelID, denom1), halfCoin.Amount)
	_, err = coreumChain.ExecuteIBCTransfer(
		ctx,
		t,
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(&ibctransfertypes.MsgTransfer{})),
		coreumSender,
		halfCoin,
		gaiaChain.ChainContext,
		gaiaRecipient,
	)
	requireT.NoError(err)
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaRecipient, ibcCoin))

	// place order should fail because of insufficient funds
	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      coreumSender.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sendCoin.Amount,
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumSender),
		coreumChain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.ErrorContains(err, cosmoserrors.ErrInsufficientFunds.Error())

	balanceRes, err := assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: coreumSender.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(halfCoin.Amount.String(), balanceRes.Balance.String())

	// fund the remaining needed amount
	msgSend = &banktypes.MsgSend{
		FromAddress: coreumIssuer.String(),
		ToAddress:   coreumSender.String(),
		Amount:      sdk.NewCoins(halfCoin),
	}
	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.NoError(err)

	// place order should succeed
	placeSellOrderMsg = &dextypes.MsgPlaceOrder{
		Sender:      coreumSender.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sendCoin.Amount,
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumSender),
		coreumChain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: coreumSender.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sendCoin.Amount.String(), balanceRes.Balance.String())
	requireT.Equal(sendCoin.Amount.String(), balanceRes.LockedInDEX.String())

	// ibc transfer should fail because of insufficient funds
	ibcCoin = sdk.NewCoin(ConvertToIBCDenom(gaiaToCoreumChannelID, denom1), halfCoin.Amount)
	_, err = coreumChain.ExecuteIBCTransfer(
		ctx,
		t,
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(&ibctransfertypes.MsgTransfer{})),
		coreumSender,
		halfCoin,
		gaiaChain.ChainContext,
		gaiaRecipient,
	)
	requireT.ErrorContains(err, cosmoserrors.ErrInsufficientFunds.Error())
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaRecipient, ibcCoin))
}

func issueFT(
	ctx context.Context,
	t *testing.T,
	chain integration.CoreumChain,
	issuer sdk.AccAddress,
	initialAmount sdkmath.Int,
	features ...assetfttypes.Feature,
) string {
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "TKN" + uuid.NewString()[:4],
		Subunit:       "tkn" + uuid.NewString()[:4],
		Precision:     5,
		InitialAmount: initialAmount,
		Features:      features,
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	require.NoError(t, err)
	return assetfttypes.BuildDenom(issueMsg.Subunit, issuer)
}
