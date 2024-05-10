//go:build integrationtests

package modules

import (
	"encoding/json"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
	moduleswasm "github.com/CoreumFoundation/coreum/v4/integration-tests/contracts/modules"
	"github.com/CoreumFoundation/coreum/v4/pkg/client"
	"github.com/CoreumFoundation/coreum/v4/testutil/event"
	"github.com/CoreumFoundation/coreum/v4/testutil/integration"
	assetftkeeper "github.com/CoreumFoundation/coreum/v4/x/asset/ft/keeper"
	testcontracts "github.com/CoreumFoundation/coreum/v4/x/asset/ft/keeper/test-contracts"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

const (
	MagicAmountDisallowed = 7
	MagicAmountBurning    = 101
	MagicAmountMinting    = 105
)

// TestAssetFTExtensionIssue tests extension issue functionality of fungible tokens.
func TestAssetFTExtensionIssue(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	assetFTClient := assetfttypes.NewQueryClient(chain.ClientContext)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	wasmClient := wasmtypes.NewQueryClient(chain.ClientContext)
	requireT := require.New(t)

	issuer := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
			Add(sdk.NewInt(1_000_000)), // one million added for uploading wasm code
	})

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactory().WithSimulateAndExecute(true), issuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)

	attachedFund := chain.NewCoin(sdk.NewInt(10))
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_extension,
		},
		URI:     "https://my-class-meta.invalid/1",
		URIHash: "content-hash",
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Funds:  sdk.NewCoins(attachedFund),
			Label:  "testing-issuance",
		},
	}

	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	// assert that attached funds are transferred to the contract
	token, err := assetFTClient.Token(ctx, &assetfttypes.QueryTokenRequest{Denom: denom})
	requireT.NoError(err)
	contractBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: token.Token.ExtensionCWAddress,
		Denom:   chain.ChainSettings.Denom,
	})
	requireT.NoError(err)
	requireT.EqualValues(contractBalance.GetBalance().String(), attachedFund.String())

	// assert correct label is applied.
	contractInfo, err := wasmClient.ContractInfo(
		ctx, &wasmtypes.QueryContractInfoRequest{Address: token.Token.ExtensionCWAddress},
	)
	requireT.NoError(err)
	requireT.EqualValues(issueMsg.ExtensionSettings.Label, contractInfo.Label)

	recipient := chain.GenAccount()
	// sending 1 will succeed
	sendMsg := &banktypes.MsgSend{
		FromAddress: issueMsg.Issuer,
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(12))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	balance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues("12", balance.Balance.Amount.String())

	// sending 7 will fail
	sendMsg.Amount = sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(MagicAmountDisallowed)))
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.ErrorIs(err, assetfttypes.ErrExtensionCallFailed)
	balance, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues("12", balance.Balance.Amount.String())

	// call directly from the user
	// sending 1 will succeed
	chain.FundAccountWithOptions(ctx, t, recipient, integration.BalancesOptions{
		Amount: sdk.NewInt(1000_000),
	})

	recipient2 := chain.GenAccount()
	contractMsg := map[string]interface{}{
		assetftkeeper.ExtenstionTransferMethod: assetftkeeper.ExtensionTransferMsg{
			Amount:    sdk.NewInt(1),
			Recipient: recipient2.String(),
		},
	}
	contractMsgBytes, err := json.Marshal(contractMsg)
	requireT.NoError(err)
	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		recipient,
		token.Token.ExtensionCWAddress,
		contractMsgBytes,
		sdk.NewCoin(denom, sdk.NewInt(1)),
	)
	requireT.NoError(err)

	balance, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient2.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues("1", balance.Balance.Amount.String())

	// sending 7 will fail
	requireT.NoError(err)
	contractMsg = map[string]interface{}{
		assetftkeeper.ExtenstionTransferMethod: assetftkeeper.ExtensionTransferMsg{
			Amount:    sdk.NewInt(MagicAmountDisallowed),
			Recipient: recipient2.String(),
		},
	}
	contractMsgBytes, err = json.Marshal(contractMsg)
	requireT.NoError(err)
	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		recipient,
		token.Token.ExtensionCWAddress,
		contractMsgBytes,
		sdk.NewCoin(denom, sdk.NewInt(MagicAmountDisallowed)),
	)
	requireT.Error(err)
}

