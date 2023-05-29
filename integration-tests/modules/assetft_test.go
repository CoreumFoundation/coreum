//go:build integrationtests

package modules

import (
	"context"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// TestAssetFTQueryParams queries parameters of asset/ft module.
func TestAssetFTQueryParams(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	issueFee := getIssueFee(ctx, t, chain.ClientContext)

	assert.True(t, issueFee.Amount.GT(sdk.ZeroInt()))
	assert.Equal(t, chain.ChainSettings.Denom, issueFee.Denom)
}

// TestAssetFTIssue tests issue functionality of fungible tokens.
func TestAssetFTIssue(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdk.NewInt(1000),
		Features:      []assetfttypes.Feature{},
	}

	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)

	// verify issue fee was burnt
	burntStr, err := event.FindStringEventAttribute(res.Events, banktypes.EventTypeCoinBurn, sdk.AttributeKeyAmount)
	requireT.NoError(err)
	requireT.Equal(getIssueFee(ctx, t, chain.ClientContext).String(), burntStr)

	// check that balance is 0 meaning issue fee was taken

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: issuer.String(),
		Denom:   chain.ChainSettings.Denom,
	})
	requireT.NoError(err)
	requireT.Equal(chain.NewCoin(sdk.ZeroInt()).String(), resp.Balance.String())
}

// TestAssetFTIssueFeeProposal tests proposal upgrading issue fee.
func TestAssetFTIssueFeeProposal(t *testing.T) {
	// This test can't be run together with other tests because it affects balances due to unexpected issue fee.
	// That's why t.Parallel() is not here.

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	origIssueFee := getIssueFee(ctx, t, chain.ClientContext)

	chain.Governance.UpdateParams(ctx, t, "Propose changing IssueFee in the assetft module",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(assetfttypes.ModuleName, string(assetfttypes.KeyIssueFee), string(must.Bytes(tmjson.Marshal(sdk.NewCoin(origIssueFee.Denom, sdk.ZeroInt()))))),
		})

	issuer := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
		},
	})

	// Issue token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdk.NewInt(1000),
		Features:      []assetfttypes.Feature{},
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)

	// Revert to original issue fee
	chain.Governance.UpdateParams(ctx, t, "Propose changing IssueFee in the assetft module",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(assetfttypes.ModuleName, string(assetfttypes.KeyIssueFee), string(must.Bytes(tmjson.Marshal(origIssueFee)))),
		})
}

// TestAssetIssueAndQueryTokens checks that tokens query works as expected.
func TestAssetIssueAndQueryTokens(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	clientCtx := chain.ClientContext

	ftClient := assetfttypes.NewQueryClient(clientCtx)

	issueFee := getIssueFee(ctx, t, chain.ClientContext).Amount

	issuer1 := chain.GenAccount()
	chain.FundAccountsWithOptions(ctx, t, issuer1, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&assetfttypes.MsgIssue{}},
		Amount:   issueFee,
	})

	issuer2 := chain.GenAccount()
	chain.FundAccountsWithOptions(ctx, t, issuer2, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&assetfttypes.MsgIssue{}},
		Amount:   issueFee,
	})

	// issue the new fungible token form issuer1
	msg1 := &assetfttypes.MsgIssue{
		Issuer:             issuer1.String(),
		Symbol:             "WBTC",
		Subunit:            "wsatoshi",
		Precision:          8,
		Description:        "Wrapped BTC",
		InitialAmount:      sdk.NewInt(777),
		BurnRate:           sdk.NewDec(0),
		SendCommissionRate: sdk.NewDec(0),
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg1)),
		msg1,
	)
	require.NoError(t, err)

	// issue the new fungible token form issuer2
	msg2 := msg1
	msg2.Issuer = issuer2.String()
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer2),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg2)),
		msg1,
	)

	require.NoError(t, err)

	// query for the tokens
	gotToken, err := ftClient.Tokens(ctx, &assetfttypes.QueryTokensRequest{
		Issuer: issuer1.String(),
	})
	requireT.NoError(err)

	denom := assetfttypes.BuildDenom(msg1.Subunit, issuer1)
	requireT.Equal(1, len(gotToken.Tokens))
	requireT.Equal(assetfttypes.Token{
		Denom:              denom,
		Issuer:             issuer1.String(),
		Symbol:             msg1.Symbol,
		Subunit:            "wsatoshi",
		Precision:          8,
		Description:        msg1.Description,
		BurnRate:           msg1.BurnRate,
		SendCommissionRate: msg1.SendCommissionRate,
	}, gotToken.Tokens[0])
}

// TestBalanceQuery tests balance query.
func TestBalanceQuery(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	assertT := assert.New(t)

	issueFee := getIssueFee(ctx, t, chain.ClientContext).Amount

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgSetWhitelistedLimit{},
			&assetfttypes.MsgFreeze{},
			&banktypes.MsgSend{},
		},
		Amount: issueFee,
	})

	// issue the new fungible token form issuer
	msgIssue := &assetfttypes.MsgIssue{
		Issuer:             issuer.String(),
		Symbol:             "WBTC",
		Subunit:            "wsatoshi",
		Precision:          8,
		Description:        "Wrapped BTC",
		InitialAmount:      sdk.NewInt(200),
		BurnRate:           sdk.NewDec(0),
		SendCommissionRate: sdk.NewDec(0),
		Features:           []assetfttypes.Feature{assetfttypes.Feature_freezing, assetfttypes.Feature_whitelisting},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgIssue)),
		msgIssue,
	)
	require.NoError(t, err)

	denom := assetfttypes.BuildDenom(msgIssue.Subunit, issuer)
	whitelistedCoin := sdk.NewInt64Coin(denom, 30)
	frozenCoin := sdk.NewInt64Coin(denom, 20)
	sendCoin := sdk.NewInt64Coin(denom, 10)

	msgWhitelist := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    whitelistedCoin,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgWhitelist)),
		msgWhitelist,
	)
	require.NoError(t, err)

	msgFreeze := &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    frozenCoin,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgFreeze)),
		msgFreeze,
	)
	require.NoError(t, err)

	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sendCoin),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	require.NoError(t, err)

	ftClient := assetfttypes.NewQueryClient(chain.ClientContext)
	resp, err := ftClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	require.NoError(t, err)

	assertT.Equal(whitelistedCoin.Amount.String(), resp.Whitelisted.String())
	assertT.Equal(frozenCoin.Amount.String(), resp.Frozen.String())
	assertT.Equal(sendCoin.Amount.String(), resp.Balance.String())
	assertT.Equal("0", resp.Locked.String())
}

