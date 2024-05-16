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
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
	"github.com/CoreumFoundation/coreum/v4/pkg/client"
	"github.com/CoreumFoundation/coreum/v4/testutil/integration"
	testcontracts "github.com/CoreumFoundation/coreum/v4/x/asset/ft/keeper/test-contracts"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

func TestExtensionIBCFailsIfNotEnabled(t *testing.T) {
	t.Parallel()

	requireT := require.New(t)

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	coreumChain := chains.Coreum
	coreumIssuer := coreumChain.GenAccount()

	issueFee := coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount
	coreumChain.FundAccountWithOptions(ctx, t, coreumIssuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&ibctransfertypes.MsgTransfer{},
		},
		Amount: issueFee.Add(sdk.NewInt(1_000_000)), // added one million for contract upload.
	})

	codeID, err := chains.Coreum.Wasm.DeployWASMContract(
		ctx, chains.Coreum.TxFactory().WithSimulateAndExecute(true), coreumIssuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        coreumIssuer.String(),
		Symbol:        "mysymbol",
		Subunit:       "mysubunit",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(1_000_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Funds:  sdk.NewCoins(),
			Label:  "testing-ibc",
		},
	}
	_, err = client.BroadcastTx(
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
	requireT.ErrorContains(err, "IBC feature is disabled.")
}

func TestExtensionIBCAssetFTWhitelisting(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia
	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(
		ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID,
	)

	coreumIssuer := coreumChain.GenAccount()
	coreumRecipientBlocked := coreumChain.GenAccount()
	coreumRecipientWhitelisted := coreumChain.GenAccount()
	gaiaRecipient := gaiaChain.GenAccount()

	gaiaChain.Faucet.FundAccounts(ctx, t, integration.FundedAccount{
		Address: gaiaRecipient,
		Amount:  gaiaChain.NewCoin(sdkmath.NewInt(1000000)), // coin for the fees
	})

	issueFee := coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount
	coreumChain.FundAccountWithOptions(ctx, t, coreumIssuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgSetWhitelistedLimit{},
			&ibctransfertypes.MsgTransfer{},
			&ibctransfertypes.MsgTransfer{},
		},
		Amount: issueFee.Add(sdk.NewInt(1_000_000)), // added one million for contract upload
	})

	codeID, err := chains.Coreum.Wasm.DeployWASMContract(
		ctx, chains.Coreum.TxFactory().WithSimulateAndExecute(true), coreumIssuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        coreumIssuer.String(),
		Symbol:        "mysymbol",
		Subunit:       "mysubunit",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(1_000_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_block_smart_contracts,
			assetfttypes.Feature_ibc,
			assetfttypes.Feature_whitelisting,
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Funds:  sdk.NewCoins(),
			Label:  "testing-ibc",
		},
	}
	_, err = client.BroadcastTx(
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
	_, err = gaiaChain.ExecuteIBCTransfer(
		ctx, t, gaiaRecipient, ibcSendCoin, coreumChain.ChainContext, coreumRecipientBlocked,
	)
	requireT.NoError(err)
	_, err = gaiaChain.ExecuteIBCTransfer(
		ctx, t, gaiaRecipient, ibcSendCoin, coreumChain.ChainContext, coreumRecipientWhitelisted,
	)
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

func TestExtensionIBCAssetFTFreezing(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	assertT := assert.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(
		ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID,
	)

	coreumIssuer := coreumChain.GenAccount()
	coreumSender := coreumChain.GenAccount()
	gaiaRecipient := gaiaChain.GenAccount()

	gaiaChain.Faucet.FundAccounts(ctx, t, integration.FundedAccount{
		Address: gaiaRecipient,
		Amount:  gaiaChain.NewCoin(sdkmath.NewInt(1000000)), // coin for the fees
	})

	issueFee := coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount
	coreumChain.FundAccountWithOptions(ctx, t, coreumIssuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgFreeze{},
		},
		Amount: issueFee.Add(sdk.NewInt(1_000_000)), // added one million for contract upload
	})
	coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&ibctransfertypes.MsgTransfer{},
			&ibctransfertypes.MsgTransfer{},
		},
	})

	codeID, err := chains.Coreum.Wasm.DeployWASMContract(
		ctx, chains.Coreum.TxFactory().WithSimulateAndExecute(true), coreumIssuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        coreumIssuer.String(),
		Symbol:        "mysymbol",
		Subunit:       "mysubunit",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(1_000_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_block_smart_contracts,
			assetfttypes.Feature_ibc,
			assetfttypes.Feature_freezing,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Funds:  sdk.NewCoins(),
			Label:  "testing-ibc",
		},
	}
	_, err = client.BroadcastTx(
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

func TestExtensionEscrowAddressIsResistantToFreezingAndWhitelisting(t *testing.T) {
	t.Parallel()

	requireT := require.New(t)

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(
		ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID,
	)

	coreumIssuer := coreumChain.GenAccount()
	gaiaRecipient := gaiaChain.GenAccount()

	gaiaChain.Faucet.FundAccounts(ctx, t, integration.FundedAccount{
		Address: gaiaRecipient,
		Amount:  gaiaChain.NewCoin(sdkmath.NewInt(1000000)), // coin for the fees
	})

	issueFee := coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount
	coreumChain.FundAccountWithOptions(ctx, t, coreumIssuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgFreeze{},
			&assetfttypes.MsgSetWhitelistedLimit{},
			&ibctransfertypes.MsgTransfer{},
			&ibctransfertypes.MsgTransfer{},
		},
		Amount: issueFee.Add(sdk.NewInt(1_000_000)), // added one million for contract upload
	})

	codeID, err := chains.Coreum.Wasm.DeployWASMContract(
		ctx, chains.Coreum.TxFactory().WithSimulateAndExecute(true), coreumIssuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        coreumIssuer.String(),
		Symbol:        "mysymbol",
		Subunit:       "mysubunit",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(1_000_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_block_smart_contracts,
			assetfttypes.Feature_ibc,
			assetfttypes.Feature_freezing,
			assetfttypes.Feature_whitelisting,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Funds:  sdk.NewCoins(),
			Label:  "testing-ibc",
		},
	}
	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	require.NoError(t, err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, coreumIssuer)
	sendCoin := sdk.NewCoin(denom, issueMsg.InitialAmount)

	coreumToGaiaChannelID := coreumChain.AwaitForIBCChannelID(
		ctx, t, ibctransfertypes.PortID, gaiaChain.ChainSettings.ChainID,
	)

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
	coreumRecipient := chains.Coreum.GenAccount()
	whitelistMsg := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  coreumIssuer.String(),
		Account: coreumRecipient.String(),
		Coin:    sendCoin,
	}
	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	require.NoError(t, err)
	ibcSendCoin := sdk.NewCoin(ibcDenom, sendCoin.Amount)
	_, err = gaiaChain.ExecuteIBCTransfer(ctx, t, gaiaRecipient, ibcSendCoin, coreumChain.ChainContext, coreumRecipient)
	requireT.NoError(err)
	requireT.NoError(coreumChain.AwaitForBalance(ctx, t, coreumRecipient, sendCoin))
}