// TestAssetFTExtensionWhitelist checks extension whitelist functionality of fungible tokens.
func TestAssetFTExtensionWhitelist(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	clientCtx := chain.ClientContext

	ftClient := assetfttypes.NewQueryClient(clientCtx)
	bankClient := banktypes.NewQueryClient(clientCtx)

	issuer := chain.GenAccount()
	nonIssuer := chain.GenAccount()
	recipient := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgSetWhitelistedLimit{},
			&assetfttypes.MsgSetWhitelistedLimit{},
			&assetfttypes.MsgSetWhitelistedLimit{},
			&banktypes.MsgSend{},
			&banktypes.MsgMultiSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
			Add(sdk.NewInt(1_000_000)), // added 1 million for smart contract upload
	})
	chain.FundAccountWithOptions(ctx, t, nonIssuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgSetWhitelistedLimit{},
		},
	})
	chain.FundAccountWithOptions(ctx, t, recipient, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactory().WithSimulateAndExecute(true), issuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)

	// Issue the new fungible token
	amount := sdkmath.NewInt(20000)
	subunit := "uabd"
	denom := assetfttypes.BuildDenom(subunit, issuer)
	attachedFund := chain.NewCoin(sdk.NewInt(10))
	msg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABD",
		Subunit:       "uabd",
		Precision:     6,
		Description:   "ABD Description",
		InitialAmount: amount,
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_whitelisting,
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Funds:  sdk.NewCoins(attachedFund),
			Label:  "testing-whitelisting",
		},
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)

	requireT.NoError(err)

	// try to send to recipient before it is whitelisted (balance 0, whitelist limit 0)
	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(10)))
	// send
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      coinsToSend,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.ErrorContains(err, "Whitelisted limit exceeded.")

	// multi-send
	multiSendMsg := &banktypes.MsgMultiSend{
		Inputs:  []banktypes.Input{{Address: issuer.String(), Coins: coinsToSend}},
		Outputs: []banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}},
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(multiSendMsg)),
		multiSendMsg,
	)
	requireT.ErrorContains(err, "Whitelisted limit exceeded.")

	// whitelist 400 tokens
	whitelistMsg := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(400)),
	}
	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	requireT.NoError(err)
	assertT.EqualValues(res.GasUsed, chain.GasLimitByMsgs(whitelistMsg))

	// query whitelisted tokens
	whitelistedBalance, err := ftClient.WhitelistedBalance(ctx, &assetfttypes.QueryWhitelistedBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdkmath.NewInt(400)), whitelistedBalance.Balance)

	whitelistedBalances, err := ftClient.WhitelistedBalances(ctx, &assetfttypes.QueryWhitelistedBalancesRequest{
		Account: recipient.String(),
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(400))), whitelistedBalances.Balances)

	// try to receive more than whitelisted (600) (possible 400)
	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(600))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.ErrorContains(err, "Whitelisted limit exceeded.")

	// try to send whitelisted balance (400)
	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(400))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	balance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(400)).String(), balance.GetBalance().String())

	// try to send one more
	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(1))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.ErrorContains(err, "Whitelisted limit exceeded.")

	// whitelist one more
	whitelistMsg = &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(401)),
	}
	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	requireT.NoError(err)
	assertT.EqualValues(res.GasUsed, chain.GasLimitByMsgs(whitelistMsg))

	// query whitelisted tokens
	whitelistedBalance, err = ftClient.WhitelistedBalance(ctx, &assetfttypes.QueryWhitelistedBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdkmath.NewInt(401)), whitelistedBalance.Balance)

	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(1))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	balance, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(401)).String(), balance.GetBalance().String())

	// Verify that issuer has no whitelisted balance
	whitelistedBalance, err = ftClient.WhitelistedBalance(ctx, &assetfttypes.QueryWhitelistedBalanceRequest{
		Account: issuer.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdkmath.ZeroInt()), whitelistedBalance.Balance)

	// Send something to issuer, it should succeed despite the fact that issuer is not whitelisted
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   issuer.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(10))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	balance, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: issuer.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(19609)).String(), balance.GetBalance().String())

	// Set whitelisted balance to 0 for recipient
	whitelistMsg = &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.ZeroInt()),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	requireT.NoError(err)

	// Transfer to recipient should fail now
	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.OneInt())),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.ErrorContains(err, "Whitelisted limit exceeded.")
}