// TestEmptyBalanceQuery tests balance query.
func TestEmptyBalanceQuery(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	assertT := assert.New(t)

	account := chain.GenAccount()

	ftClient := assetfttypes.NewQueryClient(chain.ClientContext)
	resp, err := ftClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: account.String(),
		Denom:   "nonexistingdenom",
	})
	require.NoError(t, err)

	assertT.Equal("0", resp.Whitelisted.String())
	assertT.Equal("0", resp.Frozen.String())
	assertT.Equal("0", resp.Balance.String())
	assertT.Equal("0", resp.Locked.String())
}

// TestAssetFTMint tests mint functionality of fungible tokens.
func TestAssetFTMint(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	randomAddress := chain.GenAccount()
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgMint{},
			&assetfttypes.MsgMint{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount.MulRaw(2),
	})

	chain.FundAccountsWithOptions(ctx, t, randomAddress, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgMint{},
		},
	})

	// Issue an unmintable fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABCNotMintable",
		Subunit:       "uabcnotmintable",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdk.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_burning,
			assetfttypes.Feature_freezing,
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
	mintMsg := &assetfttypes.MsgMint{
		Sender: issuer.String(),
		Coin: sdk.Coin{
			Denom:  unmintableDenom,
			Amount: sdk.NewInt(1000),
		},
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.True(assetfttypes.ErrFeatureDisabled.Is(err))

	// Issue a mintable fungible token
	issueMsg = &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABCMintable",
		Subunit:       "uabcmintable",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdk.NewInt(1000),
		Features:      []assetfttypes.Feature{assetfttypes.Feature_minting},
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

	// try to pass non-issuer signature to msg
	mintMsg = &assetfttypes.MsgMint{
		Sender: randomAddress.String(),
		Coin:   sdk.NewCoin(mintableDenom, sdk.NewInt(1000)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(randomAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.Error(err)
	assertT.True(sdkerrors.ErrUnauthorized.Is(err))

	// mint tokens and check balance and total supply
	oldSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: mintableDenom})
	requireT.NoError(err)
	mintCoin := sdk.NewCoin(mintableDenom, sdk.NewInt(1600))
	mintMsg = &assetfttypes.MsgMint{
		Sender: issuer.String(),
		Coin:   mintCoin,
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
	assertT.EqualValues(mintCoin.Add(sdk.NewCoin(mintableDenom, sdk.NewInt(1000))).String(), balance.GetBalance().String())

	newSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: mintableDenom})
	requireT.NoError(err)
	assertT.EqualValues(mintCoin, newSupply.GetAmount().Sub(oldSupply.GetAmount()))
}

// TestAssetFTBurn tests burn functionality of fungible tokens.
func TestAssetFTBurn(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgBurn{},
			&assetfttypes.MsgBurn{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount.MulRaw(2),
	})

	chain.FundAccountsWithOptions(ctx, t, recipient, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgBurn{},
			&assetfttypes.MsgBurn{},
		},
	})

	// Issue an unburnable fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABCNotBurnable",
		Subunit:       "uabcnotburnable",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdk.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_minting,
			assetfttypes.Feature_freezing,
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
	burnMsg := &assetfttypes.MsgBurn{
		Sender: issuer.String(),
		Coin: sdk.Coin{
			Denom:  unburnable,
			Amount: sdk.NewInt(900),
		},
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
		Amount:      sdk.NewCoins(sdk.NewCoin(unburnable, sdk.NewInt(100))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	// try to burn unburnable token from recipient
	burnMsg = &assetfttypes.MsgBurn{
		Sender: recipient.String(),
		Coin: sdk.Coin{
			Denom:  unburnable,
			Amount: sdk.NewInt(1000),
		},
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.True(assetfttypes.ErrFeatureDisabled.Is(err))

	// Issue a burnable fungible token
	issueMsg = &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABCBurnable",
		Subunit:       "uabcburnable",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdk.NewInt(1000),
		Features:      []assetfttypes.Feature{assetfttypes.Feature_burning},
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
		Amount:      sdk.NewCoins(sdk.NewCoin(burnableDenom, sdk.NewInt(100))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	// try to pass non-issuer signature to msg
	burnMsg = &assetfttypes.MsgBurn{
		Sender: recipient.String(),
		Coin:   sdk.NewCoin(burnableDenom, sdk.NewInt(100)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.NoError(err)

	// burn tokens and check balance and total supply
	oldSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: burnableDenom})
	requireT.NoError(err)
	burnCoin := sdk.NewCoin(burnableDenom, sdk.NewInt(600))

	burnMsg = &assetfttypes.MsgBurn{
		Sender: issuer.String(),
		Coin:   burnCoin,
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
	assertT.EqualValues(sdk.NewCoin(burnableDenom, sdk.NewInt(300)).String(), balance.GetBalance().String())

	newSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: burnableDenom})
	requireT.NoError(err)
	assertT.EqualValues(burnCoin, oldSupply.GetAmount().Sub(newSupply.GetAmount()))
}

// TestAssetFTBurnRate tests burn rate functionality of fungible tokens.
//
//nolint:dupl
func TestAssetFTBurnRate(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()
	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})
	chain.FundAccountsWithOptions(ctx, t, recipient1, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})
	chain.FundAccountsWithOptions(ctx, t, recipient2, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})

	// Issue a fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             issuer.String(),
		Symbol:             "ABC",
		Subunit:            "abc",
		Precision:          6,
		InitialAmount:      sdk.NewInt(1000),
		Description:        "ABC Description",
		Features:           []assetfttypes.Feature{},
		BurnRate:           sdk.MustNewDecFromStr("0.10"),
		SendCommissionRate: sdk.NewDec(0),
	}

	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	tokenIssuedEvents, err := event.FindTypedEvents[*assetfttypes.EventIssued](res.Events)
	requireT.NoError(err)
	denom := tokenIssuedEvents[0].Denom

	// send from issuer to recipient1 (burn must not apply)
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient1.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(400))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&issuer:     600,
		&recipient1: 400,
	})

	// send from recipient1 to recipient2 (burn must apply)
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient1.String(),
		ToAddress:   recipient2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&issuer:     600,
		&recipient1: 290,
		&recipient2: 100,
	})

	// send from recipient2 to issuer (burn must not apply)
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient2.String(),
		ToAddress:   issuer.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient2),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&issuer:     700,
		&recipient1: 290,
		&recipient2: 0,
	})

	// multi send from recipient1 to issuer and recipient2
	// (burn must apply to one of outputs, deducted from recipient 1)
	multiSendMsg := &banktypes.MsgMultiSend{
		Inputs: []banktypes.Input{
			{Address: recipient1.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(200)))},
		},
		Outputs: []banktypes.Output{
			{Address: issuer.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100)))},
			{Address: recipient2.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100)))},
		},
	}

	chain.FundAccountsWithOptions(ctx, t, recipient1, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			multiSendMsg,
		},
	})

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(multiSendMsg)),
		multiSendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&issuer:     800,
		&recipient1: 80,
		&recipient2: 100,
	})
}

