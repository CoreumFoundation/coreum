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

	integrationtests "github.com/CoreumFoundation/coreum/v6/integration-tests"
	moduleswasm "github.com/CoreumFoundation/coreum/v6/integration-tests/contracts/modules"
	"github.com/CoreumFoundation/coreum/v6/pkg/client"
	"github.com/CoreumFoundation/coreum/v6/testutil/event"
	"github.com/CoreumFoundation/coreum/v6/testutil/integration"
	testcontracts "github.com/CoreumFoundation/coreum/v6/x/asset/ft/keeper/test-contracts"
	assetfttypes "github.com/CoreumFoundation/coreum/v6/x/asset/ft/types"
	dextypes "github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

var (
	AmountBlockSmartContractTrigger = sdkmath.NewInt(testcontracts.AmountBlockSmartContractTrigger)
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
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
			Add(sdkmath.NewInt(1_000_000)).
			Add(sdkmath.NewInt(3 * 500_000)), // give 500k gas for each message since extensions are nondeterministic
	})

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactoryAuto(), issuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)

	//nolint:tagliatelle // these will be exposed to rust and must be snake case.
	issuanceMsg := struct {
		ExtraData string `json:"extra_data"`
	}{
		ExtraData: "test",
	}

	issuanceMsgBytes, err := json.Marshal(issuanceMsg)
	requireT.NoError(err)

	attachedFund := chain.NewCoin(sdkmath.NewInt(10))
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
			CodeId:      codeID,
			Funds:       sdk.NewCoins(attachedFund),
			Label:       "testing-issuance",
			IssuanceMsg: issuanceMsgBytes,
		},
	}

	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactoryAuto(),
		issueMsg,
	)
	requireT.NoError(err)
	requireT.NotEqualValues(chain.GasLimitByMsgs(issueMsg), res.GasUsed)

	// assert that attached funds are transferred to the contract
	token, err := assetFTClient.Token(ctx, &assetfttypes.QueryTokenRequest{Denom: denom})
	requireT.NoError(err)
	contractBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: token.Token.ExtensionCWAddress,
		Denom:   chain.ChainSettings.Denom,
	})
	requireT.NoError(err)
	requireT.Equal(contractBalance.GetBalance().String(), attachedFund.String())

	// assert correct label is applied.
	contractInfo, err := wasmClient.ContractInfo(
		ctx, &wasmtypes.QueryContractInfoRequest{Address: token.Token.ExtensionCWAddress},
	)
	requireT.NoError(err)
	requireT.Equal(issueMsg.ExtensionSettings.Label, contractInfo.Label)

	recipient := chain.GenAccount()
	// sending 1 will succeed
	sendMsg := &banktypes.MsgSend{
		FromAddress: issueMsg.Issuer,
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(12))),
	}
	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactoryAuto(),
		sendMsg,
	)

	requireT.NoError(err)
	requireT.NotEqualValues(chain.GasLimitByMsgs(sendMsg), res.GasUsed)
	balance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal("12", balance.Balance.Amount.String())

	// sending 7 will fail
	sendMsg.Amount = sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(testcontracts.AmountDisallowedTrigger)))
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(500_000),
		sendMsg,
	)
	requireT.ErrorIs(err, assetfttypes.ErrExtensionCallFailed)
	balance, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal("12", balance.Balance.Amount.String())

	queryRes, err := wasmClient.SmartContractState(ctx, &wasmtypes.QuerySmartContractStateRequest{
		Address:   token.Token.ExtensionCWAddress,
		QueryData: []byte(`{"query_issuance_msg":{}}`),
	})
	requireT.NoError(err)
	requireT.NoError(json.Unmarshal(queryRes.Data, &issuanceMsg))
	requireT.Equal("test", issuanceMsg.ExtraData)

	// sending 7 will to contract address will succeed
	sendMsg.ToAddress = token.Token.ExtensionCWAddress
	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(500_000),
		sendMsg,
	)
	requireT.NoError(err)
	balance, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: token.Token.ExtensionCWAddress,
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal("7", balance.Balance.Amount.String())
	skipChecksStr, err := event.FindStringEventAttribute(res.Events, "wasm", "skip_checks")
	requireT.NoError(err)
	requireT.Equal("self_recipient", skipChecksStr)
}

