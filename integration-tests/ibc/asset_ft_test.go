//go:build integrationtests

package ibc

import (
	"context"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	"github.com/CoreumFoundation/coreum/v3/pkg/client"
	assetfttypes "github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
)

func TestIBCFailsIfNotEnabled(t *testing.T) {
	t.Parallel()

	requireT := require.New(t)

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	coreumChain := chains.Coreum
	coreumIssuer := coreumChain.GenAccount()

	issueFee := coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount
	coreumChain.FundAccountWithOptions(ctx, t, coreumIssuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&ibctransfertypes.MsgTransfer{},
		},
		Amount: issueFee,
	})

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        coreumIssuer.String(),
		Symbol:        "mysymbol",
		Subunit:       "mysubunit",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(1_000_000),
	}
	_, err := client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	require.NoError(t, err)

	gaiaChain := chains.Gaia
	_, err = coreumChain.ExecuteIBCTransfer(
		ctx,
		t,
		coreumIssuer,
		sdk.NewCoin(assetfttypes.BuildDenom(issueMsg.Subunit, coreumIssuer), sdkmath.NewInt(1000)),
		gaiaChain.ChainContext,
		gaiaChain.GenAccount(),
	)
	requireT.ErrorContains(err, "unauthorized")
}

func TestIBCAssetFTSendCommissionAndBurnRate(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)

	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia
	osmosisChain := chains.Osmosis

	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID)
	coreumToGaiaChannelID := coreumChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, gaiaChain.ChainSettings.ChainID)
	osmosisToCoreumChannelID := osmosisChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID)
	coreumToOsmosisChannelID := coreumChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, osmosisChain.ChainSettings.ChainID)

	coreumToGaiaEscrowAddress := ibctransfertypes.GetEscrowAddress(ibctransfertypes.PortID, coreumToGaiaChannelID)
	coreumToOsmosisEscrowAddress := ibctransfertypes.GetEscrowAddress(ibctransfertypes.PortID, coreumToOsmosisChannelID)

	coreumSender := coreumChain.GenAccount()
	gaiaRecipient1 := gaiaChain.GenAccount()
	gaiaRecipient2 := gaiaChain.GenAccount()
	osmosisRecipient1 := osmosisChain.GenAccount()
	osmosisRecipient2 := osmosisChain.GenAccount()

	gaiaChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: gaiaRecipient1,
		Amount:  gaiaChain.NewCoin(sdkmath.NewInt(1000000)), // coin for the fees
	}, integrationtests.FundedAccount{
		Address: gaiaRecipient2,
		Amount:  gaiaChain.NewCoin(sdkmath.NewInt(1000000)), // coin for the fees
	})

	osmosisChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: osmosisRecipient1,
		Amount:  gaiaChain.NewCoin(sdkmath.NewInt(1000000)), // coin for the fees
	}, integrationtests.FundedAccount{
		Address: osmosisRecipient2,
		Amount:  gaiaChain.NewCoin(sdkmath.NewInt(1000000)), // coin for the fees
	})

	coreumIssuer := coreumChain.GenAccount()
	issueFee := coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount
	coreumChain.FundAccountWithOptions(ctx, t, coreumIssuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&assetfttypes.MsgIssue{},
			&ibctransfertypes.MsgTransfer{},
			&ibctransfertypes.MsgTransfer{},
		},
		Amount: issueFee,
	})

	coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&ibctransfertypes.MsgTransfer{},
			&ibctransfertypes.MsgTransfer{},
		},
	})

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             coreumIssuer.String(),
		Symbol:             "mysymbol",
		Subunit:            "mysubunit",
		Precision:          8,
		InitialAmount:      sdkmath.NewInt(1_000_000),
		BurnRate:           sdk.MustNewDecFromStr("0.1"),
		SendCommissionRate: sdk.MustNewDecFromStr("0.2"),
		Features:           []assetfttypes.Feature{assetfttypes.Feature_ibc},
	}
	_, err := client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	require.NoError(t, err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, coreumIssuer)

	sendCoin := sdk.NewCoin(denom, sdkmath.NewInt(1000))
	burntAmount := issueMsg.BurnRate.Mul(sdk.NewDecFromInt(sendCoin.Amount)).TruncateInt()
	sendCommissionAmount := issueMsg.SendCommissionRate.Mul(sdk.NewDecFromInt(sendCoin.Amount)).TruncateInt()
	extraAmount := sdkmath.NewInt(77) // some amount to be left at the end
	msgSend := &banktypes.MsgSend{
		FromAddress: coreumIssuer.String(),
		ToAddress:   coreumSender.String(),
		// amount to send + burn rate + send commission rate + some amount to test with none-zero reminder
		Amount: sdk.NewCoins(sdk.NewCoin(denom,
			sendCoin.Amount.MulRaw(2).
				Add(burntAmount.MulRaw(2)).
				Add(sendCommissionAmount.MulRaw(2)).
				Add(extraAmount)),
		),
	}

	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.NoError(err)

	receiveCoinGaia := sdk.NewCoin(ConvertToIBCDenom(gaiaToCoreumChannelID, sendCoin.Denom), sendCoin.Amount)
	receiveCoinOsmosis := sdk.NewCoin(ConvertToIBCDenom(osmosisToCoreumChannelID, sendCoin.Denom), sendCoin.Amount)

	// ********** Coreum to Gaia **********
	// IBC transfer from FT issuer address.
	ibcTransferAndAssertBalanceChanges(
		ctx,
		t,
		coreumChain.ChainContext,
		coreumIssuer,
		sendCoin,
		gaiaChain.ChainContext,
		gaiaRecipient1,
		receiveCoinGaia,
		map[string]sdkmath.Int{
			coreumChain.MustConvertToBech32Address(coreumIssuer):              sendCoin.Amount.Neg(),
			coreumChain.MustConvertToBech32Address(coreumToGaiaEscrowAddress): sendCoin.Amount,
		},
		map[string]sdkmath.Int{
			gaiaChain.MustConvertToBech32Address(gaiaRecipient1): sendCoin.Amount,
		},
	)

	// IBC transfer from non-FT issuer address.
	ibcTransferAndAssertBalanceChanges(
		ctx,
		t,
		coreumChain.ChainContext,
		coreumSender,
		sendCoin,
		gaiaChain.ChainContext,
		gaiaRecipient2,
		receiveCoinGaia,
		map[string]sdkmath.Int{
			coreumChain.MustConvertToBech32Address(coreumSender):              sendCoin.Amount.Add(sendCommissionAmount).Add(burntAmount).Neg(),
			coreumChain.MustConvertToBech32Address(coreumIssuer):              sendCommissionAmount,
			coreumChain.MustConvertToBech32Address(coreumToGaiaEscrowAddress): sendCoin.Amount,
		},
		map[string]sdkmath.Int{
			gaiaChain.MustConvertToBech32Address(gaiaRecipient2): sendCoin.Amount,
		},
	)

	// ********** Coreum to Osmosis **********
	// IBC transfer from FT issuer address.
	ibcTransferAndAssertBalanceChanges(
		ctx,
		t,
		coreumChain.ChainContext,
		coreumIssuer,
		sendCoin,
		osmosisChain.ChainContext,
		osmosisRecipient1,
		receiveCoinOsmosis,
		map[string]sdkmath.Int{
			coreumChain.MustConvertToBech32Address(coreumIssuer):                 sendCoin.Amount.Neg(),
			coreumChain.MustConvertToBech32Address(coreumToOsmosisEscrowAddress): sendCoin.Amount,
		},
		map[string]sdkmath.Int{
			osmosisChain.MustConvertToBech32Address(osmosisRecipient1): sendCoin.Amount,
		},
	)

	// IBC transfer from non-FT issuer address.
	ibcTransferAndAssertBalanceChanges(
		ctx,
		t,
		coreumChain.ChainContext,
		coreumSender,
		sendCoin,
		osmosisChain.ChainContext,
		osmosisRecipient2,
		receiveCoinOsmosis,
		map[string]sdkmath.Int{
			coreumChain.MustConvertToBech32Address(coreumSender):                 sendCoin.Amount.Add(sendCommissionAmount).Add(burntAmount).Neg(),
			coreumChain.MustConvertToBech32Address(coreumIssuer):                 sendCommissionAmount,
			coreumChain.MustConvertToBech32Address(coreumToOsmosisEscrowAddress): sendCoin.Amount,
		},
		map[string]sdkmath.Int{
			osmosisChain.MustConvertToBech32Address(osmosisRecipient2): sendCoin.Amount,
		},
	)

	// ********** Gaia to Coreum (send back) **********
	// IBC transfer back to issuer address.
	ibcTransferAndAssertBalanceChanges(
		ctx,
		t,
		gaiaChain.ChainContext,
		gaiaRecipient1,
		receiveCoinGaia,
		coreumChain.ChainContext,
		coreumIssuer,
		sendCoin,
		map[string]sdkmath.Int{
			gaiaChain.MustConvertToBech32Address(gaiaRecipient1): sendCoin.Amount.Neg(),
		},
		map[string]sdkmath.Int{
			coreumChain.MustConvertToBech32Address(coreumToGaiaEscrowAddress): sendCoin.Amount.Neg(),
			coreumChain.MustConvertToBech32Address(coreumIssuer):              sendCoin.Amount,
		},
	)

	// IBC transfer back to non-issuer address.
	ibcTransferAndAssertBalanceChanges(
		ctx,
		t,
		gaiaChain.ChainContext,
		gaiaRecipient2,
		receiveCoinGaia,
		coreumChain.ChainContext,
		coreumSender,
		sendCoin,
		map[string]sdkmath.Int{
			gaiaChain.MustConvertToBech32Address(gaiaRecipient2): sendCoin.Amount.Neg(),
		},
		map[string]sdkmath.Int{
			coreumChain.MustConvertToBech32Address(coreumToGaiaEscrowAddress): sendCoin.Amount.Neg(),
			coreumChain.MustConvertToBech32Address(coreumSender):              sendCoin.Amount,
			coreumChain.MustConvertToBech32Address(coreumIssuer):              sdk.ZeroInt(),
		},
	)

	// ********** Osmosis to Coreum (send back) **********
	// IBC transfer back to issuer address.
	ibcTransferAndAssertBalanceChanges(
		ctx,
		t,
		osmosisChain.ChainContext,
		osmosisRecipient1,
		receiveCoinOsmosis,
		coreumChain.ChainContext,
		coreumIssuer,
		sendCoin,
		map[string]sdkmath.Int{
			osmosisChain.MustConvertToBech32Address(osmosisRecipient1): sendCoin.Amount.Neg(),
		},
		map[string]sdkmath.Int{
			coreumChain.MustConvertToBech32Address(coreumToOsmosisEscrowAddress): sendCoin.Amount.Neg(),
			coreumChain.MustConvertToBech32Address(coreumIssuer):                 sendCoin.Amount,
		},
	)

	// IBC transfer back to non-issuer address.
	ibcTransferAndAssertBalanceChanges(
		ctx,
		t,
		osmosisChain.ChainContext,
		osmosisRecipient2,
		receiveCoinOsmosis,
		coreumChain.ChainContext,
		coreumSender,
		sendCoin,
		map[string]sdkmath.Int{
			osmosisChain.MustConvertToBech32Address(osmosisRecipient2): sendCoin.Amount.Neg(),
		},
		map[string]sdkmath.Int{
			coreumChain.MustConvertToBech32Address(coreumToOsmosisEscrowAddress): sendCoin.Amount.Neg(),
			coreumChain.MustConvertToBech32Address(coreumSender):                 sendCoin.Amount,
			coreumChain.MustConvertToBech32Address(coreumIssuer):                 sdk.ZeroInt(),
		},
	)
}