// TestAssetFTSendCommissionRate tests send commission rate functionality of fungible tokens.
//
//nolint:dupl
func TestAssetFTSendCommissionRate(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()
	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})
	chain.FundAccountsWithOptions(ctx, t, recipient1, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})
	chain.FundAccountsWithOptions(ctx, t, recipient2, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})

	// Issue a fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             issuer.String(),
		Symbol:             "ABC",
		Subunit:            "abc",
		Precision:          6,
		InitialAmount:      sdk.NewInt(1000),
		Description:        "ABC Description",
		Features:           []assetfttypes.Feature{},
		BurnRate:           sdk.NewDec(0),
		SendCommissionRate: sdk.MustNewDecFromStr("0.10"),
	}

	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	tokenIssuedEvents, err := event.FindTypedEvents[*assetfttypes.EventIssued](res.Events)
	requireT.NoError(err)
	denom := tokenIssuedEvents[0].Denom

	// send from issuer to recipient1 (send commission rate must not apply)
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient1.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(400))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&issuer:     600,
		&recipient1: 400,
	})

	// send from recipient1 to recipient2 (send commission rate must apply)
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient1.String(),
		ToAddress:   recipient2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&issuer:     610,
		&recipient1: 290,
		&recipient2: 100,
	})

	// send from recipient2 to issuer (send commission rate must not apply)
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient2.String(),
		ToAddress:   issuer.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient2),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&issuer:     710,
		&recipient1: 290,
		&recipient2: 0,
	})

	// multi send from recipient1 to issuer and recipient2
	// (send commission rate must apply to one of transfers)
	multiSendMsg := &banktypes.MsgMultiSend{
		Inputs: []banktypes.Input{
			{Address: recipient1.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(200)))},
		},
		Outputs: []banktypes.Output{
			{Address: issuer.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100)))},
			{Address: recipient2.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100)))},
		},
	}

	chain.FundAccountsWithOptions(ctx, t, recipient1, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			multiSendMsg,
		},
	})

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(multiSendMsg)),
		multiSendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&issuer:     820,
		&recipient1: 80,
		&recipient2: 100,
	})
}

// TestAssetFTFreeze checks freeze functionality of fungible tokens.
func TestAssetFTFreeze(t *testing.T) {
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
	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgFreeze{},
			&assetfttypes.MsgFreeze{},
			&assetfttypes.MsgUnfreeze{},
			&assetfttypes.MsgUnfreeze{},
			&assetfttypes.MsgUnfreeze{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})
	chain.FundAccountsWithOptions(ctx, t, recipient, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&banktypes.MsgMultiSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgMultiSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
		},
	})
	chain.FundAccountsWithOptions(ctx, t, randomAddress, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgFreeze{},
		},
	})

	// Issue the new fungible token
	msg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdk.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_freezing,
		},
	}

	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(assetfttypes.BuildDenom(msg.Subunit, issuer), sdk.NewInt(1000)),
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

	// try to pass non-issuer signature to freeze msg
	freezeMsg := &assetfttypes.MsgFreeze{
		Sender:  randomAddress.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(1000)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(randomAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	requireT.Error(err)
	assertT.True(sdkerrors.ErrUnauthorized.Is(err))

	// freeze 400 tokens
	freezeMsg = &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(400)),
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
		PreviousAmount: sdk.NewInt(0),
		CurrentAmount:  sdk.NewInt(400),
	}, fungibleTokenFreezeEvts[0])

	// query frozen tokens
	frozenBalance, err := ftClient.FrozenBalance(ctx, &assetfttypes.QueryFrozenBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(400)), frozenBalance.Balance)

	frozenBalances, err := ftClient.FrozenBalances(ctx, &assetfttypes.QueryFrozenBalancesRequest{
		Account: recipient.String(),
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(400))), frozenBalances.Balances)

	// try to send more than available (650) (600 is available)
	recipient2 := chain.GenAccount()
	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(650)))
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
	requireT.Error(err)
	assertT.True(sdkerrors.ErrInsufficientFunds.Is(err))
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
	requireT.Error(err)
	assertT.True(sdkerrors.ErrInsufficientFunds.Is(err))

	// try to send available tokens (300 + 300)
	coinsToSend = sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(300)))
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
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(700)).String(), balance1.GetBalance().String())
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
	requireT.NoError(err)
	balance1, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(400)).String(), balance1.GetBalance().String())

	// unfreeze 200 tokens and try to send 250 tokens
	unfreezeMsg := &assetfttypes.MsgUnfreeze{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(200)),
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
		PreviousAmount: sdk.NewInt(400),
		CurrentAmount:  sdk.NewInt(200),
	}, fungibleTokenFreezeEvts[0])

	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   recipient2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(250))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.Error(err)
	assertT.True(sdkerrors.ErrInsufficientFunds.Is(err))

	// send available tokens (200)
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   recipient2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(200))),
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
		Coin:    sdk.NewCoin(denom, sdk.NewInt(400)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(unfreezeMsg)),
		unfreezeMsg,
	)
	requireT.True(sdkerrors.ErrInsufficientFunds.Is(err))

	// unfreeze 200 tokens and observer current frozen amount is zero
	unfreezeMsg = &assetfttypes.MsgUnfreeze{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(200)),
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
		PreviousAmount: sdk.NewInt(200),
		CurrentAmount:  sdk.NewInt(0),
	}, fungibleTokenFreezeEvts[0])
}