// TestAssetFTExtensionWhitelist checks extension whitelist functionality of fungible tokens.
func TestAssetFTExtensionWhitelist(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	clientCtx := chain.ClientContext

	ftClient := assetfttypes.NewQueryClient(clientCtx)

	issuer := chain.GenAccount()
	nonIssuer := chain.GenAccount()
	recipient := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: issuer,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&assetfttypes.MsgSetWhitelistedLimit{},
				},
				Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.Mul(sdkmath.NewInt(2)).
					Add(sdkmath.NewInt(1_000_000)).    // added 1 million for smart contract upload
					Add(sdkmath.NewInt(10 * 500_000)), // give 500k gas for each message since extensions are nondeterministic
			},
		}, {
			Acc: nonIssuer,
			Options: integration.BalancesOptions{
				Amount: sdkmath.NewInt(1 * 500_000), // give 500k gas for each message since extensions are nondeterministic
			},
		}, {
			Acc: recipient,
			Options: integration.BalancesOptions{
				Amount: sdkmath.NewInt(1 * 500_000), // give 500k gas for each message since extensions are nondeterministic
			},
		},
	})

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactoryAuto(), issuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)

	// Issue the new fungible token
	amount := sdkmath.NewInt(20000)
	subunit := "uabd"
	denom := assetfttypes.BuildDenom(subunit, issuer)
	attachedFund := chain.NewCoin(sdkmath.NewInt(10))
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
	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactoryAuto(),
		msg,
	)
	requireT.NoError(err)
	requireT.NotEqualValues(chain.GasLimitByMsgs(msg), res.GasUsed)

	// Issue the new fungible token without extension
	subunit = "uabe"
	denomWithoutExtension := assetfttypes.BuildDenom(subunit, issuer)
	msg = &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABE",
		Subunit:       "uabe",
		Precision:     6,
		Description:   "ABE Description",
		InitialAmount: amount,
		Features:      []assetfttypes.Feature{},
	}
	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)

	requireT.NoError(err)
	requireT.EqualValues(chain.GasLimitByMsgs(msg), res.GasUsed)

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
		chain.TxFactory().WithGas(500_000),
		sendMsg,
	)
	requireT.ErrorIs(err, assetfttypes.ErrWhitelistedLimitExceeded)

	// multi-send
	multiSendMsg := &banktypes.MsgMultiSend{
		Inputs:  []banktypes.Input{{Address: issuer.String(), Coins: coinsToSend}},
		Outputs: []banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}},
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(500_000),
		multiSendMsg,
	)
	requireT.ErrorIs(err, assetfttypes.ErrWhitelistedLimitExceeded)

	// multi-send tokens with and without extension
	multiSendMsg = &banktypes.MsgMultiSend{
		Inputs: []banktypes.Input{{Address: issuer.String(), Coins: sdk.NewCoins(
			sdk.NewCoin(denom, sdkmath.NewInt(10)),
			sdk.NewCoin(denomWithoutExtension, sdkmath.NewInt(10)),
		)}},
		Outputs: []banktypes.Output{{Address: recipient.String(), Coins: sdk.NewCoins(
			sdk.NewCoin(denom, sdkmath.NewInt(10)),
			sdk.NewCoin(denomWithoutExtension, sdkmath.NewInt(10)),
		)}},
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(500_000),
		multiSendMsg,
	)
	requireT.ErrorIs(err, assetfttypes.ErrWhitelistedLimitExceeded)

	// whitelist 400 tokens
	whitelistMsg := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(400)),
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
	whitelistedBalance, err := ftClient.WhitelistedBalance(ctx, &assetfttypes.QueryWhitelistedBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(400)), whitelistedBalance.Balance)

	whitelistedBalances, err := ftClient.WhitelistedBalances(ctx, &assetfttypes.QueryWhitelistedBalancesRequest{
		Account: recipient.String(),
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(400))), whitelistedBalances.Balances)

	// reverse whitelisted amount
	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(400))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactoryAuto(),
		sendMsg,
	)
	requireT.NoError(err)
}