// TestAssetFTExtensionFreeze checks extension freeze functionality of fungible tokens.
func TestAssetFTExtensionFreeze(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	clientCtx := chain.ClientContext

	ftClient := assetfttypes.NewQueryClient(clientCtx)
	bankClient := banktypes.NewQueryClient(clientCtx)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	randomAddress := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgFreeze{},
			&assetfttypes.MsgFreeze{},
			&assetfttypes.MsgUnfreeze{},
			&assetfttypes.MsgUnfreeze{},
			&assetfttypes.MsgUnfreeze{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
			Add(sdk.NewInt(1_000_000)), // added 1 million for smart contract upload
	})
	chain.FundAccountWithOptions(ctx, t, recipient, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&banktypes.MsgMultiSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgMultiSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
		},
	})
	chain.FundAccountWithOptions(ctx, t, randomAddress, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgFreeze{},
		},
	})

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactory().WithSimulateAndExecute(true), issuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)
	attachedFund := chain.NewCoin(sdk.NewInt(10))

	// Issue the new fungible token
	msg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(1010),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_freezing,
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Funds:  sdk.NewCoins(attachedFund),
			Label:  "testing-freezing",
		},
	}

	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(assetfttypes.BuildDenom(msg.Subunit, issuer), sdkmath.NewInt(1000)),
		),
	}

	msgList := []sdk.Msg{
		msg, msgSend,
	}

	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgList...)),
		msgList...,
	)

	requireT.NoError(err)
	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assetfttypes.EventIssued](res.Events)
	requireT.NoError(err)
	denom := fungibleTokenIssuedEvts[0].Denom

	// freeze 400 tokens
	freezeMsg := &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(400)),
	}
	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	requireT.NoError(err)
	assertT.EqualValues(res.GasUsed, chain.GasLimitByMsgs(freezeMsg))

	fungibleTokenFreezeEvts, err := event.FindTypedEvents[*assetfttypes.EventFrozenAmountChanged](res.Events)
	requireT.NoError(err)
	assertT.EqualValues(&assetfttypes.EventFrozenAmountChanged{
		Account:        recipient.String(),
		Denom:          denom,
		PreviousAmount: sdkmath.NewInt(0),
		CurrentAmount:  sdkmath.NewInt(400),
	}, fungibleTokenFreezeEvts[0])

	// query frozen tokens
	frozenBalance, err := ftClient.FrozenBalance(ctx, &assetfttypes.QueryFrozenBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdkmath.NewInt(400)), frozenBalance.Balance)

	frozenBalances, err := ftClient.FrozenBalances(ctx, &assetfttypes.QueryFrozenBalancesRequest{
		Account: recipient.String(),
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(400))), frozenBalances.Balances)

	// try to send more than available (650) (600 is available)
	recipient2 := chain.GenAccount()
	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(650)))
	// send
	sendMsg := &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   recipient2.String(),
		Amount:      coinsToSend,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.ErrorContains(err, "Requested transfer token is frozen.")
	// multi-send
	multiSendMsg := &banktypes.MsgMultiSend{
		Inputs:  []banktypes.Input{{Address: recipient.String(), Coins: coinsToSend}},
		Outputs: []banktypes.Output{{Address: recipient2.String(), Coins: coinsToSend}},
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(multiSendMsg)),
		multiSendMsg,
	)
	requireT.ErrorContains(err, "Requested transfer token is frozen.")

	// try to send available tokens (300 + 300)
	coinsToSend = sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(300)))
	// send
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   recipient2.String(),
		Amount:      coinsToSend,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	balance1, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(700)).String(), balance1.GetBalance().String())
	// multi-send
	multiSendMsg = &banktypes.MsgMultiSend{
		Inputs:  []banktypes.Input{{Address: recipient.String(), Coins: coinsToSend}},
		Outputs: []banktypes.Output{{Address: recipient2.String(), Coins: coinsToSend}},
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(multiSendMsg)),
		multiSendMsg,
	)
	requireT.NoError(err) // Requested transfer token is frozen
	balance1, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(400)).String(), balance1.GetBalance().String())

	// unfreeze 200 tokens and try to send 250 tokens
	unfreezeMsg := &assetfttypes.MsgUnfreeze{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(200)),
	}
	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(unfreezeMsg)),
		unfreezeMsg,
	)
	requireT.NoError(err)
	assertT.EqualValues(res.GasUsed, chain.GasLimitByMsgs(unfreezeMsg))

	fungibleTokenFreezeEvts, err = event.FindTypedEvents[*assetfttypes.EventFrozenAmountChanged](res.Events)
	requireT.NoError(err)
	assertT.EqualValues(&assetfttypes.EventFrozenAmountChanged{
		Account:        recipient.String(),
		Denom:          denom,
		PreviousAmount: sdkmath.NewInt(400),
		CurrentAmount:  sdkmath.NewInt(200),
	}, fungibleTokenFreezeEvts[0])

	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   recipient2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(250))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.ErrorContains(err, "Requested transfer token is frozen.")

	// send available tokens (200)
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   recipient2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(200))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	// unfreeze 400 tokens (frozen balance is 200), it should give error
	unfreezeMsg = &assetfttypes.MsgUnfreeze{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(400)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(unfreezeMsg)),
		unfreezeMsg,
	)
	requireT.ErrorIs(err, cosmoserrors.ErrInsufficientFunds)

	// set absolute frozen amount to 250
	setFrozenMsg := &assetfttypes.MsgSetFrozen{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(250)),
	}
	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(setFrozenMsg)),
		setFrozenMsg,
	)
	requireT.NoError(err)
	assertT.EqualValues(res.GasUsed, chain.GasLimitByMsgs(setFrozenMsg))
	fungibleTokenFreezeEvts, err = event.FindTypedEvents[*assetfttypes.EventFrozenAmountChanged](res.Events)
	requireT.NoError(err)
	assertT.EqualValues(&assetfttypes.EventFrozenAmountChanged{
		Account:        recipient.String(),
		Denom:          denom,
		PreviousAmount: sdkmath.NewInt(200),
		CurrentAmount:  sdkmath.NewInt(250),
	}, fungibleTokenFreezeEvts[0])

	frozenBalance, err = ftClient.FrozenBalance(ctx, &assetfttypes.QueryFrozenBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdkmath.NewInt(250)), frozenBalance.Balance)

	// unfreeze 250 tokens and observer current frozen amount is zero
	unfreezeMsg = &assetfttypes.MsgUnfreeze{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(250)),
	}
	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(unfreezeMsg)),
		unfreezeMsg,
	)
	requireT.NoError(err)
	assertT.EqualValues(res.GasUsed, chain.GasLimitByMsgs(unfreezeMsg))

	fungibleTokenFreezeEvts, err = event.FindTypedEvents[*assetfttypes.EventFrozenAmountChanged](res.Events)
	requireT.NoError(err)
	assertT.EqualValues(&assetfttypes.EventFrozenAmountChanged{
		Account:        recipient.String(),
		Denom:          denom,
		PreviousAmount: sdkmath.NewInt(250),
		CurrentAmount:  sdkmath.NewInt(0),
	}, fungibleTokenFreezeEvts[0])
}