// TestAssetFTFreezeUnfreezable checks freeze functionality on unfreezable fungible tokens.
func TestAssetFTFreezeUnfreezable(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgFreeze{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})

	// Issue an unfreezable fungible token
	msg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABCNotFreezable",
		Subunit:       "uabcnotfreezable",
		Description:   "ABC Description",
		Precision:     1,
		InitialAmount: sdk.NewInt(1000),
		Features:      []assetfttypes.Feature{},
	}

	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)

	requireT.NoError(err)
	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assetfttypes.EventIssued](res.Events)
	requireT.NoError(err)
	unfreezableDenom := fungibleTokenIssuedEvts[0].Denom

	// try to freeze unfreezable token
	freezeMsg := &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(unfreezableDenom, sdk.NewInt(1000)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	assertT.True(assetfttypes.ErrFeatureDisabled.Is(err))
}

// TestAssetFTFreezeIssuerAccount checks that freezing the issuer account is not possible.
func TestAssetFTFreezeIssuerAccount(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()
	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgFreeze{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})

	// Issue an freezable fungible token
	msg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     1,
		Description:   "ABC Description",
		InitialAmount: sdk.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_freezing,
		},
	}

	denom := assetfttypes.BuildDenom(msg.Subunit, issuer)
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)
	requireT.NoError(err)

	// try to freeze issuer account
	freezeMsg := &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: issuer.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(1000)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	requireT.ErrorIs(err, sdkerrors.ErrUnauthorized)
}

// TestAssetFTGloballyFreeze checks global freeze functionality of fungible tokens.
func TestAssetFTGloballyFreeze(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgGloballyFreeze{},
			&banktypes.MsgSend{},
			&banktypes.MsgMultiSend{},
			&assetfttypes.MsgGloballyUnfreeze{},
			&banktypes.MsgSend{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})
	chain.FundAccountsWithOptions(ctx, t, recipient, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&banktypes.MsgMultiSend{},
			&banktypes.MsgSend{},
		},
	})

	// Issue the new fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "FREEZE",
		Subunit:       "freeze",
		Precision:     6,
		Description:   "FREEZE Description",
		InitialAmount: sdk.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_freezing,
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
	denom := fungibleTokenIssuedEvts[0].Denom

	// Globally freeze Token.
	globFreezeMsg := &assetfttypes.MsgGloballyFreeze{
		Sender: issuer.String(),
		Denom:  denom,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(globFreezeMsg)),
		globFreezeMsg,
	)
	requireT.NoError(err)

	// Try to send Token.
	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(50)))
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
	requireT.NoError(err)

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
	requireT.NoError(err)

	// send from recipient
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   issuer.String(),
		Amount:      coinsToSend,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.True(assetfttypes.ErrGloballyFrozen.Is(err))

	// multi-send
	multiSendMsg = &banktypes.MsgMultiSend{
		Inputs:  []banktypes.Input{{Address: recipient.String(), Coins: coinsToSend}},
		Outputs: []banktypes.Output{{Address: issuer.String(), Coins: coinsToSend}},
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(multiSendMsg)),
		multiSendMsg,
	)
	requireT.True(assetfttypes.ErrGloballyFrozen.Is(err))

	// Globally unfreeze Token.
	globUnfreezeMsg := &assetfttypes.MsgGloballyUnfreeze{
		Sender: issuer.String(),
		Denom:  denom,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(globUnfreezeMsg)),
		globUnfreezeMsg,
	)
	requireT.NoError(err)

	// Try to send Token from issuer.
	sendMsg = &banktypes.MsgSend{
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
	requireT.NoError(err)

	// Try to send Token from recipient.
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   issuer.String(),
		Amount:      coinsToSend,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
}