// TestAssetFTExtensionFreeze checks extension freeze functionality of fungible tokens.
func TestAssetFTExtensionFreeze(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	clientCtx := chain.ClientContext

	ftClient := assetfttypes.NewQueryClient(clientCtx)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	randomAddress := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: issuer,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&assetfttypes.MsgFreeze{},
					&assetfttypes.MsgFreeze{},
					&assetfttypes.MsgUnfreeze{},
					&assetfttypes.MsgUnfreeze{},
					&assetfttypes.MsgUnfreeze{},
				},
				Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
					Add(sdkmath.NewInt(1_000_000)).   // added 1 million for smart contract upload
					Add(sdkmath.NewInt(2 * 500_000)), // add 500k for each message with extension transfer
			},
		}, {
			Acc: recipient,
			Options: integration.BalancesOptions{
				Amount: sdkmath.NewInt(6 * 500_000), // add 500k for each message with extension transfer
			},
		}, {
			Acc: randomAddress,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&assetfttypes.MsgFreeze{},
				},
			},
		},
	})

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactoryAuto(), issuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)
	attachedFund := chain.NewCoin(sdkmath.NewInt(10))

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
		chain.TxFactoryAuto(),
		msgList...,
	)

	requireT.NoError(err)
	requireT.NotEqualValues(chain.GasLimitByMsgs(msgList...), res.GasUsed)
	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assetfttypes.EventIssued](res.Events)
	requireT.NoError(err)
	denom := fungibleTokenIssuedEvts[0].Denom

	// freeze 400 tokens
	freezeMsg := &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(925)),
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
	assertT.Equal(&assetfttypes.EventFrozenAmountChanged{
		Account:        recipient.String(),
		Denom:          denom,
		PreviousAmount: sdkmath.NewInt(0),
		CurrentAmount:  sdkmath.NewInt(925),
	}, fungibleTokenFreezeEvts[0])

	// query frozen tokens
	frozenBalance, err := ftClient.FrozenBalance(ctx, &assetfttypes.QueryFrozenBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(925)), frozenBalance.Balance)

	// try to send more than available (76) (75 is available)
	recipient2 := chain.GenAccount()
	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(76)))
	// send
	sendMsg := &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   recipient2.String(),
		Amount:      coinsToSend,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(500_000),
		sendMsg,
	)
	requireT.ErrorIs(err, cosmoserrors.ErrInsufficientFunds)
	// multi-send
	multiSendMsg := &banktypes.MsgMultiSend{
		Inputs:  []banktypes.Input{{Address: recipient.String(), Coins: coinsToSend}},
		Outputs: []banktypes.Output{{Address: recipient2.String(), Coins: coinsToSend}},
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(500_000),
		multiSendMsg,
	)
	requireT.ErrorIs(err, cosmoserrors.ErrInsufficientFunds)
	// send allowed amount
	coinsToSend = sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(75)))
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   recipient2.String(),
		Amount:      coinsToSend,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactoryAuto(),
		sendMsg,
	)
	requireT.NoError(err)
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

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: issuer,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{},
				Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.MulRaw(2).
					Add(sdkmath.NewInt(1_000_000)).   // added 1 million for smart contract upload
					Add(sdkmath.NewInt(6 * 500_000)), // add 500k for each message with extension transfer
			},
		}, {
			Acc: recipient,
			Options: integration.BalancesOptions{
				Amount: sdkmath.NewInt(2 * 500_000), // add 500k for each message with extension transfer
			},
		},
	})

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactoryAuto(), issuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)
	attachedFund := chain.NewCoin(sdkmath.NewInt(10))

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
		chain.TxFactoryAuto(),
		issueMsg,
	)

	requireT.NoError(err)
	requireT.NotEqualValues(chain.GasLimitByMsgs(issueMsg), res.GasUsed)
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
		chain.TxFactoryAuto(),
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
		chain.TxFactoryAuto(),
		sendMsg,
	)
	requireT.NoError(err)

	// try to burn unburnable token from recipient, it should be possible if extension decides
	burnMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   issuer.String(),
		Amount: sdk.NewCoins(sdk.Coin{
			Denom:  unburnable,
			Amount: sdkmath.NewInt(testcontracts.AmountBurningTrigger),
		}),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactoryAuto(),
		burnMsg,
	)
	requireT.NoError(err)

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
		chain.TxFactoryAuto(),
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
		chain.TxFactoryAuto(),
		sendMsg,
	)
	requireT.NoError(err)

	// burn tokens and check balance and total supply
	oldSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: burnableDenom})
	requireT.NoError(err)
	burnCoin := sdk.NewCoin(burnableDenom, sdkmath.NewInt(testcontracts.AmountBurningTrigger))

	burnMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   issuer.String(),
		Amount:      sdk.NewCoins(burnCoin),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactoryAuto(),
		burnMsg,
	)
	requireT.NoError(err)

	balance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{Address: issuer.String(), Denom: burnableDenom})
	requireT.NoError(err)
	assertT.Equal(sdk.NewCoin(burnableDenom, sdkmath.NewInt(797)).String(), balance.GetBalance().String())

	newSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: burnableDenom})
	requireT.NoError(err)
	assertT.Equal(burnCoin.String(), oldSupply.GetAmount().Sub(newSupply.GetAmount()).String())
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

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: issuer,
			Options: integration.BalancesOptions{
				Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.MulRaw(2).
					Add(sdkmath.NewInt(1_000_000)).   // added 1 million for smart contract upload
					Add(sdkmath.NewInt(7 * 500_000)), // add 500k for each message with extension transfer
			},
		}, {
			Acc: randomAddress,
			Options: integration.BalancesOptions{
				Amount: sdkmath.NewInt(1 * 500_000), // add 500k for each message with extension transfer
			},
		}, {
			Acc: admin,
			Options: integration.BalancesOptions{
				Amount: sdkmath.NewInt(1_000_000),
			},
		},
	})

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactoryAuto(), issuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)
	attachedFund := chain.NewCoin(sdkmath.NewInt(10))

	// Issue an unmintable fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABCNotMintable",
		Subunit:       "uabcnotmintable",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(1000),
		Features: []assetfttypes.Feature{
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
		chain.TxFactoryAuto(),
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
			Amount: sdkmath.NewInt(testcontracts.AmountMintingTrigger),
		}),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactoryAuto(),
		mintMsg,
	)
	requireT.ErrorContains(err, "feature minting is disabled")

	// Issue a mintable fungible token
	issueMsg = &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABCMintable",
		Subunit:       "uabcmintable",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(1000),
		Features: []assetfttypes.Feature{
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
		chain.TxFactoryAuto(),
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
		chain.TxFactoryAuto(),
		sendMsg,
	)
	requireT.NoError(err)

	// mint tokens and check balance and total supply
	oldSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: mintableDenom})
	requireT.NoError(err)
	mintCoin := sdk.NewCoin(mintableDenom, sdkmath.NewInt(testcontracts.AmountMintingTrigger))
	mintMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   issuer.String(),
		Amount:      sdk.NewCoins(mintCoin),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactoryAuto(),
		mintMsg,
	)
	requireT.NoError(err)

	balance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{Address: issuer.String(), Denom: mintableDenom})
	requireT.NoError(err)
	assertT.Equal(
		mintCoin.Add(sdk.NewCoin(mintableDenom, sdkmath.NewInt(790))).String(),
		balance.GetBalance().String(),
	)

	newSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: mintableDenom})
	requireT.NoError(err)
	assertT.Equal(mintCoin, newSupply.GetAmount().Sub(oldSupply.GetAmount()))

	// mint tokens to recipient
	mintCoin = sdk.NewCoin(mintableDenom, sdkmath.NewInt(testcontracts.AmountMintingTrigger))
	mintMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(mintCoin),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactoryAuto(),
		mintMsg,
	)
	requireT.NoError(err)

	balance, err = bankClient.Balance(
		ctx,
		&banktypes.QueryBalanceRequest{Address: recipient.String(), Denom: mintableDenom},
	)
	requireT.NoError(err)
	assertT.Equal(mintCoin.String(), balance.GetBalance().String())

	newSupply2, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: mintableDenom})
	requireT.NoError(err)
	assertT.Equal(mintCoin.String(), newSupply2.GetAmount().Sub(newSupply.GetAmount()).String())

	// sending to smart contract is blocked so minting to it should fail
	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		chain.TxFactoryAuto(),
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
		Amount:      sdk.NewCoins(sdk.NewCoin(mintableDenom, AmountBlockSmartContractTrigger)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(500_000),
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

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
			Add(sdkmath.NewInt(1_000_000)),
	})

	clientCtx := chain.ClientContext

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactoryAuto(), issuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)
	attachedFund := chain.NewCoin(sdkmath.NewInt(10))

	// Issue a fungible token which cannot be sent to the smart contract
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     6,
		InitialAmount: sdkmath.NewInt(1000),
		Description:   "ABC Description",
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Funds:  sdk.NewCoins(attachedFund),
			Label:  "testing-block-smart-contract",
		},
		BurnRate:           sdkmath.LegacyZeroDec(),
		SendCommissionRate: sdkmath.LegacyZeroDec(),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactoryAuto(),
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
		chain.TxFactoryAuto(),
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
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, AmountBlockSmartContractTrigger)),
	}
	_, err = client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(issuer),
		chain.TxFactoryAuto(),
		sendMsg,
	)
	requireT.ErrorContains(err, "Transferring to or from smart contracts are prohibited.")

	multiSendMsg := &banktypes.MsgMultiSend{
		Inputs: []banktypes.Input{
			{
				Address: issuer.String(),
				Coins:   sdk.NewCoins(sdk.NewCoin(denom, AmountBlockSmartContractTrigger)),
			},
		},
		Outputs: []banktypes.Output{
			{
				Address: contractAddr,
				Coins:   sdk.NewCoins(sdk.NewCoin(denom, AmountBlockSmartContractTrigger)),
			},
		},
	}
	_, err = client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(issuer),
		chain.TxFactoryAuto(),
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
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
			Add(sdkmath.NewInt(1_000_000)),
	})

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactoryAuto(), issuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)
	attachedFund := chain.NewCoin(sdkmath.NewInt(10))

	txf := chain.TxFactoryAuto()

	// Issue a fungible token which cannot be sent to the smart contract
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     6,
		InitialAmount: sdkmath.NewInt(1000),
		Description:   "ABC Description",
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Funds:  sdk.NewCoins(attachedFund),
			Label:  "block-smart-contract",
		},
		BurnRate:           sdkmath.LegacyZeroDec(),
		SendCommissionRate: sdkmath.LegacyZeroDec(),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactoryAuto(),
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
	_, err = chain.Wasm.ExecuteWASMContract(
		ctx, txf, issuer, contractAddr, incrementPayload, sdk.NewCoin(denom, AmountBlockSmartContractTrigger),
	)
	requireT.ErrorContains(err, "Transferring to or from smart contracts are prohibited.")
}