// TestAssetFTExtensionBurn checks extension burn functionality of fungible tokens.
func TestAssetFTExtensionBurn(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.MulRaw(2).
			Add(sdk.NewInt(1_000_000)), // added 1 million for smart contract upload
	})

	chain.FundAccountWithOptions(ctx, t, recipient, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
		},
	})

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactory().WithSimulateAndExecute(true), issuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)
	attachedFund := chain.NewCoin(sdk.NewInt(10))

	// Issue an unburnable fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABCNotBurnable",
		Subunit:       "uabcnotburnable",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(1012),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_minting,
			assetfttypes.Feature_freezing,
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Funds:  sdk.NewCoins(attachedFund),
			Label:  "testing-burning",
		},
	}

	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assetfttypes.EventIssued](res.Events)
	requireT.NoError(err)
	unburnable := fungibleTokenIssuedEvts[0].Denom

	// try to burn unburnable token from issuer
	burnMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   issuer.String(),
		Amount: sdk.NewCoins(sdk.Coin{
			Denom:  unburnable,
			Amount: sdkmath.NewInt(900),
		}),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.NoError(err)

	// send some coins to the recipient
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(unburnable, sdkmath.NewInt(102))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	// try to burn unburnable token from recipient
	burnMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   issuer.String(),
		Amount: sdk.NewCoins(sdk.Coin{
			Denom:  unburnable,
			Amount: sdkmath.NewInt(MagicAmountBurning),
		}),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.ErrorContains(err, "Feature disabled.")

	// Issue a burnable fungible token
	issueMsg = &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABCBurnable",
		Subunit:       "uabcburnable",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_burning,
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Funds:  sdk.NewCoins(attachedFund),
			Label:  "testing-burning",
		},
	}

	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	fungibleTokenIssuedEvts, err = event.FindTypedEvents[*assetfttypes.EventIssued](res.Events)
	requireT.NoError(err)
	burnableDenom := fungibleTokenIssuedEvts[0].Denom

	// send some coins to the recipient
	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(burnableDenom, sdkmath.NewInt(102))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	// burn tokens and check balance and total supply
	oldSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: burnableDenom})
	requireT.NoError(err)
	burnCoin := sdk.NewCoin(burnableDenom, sdkmath.NewInt(MagicAmountBurning))

	burnMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   issuer.String(),
		Amount:      sdk.NewCoins(burnCoin),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.NoError(err)

	balance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{Address: issuer.String(), Denom: burnableDenom})
	requireT.NoError(err)
	assertT.EqualValues(sdk.NewCoin(burnableDenom, sdkmath.NewInt(797)).String(), balance.GetBalance().String())

	newSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: burnableDenom})
	requireT.NoError(err)
	assertT.EqualValues(burnCoin.String(), oldSupply.GetAmount().Sub(newSupply.GetAmount()).String())
}