// TestAssetCommissionRateExceedFreeze checks tx will fail if send commission causes
// breach of freeze limit functionality.
func TestAssetCommissionRateExceedFreeze(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgFreeze{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})
	chain.FundAccountsWithOptions(ctx, t, recipient, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})

	// Issue the new fungible token
	msgIssue := &assetfttypes.MsgIssue{
		Issuer:             issuer.String(),
		Symbol:             "ABC",
		Subunit:            "uabc",
		Precision:          6,
		Description:        "ABC Description",
		InitialAmount:      sdk.NewInt(1000),
		SendCommissionRate: sdk.MustNewDecFromStr("0.3"),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_freezing,
		},
	}
	denom := assetfttypes.BuildDenom(msgIssue.Subunit, issuer)
	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denom, sdk.NewInt(1000)),
		),
	}

	msgFreeze := &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(650)),
	}

	msgList := []sdk.Msg{
		msgIssue, msgSend, msgFreeze,
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgList...)),
		msgList...,
	)
	requireT.NoError(err)

	// try to send more than available (300 + 60) (1000 - 650(frozen) = 350 is available)
	recipient2 := chain.GenAccount()
	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(300)))
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
	requireT.Error(err)
	assertT.True(sdkerrors.ErrInsufficientFunds.Is(err))
}

// TestSendCoreTokenWithRestrictedToken checks tx will fail if try to send core token
// alongside restricted user issued token.
func TestSendCoreTokenWithRestrictedToken(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgFreeze{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})
	chain.FundAccountsWithOptions(ctx, t, recipient, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
		Amount: sdk.NewInt(1000),
	})

	// Issue the new fungible token
	msgIssue := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdk.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_freezing,
		},
	}
	denom := assetfttypes.BuildDenom(msgIssue.Subunit, issuer)
	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denom, sdk.NewInt(1000)),
		),
	}

	msgFreeze := &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(800)),
	}

	msgList := []sdk.Msg{
		msgIssue, msgSend, msgFreeze,
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgList...)),
		msgList...,
	)
	requireT.NoError(err)

	// try to send core token and minted token with freezing violation
	recipient2 := chain.GenAccount()
	coinsToSend := sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(210)),
		chain.NewCoin(sdk.NewInt(1000)),
	)

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
	requireT.Error(err)
	assertT.True(sdkerrors.ErrInsufficientFunds.Is(err))
}

// TestNotEnoughBalanceForBurnRate checks tx will fail if there is not enough balance to cover burn rate.
//
//nolint:dupl // we expect code duplication in tests
func TestNotEnoughBalanceForBurnRate(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})
	chain.FundAccountsWithOptions(ctx, t, recipient, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})

	// Issue the new fungible token
	msgIssue := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdk.NewInt(1000),
		BurnRate:      sdk.MustNewDecFromStr("0.1"),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_freezing,
		},
	}
	denom := assetfttypes.BuildDenom(msgIssue.Subunit, issuer)
	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denom, sdk.NewInt(1000)),
		),
	}

	msgList := []sdk.Msg{
		msgIssue, msgSend,
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgList...)),
		msgList...,
	)
	requireT.NoError(err)

	// try to send, it should fail (920 + 92 = 1012 > 1000)
	recipient2 := chain.GenAccount()
	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(920)))

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
	requireT.Error(err)
	assertT.True(sdkerrors.ErrInsufficientFunds.Is(err))
}

// TestNotEnoughBalanceForCommissionRate checks tx will fail if there is not enough balance to cover commission rate.
//
//nolint:dupl // we expect code duplication in tests
func TestNotEnoughBalanceForCommissionRate(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})
	chain.FundAccountsWithOptions(ctx, t, recipient, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})

	// Issue the new fungible token
	msgIssue := &assetfttypes.MsgIssue{
		Issuer:             issuer.String(),
		Symbol:             "ABC",
		Subunit:            "uabc",
		Precision:          6,
		Description:        "ABC Description",
		InitialAmount:      sdk.NewInt(1000),
		SendCommissionRate: sdk.MustNewDecFromStr("0.1"),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_freezing,
		},
	}
	denom := assetfttypes.BuildDenom(msgIssue.Subunit, issuer)
	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denom, sdk.NewInt(1000)),
		),
	}

	msgList := []sdk.Msg{
		msgIssue, msgSend,
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgList...)),
		msgList...,
	)
	requireT.NoError(err)

	// try to send, it should fail (920 + 92 = 1012 > 1000)
	recipient2 := chain.GenAccount()
	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(920)))

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
	requireT.Error(err)
	assertT.True(sdkerrors.ErrInsufficientFunds.Is(err))
}

// TestAssetFTWhitelist checks whitelist functionality of fungible tokens.
func TestAssetFTWhitelist(t *testing.T) {
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
	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
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
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})
	chain.FundAccountsWithOptions(ctx, t, nonIssuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgSetWhitelistedLimit{},
		},
	})
	chain.FundAccountsWithOptions(ctx, t, recipient, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})

	// Issue the new fungible token
	amount := sdk.NewInt(20000)
	subunit := "uabc"
	denom := assetfttypes.BuildDenom(subunit, issuer)
	msg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: amount,
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_whitelisting,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)

	requireT.NoError(err)

	// try to pass non-issuer signature to whitelist msg
	whitelistMsg := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  nonIssuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(400)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(nonIssuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	assertT.True(sdkerrors.ErrUnauthorized.Is(err))

	// try to send to recipient before it is whitelisted (balance 0, whitelist limit 0)
	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(10)))
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
	assertT.True(assetfttypes.ErrWhitelistedLimitExceeded.Is(err))

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
	assertT.True(assetfttypes.ErrWhitelistedLimitExceeded.Is(err))

	// whitelist 400 tokens
	whitelistMsg = &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(400)),
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
	requireT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(400)), whitelistedBalance.Balance)

	whitelistedBalances, err := ftClient.WhitelistedBalances(ctx, &assetfttypes.QueryWhitelistedBalancesRequest{
		Account: recipient.String(),
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(400))), whitelistedBalances.Balances)

	// try to receive more than whitelisted (600) (possible 400)
	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(600))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	assertT.True(assetfttypes.ErrWhitelistedLimitExceeded.Is(err))

	// try to send whitelisted balance (400)
	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(400))),
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
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(400)).String(), balance.GetBalance().String())

	// try to send one more
	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	assertT.True(assetfttypes.ErrWhitelistedLimitExceeded.Is(err))

	// whitelist one more
	whitelistMsg = &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(401)),
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
	requireT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(401)), whitelistedBalance.Balance)

	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1))),
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
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(401)).String(), balance.GetBalance().String())

	// Verify that issuer has no whitelisted balance
	whitelistedBalance, err = ftClient.WhitelistedBalance(ctx, &assetfttypes.QueryWhitelistedBalanceRequest{
		Account: issuer.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdk.ZeroInt()), whitelistedBalance.Balance)

	// Send something to issuer, it should succeed despite the fact that issuer is not whitelisted
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   issuer.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(10))),
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
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(19609)).String(), balance.GetBalance().String())

	// Set whitelisted balance to 0 for recipient
	whitelistMsg = &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.ZeroInt()),
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
	assertT.True(assetfttypes.ErrWhitelistedLimitExceeded.Is(err))
}