func TestIBCAssetFTWhitelisting(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia
	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID)

	coreumIssuer := coreumChain.GenAccount()
	coreumRecipientBlocked := coreumChain.GenAccount()
	coreumRecipientWhitelisted := coreumChain.GenAccount()
	gaiaRecipient := gaiaChain.GenAccount()

	gaiaChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: gaiaRecipient,
		Amount:  gaiaChain.NewCoin(sdkmath.NewInt(1000000)), // coin for the fees
	})

	issueFee := coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount
	coreumChain.FundAccountWithOptions(ctx, t, coreumIssuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgSetWhitelistedLimit{},
			&ibctransfertypes.MsgTransfer{},
			&ibctransfertypes.MsgTransfer{},
		},
		Amount: issueFee,
	})

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        coreumIssuer.String(),
		Symbol:        "mysymbol",
		Subunit:       "mysubunit",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(1_000_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_ibc,
			assetfttypes.Feature_whitelisting,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	require.NoError(t, err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, coreumIssuer)
	sendBackCoin := sdk.NewCoin(denom, sdkmath.NewInt(1000))
	sendCoin := sdk.NewCoin(denom, sendBackCoin.Amount.MulRaw(2))

	whitelistMsg := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  coreumIssuer.String(),
		Account: coreumRecipientWhitelisted.String(),
		Coin:    sendBackCoin,
	}
	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	require.NoError(t, err)

	// send minted coins to gaia
	_, err = coreumChain.ExecuteIBCTransfer(ctx, t, coreumIssuer, sendCoin, gaiaChain.ChainContext, gaiaRecipient)
	requireT.NoError(err)

	ibcDenom := ConvertToIBCDenom(gaiaToCoreumChannelID, denom)
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaRecipient, sdk.NewCoin(ibcDenom, sendCoin.Amount)))

	// send coins back to two accounts, one blocked, one whitelisted
	ibcSendCoin := sdk.NewCoin(ibcDenom, sendBackCoin.Amount)
	_, err = gaiaChain.ExecuteIBCTransfer(ctx, t, gaiaRecipient, ibcSendCoin, coreumChain.ChainContext, coreumRecipientBlocked)
	requireT.NoError(err)
	_, err = gaiaChain.ExecuteIBCTransfer(ctx, t, gaiaRecipient, ibcSendCoin, coreumChain.ChainContext, coreumRecipientWhitelisted)
	requireT.NoError(err)

	// transfer to whitelisted account is expected to succeed
	requireT.NoError(coreumChain.AwaitForBalance(ctx, t, coreumRecipientWhitelisted, sendBackCoin))

	// transfer to blocked account is expected to fail and funds should be returned back
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaRecipient, sdk.NewCoin(ibcDenom, sendBackCoin.Amount)))

	bankClient := banktypes.NewQueryClient(coreumChain.ClientContext)
	balanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumRecipientBlocked.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdk.ZeroInt()).String(), balanceRes.Balance.String())
}