// TestAssetFTExtensionIssuingSmartContractIsAllowedToReceive verifies that issuing smart contract is allowed to
// receive coins even if sending them to smart contract is disabled.
func TestAssetFTExtensionIssuingSmartContractIsAllowedToSendAndReceive(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	admin := chain.GenAccount()
	recipient := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: admin,
			Options: integration.BalancesOptions{
				Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.MulRaw(2).
					Add(sdkmath.NewInt(1_000_000)),
			},
		}, {
			Acc: recipient,
			Options: integration.BalancesOptions{
				Amount: sdkmath.NewInt(500_000),
			},
		},
	})

	txf := chain.TxFactoryAuto()

	codeID, err := chain.Wasm.DeployWASMContract(ctx, txf, admin, testcontracts.AssetExtensionWasm)
	requireT.NoError(err)

	//nolint:tagliatelle // these will be exposed to rust and must be snake case.
	issuanceMsg := struct {
		ExtraData string `json:"extra_data"`
	}{
		ExtraData: "test",
	}

	issuanceMsgBytes, err := json.Marshal(issuanceMsg)
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
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId:      codeID,
			Label:       "smart-contract",
			IssuanceMsg: wasmtypes.RawContractMessage(issuanceMsgBytes),
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
	amountToMint = AmountBlockSmartContractTrigger
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
		chain.TxFactoryAuto(),
		msgSend,
	)
	require.NoError(t, err)
}