// TestAssetFTWhitelistUnwhitelistable checks whitelist functionality on unwhitelistable fungible tokens.
func TestAssetFTWhitelistUnwhitelistable(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgSetWhitelistedLimit{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})

	// Issue an unwhitelistable fungible token
	subunit := "uabcnotwhitelistable"
	unwhitelistableDenom := assetfttypes.BuildDenom(subunit, issuer)
	amount := sdk.NewInt(1000)
	msg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABCNotWhitelistable",
		Subunit:       "uabcnotwhitelistable",
		Precision:     1,
		Description:   "ABC Description",
		InitialAmount: amount,
		Features:      []assetfttypes.Feature{},
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)

	requireT.NoError(err)

	// try to whitelist unwhitelistable token
	whitelistMsg := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(unwhitelistableDenom, sdk.NewInt(1000)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	assertT.True(assetfttypes.ErrFeatureDisabled.Is(err))
}

// TestAssetFTWhitelistIssuer checks whitelisting on issuer account is not possible.
func TestAssetFTWhitelistIssuerAccount(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()
	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgSetWhitelistedLimit{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})

	// Issue an whitelistable fungible token
	subunit := "uabcwhitelistable"
	denom := assetfttypes.BuildDenom(subunit, issuer)
	amount := sdk.NewInt(1000)
	msg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABCWhitelistable",
		Subunit:       subunit,
		Description:   "ABC Description",
		Precision:     1,
		InitialAmount: amount,
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_whitelisting,
		},
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)

	requireT.NoError(err)

	// try to whitelist issuer account
	whitelistMsg := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: issuer.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(1000)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)

	requireT.ErrorIs(err, sdkerrors.ErrUnauthorized)
}

// TestBareToken checks non of the features will work if the flags are not set.
func TestBareToken(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgMint{},
			&assetfttypes.MsgBurn{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgFreeze{},
			&assetfttypes.MsgGloballyFreeze{},
			&assetfttypes.MsgSetWhitelistedLimit{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})
	chain.FundAccountsWithOptions(ctx, t, recipient, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgBurn{},
		},
	})

	// Issue a bare token
	amount := sdk.NewInt(1000)
	msg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "baretoken",
		Subunit:       "baretoken",
		InitialAmount: amount,
		Precision:     10,
	}
	denom := assetfttypes.BuildDenom(msg.Subunit, issuer)

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)

	requireT.NoError(err)

	// try to mint
	mintMsg := &assetfttypes.MsgMint{
		Sender: issuer.String(),
		Coin:   sdk.NewCoin(denom, sdk.NewInt(1000)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	assertT.True(assetfttypes.ErrFeatureDisabled.Is(err))

	// try to burn from issuer account (must succeed)
	burnMsg := &assetfttypes.MsgBurn{
		Sender: issuer.String(),
		Coin:   sdk.NewCoin(denom, sdk.NewInt(10)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	assertT.NoError(err)

	// try to burn from non-issuer account (must fail)
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(10))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	assertT.NoError(err)

	burnMsg = &assetfttypes.MsgBurn{
		Sender: recipient.String(),
		Coin:   sdk.NewCoin(denom, sdk.NewInt(10)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	assertT.ErrorIs(err, assetfttypes.ErrFeatureDisabled)

	// try to whitelist
	whitelistMsg := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(1000)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	assertT.ErrorIs(err, assetfttypes.ErrFeatureDisabled)

	// try to freeze
	freezeMsg := &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(1000)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	assertT.ErrorIs(err, assetfttypes.ErrFeatureDisabled)

	// try to globally freeze
	globalFreezeMsg := &assetfttypes.MsgGloballyFreeze{
		Sender: issuer.String(),
		Denom:  denom,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(globalFreezeMsg)),
		globalFreezeMsg,
	)
	assertT.ErrorIs(err, assetfttypes.ErrFeatureDisabled)
}

// TestAuthz tests the authz module works well with assetft module.
func TestAuthzWithAssetFT(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	assetftClient := assetfttypes.NewQueryClient(chain.ClientContext)
	authzClient := authztypes.NewQueryClient(chain.ClientContext)

	granter := chain.GenAccount()
	grantee := chain.GenAccount()
	recipient := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, granter, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&authztypes.MsgGrant{},
			&authztypes.MsgGrant{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})

	// mint and grant authorization
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        granter.String(),
		Symbol:        "symbol",
		Subunit:       "subunit",
		Precision:     1,
		InitialAmount: sdk.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_freezing,
			assetfttypes.Feature_whitelisting,
		},
	}
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, granter)
	grantFreezeMsg, err := authztypes.NewMsgGrant(
		granter,
		grantee,
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&assetfttypes.MsgFreeze{})),
		time.Now().Add(time.Minute),
	)
	require.NoError(t, err)

	grantWhitelistMsg, err := authztypes.NewMsgGrant(
		granter,
		grantee,
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&assetfttypes.MsgSetWhitelistedLimit{})),
		time.Now().Add(time.Minute),
	)
	require.NoError(t, err)

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantFreezeMsg, grantWhitelistMsg, issueMsg)),
		grantFreezeMsg, grantWhitelistMsg, issueMsg,
	)
	requireT.NoError(err)

	// assert granted
	gransRes, err := authzClient.Grants(ctx, &authztypes.QueryGrantsRequest{
		Granter: granter.String(),
		Grantee: grantee.String(),
	})
	requireT.NoError(err)
	requireT.Equal(2, len(gransRes.Grants))

	// try to whitelist and freeze using the authz
	msgFreeze := &assetfttypes.MsgFreeze{
		Sender:  granter.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(240)),
	}

	msgWhitelist := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  granter.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(921)),
	}

	execMsg := authztypes.NewMsgExec(grantee, []sdk.Msg{msgFreeze, msgWhitelist})
	chain.FundAccountsWithOptions(ctx, t, grantee, integrationtests.BalancesOptions{
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

	freezingRes, err := assetftClient.FrozenBalance(ctx, &assetfttypes.QueryFrozenBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues("240", freezingRes.GetBalance().Amount.String())

	whitelistingRes, err := assetftClient.WhitelistedBalance(ctx, &assetfttypes.QueryWhitelistedBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues("921", whitelistingRes.GetBalance().Amount.String())
}

// TestAssetFTBurnRate_OnMinting verifies both burn rate and send commission rate are not applied on received minted tokens.
func TestAssetFT_RatesAreNotApplied_OnMinting(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	issuer := chain.GenAccount()

	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgMint{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})

	// Issue a fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             issuer.String(),
		Symbol:             "ABC",
		Subunit:            "abc",
		Precision:          6,
		InitialAmount:      sdk.NewInt(1000),
		Description:        "ABC Description",
		Features:           []assetfttypes.Feature{assetfttypes.Feature_minting},
		BurnRate:           sdk.MustNewDecFromStr("0.10"),
		SendCommissionRate: sdk.MustNewDecFromStr("0.10"),
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

	// mint tokens
	requireT.NoError(err)
	mintCoin := sdk.NewCoin(denom, sdk.NewInt(500))
	mintMsg := &assetfttypes.MsgMint{
		Sender: issuer.String(),
		Coin:   mintCoin,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.NoError(err)

	// verify balance of token was not affected by either burn rate or send commission rate
	balance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{Address: issuer.String(), Denom: denom})
	requireT.NoError(err)
	assertT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(1500)).String(), balance.GetBalance().String())
}