func TestIBCAssetFTFreezing(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	assertT := assert.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID)

	coreumIssuer := coreumChain.GenAccount()
	coreumSender := coreumChain.GenAccount()
	gaiaRecipient := gaiaChain.GenAccount()

	gaiaChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: gaiaRecipient,
		Amount:  gaiaChain.NewCoin(sdkmath.NewInt(1000000)), // coin for the fees
	})

	issueFee := coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount
	coreumChain.FundAccountWithOptions(ctx, t, coreumIssuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgFreeze{},
		},
		Amount: issueFee,
	})
	coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&ibctransfertypes.MsgTransfer{},
			&ibctransfertypes.MsgTransfer{},
		},
	})

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        coreumIssuer.String(),
		Symbol:        "mysymbol",
		Subunit:       "mysubunit",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(1_000_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_ibc,
			assetfttypes.Feature_freezing,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	require.NoError(t, err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, coreumIssuer)

	sendCoin := sdk.NewCoin(denom, sdkmath.NewInt(1000))
	halfCoin := sdk.NewCoin(denom, sdkmath.NewInt(500))
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

	freezeMsg := &assetfttypes.MsgFreeze{
		Sender:  coreumIssuer.String(),
		Account: coreumSender.String(),
		Coin:    halfCoin,
	}
	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	require.NoError(t, err)

	// send more than allowed, should fail
	_, err = coreumChain.ExecuteIBCTransfer(ctx, t, coreumSender, sendCoin, gaiaChain.ChainContext, gaiaRecipient)
	requireT.Error(err)
	assertT.Contains(err.Error(), cosmoserrors.ErrInsufficientFunds.Error())

	// send up to the limit, should succeed
	ibcCoin := sdk.NewCoin(ConvertToIBCDenom(gaiaToCoreumChannelID, denom), halfCoin.Amount)
	_, err = coreumChain.ExecuteIBCTransfer(ctx, t, coreumSender, halfCoin, gaiaChain.ChainContext, gaiaRecipient)
	requireT.NoError(err)
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaRecipient, ibcCoin))

	// send it back, frozen limit should not affect it
	_, err = gaiaChain.ExecuteIBCTransfer(ctx, t, gaiaRecipient, ibcCoin, coreumChain.ChainContext, coreumSender)
	requireT.NoError(err)
	requireT.NoError(coreumChain.AwaitForBalance(ctx, t, coreumSender, sendCoin))
}

func TestEscrowAddressIsResistantToFreezingAndWhitelisting(t *testing.T) {
	t.Parallel()

	requireT := require.New(t)

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID)

	coreumIssuer := coreumChain.GenAccount()
	gaiaRecipient := gaiaChain.GenAccount()

	gaiaChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: gaiaRecipient,
		Amount:  gaiaChain.NewCoin(sdkmath.NewInt(1000000)), // coin for the fees
	})

	issueFee := coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount
	coreumChain.FundAccountWithOptions(ctx, t, coreumIssuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgFreeze{},
			&ibctransfertypes.MsgTransfer{},
			&ibctransfertypes.MsgTransfer{},
		},
		Amount: issueFee,
	})

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        coreumIssuer.String(),
		Symbol:        "mysymbol",
		Subunit:       "mysubunit",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(1_000_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_ibc,
			assetfttypes.Feature_freezing,
			assetfttypes.Feature_whitelisting,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	require.NoError(t, err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, coreumIssuer)
	sendCoin := sdk.NewCoin(denom, issueMsg.InitialAmount)

	coreumToGaiaChannelID := coreumChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, gaiaChain.ChainSettings.ChainID)

	// send minted coins to gaia
	_, err = coreumChain.ExecuteIBCTransfer(ctx, t, coreumIssuer, sendCoin, gaiaChain.ChainContext, gaiaRecipient)
	requireT.NoError(err)

	ibcDenom := ConvertToIBCDenom(gaiaToCoreumChannelID, denom)
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaRecipient, sdk.NewCoin(ibcDenom, sendCoin.Amount)))

	// freeze escrow account
	coreumToGaiaEscrowAddress := ibctransfertypes.GetEscrowAddress(ibctransfertypes.PortID, coreumToGaiaChannelID)
	freezeMsg := &assetfttypes.MsgFreeze{
		Sender:  coreumIssuer.String(),
		Account: coreumToGaiaEscrowAddress.String(),
		Coin:    sendCoin,
	}
	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	require.NoError(t, err)

	// send coins back to coreum, it should succeed despite frozen escrow address
	ibcSendCoin := sdk.NewCoin(ibcDenom, sendCoin.Amount)
	_, err = gaiaChain.ExecuteIBCTransfer(ctx, t, gaiaRecipient, ibcSendCoin, coreumChain.ChainContext, coreumIssuer)
	requireT.NoError(err)
	requireT.NoError(coreumChain.AwaitForBalance(ctx, t, coreumIssuer, sendCoin))
}