// TestAssetFTExtensionAttachingToSmartContractIsDenied verifies that this is not possible to attach token to smart
// contract instantiation if issuer blocked this operation.
func TestAssetFTExtensionAttachingToSmartContractInstantiationIsDenied(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issuer := chain.GenAccount()

	requireT := require.New(t)
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
			Add(sdkmath.NewInt(1_000_000)),
	})

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactory().WithSimulateAndExecute(true), issuer, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)
	attachedFund := chain.NewCoin(sdkmath.NewInt(10))

	// Issue a fungible token which cannot be sent to the smart contract
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     6,
		InitialAmount: sdkmath.NewInt(1000),
		Description:   "ABC Description",
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Funds:  sdk.NewCoins(attachedFund),
			Label:  "block-smart-contract",
		},
		BurnRate:           sdkmath.LegacyZeroDec(),
		SendCommissionRate: sdkmath.LegacyZeroDec(),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactoryAuto(),
		issueMsg,
	)

	requireT.NoError(err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

	initialPayload, err := json.Marshal(moduleswasm.SimpleState{
		Count: 1337,
	})
	requireT.NoError(err)

	// This operation should fail due to coins being attached to it
	_, _, err = chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		chain.TxFactory().WithGas(3_000_000),
		issuer,
		moduleswasm.SimpleStateWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayload,
			Amount:     sdk.NewCoin(denom, AmountBlockSmartContractTrigger),
			Label:      "simple_state",
		},
	)
	requireT.ErrorContains(err, "Transferring to or from smart contracts are prohibited.")
}