// TestAssetFTExtensionMint checks extension mint functionality of fungible tokens.
func TestAssetFTExtensionMint(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	admin := chain.GenAccount()
	randomAddress := chain.GenAccount()
	recipient := chain.GenAccount()
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.MulRaw(2).
			Add(sdk.NewInt(1_000_000)), // added 1 million for smart contract upload
	})

	chain.FundAccountWithOptions(ctx, t, randomAddress, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})

	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactory().WithSimulateAndExecute(true), issuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)
	attachedFund := chain.NewCoin(sdk.NewInt(10))

	// Issue an unmintable fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABCNotMintable",
		Subunit:       "uabcnotmintable",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_block_smart_contracts,
			assetfttypes.Feature_burning,
			assetfttypes.Feature_freezing,
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Funds:  sdk.NewCoins(attachedFund),
			Label:  "testing-minting",
		},
	}

	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assetfttypes.EventIssued](res.Events)
	requireT.NoError(err)
	unmintableDenom := fungibleTokenIssuedEvts[0].Denom

	// try to mint unmintable token
	mintMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   issuer.String(),
		Amount: sdk.NewCoins(sdk.Coin{
			Denom:  unmintableDenom,
			Amount: sdkmath.NewInt(MagicAmountMinting),
		}),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.ErrorContains(err, "Feature disabled.")

	// Issue a mintable fungible token
	issueMsg = &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABCMintable",
		Subunit:       "uabcmintable",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_block_smart_contracts,
			assetfttypes.Feature_minting,
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Funds:  sdk.NewCoins(attachedFund),
			Label:  "testing-minting",
		},
	}

	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	fungibleTokenIssuedEvts, err = event.FindTypedEvents[*assetfttypes.EventIssued](res.Events)
	requireT.NoError(err)
	mintableDenom := fungibleTokenIssuedEvts[0].Denom

	// send some coins to the account
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   randomAddress.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(mintableDenom, sdkmath.NewInt(210))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	// mint tokens and check balance and total supply
	oldSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: mintableDenom})
	requireT.NoError(err)
	mintCoin := sdk.NewCoin(mintableDenom, sdkmath.NewInt(MagicAmountMinting))
	mintMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   issuer.String(),
		Amount:      sdk.NewCoins(mintCoin),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.NoError(err)

	balance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{Address: issuer.String(), Denom: mintableDenom})
	requireT.NoError(err)
	assertT.EqualValues(
		mintCoin.Add(sdk.NewCoin(mintableDenom, sdkmath.NewInt(790))).String(),
		balance.GetBalance().String(),
	)

	newSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: mintableDenom})
	requireT.NoError(err)
	assertT.EqualValues(mintCoin, newSupply.GetAmount().Sub(oldSupply.GetAmount()))

	// mint tokens to recipient
	mintCoin = sdk.NewCoin(mintableDenom, sdkmath.NewInt(MagicAmountMinting))
	mintMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(mintCoin),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.NoError(err)

	balance, err = bankClient.Balance(
		ctx,
		&banktypes.QueryBalanceRequest{Address: recipient.String(), Denom: mintableDenom},
	)
	requireT.NoError(err)
	assertT.EqualValues(mintCoin.String(), balance.GetBalance().String())

	newSupply2, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: mintableDenom})
	requireT.NoError(err)
	assertT.EqualValues(mintCoin.String(), newSupply2.GetAmount().Sub(newSupply.GetAmount()).String())

	// sending to smart contract is blocked so minting to it should fail
	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		admin,
		moduleswasm.BankSendWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    moduleswasm.EmptyPayload,
			Label:      "bank_send",
		},
	)
	requireT.NoError(err)

	mintMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   contractAddr,
		Amount:      sdk.NewCoins(mintCoin),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.ErrorContains(err, "Transferring to or from smart contracts are prohibited.")
}