func TestIBCGlobalFreeze(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID)

	coreumIssuer := coreumChain.GenAccount()
	coreumSender := coreumChain.GenAccount()
	coreumRecipient := coreumChain.GenAccount()
	gaiaRecipient := gaiaChain.GenAccount()

	gaiaChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: gaiaRecipient,
		Amount:  gaiaChain.NewCoin(sdkmath.NewInt(1000000)), // coin for the fees
	})

	issueFee := coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount
	coreumChain.FundAccountWithOptions(ctx, t, coreumIssuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgGloballyFreeze{},
			&banktypes.MsgSend{},
			&ibctransfertypes.MsgTransfer{},
			&assetfttypes.MsgGloballyUnfreeze{},
		},
		Amount: issueFee,
	})
	coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&ibctransfertypes.MsgTransfer{},
		},
	})

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        coreumIssuer.String(),
		Symbol:        "mysymbol",
		Subunit:       "mysubunit",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(1_000_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_ibc,
			assetfttypes.Feature_freezing,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	require.NoError(t, err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, coreumIssuer)
	sendCoin := sdk.NewCoin(denom, issueMsg.InitialAmount.QuoRaw(2))
	ibcSendCoin := sdk.NewCoin(ConvertToIBCDenom(gaiaToCoreumChannelID, denom), sendCoin.Amount)
	sendCoinBack := sdk.NewCoin(denom, issueMsg.InitialAmount.QuoRaw(10))
	ibcSendCoinBack := sdk.NewCoin(ConvertToIBCDenom(gaiaToCoreumChannelID, denom), sendCoinBack.Amount)

	// set global freeze
	freezeMsg := &assetfttypes.MsgGloballyFreeze{
		Sender: coreumIssuer.String(),
		Denom:  denom,
	}
	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	require.NoError(t, err)

	// send some coins to the other account, should work despite global freeze
	sendMsg := &banktypes.MsgSend{
		FromAddress: coreumIssuer.String(),
		ToAddress:   coreumSender.String(),
		Amount:      sdk.NewCoins(sendCoin),
	}
	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	require.NoError(t, err)

	// send minted coins to gaia from the other account, should fail due to global freeze
	_, err = coreumChain.ExecuteIBCTransfer(ctx, t, coreumSender, sendCoin, gaiaChain.ChainContext, gaiaRecipient)
	requireT.ErrorContains(err, assetfttypes.ErrGloballyFrozen.Error())

	// send minted coins to gaia from issuer, should succeed despite global freeze
	_, err = coreumChain.ExecuteIBCTransfer(ctx, t, coreumIssuer, sendCoin, gaiaChain.ChainContext, gaiaRecipient)
	requireT.NoError(err)
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaRecipient, ibcSendCoin))

	// send coins back to issuer on coreum, it should fail
	_, err = gaiaChain.ExecuteIBCTransfer(ctx, t, gaiaRecipient, ibcSendCoinBack, coreumChain.ChainContext, coreumIssuer)
	requireT.NoError(err)
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaRecipient, ibcSendCoin))

	bankClient := banktypes.NewQueryClient(coreumChain.ClientContext)
	balanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumIssuer.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdk.ZeroInt()).String(), balanceRes.Balance.String())

	// send coins back to recipient on coreum, it should fail
	_, err = gaiaChain.ExecuteIBCTransfer(ctx, t, gaiaRecipient, ibcSendCoinBack, coreumChain.ChainContext, coreumRecipient)
	requireT.NoError(err)
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaRecipient, ibcSendCoin))

	balanceRes, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumRecipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdk.ZeroInt()).String(), balanceRes.Balance.String())

	// remove global freeze
	unfreezeMsg := &assetfttypes.MsgGloballyUnfreeze{
		Sender: coreumIssuer.String(),
		Denom:  denom,
	}
	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(unfreezeMsg)),
		unfreezeMsg,
	)
	require.NoError(t, err)

	// send coins back to issuer on coreum again, it should succeed
	_, err = gaiaChain.ExecuteIBCTransfer(ctx, t, gaiaRecipient, ibcSendCoinBack, coreumChain.ChainContext, coreumIssuer)
	requireT.NoError(err)
	requireT.NoError(coreumChain.AwaitForBalance(ctx, t, coreumIssuer, sendCoinBack))

	// send coins back to recipient on coreum again, it should succeed
	_, err = gaiaChain.ExecuteIBCTransfer(ctx, t, gaiaRecipient, ibcSendCoinBack, coreumChain.ChainContext, coreumRecipient)
	requireT.NoError(err)
	requireT.NoError(coreumChain.AwaitForBalance(ctx, t, coreumRecipient, sendCoinBack))
}