// TestAssetFTExtensionMintingAndSendingOnBehalfOfIssuingSmartContractIsPossibleEvenIfSmartContractsAreBlocked verifies
// that it is possible to use authz to mint and send the token on behalf of the issuing smart contract if smart
// contracts are blocked.
func TestAssetFTExtensionMintingAndSendingOnBehalfOfIssuingSmartContractIsPossibleEvenIfSmartContractsAreBlocked(
	t *testing.T,
) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	admin := chain.GenAccount()
	grantee := chain.GenAccount()
	recipient := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: admin,
			Options: integration.BalancesOptions{
				Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
					Add(sdkmath.NewInt(1_000_000)),
			},
		}, {
			Acc: grantee,
			Options: integration.BalancesOptions{
				Amount: sdkmath.NewInt(3 * 500_000),
			},
		},
	})

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactoryAuto(), admin, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)

	// instantiate new contract
	initialPayload, err := json.Marshal(struct{}{})
	requireT.NoError(err)

	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		chain.TxFactoryAuto(),
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
		chain.TxFactoryAuto(),
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
			Amount:      sdk.NewCoins(sdk.NewCoin(denom, AmountBlockSmartContractTrigger)),
		},
		&assetfttypes.MsgMint{
			Sender:    contractAddr,
			Recipient: recipient.String(),
			Coin:      sdk.NewCoin(denom, AmountBlockSmartContractTrigger),
		},
	})

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(grantee),
		chain.TxFactoryAuto(),
		&execMsg,
	)
	requireT.NoError(err)

	// check balances

	contract := sdk.MustAccAddressFromBech32(contractAddr)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&contract:  889,
		&recipient: 222,
	})
}

// TestAssetFTExtensionBurnRate checks extension burn rate functionality of fungible tokens.
func TestAssetFTExtensionBurnRate(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	admin := chain.GenAccount()
	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: admin,
			Options: integration.BalancesOptions{
				Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
					Add(sdkmath.NewInt(1_000_000)). // added 1 million for smart contract upload
					Add(sdkmath.NewInt(2 * 500_000)),
			},
		}, {
			Acc: recipient1,
			Options: integration.BalancesOptions{
				Amount: sdkmath.NewInt(2 * 500_000),
			},
		}, {
			Acc: recipient2,
			Options: integration.BalancesOptions{
				Amount: sdkmath.NewInt(500_000),
			},
		},
	})

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactory().WithSimulateAndExecute(true), admin, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)

	// Issue a fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        admin.String(),
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     6,
		InitialAmount: sdkmath.NewInt(1000),
		Description:   "ABC Description",
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Label:  "testing-burn-rate",
		},
		BurnRate:           sdkmath.LegacyMustNewDecFromStr("0.10"),
		SendCommissionRate: sdkmath.LegacyNewDec(0),
	}

	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactoryAuto(),
		issueMsg,
	)

	requireT.NoError(err)
	tokenIssuedEvents, err := event.FindTypedEvents[*assetfttypes.EventIssued](res.Events)
	requireT.NoError(err)
	denom := tokenIssuedEvents[0].Denom

	// send from admin to recipient1 (burn must apply if the extension decides)
	sendMsg := &banktypes.MsgSend{
		FromAddress: admin.String(),
		ToAddress:   recipient1.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(400))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactoryAuto(),
		sendMsg,
	)
	requireT.NoError(err)

	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&admin:      560, // 1000 - 400 - 40 (10% burn rate)
		&recipient1: 400,
	})

	// send trigger amount from recipient1 to recipient2 (burn must not apply if the extension decides)
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient1.String(),
		ToAddress:   recipient2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(testcontracts.AmountIgnoreBurnRateTrigger))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient1),
		chain.TxFactoryAuto(),
		sendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&admin:      560,
		&recipient1: 292, // 400 - 108 (AmountIgnoreBurnRateTrigger)
		&recipient2: 108, // AmountIgnoreBurnRateTrigger
	})

	// send from recipient1 to recipient2 (burn must apply)
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient1.String(),
		ToAddress:   recipient2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient1),
		chain.TxFactoryAuto(),
		sendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&admin:      560,
		&recipient1: 182, // 292 - 100 - 10 (10% burn rate)
		&recipient2: 208, // 108 + 100
	})

	// send from recipient2 to admin (burn must not apply)
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient2.String(),
		ToAddress:   admin.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient2),
		chain.TxFactoryAuto(),
		sendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&admin:      660, // 560 + 100
		&recipient1: 182,
		&recipient2: 98, // 208 - 100 - 10 (10% burn rate)
	})
}