// TestAssetFTExtensionSendingToSmartContractIsDenied verifies that this is not possible to send token to smart contract
// if issuer blocked this operation.
func TestAssetFTExtensionSendingToSmartContractIsDenied(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issuer := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(issuer, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

	clientCtx := chain.ClientContext

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactory().WithSimulateAndExecute(true), issuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)
	attachedFund := chain.NewCoin(sdk.NewInt(10))

	// Issue a fungible token which cannot be sent to the smart contract
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     6,
		InitialAmount: sdkmath.NewInt(1000),
		Description:   "ABC Description",
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_block_smart_contracts,
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Funds:  sdk.NewCoins(attachedFund),
			Label:  "testing-block-smart-contract",
		},
		BurnRate:           sdk.ZeroDec(),
		SendCommissionRate: sdk.ZeroDec(),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

	initialPayload, err := json.Marshal(moduleswasm.SimpleState{
		Count: 1337,
	})
	requireT.NoError(err)

	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		issuer,
		moduleswasm.SimpleStateWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayload,
			Label:      "simple_state",
		},
	)
	requireT.NoError(err)

	// sending coins to the smart contract should fail
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   contractAddr,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(denom, 100)),
	}
	_, err = client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.ErrorContains(err, "Transferring to or from smart contracts are prohibited.")

	multiSendMsg := &banktypes.MsgMultiSend{
		Inputs: []banktypes.Input{
			{
				Address: issuer.String(),
				Coins:   sdk.NewCoins(sdk.NewInt64Coin(denom, 100)),
			},
		},
		Outputs: []banktypes.Output{
			{
				Address: contractAddr,
				Coins:   sdk.NewCoins(sdk.NewInt64Coin(denom, 100)),
			},
		},
	}
	_, err = client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(multiSendMsg)),
		multiSendMsg,
	)
	requireT.ErrorContains(err, "Transferring to or from smart contracts are prohibited.")
}