func TestIBCAssetFTTimedOutTransfer(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Osmosis

	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID)

	retryCtx, retryCancel := context.WithTimeout(ctx, 2*time.Minute)
	defer retryCancel()

	// This is the retry loop where we try to trigger a timeout condition for IBC transfer.
	// We can't reproduce it with 100% probability, so we may need to try it many times.
	// On every trial we send funds from one chain to the other. Then we observe accounts on both chains
	// to find if IBC transfer completed successfully or timed out. If tokens were delivered to the recipient
	// we must retry. Otherwise, if tokens were returned back to the sender, we might continue the test.
	issueFee := coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount
	err := retry.Do(retryCtx, time.Millisecond, func() error {
		coreumSender := coreumChain.GenAccount()
		gaiaRecipient := gaiaChain.GenAccount()

		coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assetfttypes.MsgIssue{},
				&ibctransfertypes.MsgTransfer{},
			},
			Amount: issueFee,
		})

		issueMsg := &assetfttypes.MsgIssue{
			Issuer:        coreumSender.String(),
			Symbol:        "mysymbol",
			Subunit:       "mysubunit",
			Precision:     8,
			InitialAmount: sdkmath.NewInt(1_000_000),
			Features: []assetfttypes.Feature{
				assetfttypes.Feature_ibc,
			},
		}
		_, err := client.BroadcastTx(
			ctx,
			coreumChain.ClientContext.WithFromAddress(coreumSender),
			coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(issueMsg)),
			issueMsg,
		)
		require.NoError(t, err)
		denom := assetfttypes.BuildDenom(issueMsg.Subunit, coreumSender)
		sendToGaiaCoin := sdk.NewCoin(denom, issueMsg.InitialAmount)

		_, err = coreumChain.ExecuteTimingOutIBCTransfer(ctx, t, coreumSender, sendToGaiaCoin, gaiaChain.ChainContext, gaiaRecipient)
		switch {
		case err == nil:
		case strings.Contains(err.Error(), ibcchanneltypes.ErrPacketTimeout.Error()):
			return retry.Retryable(err)
		default:
			requireT.NoError(err)
		}

		parallelCtx, parallelCancel := context.WithCancel(ctx)
		defer parallelCancel()
		errCh := make(chan error, 1)
		go func() {
			// In this goroutine we check if funds were returned back to the sender.
			// If this happens it means timeout occurred.

			defer parallelCancel()
			if err := coreumChain.AwaitForBalance(parallelCtx, t, coreumSender, sendToGaiaCoin); err != nil {
				select {
				case errCh <- retry.Retryable(err):
				default:
				}
			} else {
				errCh <- nil
			}
		}()
		go func() {
			// In this goroutine we check if funds were delivered to the other chain.
			// If this happens it means timeout didn't occur and we must try again.

			if err := gaiaChain.AwaitForBalance(parallelCtx, t, gaiaRecipient, sdk.NewCoin(ConvertToIBCDenom(gaiaToCoreumChannelID, sendToGaiaCoin.Denom), sendToGaiaCoin.Amount)); err == nil {
				select {
				case errCh <- retry.Retryable(errors.New("timeout didn't happen")):
					parallelCancel()
				default:
				}
			}
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			if err != nil {
				return err
			}
		}

		// At this point we are sure that timeout happened.

		// funds should not be received on gaia
		bankClient := banktypes.NewQueryClient(gaiaChain.ClientContext)
		resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
			Address: gaiaChain.MustConvertToBech32Address(gaiaRecipient),
			Denom:   ConvertToIBCDenom(gaiaToCoreumChannelID, sendToGaiaCoin.Denom),
		})
		requireT.NoError(err)
		requireT.Equal("0", resp.Balance.Amount.String())

		return nil
	})
	requireT.NoError(err)
}

func TestIBCAssetFTRejectedTransfer(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID)

	// Bank module rejects transfers targeting some module accounts. We use this feature to test that
	// this type of IBC transfer is rejected by the receiving chain.
	moduleAddress := authtypes.NewModuleAddress(ibctransfertypes.ModuleName)
	coreumSender := coreumChain.GenAccount()
	gaiaRecipient := gaiaChain.GenAccount()

	coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&ibctransfertypes.MsgTransfer{},
			&ibctransfertypes.MsgTransfer{},
		},
		Amount: coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})
	gaiaChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: gaiaRecipient,
		Amount:  gaiaChain.NewCoin(sdk.NewIntFromUint64(100000)),
	})

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        coreumSender.String(),
		Symbol:        "mysymbol",
		Subunit:       "mysubunit",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(1_000_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_ibc,
			assetfttypes.Feature_freezing,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumSender),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	require.NoError(t, err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, coreumSender)
	sendToGaiaCoin := sdk.NewCoin(denom, issueMsg.InitialAmount)

	_, err = coreumChain.ExecuteIBCTransfer(ctx, t, coreumSender, sendToGaiaCoin, gaiaChain.ChainContext, moduleAddress)
	requireT.NoError(err)

	// funds should be returned to coreum
	requireT.NoError(coreumChain.AwaitForBalance(ctx, t, coreumSender, sendToGaiaCoin))

	// funds should not be received on gaia
	ibcGaiaDenom := ConvertToIBCDenom(gaiaToCoreumChannelID, sendToGaiaCoin.Denom)
	bankClient := banktypes.NewQueryClient(gaiaChain.ClientContext)
	resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: gaiaChain.MustConvertToBech32Address(moduleAddress),
		Denom:   ibcGaiaDenom,
	})
	requireT.NoError(err)
	requireT.Equal("0", resp.Balance.Amount.String())

	// test that the reverse transfer from gaia to coreum is blocked too

	coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
	})

	sendToCoreumCoin := sdk.NewCoin(ibcGaiaDenom, sendToGaiaCoin.Amount)
	_, err = coreumChain.ExecuteIBCTransfer(ctx, t, coreumSender, sendToGaiaCoin, gaiaChain.ChainContext, gaiaRecipient)
	requireT.NoError(err)
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaRecipient, sendToCoreumCoin))

	_, err = gaiaChain.ExecuteIBCTransfer(ctx, t, gaiaRecipient, sendToCoreumCoin, coreumChain.ChainContext, moduleAddress)
	requireT.NoError(err)

	// funds should be returned to gaia
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaRecipient, sendToCoreumCoin))

	// funds should not be received on coreum
	bankClient = banktypes.NewQueryClient(coreumChain.ClientContext)
	resp, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumChain.MustConvertToBech32Address(moduleAddress),
		Denom:   sendToGaiaCoin.Denom,
	})
	requireT.NoError(err)
	requireT.Equal("0", resp.Balance.Amount.String())
}