// TestAssetFTExtensionSendCommissionRate checks extension send commission rate functionality of fungible tokens.
func TestAssetFTExtensionSendCommissionRate(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	assetFTClient := assetfttypes.NewQueryClient(chain.ClientContext)

	requireT := require.New(t)
	admin := chain.GenAccount()
	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: admin,
			Options: integration.BalancesOptions{
				Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
					Add(sdkmath.NewInt(1_000_000)). // added 1 million for smart contract upload
					Add(sdkmath.NewInt(2 * 500_000)),
			},
		}, {
			Acc: recipient1,
			Options: integration.BalancesOptions{
				Amount: sdkmath.NewInt(2 * 500_000),
			},
		}, {
			Acc: recipient2,
			Options: integration.BalancesOptions{
				Amount: sdkmath.NewInt(1 * 500_000),
			},
		},
	})

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactory().WithSimulateAndExecute(true), admin, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)

	// Issue a fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        admin.String(),
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     6,
		InitialAmount: sdkmath.NewInt(1000),
		Description:   "ABC Description",
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Label:  "testing-send-commission-rate",
		},
		BurnRate:           sdkmath.LegacyNewDec(0),
		SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("0.10"),
	}

	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactoryAuto(),
		issueMsg,
	)

	requireT.NoError(err)
	tokenIssuedEvents, err := event.FindTypedEvents[*assetfttypes.EventIssued](res.Events)
	requireT.NoError(err)
	denom := tokenIssuedEvents[0].Denom

	token, err := assetFTClient.Token(ctx, &assetfttypes.QueryTokenRequest{Denom: denom})
	requireT.NoError(err)
	extension := sdk.MustAccAddressFromBech32(token.GetToken().ExtensionCWAddress)

	// send from admin to recipient1 (send commission rate must apply if the extension decides)
	sendMsg := &banktypes.MsgSend{
		FromAddress: admin.String(),
		ToAddress:   recipient1.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(400))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactoryAuto(),
		sendMsg,
	)
	requireT.NoError(err)

	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&admin:      580, // 1000 - 400 - 40 (10% commission from sender) + 20 (50% of the commission to the admin)
		&recipient1: 400,
		&extension:  20, // 50% of the commission to the extension
	})

	// send trigger amount from recipient1 to recipient2 (send commission rate must not apply)
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient1.String(),
		ToAddress:   recipient2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(testcontracts.AmountIgnoreSendCommissionRateTrigger))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient1),
		chain.TxFactoryAuto(),
		sendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&admin:      580,
		&recipient1: 291, // 400 - 109 (AmountIgnoreSendCommissionRateTrigger)
		&recipient2: 109, // AmountIgnoreSendCommissionRateTrigger
		&extension:  20,
	})

	// send from recipient1 to recipient2 (send commission rate must apply)
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient1.String(),
		ToAddress:   recipient2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient1),
		chain.TxFactoryAuto(),
		sendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&admin:      585, // 580 + 5 (50% of the commission to the admin)
		&recipient1: 181, // 291 - 100 - 10 (10% commission from sender)
		&recipient2: 209, // 109 + 100
		&extension:  25,  // 20 + 5 (50% of the commission to the extension)
	})

	// send from recipient2 to admin (send commission rate must apply if the extension decides)
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient2.String(),
		ToAddress:   admin.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient2),
		chain.TxFactoryAuto(),
		sendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&admin:      690, // 585 + 100 + 5 (50% of the commission to the admin)
		&recipient1: 181,
		&recipient2: 99, // 209 - 100 - 10 (10% commission from sender)
		&extension:  30, // 25 + 5 (50% of the commission to the extension)
	})
}