func TestExtensionIBCGlobalFreeze(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(
		ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID,
	)

	coreumIssuer := coreumChain.GenAccount()
	coreumSender := coreumChain.GenAccount()
	coreumRecipient := coreumChain.GenAccount()
	gaiaRecipient := gaiaChain.GenAccount()

	gaiaChain.Faucet.FundAccounts(ctx, t, integration.FundedAccount{
		Address: gaiaRecipient,
		Amount:  gaiaChain.NewCoin(sdkmath.NewInt(1000000)), // coin for the fees
	})

	issueFee := coreumChain.QueryAssetFTParams(ctx, t).IssueFee.Amount
	coreumChain.FundAccountWithOptions(ctx, t, coreumIssuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgGloballyFreeze{},
			&banktypes.MsgSend{},
			&ibctransfertypes.MsgTransfer{},
			&assetfttypes.MsgGloballyUnfreeze{},
		},
		Amount: issueFee,
	})
	coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integration.BalancesOptions{
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
			assetfttypes.Feature_block_smart_contracts,
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
	_, err = gaiaChain.ExecuteIBCTransfer(
		ctx, t, gaiaRecipient, ibcSendCoinBack, coreumChain.ChainContext, coreumRecipient,
	)
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
	_, err = gaiaChain.ExecuteIBCTransfer(
		ctx, t, gaiaRecipient, ibcSendCoinBack, coreumChain.ChainContext, coreumIssuer,
	)
	requireT.NoError(err)
	requireT.NoError(coreumChain.AwaitForBalance(ctx, t, coreumIssuer, sendCoinBack))

	// send coins back to recipient on coreum again, it should succeed
	_, err = gaiaChain.ExecuteIBCTransfer(
		ctx, t, gaiaRecipient, ibcSendCoinBack, coreumChain.ChainContext, coreumRecipient,
	)
	requireT.NoError(err)
	requireT.NoError(coreumChain.AwaitForBalance(ctx, t, coreumRecipient, sendCoinBack))
}

func TestExtensionIBCAssetFTTimedOutTransfer(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Osmosis

	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(
		ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID,
	)

	retryCtx, retryCancel := context.WithTimeout(ctx, 5*integration.AwaitForBalanceTimeout)
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

		coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integration.BalancesOptions{
			Messages: []sdk.Msg{
				&assetfttypes.MsgIssue{},
				&ibctransfertypes.MsgTransfer{},
			},
			Amount: issueFee.Add(sdk.NewInt(1_000_000)), // added one million for contract upload
		})

		codeID, err := chains.Coreum.Wasm.DeployWASMContract(
			ctx, chains.Coreum.TxFactory().WithSimulateAndExecute(true), coreumSender, testcontracts.AssetExtensionWasm,
		)
		requireT.NoError(err)

		issueMsg := &assetfttypes.MsgIssue{
			Issuer:        coreumSender.String(),
			Symbol:        "mysymbol",
			Subunit:       "mysubunit",
			Precision:     8,
			InitialAmount: sdkmath.NewInt(1_000_000),
			Features: []assetfttypes.Feature{
				assetfttypes.Feature_block_smart_contracts,
				assetfttypes.Feature_ibc,
			},
			ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
				CodeId: codeID,
				Funds:  sdk.NewCoins(),
				Label:  "testing-ibc",
			},
		}
		_, err = client.BroadcastTx(
			ctx,
			coreumChain.ClientContext.WithFromAddress(coreumSender),
			coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(issueMsg)),
			issueMsg,
		)
		require.NoError(t, err)
		denom := assetfttypes.BuildDenom(issueMsg.Subunit, coreumSender)
		sendToGaiaCoin := sdk.NewCoin(denom, issueMsg.InitialAmount)

		_, err = coreumChain.ExecuteTimingOutIBCTransfer(
			ctx, t, coreumSender, sendToGaiaCoin, gaiaChain.ChainContext, gaiaRecipient,
		)
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

			if err := gaiaChain.AwaitForBalance(
				parallelCtx,
				t,
				gaiaRecipient,
				sdk.NewCoin(ConvertToIBCDenom(gaiaToCoreumChannelID, sendToGaiaCoin.Denom), sendToGaiaCoin.Amount),
			); err == nil {
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