// TestAssetFTExtensionAttachingToSmartContractIsDenied verifies that this is not possible to attach token to smart
// contract call if issuer blocked this operation.
func TestAssetFTExtensionAttachingToSmartContractCallIsDenied(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issuer := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(issuer, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactory().WithSimulateAndExecute(true), issuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)
	attachedFund := chain.NewCoin(sdk.NewInt(10))

	txf := chain.TxFactory().
		WithSimulateAndExecute(true)

	// Issue a fungible token which cannot be sent to the smart contract
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     6,
		InitialAmount: sdkmath.NewInt(1000),
		Description:   "ABC Description",
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_block_smart_contracts,
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Funds:  sdk.NewCoins(attachedFund),
			Label:  "block-smart-contract",
		},
		BurnRate:           sdk.ZeroDec(),
		SendCommissionRate: sdk.ZeroDec(),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

	initialPayload, err := json.Marshal(moduleswasm.SimpleState{
		Count: 1337,
	})
	requireT.NoError(err)

	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		issuer,
		moduleswasm.SimpleStateWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayload,
			Label:      "simple_state",
		},
	)
	requireT.NoError(err)

	// Executing smart contract - this operation should fail because coins are attached to it
	incrementPayload, err := moduleswasm.MethodToEmptyBodyPayload(moduleswasm.SimpleIncrement)
	requireT.NoError(err)
	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, issuer, contractAddr, incrementPayload, sdk.NewInt64Coin(denom, 100))
	requireT.ErrorContains(err, "Transferring to or from smart contracts are prohibited.")
}