func TestIBCRejectedTransferWithWhitelistingAndFreezing(t *testing.T) {
	t.Parallel()

	requireT := require.New(t)

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	coreumIssuer := coreumChain.GenAccount()
	coreumSender := coreumChain.GenAccount()
	// Bank module rejects transfers targeting some module accounts. We use this feature to test that
	// this type of IBC transfer is rejected by the receiving chain.
	moduleAddress := authtypes.NewModuleAddress(ibctransfertypes.ModuleName)

	issueFee := coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount
	coreumChain.FundAccountWithOptions(ctx, t, coreumIssuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgFreeze{},
			&assetfttypes.MsgSetWhitelistedLimit{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgSetWhitelistedLimit{},
		},
		Amount: issueFee,
	})
	coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&ibctransfertypes.MsgTransfer{},
		},
	})

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        coreumIssuer.String(),
		Symbol:        "mysymbol",
		Subunit:       "mysubunit",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(1_000_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_ibc,
			assetfttypes.Feature_freezing,
			assetfttypes.Feature_whitelisting,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	require.NoError(t, err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, coreumIssuer)
	sendCoin := sdk.NewCoin(denom, issueMsg.InitialAmount)

	coreumToGaiaChannelID := coreumChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, gaiaChain.ChainSettings.ChainID)

	// freeze escrow account
	coreumToGaiaEscrowAddress := ibctransfertypes.GetEscrowAddress(ibctransfertypes.PortID, coreumToGaiaChannelID)
	freezeMsg := &assetfttypes.MsgFreeze{
		Sender:  coreumIssuer.String(),
		Account: coreumToGaiaEscrowAddress.String(),
		Coin:    sendCoin,
	}
	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	require.NoError(t, err)

	// whitelist sender
	whitelistMsg := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  coreumIssuer.String(),
		Account: coreumSender.String(),
		Coin:    sendCoin,
	}
	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	require.NoError(t, err)

	// send coins from issuer to sender
	sendMsg := &banktypes.MsgSend{
		FromAddress: coreumIssuer.String(),
		ToAddress:   coreumSender.String(),
		Amount:      sdk.NewCoins(sendCoin),
	}
	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	require.NoError(t, err)

	// blacklist sender
	blacklistMsg := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  coreumIssuer.String(),
		Account: coreumSender.String(),
		Coin:    sdk.NewInt64Coin(sendCoin.Denom, 0),
	}
	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(blacklistMsg)),
		blacklistMsg,
	)
	require.NoError(t, err)

	// send coins from sender to blocked address on gaia
	_, err = coreumChain.ExecuteIBCTransfer(ctx, t, coreumSender, sendCoin, gaiaChain.ChainContext, moduleAddress)
	requireT.NoError(err)

	// gaia should reject the IBC transfers and funds should be returned back to coreum, despite:
	// - escrow address being frozen
	// - sender account not being whitelisted
	requireT.NoError(coreumChain.AwaitForBalance(ctx, t, coreumSender, sendCoin))
}

func TestIBCTimedOutTransferWithWhitelistingAndFreezing(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID)

	retryCtx, retryCancel := context.WithTimeout(ctx, 2*time.Minute)
	defer retryCancel()

	// This is the retry loop where we try to trigger a timeout condition for IBC transfer.
	// We can't reproduce it with 100% probability, so we may need to try it many times.
	// On every trial we send funds from one chain to the other. Then we observe accounts on both chains
	// to find if IBC transfer completed successfully or timed out. If tokens were delivered to the recipient
	// we must retry. Otherwise, if tokens were returned back to the sender, we might continue the test.
	issueFee := coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount
	err := retry.Do(retryCtx, time.Millisecond, func() error {
		coreumIssuer := coreumChain.GenAccount()
		coreumSender := coreumChain.GenAccount()
		gaiaRecipient := gaiaChain.GenAccount()

		coreumChain.FundAccountWithOptions(ctx, t, coreumIssuer, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assetfttypes.MsgIssue{},
				&assetfttypes.MsgFreeze{},
				&assetfttypes.MsgSetWhitelistedLimit{},
				&banktypes.MsgSend{},
				&assetfttypes.MsgSetWhitelistedLimit{},
			},
			Amount: issueFee,
		})
		coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&ibctransfertypes.MsgTransfer{},
			},
		})

		issueMsg := &assetfttypes.MsgIssue{
			Issuer:        coreumIssuer.String(),
			Symbol:        "mysymbol",
			Subunit:       "mysubunit",
			Precision:     8,
			InitialAmount: sdkmath.NewInt(1_000_000),
			Features: []assetfttypes.Feature{
				assetfttypes.Feature_ibc,
				assetfttypes.Feature_whitelisting,
				assetfttypes.Feature_freezing,
			},
		}
		_, err := client.BroadcastTx(
			ctx,
			coreumChain.ClientContext.WithFromAddress(coreumIssuer),
			coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(issueMsg)),
			issueMsg,
		)
		require.NoError(t, err)
		denom := assetfttypes.BuildDenom(issueMsg.Subunit, coreumIssuer)
		sendToGaiaCoin := sdk.NewCoin(denom, issueMsg.InitialAmount)

		coreumToGaiaChannelID := coreumChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, gaiaChain.ChainSettings.ChainID)

		// freeze escrow account
		coreumToGaiaEscrowAddress := ibctransfertypes.GetEscrowAddress(ibctransfertypes.PortID, coreumToGaiaChannelID)
		freezeMsg := &assetfttypes.MsgFreeze{
			Sender:  coreumIssuer.String(),
			Account: coreumToGaiaEscrowAddress.String(),
			Coin:    sendToGaiaCoin,
		}
		_, err = client.BroadcastTx(
			ctx,
			coreumChain.ClientContext.WithFromAddress(coreumIssuer),
			coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(freezeMsg)),
			freezeMsg,
		)
		require.NoError(t, err)

		// whitelist sender
		whitelistMsg := &assetfttypes.MsgSetWhitelistedLimit{
			Sender:  coreumIssuer.String(),
			Account: coreumSender.String(),
			Coin:    sendToGaiaCoin,
		}
		_, err = client.BroadcastTx(
			ctx,
			coreumChain.ClientContext.WithFromAddress(coreumIssuer),
			coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(whitelistMsg)),
			whitelistMsg,
		)
		require.NoError(t, err)

		// send coins from issuer to sender
		sendMsg := &banktypes.MsgSend{
			FromAddress: coreumIssuer.String(),
			ToAddress:   coreumSender.String(),
			Amount:      sdk.NewCoins(sendToGaiaCoin),
		}
		_, err = client.BroadcastTx(
			ctx,
			coreumChain.ClientContext.WithFromAddress(coreumIssuer),
			coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(sendMsg)),
			sendMsg,
		)
		require.NoError(t, err)

		// blacklist sender
		blacklistMsg := &assetfttypes.MsgSetWhitelistedLimit{
			Sender:  coreumIssuer.String(),
			Account: coreumSender.String(),
			Coin:    sdk.NewInt64Coin(sendToGaiaCoin.Denom, 0),
		}
		_, err = client.BroadcastTx(
			ctx,
			coreumChain.ClientContext.WithFromAddress(coreumIssuer),
			coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(blacklistMsg)),
			blacklistMsg,
		)
		require.NoError(t, err)

		_, err = coreumChain.ExecuteTimingOutIBCTransfer(ctx, t, coreumSender, sendToGaiaCoin, gaiaChain.ChainContext, gaiaRecipient)
		switch {
		case err == nil:
		case strings.Contains(err.Error(), ibcchanneltypes.ErrPacketTimeout.Error()):
			return retry.Retryable(err)
		default:
			requireT.NoError(err)
		}

		parallelCtx, parallelCancel := context.WithCancel(ctx)
		defer parallelCancel()
		errCh := make(chan error, 1)
		go func() {
			// In this goroutine we check if funds were returned back to the sender.
			// If this happens it means timeout occurred.

			defer parallelCancel()
			if err := coreumChain.AwaitForBalance(parallelCtx, t, coreumSender, sendToGaiaCoin); err != nil {
				select {
				case errCh <- retry.Retryable(err):
				default:
				}
			} else {
				errCh <- nil
			}
		}()
		go func() {
			// In this goroutine we check if funds were delivered to the other chain.
			// If this happens it means timeout didn't occur and we must try again.

			if err := gaiaChain.AwaitForBalance(parallelCtx, t, gaiaRecipient, sdk.NewCoin(ConvertToIBCDenom(gaiaToCoreumChannelID, sendToGaiaCoin.Denom), sendToGaiaCoin.Amount)); err == nil {
				select {
				case errCh <- retry.Retryable(errors.New("timeout didn't happen")):
					parallelCancel()
				default:
				}
			}
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			if err != nil {
				return err
			}
		}

		// At this point we are sure that timeout happened and coins has been sent back to the sender.
		return nil
	})
	requireT.NoError(err)
}