// TestAssetFTBurnRate_OnBurning verifies that both burn rate and send commission rate are not applied when a token is burnt.
func TestAssetFTBurnRate_SendCommissionRate_OnBurning(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	recipient := chain.GenAccount()

	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&assetfttypes.MsgIssue{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})

	chain.FundAccountsWithOptions(ctx, t, recipient, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgBurn{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})

	// Issue a fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             issuer.String(),
		Symbol:             "ABC",
		Subunit:            "abc",
		Precision:          6,
		InitialAmount:      sdk.NewInt(1000),
		Description:        "ABC Description",
		Features:           []assetfttypes.Feature{assetfttypes.Feature_burning},
		BurnRate:           sdk.MustNewDecFromStr("0.20"),
		SendCommissionRate: sdk.MustNewDecFromStr("0.10"),
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

	// send some coins to the recipient
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(200))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	// recipient burns tokens. Then check recipient and issuer balance, as well as total supply
	oldSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: denom})
	requireT.NoError(err)
	burnCoin := sdk.NewCoin(denom, sdk.NewInt(100))

	burnMsg := &assetfttypes.MsgBurn{
		Sender: recipient.String(),
		Coin:   burnCoin,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.NoError(err)

	issuerBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{Address: issuer.String(), Denom: denom})
	requireT.NoError(err)
	recipientBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{Address: recipient.String(), Denom: denom})
	requireT.NoError(err)
	// verify issuer balance after burning was not affected by the send commission rate
	assertT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(800)).String(), issuerBalance.GetBalance().String())
	// verify recipient balance after burning was not affected by the burn rate
	assertT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(100)).String(), recipientBalance.GetBalance().String())

	newSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: denom})
	requireT.NoError(err)
	// verify the total supply
	assertT.EqualValues(burnCoin, oldSupply.GetAmount().Sub(newSupply.GetAmount()))
}

// TestAssetFTFreezeAndBurn verifies that it is not possible to burn more tokens - outside of freezing limit.
func TestAssetFTFreezeAndBurn(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	recipient := chain.GenAccount()

	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgFreeze{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})

	chain.FundAccountsWithOptions(ctx, t, recipient, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgBurn{},
			&assetfttypes.MsgBurn{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})

	// Issue a fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             issuer.String(),
		Symbol:             "ABC",
		Subunit:            "abc",
		Precision:          6,
		InitialAmount:      sdk.NewInt(1000),
		Description:        "ABC Description",
		Features:           []assetfttypes.Feature{assetfttypes.Feature_burning, assetfttypes.Feature_freezing},
		BurnRate:           sdk.NewDec(0),
		SendCommissionRate: sdk.NewDec(0),
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

	// send some coins to the recipient
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(550))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	// freeze 300 tokens
	freezeMsg := &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(300)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	requireT.NoError(err)

	// recipient burns tokens within allowed unfrozen limit
	burnCoin := sdk.NewCoin(denom, sdk.NewInt(200))

	burnMsg := &assetfttypes.MsgBurn{
		Sender: recipient.String(),
		Coin:   burnCoin,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.NoError(err)

	recipientBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{Address: recipient.String(), Denom: denom})
	requireT.NoError(err)
	// verify recipient balance after burning
	assertT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(350)).String(), recipientBalance.GetBalance().String())

	// recipient tries to burn more token than allowed (left from unfrozen balance). Tx should fail
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.Error(err)
	assertT.True(sdkerrors.ErrInsufficientFunds.Is(err))
	// verify recipient balance did not change
	recipientBalance, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{Address: recipient.String(), Denom: denom})
	requireT.NoError(err)
	assertT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(350)).String(), recipientBalance.GetBalance().String())
}