// TestAssetFTExtensionIssuingSmartContractIsAllowedToReceive verifies that issuing smart contract is allowed to
// receive coins even if sending them to smart contract is disabled.
func TestAssetFTExtensionIssuingSmartContractIsAllowedToSendAndReceive(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	admin := chain.GenAccount()
	recipient := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(50000000000))),
	)
	chain.FundAccountWithOptions(ctx, t, recipient, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})

	txf := chain.TxFactory().
		WithSimulateAndExecute(true)

	codeID, err := chain.Wasm.DeployWASMContract(ctx, txf, admin, testcontracts.AssetExtensionWasm)
	requireT.NoError(err)

	issuanceAmount := sdkmath.NewInt(10_000)
	issuanceReq := issueFTRequest{
		Symbol:        "symbol",
		Subunit:       "subunit",
		Precision:     6,
		InitialAmount: issuanceAmount.String(),
		Description:   "my wasm fungible token",
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_minting,
			assetfttypes.Feature_block_smart_contracts,
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Label:  "block-smart-contract",
		},
		BurnRate:           "0",
		SendCommissionRate: "0",
	}
	issuerFTInstantiatePayload, err := json.Marshal(issuanceReq)
	requireT.NoError(err)

	// instantiate new contract
	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		admin,
		moduleswasm.FTWASM,
		integration.InstantiateConfig{
			// we add the initial amount to let the contract issue the token on behalf of it
			Amount:     chain.QueryAssetFTParams(ctx, t).IssueFee,
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    issuerFTInstantiatePayload,
			Label:      "fungible_token",
		},
	)
	requireT.NoError(err)

	denom := assetfttypes.BuildDenom(issuanceReq.Subunit, sdk.MustAccAddressFromBech32(contractAddr))

	// mint to itself
	amountToMint := sdkmath.NewInt(500)
	mintPayload, err := json.Marshal(map[ftMethod]amountRecipientBodyFTRequest{
		ftMethodMint: {
			Amount:    amountToMint.String(),
			Recipient: contractAddr,
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, mintPayload, sdk.Coin{})
	requireT.NoError(err)

	// mint to someone else
	amountToMint = sdkmath.NewInt(100)
	mintPayload, err = json.Marshal(map[ftMethod]amountRecipientBodyFTRequest{
		ftMethodMint: {
			Amount:    amountToMint.String(),
			Recipient: recipient.String(),
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, mintPayload, sdk.Coin{})
	requireT.NoError(err)

	// send back to smart contract
	msgSend := &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   contractAddr,
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, amountToMint)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	require.NoError(t, err)
}

// TestAssetFTExtensionMintingAndSendingOnBehalfOfIssuingSmartContractIsPossibleEvenIfSmartContractsAreBlocked verifies
// that it is possible to use authz to mint and send the token on behalf of the issuing smart contract if smart
// contracts are blocked.
func TestAssetFTExtensionMintingAndSendingOnBehalfOfIssuingSmartContractIsPossibleEvenIfSmartContractsAreBlocked(
	t *testing.T,
) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	admin := chain.GenAccount()
	grantee := chain.GenAccount()
	recipient := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactory().WithSimulateAndExecute(true), admin, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)

	// instantiate new contract
	initialPayload, err := json.Marshal(struct{}{})
	requireT.NoError(err)

	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		admin,
		moduleswasm.AuthzStargateWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayload,
			Label:      "authzStargate",
			Amount:     chain.QueryAssetFTParams(ctx, t).IssueFee,
		},
	)
	requireT.NoError(err)

	// grant authorizations
	grantIssueMsg, err := authztypes.NewMsgGrant(
		sdk.MustAccAddressFromBech32(contractAddr),
		grantee,
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&assetfttypes.MsgIssue{})),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	grantMintMsg, err := authztypes.NewMsgGrant(
		sdk.MustAccAddressFromBech32(contractAddr),
		grantee,
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&assetfttypes.MsgMint{})),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	grantSendMsg, err := authztypes.NewMsgGrant(
		sdk.MustAccAddressFromBech32(contractAddr),
		grantee,
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	grantIssueMsgAny, err := codectypes.NewAnyWithValue(grantIssueMsg)
	requireT.NoError(err)
	grantMintMsgAny, err := codectypes.NewAnyWithValue(grantMintMsg)
	requireT.NoError(err)
	grantSendMsgAny, err := codectypes.NewAnyWithValue(grantSendMsg)
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		admin,
		contractAddr,
		moduleswasm.AuthZExecuteStargateRequest(&authztypes.MsgExec{
			Grantee: contractAddr,
			Msgs: []*codectypes.Any{
				grantIssueMsgAny,
				grantSendMsgAny,
				grantMintMsgAny,
			},
		}),
		sdk.Coin{},
	)
	requireT.NoError(err)

	// issue, send and mint on behalf of the smart contract
	subunit := "uabe"
	denom := assetfttypes.BuildDenom(subunit, sdk.MustAccAddressFromBech32(contractAddr))
	execMsg := authztypes.NewMsgExec(grantee, []sdk.Msg{
		&assetfttypes.MsgIssue{
			Issuer:        contractAddr,
			Symbol:        "ABE",
			Subunit:       subunit,
			Precision:     6,
			Description:   "ABE Description",
			InitialAmount: sdkmath.NewInt(1000),
			Features: []assetfttypes.Feature{
				assetfttypes.Feature_minting,
				assetfttypes.Feature_block_smart_contracts,
				assetfttypes.Feature_extension,
			},
			ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
				CodeId: codeID,
				Label:  "block-smart-contract",
			},
		},
		&banktypes.MsgSend{
			FromAddress: contractAddr,
			ToAddress:   recipient.String(),
			Amount:      sdk.NewCoins(sdk.NewInt64Coin(denom, 100)),
		},
		&assetfttypes.MsgMint{
			Sender:    contractAddr,
			Recipient: recipient.String(),
			Coin:      sdk.NewInt64Coin(denom, 100),
		},
	})

	chain.FundAccountWithOptions(ctx, t, grantee, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&execMsg,
		},
	})

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(grantee),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(&execMsg)),
		&execMsg,
	)
	requireT.NoError(err)

	// check balances

	contract := sdk.MustAccAddressFromBech32(contractAddr)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&contract:  900,
		&recipient: 200,
	})
}