// TestAssetFTExtensionDEX checks extension works with the DEX.
func TestAssetFTExtensionDEX(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	assetFTClint := assetfttypes.NewQueryClient(chain.ClientContext)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)
	dexReserver := dexParamsRes.Params.OrderReserve

	admin := chain.GenAccount()
	acc1 := chain.GenAccount()
	acc2 := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: admin,
			Options: integration.BalancesOptions{
				Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
					AddRaw(1_000_000). // added 1 million for smart contract upload
					AddRaw(2 * 500_000),
			},
		}, {
			Acc: acc1,
			Options: integration.BalancesOptions{
				Amount: sdkmath.NewInt(500_000).Add(dexReserver.Amount), // message + order reserve
			},
		}, {
			Acc: acc2,
			Options: integration.BalancesOptions{
				Amount: sdkmath.NewInt(500_000).
					AddRaw(200_000_000).
					Add(dexReserver.Amount), // message  + balance to place an order + order reserve
			},
		},
	})

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactory().WithSimulateAndExecute(true), admin, testcontracts.AssetExtensionWasm,
	)
	requireT.NoError(err)

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        admin.String(),
		Symbol:        "EXABC",
		Subunit:       "extabc",
		Precision:     6,
		InitialAmount: sdkmath.NewInt(1_000_000_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Label:  "testing-dex",
		},
	}

	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactoryAuto(),
		issueMsg,
	)
	requireT.NoError(err)
	tokenIssuedEvents, err := event.FindTypedEvents[*assetfttypes.EventIssued](res.Events)
	requireT.NoError(err)
	denomWithExtension := tokenIssuedEvents[0].Denom

	// send from admin to acc1 to place an order
	sendMsg := &banktypes.MsgSend{
		FromAddress: admin.String(),
		ToAddress:   acc1.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denomWithExtension, sdkmath.NewInt(400_000_000))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactoryAuto(),
		sendMsg,
	)
	requireT.NoError(err)

	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denomWithExtension,
		QuoteDenom:  chain.ChainSettings.Denom,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1")),
		Quantity:    sdkmath.NewInt(testcontracts.AmountDEXExpectToReceiveTrigger),
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.ErrorContains(err, "wasm error: DEX order placement is failed")

	// update to allowed
	placeSellOrderMsg.Quantity = sdkmath.NewInt(100_000_000)
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	acc1BalanceRes, err := assetFTClint.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denomWithExtension,
	})
	requireT.NoError(err)
	requireT.True(acc1BalanceRes.ExpectedToReceiveInDEX.IsZero())
	requireT.Equal(sdkmath.NewInt(100_000_000).String(), acc1BalanceRes.LockedInDEX.String())

	// place buy order from acc2

	acc2BalanceRes, err := assetFTClint.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc2.String(),
		Denom:   denomWithExtension,
	})
	requireT.NoError(err)
	// no coins of the denomWithExtension
	requireT.True(acc2BalanceRes.Balance.IsZero())

	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc2.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denomWithExtension,
		QuoteDenom:  chain.ChainSettings.Denom,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1")),
		Quantity:    sdkmath.NewInt(testcontracts.AmountDEXExpectToSpendTrigger),
		Side:        dextypes.SIDE_BUY,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactoryAuto(),
		placeBuyOrderMsg,
	)
	requireT.ErrorContains(err, "wasm error: DEX order placement is failed")

	// update to allowed
	placeBuyOrderMsg.Quantity = sdkmath.NewInt(100_000_000)
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactoryAuto(),
		placeBuyOrderMsg,
	)
	requireT.NoError(err)

	// both order are executed and closed
	acc2BalanceRes, err = assetFTClint.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc2.String(),
		Denom:   denomWithExtension,
	})
	requireT.NoError(err)
	// bought expected quantity
	requireT.Equal(sdkmath.NewInt(100_000_000).String(), acc2BalanceRes.Balance.String())
	requireT.True(acc2BalanceRes.LockedInDEX.IsZero())
	requireT.True(acc2BalanceRes.ExpectedToReceiveInDEX.IsZero())

	acc1BalanceRes, err = assetFTClint.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denomWithExtension,
	})
	requireT.NoError(err)
	requireT.True(acc1BalanceRes.LockedInDEX.IsZero())
	requireT.True(acc1BalanceRes.ExpectedToReceiveInDEX.IsZero())
}