func TestIBCRejectedTransferWithBurnRateAndSendCommission(t *testing.T) {
	t.Parallel()

	requireT := require.New(t)

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	coreumIssuer := coreumChain.GenAccount()
	coreumSender := coreumChain.GenAccount()
	// Bank module rejects transfers targeting some module accounts. We use this feature to test that
	// this type of IBC transfer is rejected by the receiving chain.
	moduleAddress := authtypes.NewModuleAddress(ibctransfertypes.ModuleName)

	issueFee := coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount
	coreumChain.FundAccountWithOptions(ctx, t, coreumIssuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
		},
		Amount: issueFee,
	})
	coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&ibctransfertypes.MsgTransfer{},
		},
	})

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        coreumIssuer.String(),
		Symbol:        "mysymbol",
		Subunit:       "mysubunit",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(910_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_ibc,
		},
		BurnRate:           sdk.MustNewDecFromStr("0.1"),
		SendCommissionRate: sdk.MustNewDecFromStr("0.2"),
	}
	_, err := client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	require.NoError(t, err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, coreumIssuer)

	// send coins from issuer to sender
	sendMsg := &banktypes.MsgSend{
		FromAddress: coreumIssuer.String(),
		ToAddress:   coreumSender.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, issueMsg.InitialAmount)),
	}
	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	require.NoError(t, err)

	// Send coins from sender to blocked address on Gaia.
	// We send everything except amount required to cover burn rate and send commission.
	sendCoin := sdk.NewCoin(denom, sdk.NewDecFromInt(issueMsg.InitialAmount).Quo(sdk.OneDec().Add(issueMsg.BurnRate).Add(issueMsg.SendCommissionRate)).TruncateInt())
	_, err = coreumChain.ExecuteIBCTransfer(ctx, t, coreumSender, sendCoin, gaiaChain.ChainContext, moduleAddress)
	requireT.NoError(err)

	// Gaia should reject the IBC transfers and funds should be returned back to Coreum.
	// Burn rate and send commission should be charged only once when IBC transfer is requested (we will probably change this in the future),
	// but when IBC transfer is rolled back, rates should not be charged again.
	requireT.NoError(coreumChain.AwaitForBalance(ctx, t, coreumSender, sendCoin))

	// Balance on escrow address should be 0.
	coreumToGaiaChannelID := coreumChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, gaiaChain.ChainSettings.ChainID)
	coreumToGaiaEscrowAddress := ibctransfertypes.GetEscrowAddress(ibctransfertypes.PortID, coreumToGaiaChannelID)
	bankClient := banktypes.NewQueryClient(coreumChain.ClientContext)
	balanceResp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumToGaiaEscrowAddress.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal("0", balanceResp.Balance.Amount.String())
}