// TestAssetFTFreeze_WithRates verifies freezing with both burn and send commission rates applied
// and when one of the rates goes outside of unfrozen balance.
func TestAssetFTFreeze_WithRates(t *testing.T) {
	t.Parallel()

	testData := []struct {
		description              string
		burnRate                 sdk.Dec
		sendCommissionRate       sdk.Dec
		expectedCoinDistribution []int
	}{
		{"WithBurnRateOutOfLimit", sdk.MustNewDecFromStr("0.50"), sdk.MustNewDecFromStr("0.10"), []int{510, 340, 100}},
		{"WithSendCommissionRateOutOfLimit", sdk.MustNewDecFromStr("0.10"), sdk.MustNewDecFromStr("0.50"), []int{550, 340, 100}},
	}

	for _, tc := range testData {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()

			ctx, chain := integrationtests.NewCoreumTestingContext(t)

			requireT := require.New(t)
			assertT := assert.New(t)
			issuer := chain.GenAccount()
			recipient1 := chain.GenAccount()
			recipient2 := chain.GenAccount()

			chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
				Messages: []sdk.Msg{
					&banktypes.MsgSend{},
					&assetfttypes.MsgIssue{},
					&assetfttypes.MsgFreeze{},
				},
				Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
			})

			chain.FundAccountsWithOptions(ctx, t, recipient1, integrationtests.BalancesOptions{
				Messages: []sdk.Msg{
					&banktypes.MsgSend{},
					&banktypes.MsgSend{},
				},
				Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
			})

			// Issue a fungible token
			issueMsg := &assetfttypes.MsgIssue{
				Issuer:             issuer.String(),
				Symbol:             "ABC",
				Subunit:            "abc",
				Precision:          6,
				InitialAmount:      sdk.NewInt(1000),
				Description:        "ABC Description",
				Features:           []assetfttypes.Feature{assetfttypes.Feature_freezing},
				BurnRate:           tc.burnRate,           // set burn rate
				SendCommissionRate: tc.sendCommissionRate, // set commission rate
			}

			_, err := client.BroadcastTx(
				ctx,
				chain.ClientContext.WithFromAddress(issuer),
				chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
				issueMsg,
			)

			requireT.NoError(err)
			denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

			// send some coins to the recipient
			sendMsg := &banktypes.MsgSend{
				FromAddress: issuer.String(),
				ToAddress:   recipient1.String(),
				Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(500))),
			}

			_, err = client.BroadcastTx(
				ctx,
				chain.ClientContext.WithFromAddress(issuer),
				chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
				sendMsg,
			)
			requireT.NoError(err)

			// freeze 200 tokens
			freezeMsg := &assetfttypes.MsgFreeze{
				Sender:  issuer.String(),
				Account: recipient1.String(),
				Coin:    sdk.NewCoin(denom, sdk.NewInt(200)),
			}
			_, err = client.BroadcastTx(
				ctx,
				chain.ClientContext.WithFromAddress(issuer),
				chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
				freezeMsg,
			)
			requireT.NoError(err)

			// send from recipient1 to recipient2 (burn and commission rate must apply) - within unfrozen balance limit
			sendMsg = &banktypes.MsgSend{
				FromAddress: recipient1.String(),
				ToAddress:   recipient2.String(),
				Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100))),
			}

			_, err = client.BroadcastTx(
				ctx,
				chain.ClientContext.WithFromAddress(recipient1),
				chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
				sendMsg,
			)
			requireT.NoError(err)

			assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
				&issuer:     int64(tc.expectedCoinDistribution[0]),
				&recipient1: int64(tc.expectedCoinDistribution[1]),
				&recipient2: int64(tc.expectedCoinDistribution[2]),
			})

			// try to send from recipient1 to recipient2. Tx should fail because one of the rates
			// (in the 1st case - burn rate, in the 2nd case - send commission rate) exceeds unfrozen balance limit
			sendMsg = &banktypes.MsgSend{
				FromAddress: recipient1.String(),
				ToAddress:   recipient2.String(),
				Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100))),
			}

			_, err = client.BroadcastTx(
				ctx,
				chain.ClientContext.WithFromAddress(recipient1),
				chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
				sendMsg,
			)
			requireT.Error(err)
			assertT.True(sdkerrors.ErrInsufficientFunds.Is(err))
			// verify balances did not change
			assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
				&issuer:     int64(tc.expectedCoinDistribution[0]),
				&recipient1: int64(tc.expectedCoinDistribution[1]),
				&recipient2: int64(tc.expectedCoinDistribution[2]),
			})
		})
	}
}

func assertCoinDistribution(ctx context.Context, clientCtx client.Context, t *testing.T, denom string, dist map[*sdk.AccAddress]int64) {
	bankClient := banktypes.NewQueryClient(clientCtx)
	requireT := require.New(t)

	total := int64(0)
	for acc, expectedBalance := range dist {
		total += expectedBalance
		getBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{Address: acc.String(), Denom: denom})
		requireT.NoError(err)
		requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(expectedBalance)).String(), getBalance.Balance.String())
	}

	supply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: denom})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(total)).String(), supply.Amount.String())
}

func getIssueFee(ctx context.Context, t *testing.T, clientCtx client.Context) sdk.Coin {
	queryClient := assetfttypes.NewQueryClient(clientCtx)
	resp, err := queryClient.Params(ctx, &assetfttypes.QueryParamsRequest{})
	require.NoError(t, err)

	return resp.Params.IssueFee
}