func TestIBCTimedOutTransferWithBurnRateAndSendCommission(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID)

	retryCtx, retryCancel := context.WithTimeout(ctx, 2*time.Minute)
	defer retryCancel()

	// This is the retry loop where we try to trigger a timeout condition for IBC transfer.
	// We can't reproduce it with 100% probability, so we may need to try it many times.
	// On every trial we send funds from one chain to the other. Then we observe accounts on both chains
	// to find if IBC transfer completed successfully or timed out. If tokens were delivered to the recipient
	// we must retry. Otherwise, if tokens were returned back to the sender, we might continue the test.
	issueFee := coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount
	err := retry.Do(retryCtx, time.Millisecond, func() error {
		coreumIssuer := coreumChain.GenAccount()
		coreumSender := coreumChain.GenAccount()
		gaiaRecipient := gaiaChain.GenAccount()

		coreumChain.FundAccountWithOptions(ctx, t, coreumIssuer, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assetfttypes.MsgIssue{},
				&banktypes.MsgSend{},
			},
			Amount: issueFee,
		})
		coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&ibctransfertypes.MsgTransfer{},
			},
		})

		issueMsg := &assetfttypes.MsgIssue{
			Issuer:        coreumIssuer.String(),
			Symbol:        "mysymbol",
			Subunit:       "mysubunit",
			Precision:     8,
			InitialAmount: sdkmath.NewInt(910_000),
			Features: []assetfttypes.Feature{
				assetfttypes.Feature_ibc,
			},
			BurnRate:           sdk.MustNewDecFromStr("0.1"),
			SendCommissionRate: sdk.MustNewDecFromStr("0.2"),
		}
		_, err := client.BroadcastTx(
			ctx,
			coreumChain.ClientContext.WithFromAddress(coreumIssuer),
			coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(issueMsg)),
			issueMsg,
		)
		require.NoError(t, err)
		denom := assetfttypes.BuildDenom(issueMsg.Subunit, coreumIssuer)

		// send coins from issuer to sender
		sendMsg := &banktypes.MsgSend{
			FromAddress: coreumIssuer.String(),
			ToAddress:   coreumSender.String(),
			Amount:      sdk.NewCoins(sdk.NewCoin(denom, issueMsg.InitialAmount)),
		}
		_, err = client.BroadcastTx(
			ctx,
			coreumChain.ClientContext.WithFromAddress(coreumIssuer),
			coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(sendMsg)),
			sendMsg,
		)
		require.NoError(t, err)

		// Send coins from sender to Gaia.
		// We send everything except amount required to cover burn rate and send commission.
		sendCoin := sdk.NewCoin(denom, sdk.NewDecFromInt(issueMsg.InitialAmount).Quo(sdk.OneDec().Add(issueMsg.BurnRate).Add(issueMsg.SendCommissionRate)).TruncateInt())

		_, err = coreumChain.ExecuteTimingOutIBCTransfer(ctx, t, coreumSender, sendCoin, gaiaChain.ChainContext, gaiaRecipient)
		switch {
		case err == nil:
		case strings.Contains(err.Error(), ibcchanneltypes.ErrPacketTimeout.Error()):
			return retry.Retryable(err)
		default:
			requireT.NoError(err)
		}

		parallelCtx, parallelCancel := context.WithCancel(ctx)
		defer parallelCancel()
		errCh := make(chan error, 1)
		go func() {
			// In this goroutine we check if funds were returned back to the sender.
			// If this happens it means timeout occurred.

			defer parallelCancel()
			if err := coreumChain.AwaitForBalance(parallelCtx, t, coreumSender, sendCoin); err != nil {
				select {
				case errCh <- retry.Retryable(err):
				default:
				}
			} else {
				errCh <- nil
			}
		}()
		go func() {
			// In this goroutine we check if funds were delivered to the other chain.
			// If this happens it means timeout didn't occur and we must try again.

			if err := gaiaChain.AwaitForBalance(parallelCtx, t, gaiaRecipient, sdk.NewCoin(ConvertToIBCDenom(gaiaToCoreumChannelID, sendCoin.Denom), sendCoin.Amount)); err == nil {
				select {
				case errCh <- retry.Retryable(errors.New("timeout didn't happen")):
					parallelCancel()
				default:
				}
			}
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			if err != nil {
				return err
			}
		}

		// At this point we are sure that timeout happened and coins has been sent back to the sender.

		// Balance on escrow address should be 0.
		coreumToGaiaChannelID := coreumChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, gaiaChain.ChainSettings.ChainID)
		coreumToGaiaEscrowAddress := ibctransfertypes.GetEscrowAddress(ibctransfertypes.PortID, coreumToGaiaChannelID)
		bankClient := banktypes.NewQueryClient(coreumChain.ClientContext)
		balanceResp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
			Address: coreumToGaiaEscrowAddress.String(),
			Denom:   denom,
		})
		requireT.NoError(err)
		requireT.Equal("0", balanceResp.Balance.Amount.String())

		return nil
	})
	requireT.NoError(err)
}

func ibcTransferAndAssertBalanceChanges(
	ctx context.Context,
	t *testing.T,
	srcChainCtx integrationtests.ChainContext,
	srcSender sdk.AccAddress,
	sendCoin sdk.Coin,
	dstChainCtx integrationtests.ChainContext,
	dstChainRecipient sdk.AccAddress,
	receiveCoin sdk.Coin,
	srcExpectedBalanceChanges map[string]sdkmath.Int,
	dstExpectedBalanceChanges map[string]sdkmath.Int,
) {
	t.Helper()

	requireT := require.New(t)

	srcBalancesBeforeOperation := fetchBalanceForMultipleAddresses(ctx, t, srcChainCtx, sendCoin.Denom, lo.Keys(srcExpectedBalanceChanges))
	dstBalancesBeforeOperation := fetchBalanceForMultipleAddresses(ctx, t, dstChainCtx, receiveCoin.Denom, lo.Keys(dstExpectedBalanceChanges))

	dstBankClient := banktypes.NewQueryClient(dstChainCtx.ClientContext)
	dstChainRecipientBalanceBefore, err := dstBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: dstChainCtx.MustConvertToBech32Address(dstChainRecipient),
		Denom:   receiveCoin.Denom,
	})
	requireT.NoError(err)
	dstChainRecipientBalanceExpected := dstChainRecipientBalanceBefore.Balance.Add(receiveCoin)

	_, err = srcChainCtx.ExecuteIBCTransfer(ctx, t, srcSender, sendCoin, dstChainCtx, dstChainRecipient)
	requireT.NoError(err)
	requireT.NoError(dstChainCtx.AwaitForBalance(ctx, t, dstChainRecipient, dstChainRecipientBalanceExpected))

	srcBalancesAfterOperation := fetchBalanceForMultipleAddresses(ctx, t, srcChainCtx, sendCoin.Denom, lo.Keys(srcExpectedBalanceChanges))
	dstBalancesAfterOperation := fetchBalanceForMultipleAddresses(ctx, t, dstChainCtx, receiveCoin.Denom, lo.Keys(dstExpectedBalanceChanges))

	assertBalanceChanges(t, srcExpectedBalanceChanges, srcBalancesBeforeOperation, srcBalancesAfterOperation)
	assertBalanceChanges(t, dstExpectedBalanceChanges, dstBalancesBeforeOperation, dstBalancesAfterOperation)
}

func fetchBalanceForMultipleAddresses(
	ctx context.Context,
	t *testing.T,
	chainCtx integrationtests.ChainContext,
	denom string,
	addresses []string,
) map[string]sdkmath.Int {
	requireT := require.New(t)
	bankClient := banktypes.NewQueryClient(chainCtx.ClientContext)
	balances := make(map[string]sdkmath.Int, len(addresses))

	for _, addr := range addresses {
		balance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
			Address: addr,
			Denom:   denom,
		})
		requireT.NoError(err)
		requireT.NotNil(balance.Balance)
		balances[addr] = balance.Balance.Amount
	}

	return balances
}

func assertBalanceChanges(t *testing.T, expectedBalanceChanges, balancesBefore, balancesAfter map[string]sdkmath.Int) {
	requireT := require.New(t)

	for addr, expectedBalanceChange := range expectedBalanceChanges {
		actualBalanceChange := balancesAfter[addr].Sub(balancesBefore[addr])
		requireT.Equal(expectedBalanceChange.String(), actualBalanceChange.String())
	}
}
