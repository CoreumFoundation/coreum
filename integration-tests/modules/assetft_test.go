//go:build integrationtests

package modules

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
	moduleswasm "github.com/CoreumFoundation/coreum/v4/integration-tests/contracts/modules"
	"github.com/CoreumFoundation/coreum/v4/pkg/client"
	"github.com/CoreumFoundation/coreum/v4/testutil/event"
	"github.com/CoreumFoundation/coreum/v4/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

// TestAssetFTQueryParams queries parameters of asset/ft module.
func TestAssetFTQueryParams(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	issueFee := chain.QueryAssetFTParams(ctx, t).IssueFee

	assert.True(t, issueFee.Amount.GT(sdkmath.ZeroInt()))
	assert.Equal(t, chain.ChainSettings.Denom, issueFee.Denom)
}

// TestAssetFTIssue tests issue functionality of fungible tokens.
func TestAssetFTIssue(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	testCases := []struct {
		Name      string
		Precision uint32
	}{
		{
			Name:      "positive precision",
			Precision: 6,
		},
		{
			Name:      "0 precision",
			Precision: 0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			issuer := chain.GenAccount()
			chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
				Messages: []sdk.Msg{
					&assetfttypes.MsgIssue{},
				},
				Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
			})

			issueMsg := &assetfttypes.MsgIssue{
				Issuer:        issuer.String(),
				Symbol:        "ABC",
				Subunit:       "uabc",
				Precision:     tc.Precision,
				Description:   "ABC Description",
				InitialAmount: sdkmath.NewInt(1000),
				Features:      []assetfttypes.Feature{},
				URI:           "https://my-class-meta.invalid/1",
				URIHash:       "content-hash",
			}

			denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)
			expectedMetadata := banktypes.Metadata{
				Description: issueMsg.Description,
				DenomUnits: []*banktypes.DenomUnit{
					{Denom: denom, Exponent: 0},
					{Denom: issueMsg.Symbol, Exponent: issueMsg.Precision},
				},
				Base:    denom,
				Display: issueMsg.Symbol,
				Name:    issueMsg.Symbol,
				Symbol:  issueMsg.Symbol,
				URI:     issueMsg.URI,
				URIHash: issueMsg.URIHash,
			}

			if tc.Precision == 0 {
				expectedMetadata.DenomUnits = []*banktypes.DenomUnit{
					{Denom: denom, Exponent: 0},
				}
				expectedMetadata.Display = denom
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
			requireT.Equal(chain.QueryAssetFTParams(ctx, t).IssueFee.String(), burntStr)

			// check that balance is 0 meaning issue fee was taken

			bankClient := banktypes.NewQueryClient(chain.ClientContext)
			resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
				Address: issuer.String(),
				Denom:   chain.ChainSettings.Denom,
			})
			requireT.NoError(err)
			requireT.Equal(chain.NewCoin(sdkmath.ZeroInt()).String(), resp.Balance.String())

			// check metadata
			metadata, err := bankClient.DenomMetadata(ctx, &banktypes.QueryDenomMetadataRequest{Denom: denom})
			requireT.NoError(err)

			requireT.EqualValues(expectedMetadata, metadata.Metadata)
		})
	}
}

// TestAssetFTIssueInvalidFeatures tests issue functionality of fungible tokens with invalid features.
func TestAssetFTIssueInvalidFeatures(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgIssue{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_ibc,
			100,
			assetfttypes.Feature_minting,
			assetfttypes.Feature_burning,
		},
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.ErrorContains(err, "non-existing feature provided")

	issueMsg = &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_minting,
			assetfttypes.Feature_ibc,
			assetfttypes.Feature_whitelisting,
			assetfttypes.Feature_ibc,
		},
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.ErrorContains(err, "duplicated features in the features list")
}

// TestAssetFTIssueFeeProposal tests proposal upgrading issue fee.
func TestAssetFTIssueFeeProposal(t *testing.T) {
	// This test can't be run together with other tests because it affects balances due to unexpected issue fee.
	// That's why t.Parallel() is not here.

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	origParams := chain.QueryAssetFTParams(ctx, t)
	newParams := origParams
	newParams.IssueFee.Amount = sdk.ZeroInt()
	chain.Governance.ProposalFromMsgAndVote(
		ctx, t, nil,
		"-", "-", "-", govtypesv1.OptionYes,
		&assetfttypes.MsgUpdateParams{
			Params:    newParams,
			Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		},
	)

	issuer := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
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
		InitialAmount: sdkmath.NewInt(1000),
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
	chain.Governance.ProposalFromMsgAndVote(
		ctx, t, nil,
		"-", "-", "-", govtypesv1.OptionYes,
		&assetfttypes.MsgUpdateParams{
			Params:    origParams,
			Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		},
	)
}

// TestAssetIssueAndQueryTokens checks that tokens query works as expected.
func TestAssetIssueAndQueryTokens(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	clientCtx := chain.ClientContext

	ftClient := assetfttypes.NewQueryClient(clientCtx)

	issueFee := chain.QueryAssetFTParams(ctx, t).IssueFee.Amount

	issuer1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer1, integration.BalancesOptions{
		Messages: []sdk.Msg{&assetfttypes.MsgIssue{}},
		Amount:   issueFee,
	})

	issuer2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer2, integration.BalancesOptions{
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
		InitialAmount:      sdkmath.NewInt(777),
		BurnRate:           sdk.NewDec(0),
		SendCommissionRate: sdk.NewDec(0),
		URI:                "https://my-class-meta.invalid/1",
		URIHash:            "content-hash",
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_block_smart_contracts,
		},
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
	requireT.Len(gotToken.Tokens, 1)
	requireT.Equal(assetfttypes.Token{
		Denom:              denom,
		Issuer:             issuer1.String(),
		Symbol:             msg1.Symbol,
		Subunit:            msg1.Subunit,
		Precision:          msg1.Precision,
		Description:        msg1.Description,
		BurnRate:           msg1.BurnRate,
		SendCommissionRate: msg1.SendCommissionRate,
		Version:            gotToken.Tokens[0].Version, // test should work with all versions
		URI:                msg1.URI,
		URIHash:            msg1.URIHash,
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_block_smart_contracts,
		},
	}, gotToken.Tokens[0])
}

// TestBalanceQuery tests balance query.
func TestBalanceQuery(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	assertT := assert.New(t)

	issueFee := chain.QueryAssetFTParams(ctx, t).IssueFee.Amount

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgSetWhitelistedLimit{},
			&assetfttypes.MsgFreeze{},
			&assetfttypes.MsgGloballyFreeze{},
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
		InitialAmount:      sdkmath.NewInt(200),
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

	// freeze globally now

	msgGloballyFreeze := &assetfttypes.MsgGloballyFreeze{
		Sender: issuer.String(),
		Denom:  denom,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgGloballyFreeze)),
		msgGloballyFreeze,
	)
	require.NoError(t, err)

	resp, err = ftClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	require.NoError(t, err)

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	recipientBalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	require.NoError(t, err)

	assertT.Equal(whitelistedCoin.Amount.String(), resp.Whitelisted.String())
	assertT.Equal(recipientBalanceRes.Balance.Amount.String(), resp.Frozen.String())
	assertT.Equal(sendCoin.Amount.String(), resp.Balance.String())
	assertT.Equal("0", resp.Locked.String())
}

func TestSpendableBalanceQuery(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	issueFee := chain.QueryAssetFTParams(ctx, t).IssueFee.Amount

	issuer := chain.GenAccount()
	recipient1 := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgFreeze{},
			&assetfttypes.MsgFreeze{},
			&assetfttypes.MsgGloballyFreeze{},
		},
		Amount: issueFee.MulRaw(2),
	})

	// issue the new fungible token form issuer
	msgIssue := &assetfttypes.MsgIssue{
		Issuer:             issuer.String(),
		Symbol:             "WBTC",
		Subunit:            "wsatoshi",
		Precision:          8,
		InitialAmount:      sdkmath.NewInt(200),
		BurnRate:           sdk.NewDec(0),
		SendCommissionRate: sdk.NewDec(0),
		Features:           []assetfttypes.Feature{assetfttypes.Feature_freezing},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgIssue)),
		msgIssue,
	)
	requireT.NoError(err)

	denom1 := assetfttypes.BuildDenom(msgIssue.Subunit, issuer)
	frozenCoin1 := sdk.NewInt64Coin(denom1, 20)
	sendCoin1 := sdk.NewInt64Coin(denom1, 100)

	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient1.String(),
		Amount:      sdk.NewCoins(sendCoin1),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.NoError(err)

	recipientSpendableBalanceBeforeFreezeRes, err := bankClient.SpendableBalanceByDenom(
		ctx,
		&banktypes.QuerySpendableBalanceByDenomRequest{
			Address: recipient1.String(),
			Denom:   denom1,
		})
	requireT.NoError(err)
	requireT.Equal(sendCoin1.Amount.String(), recipientSpendableBalanceBeforeFreezeRes.Balance.Amount.String())

	msgFreeze := &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: recipient1.String(),
		Coin:    frozenCoin1,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgFreeze)),
		msgFreeze,
	)
	requireT.NoError(err)

	recipientSpendableBalanceAfterFreezeRes, err := bankClient.SpendableBalanceByDenom(
		ctx,
		&banktypes.QuerySpendableBalanceByDenomRequest{
			Address: recipient1.String(),
			Denom:   denom1,
		})
	requireT.NoError(err)
	requireT.Equal(
		sendCoin1.Amount.Sub(frozenCoin1.Amount).String(),
		recipientSpendableBalanceAfterFreezeRes.Balance.Amount.String(),
	)

	// freeze globally now
	msgGloballyFreeze := &assetfttypes.MsgGloballyFreeze{
		Sender: issuer.String(),
		Denom:  denom1,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgGloballyFreeze)),
		msgGloballyFreeze,
	)
	requireT.NoError(err)

	recipientSpendableBalanceAfterGlobalFreezeRes, err := bankClient.SpendableBalanceByDenom(
		ctx,
		&banktypes.QuerySpendableBalanceByDenomRequest{
			Address: recipient1.String(),
			Denom:   denom1,
		})
	requireT.NoError(err)
	requireT.Equal(sdkmath.ZeroInt().String(), recipientSpendableBalanceAfterGlobalFreezeRes.Balance.Amount.String())

	// issue one more token
	msgIssue = &assetfttypes.MsgIssue{
		Issuer:             issuer.String(),
		Symbol:             "WBTC2",
		Subunit:            "wsatoshi2",
		Precision:          8,
		InitialAmount:      sdkmath.NewInt(200),
		BurnRate:           sdk.NewDec(0),
		SendCommissionRate: sdk.NewDec(0),
		Features:           []assetfttypes.Feature{assetfttypes.Feature_freezing},
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgIssue)),
		msgIssue,
	)
	requireT.NoError(err)

	denom2 := assetfttypes.BuildDenom(msgIssue.Subunit, issuer)
	frozenCoin2 := sdk.NewInt64Coin(denom2, 20)
	sendCoin2 := sdk.NewInt64Coin(denom2, 100)

	msgSend = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient1.String(),
		Amount:      sdk.NewCoins(sendCoin2),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.NoError(err)

	recipientSpendableBalancesBeforeFreezeRes, err := bankClient.SpendableBalances(
		ctx,
		&banktypes.QuerySpendableBalancesRequest{
			Address: recipient1.String(),
		})
	requireT.NoError(err)
	requireT.Len(recipientSpendableBalancesBeforeFreezeRes.Balances, 2)
	requireT.Equal(sendCoin2.Amount.String(), recipientSpendableBalancesBeforeFreezeRes.Balances.AmountOf(denom2).String())

	msgFreeze = &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: recipient1.String(),
		Coin:    frozenCoin2,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgFreeze)),
		msgFreeze,
	)
	requireT.NoError(err)

	recipientSpendableBalancesBeforeFreezeRes, err = bankClient.SpendableBalances(
		ctx,
		&banktypes.QuerySpendableBalancesRequest{
			Address: recipient1.String(),
		})
	requireT.NoError(err)
	requireT.Equal(
		sendCoin2.Amount.Sub(frozenCoin2.Amount).String(),
		recipientSpendableBalancesBeforeFreezeRes.Balances.AmountOf(denom2).String(),
	)

	// check the native denom
	recipient2 := chain.GenAccount()
	amountToFund := sdkmath.NewInt(100)
	chain.FundAccountWithOptions(ctx, t, recipient2, integration.BalancesOptions{
		Amount: amountToFund,
	})
	recipient2SpendableBalance, err := bankClient.SpendableBalanceByDenom(
		ctx,
		&banktypes.QuerySpendableBalanceByDenomRequest{
			Address: recipient2.String(),
			Denom:   chain.Chain.ChainSettings.Denom,
		})
	requireT.NoError(err)
	requireT.Equal(amountToFund.String(), recipient2SpendableBalance.Balance.Amount.String())
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
	admin := chain.GenAccount()
	randomAddress := chain.GenAccount()
	recipient := chain.GenAccount()
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgMint{},
			&assetfttypes.MsgMint{},
			&assetfttypes.MsgMint{},
			&assetfttypes.MsgMint{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.MulRaw(2),
	})

	chain.FundAccountWithOptions(ctx, t, randomAddress, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgMint{},
		},
	})

	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

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
			Amount: sdkmath.NewInt(1000),
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
		InitialAmount: sdkmath.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_block_smart_contracts,
			assetfttypes.Feature_minting,
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

	// try to pass non-issuer signature to msg
	mintMsg = &assetfttypes.MsgMint{
		Sender: randomAddress.String(),
		Coin:   sdk.NewCoin(mintableDenom, sdkmath.NewInt(1000)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(randomAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.Error(err)
	assertT.True(cosmoserrors.ErrUnauthorized.Is(err))

	// mint tokens and check balance and total supply
	oldSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: mintableDenom})
	requireT.NoError(err)
	mintCoin := sdk.NewCoin(mintableDenom, sdkmath.NewInt(1600))
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
	assertT.EqualValues(
		mintCoin.Add(sdk.NewCoin(mintableDenom, sdkmath.NewInt(1000))).String(),
		balance.GetBalance().String(),
	)

	newSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: mintableDenom})
	requireT.NoError(err)
	assertT.EqualValues(mintCoin, newSupply.GetAmount().Sub(oldSupply.GetAmount()))

	// mint tokens to recipient
	mintCoin = sdk.NewCoin(mintableDenom, sdkmath.NewInt(10))
	mintMsg = &assetfttypes.MsgMint{
		Sender:    issuer.String(),
		Recipient: recipient.String(),
		Coin:      mintCoin,
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
	assertT.EqualValues(mintCoin, newSupply2.GetAmount().Sub(newSupply.GetAmount()))

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

	mintMsg = &assetfttypes.MsgMint{
		Sender:    issuer.String(),
		Recipient: contractAddr,
		Coin:      mintCoin,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)
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

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgBurn{},
			&assetfttypes.MsgBurn{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.MulRaw(2),
	})

	chain.FundAccountWithOptions(ctx, t, recipient, integration.BalancesOptions{
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
		InitialAmount: sdkmath.NewInt(1000),
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
			Amount: sdkmath.NewInt(900),
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
		Amount:      sdk.NewCoins(sdk.NewCoin(unburnable, sdkmath.NewInt(100))),
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
			Amount: sdkmath.NewInt(1000),
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
		InitialAmount: sdkmath.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_block_smart_contracts,
			assetfttypes.Feature_burning,
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
		Amount:      sdk.NewCoins(sdk.NewCoin(burnableDenom, sdkmath.NewInt(100))),
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
		Coin:   sdk.NewCoin(burnableDenom, sdkmath.NewInt(100)),
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
	burnCoin := sdk.NewCoin(burnableDenom, sdkmath.NewInt(600))

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
	assertT.EqualValues(sdk.NewCoin(burnableDenom, sdkmath.NewInt(300)).String(), balance.GetBalance().String())

	newSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: burnableDenom})
	requireT.NoError(err)
	assertT.EqualValues(burnCoin, oldSupply.GetAmount().Sub(newSupply.GetAmount()))
}

// TestAssetFTBurnRate tests burn rate functionality of fungible tokens.
func TestAssetFTBurnRate(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()
	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})
	chain.FundAccountWithOptions(ctx, t, recipient1, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})
	chain.FundAccountWithOptions(ctx, t, recipient2, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})

	// Issue a fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     6,
		InitialAmount: sdkmath.NewInt(1000),
		Description:   "ABC Description",
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_block_smart_contracts,
		},
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
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(400))),
	}

	txRes, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	// assert that we don't receive events with empty amounts
	requireT.NotContains(txRes.RawLog, `{"key":"amount"}`)

	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&issuer:     600,
		&recipient1: 400,
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
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100))),
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
			{Address: recipient1.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(200)))},
		},
		Outputs: []banktypes.Output{
			{Address: issuer.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100)))},
			{Address: recipient2.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100)))},
		},
	}

	chain.FundAccountWithOptions(ctx, t, recipient1, integration.BalancesOptions{
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
func TestAssetFTSendCommissionRate(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()
	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})
	chain.FundAccountWithOptions(ctx, t, recipient1, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})
	chain.FundAccountWithOptions(ctx, t, recipient2, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})

	// Issue a fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     6,
		InitialAmount: sdkmath.NewInt(1000),
		Description:   "ABC Description",
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_block_smart_contracts,
		},
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
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(400))),
	}

	txRes, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	// assert that we don't receive events with empty amounts
	requireT.NotContains(txRes.RawLog, `{"key":"amount"}`)

	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&issuer:     600,
		&recipient1: 400,
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
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100))),
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
			{Address: recipient1.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(200)))},
		},
		Outputs: []banktypes.Output{
			{Address: issuer.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100)))},
			{Address: recipient2.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100)))},
		},
	}

	chain.FundAccountWithOptions(ctx, t, recipient1, integration.BalancesOptions{
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

// TestAssetFTFeesAreChargedWhenSmartContractExecutesAuthZTransfer verifies that fees are correctly
// charged when smart contract executes authz transfer on behalf of regular account.
func TestAssetFTFeesAreChargedWhenSmartContractExecutesAuthZTransfer(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()
	granter := chain.GenAccount()
	receiver := chain.GenAccount()

	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(issuer, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)
	chain.FundAccountWithOptions(ctx, t, granter, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&authztypes.MsgGrant{},
		},
	})

	// Issue a fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             issuer.String(),
		Symbol:             "ABC",
		Subunit:            "abc",
		Precision:          6,
		InitialAmount:      sdkmath.NewInt(1000),
		Description:        "ABC Description",
		BurnRate:           sdk.MustNewDecFromStr("0.10"),
		SendCommissionRate: sdk.MustNewDecFromStr("0.20"),
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

	// send half of the amount to the second account
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   granter.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(500))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	// deployWASMContract and init contract with the granter.
	initialPayloadAuthzTransfer, err := json.Marshal(authz{
		Granter: granter.String(),
	})
	requireT.NoError(err)

	initialPayloadAuthzStargate, err := json.Marshal(struct{}{})
	requireT.NoError(err)

	contractAddrAuthzTransfer, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		issuer,
		moduleswasm.AuthzTransferWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayloadAuthzTransfer,
			Label:      "authzTransfer",
		},
	)
	requireT.NoError(err)

	contractAddrAuthzStargate, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		issuer,
		moduleswasm.AuthzStargateWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayloadAuthzStargate,
			Label:      "authzStargate",
		},
	)
	requireT.NoError(err)

	// grant the bank send authorization
	grantMsg, err := authztypes.NewMsgGrant(
		granter,
		sdk.MustAccAddressFromBech32(contractAddrAuthzTransfer),
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantMsg)),
		grantMsg,
	)
	requireT.NoError(err)

	grantMsg, err = authztypes.NewMsgGrant(
		issuer,
		sdk.MustAccAddressFromBech32(contractAddrAuthzStargate),
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantMsg)),
		grantMsg,
	)
	requireT.NoError(err)

	// ********** Transfer **********

	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		issuer,
		contractAddrAuthzTransfer,
		moduleswasm.AuthZExecuteTransferRequest(receiver.String(), sdk.NewInt64Coin(denom, 100)),
		sdk.Coin{},
	)
	requireT.NoError(err)

	// check balances

	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&issuer:   520,
		&granter:  370,
		&receiver: 100,
	})

	// ********** Stargate **********

	msgSendAny, err := codectypes.NewAnyWithValue(&banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   receiver.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(denom, 100)),
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		issuer,
		contractAddrAuthzStargate,
		moduleswasm.AuthZExecuteStargateRequest(&authztypes.MsgExec{
			Grantee: contractAddrAuthzStargate,
			Msgs: []*codectypes.Any{
				msgSendAny,
			},
		}),
		sdk.Coin{},
	)
	requireT.NoError(err)

	// check balances

	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&issuer:   420,
		&granter:  370,
		&receiver: 200,
	})
}

// TestAssetFTFeesAreNotChargedWhenTokensAreTransferredFromSmartContractUsingAuthZGrant verifies that fees are not
// charged when funds are transferred from smart contract using authz grant given by the contract.
func TestAssetFTFeesAreNotChargedWhenTokensAreTransferredFromSmartContractUsingAuthZGrant(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()
	grantee := chain.GenAccount()
	receiver := chain.GenAccount()

	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(issuer, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

	// Issue a fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             issuer.String(),
		Symbol:             "ABC",
		Subunit:            "abc",
		Precision:          6,
		InitialAmount:      sdkmath.NewInt(1000),
		Description:        "ABC Description",
		BurnRate:           sdk.MustNewDecFromStr("0.10"),
		SendCommissionRate: sdk.MustNewDecFromStr("0.20"),
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

	// deploy smart contract
	initialPayload, err := json.Marshal(struct{}{})
	requireT.NoError(err)

	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		issuer,
		moduleswasm.AuthzStargateWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayload,
			Label:      "authzStargate",
		},
	)
	requireT.NoError(err)

	// send tokens to the smart contract
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   contractAddr,
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, issueMsg.InitialAmount)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	// grant the bank send authorization
	grantMsg, err := authztypes.NewMsgGrant(
		sdk.MustAccAddressFromBech32(contractAddr),
		grantee,
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	grantMsgAny, err := codectypes.NewAnyWithValue(grantMsg)
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		issuer,
		contractAddr,
		moduleswasm.AuthZExecuteStargateRequest(&authztypes.MsgExec{
			Grantee: contractAddr,
			Msgs: []*codectypes.Any{
				grantMsgAny,
			},
		}),
		sdk.Coin{},
	)
	requireT.NoError(err)

	// send from smart contract using authz
	execMsg := authztypes.NewMsgExec(grantee, []sdk.Msg{&banktypes.MsgSend{
		FromAddress: contractAddr,
		ToAddress:   receiver.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(denom, 100)),
	}})

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
		&contract: 900,
		&receiver: 100,
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
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
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

	// Issue the new fungible token
	msg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_block_smart_contracts,
			assetfttypes.Feature_freezing,
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

	// try to pass non-issuer signature to freeze msg
	freezeMsg := &assetfttypes.MsgFreeze{
		Sender:  randomAddress.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(1000)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(randomAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	requireT.Error(err)
	assertT.True(cosmoserrors.ErrUnauthorized.Is(err))

	// freeze 400 tokens
	freezeMsg = &assetfttypes.MsgFreeze{
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
	requireT.Error(err)
	assertT.True(cosmoserrors.ErrInsufficientFunds.Is(err))
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
	assertT.True(cosmoserrors.ErrInsufficientFunds.Is(err))

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
	requireT.NoError(err)
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
	requireT.Error(err)
	assertT.True(cosmoserrors.ErrInsufficientFunds.Is(err))

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
	requireT.True(cosmoserrors.ErrInsufficientFunds.Is(err))

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

// TestAssetFTClawback checks clawback functionality of fungible tokens.
func TestAssetFTClawback(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	clientCtx := chain.ClientContext

	bankClient := banktypes.NewQueryClient(clientCtx)

	issuer := chain.GenAccount()
	from := chain.GenAccount()
	randomAddress := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgClawback{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})
	chain.FundAccountWithOptions(ctx, t, from, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&banktypes.MsgMultiSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgMultiSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgClawback{},
		},
	})
	chain.FundAccountWithOptions(ctx, t, randomAddress, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgFreeze{},
			&assetfttypes.MsgClawback{},
		},
	})

	// Issue the new fungible token
	msg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_clawback,
		},
	}

	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   from.String(),
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

	// query account balance before clawback
	bankRes, err := bankClient.Balance(ctx, banktypes.NewQueryBalanceRequest(from, denom))
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdkmath.NewInt(1000)).String(), bankRes.Balance.String())

	// try to pass non-issuer signature to clawback msg
	clawbackMsg := &assetfttypes.MsgClawback{
		Sender:  randomAddress.String(),
		Account: from.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(1000)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(randomAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(clawbackMsg)),
		clawbackMsg,
	)
	requireT.Error(err)
	assertT.True(cosmoserrors.ErrUnauthorized.Is(err))

	// clawback 400 tokens
	clawbackMsg = &assetfttypes.MsgClawback{
		Sender:  issuer.String(),
		Account: from.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(400)),
	}
	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(clawbackMsg)),
		clawbackMsg,
	)
	requireT.NoError(err)
	assertT.EqualValues(res.GasUsed, chain.GasLimitByMsgs(clawbackMsg))

	fungibleTokenClawbackEvts, err := event.FindTypedEvents[*assetfttypes.EventAmountClawedBack](res.Events)
	requireT.NoError(err)
	assertT.EqualValues(&assetfttypes.EventAmountClawedBack{
		Account: from.String(),
		Denom:   denom,
		Amount:  sdkmath.NewInt(400),
	}, fungibleTokenClawbackEvts[0])

	// query account balance after clawback
	bankRes, err = bankClient.Balance(ctx, banktypes.NewQueryBalanceRequest(from, denom))
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdkmath.NewInt(600)).String(), bankRes.Balance.String())

	// try to clawback more than available (650) (600 is available)
	coinsToClawback := sdk.NewCoin(denom, sdkmath.NewInt(650))

	clawbackMsg = &assetfttypes.MsgClawback{
		Sender:  issuer.String(),
		Account: from.String(),
		Coin:    coinsToClawback,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(from),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(clawbackMsg)),
		clawbackMsg,
	)
	requireT.Error(err)
	assertT.True(cosmoserrors.ErrInsufficientFunds.Is(err))
}

// TestAssetFTFreezeUnfreezable checks freeze functionality on unfreezable fungible tokens.
func TestAssetFTFreezeUnfreezable(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgFreeze{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	// Issue an unfreezable fungible token
	msg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABCNotFreezable",
		Subunit:       "uabcnotfreezable",
		Description:   "ABC Description",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(1000),
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
		Coin:    sdk.NewCoin(unfreezableDenom, sdkmath.NewInt(1000)),
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
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgFreeze{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	// Issue an freezable fungible token
	msg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     1,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(1000),
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
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(1000)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)
}

// TestAssetFTGloballyFreeze checks global freeze functionality of fungible tokens.
func TestAssetFTGloballyFreeze(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgGloballyFreeze{},
			&banktypes.MsgSend{},
			&banktypes.MsgMultiSend{},
			&assetfttypes.MsgGloballyUnfreeze{},
			&banktypes.MsgSend{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})
	chain.FundAccountWithOptions(ctx, t, recipient, integration.BalancesOptions{
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
		InitialAmount: sdkmath.NewInt(1000),
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
	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(50)))
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
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgFreeze{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})
	chain.FundAccountWithOptions(ctx, t, recipient, integration.BalancesOptions{
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
		InitialAmount:      sdkmath.NewInt(1000),
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
			sdk.NewCoin(denom, sdkmath.NewInt(1000)),
		),
	}

	msgFreeze := &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(650)),
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
	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(300)))
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
	assertT.True(cosmoserrors.ErrInsufficientFunds.Is(err))
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
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgFreeze{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})
	chain.FundAccountWithOptions(ctx, t, recipient, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
		Amount: sdkmath.NewInt(1000),
	})

	// Issue the new fungible token
	msgIssue := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_freezing,
		},
	}
	denom := assetfttypes.BuildDenom(msgIssue.Subunit, issuer)
	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denom, sdkmath.NewInt(1000)),
		),
	}

	msgFreeze := &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(800)),
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
		sdk.NewCoin(denom, sdkmath.NewInt(210)),
		chain.NewCoin(sdkmath.NewInt(1000)),
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
	assertT.True(cosmoserrors.ErrInsufficientFunds.Is(err))
}

// TestNotEnoughBalanceForBurnRate checks tx will fail if there is not enough balance to cover burn rate.
func TestNotEnoughBalanceForBurnRate(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})
	chain.FundAccountWithOptions(ctx, t, recipient, integration.BalancesOptions{
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
		InitialAmount: sdkmath.NewInt(1000),
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
			sdk.NewCoin(denom, sdkmath.NewInt(1000)),
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
	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(920)))

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
	assertT.True(cosmoserrors.ErrInsufficientFunds.Is(err))
}

// TestNotEnoughBalanceForCommissionRate checks tx will fail if there is not enough balance to cover commission rate.
func TestNotEnoughBalanceForCommissionRate(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})
	chain.FundAccountWithOptions(ctx, t, recipient, integration.BalancesOptions{
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
		InitialAmount:      sdkmath.NewInt(1000),
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
			sdk.NewCoin(denom, sdkmath.NewInt(1000)),
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
	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(920)))

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
	assertT.True(cosmoserrors.ErrInsufficientFunds.Is(err))
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
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
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

	// Issue the new fungible token
	amount := sdkmath.NewInt(20000)
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
			assetfttypes.Feature_block_smart_contracts,
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
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(400)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(nonIssuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	assertT.True(cosmoserrors.ErrUnauthorized.Is(err))

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
	assertT.True(assetfttypes.ErrWhitelistedLimitExceeded.Is(err))

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
	assertT.True(assetfttypes.ErrWhitelistedLimitExceeded.Is(err))

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
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgSetWhitelistedLimit{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	// Issue an unwhitelistable fungible token
	subunit := "uabcnotwhitelistable"
	unwhitelistableDenom := assetfttypes.BuildDenom(subunit, issuer)
	amount := sdkmath.NewInt(1000)
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
		Coin:    sdk.NewCoin(unwhitelistableDenom, sdkmath.NewInt(1000)),
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
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgSetWhitelistedLimit{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	// Issue an whitelistable fungible token
	subunit := "uabcwhitelistable"
	denom := assetfttypes.BuildDenom(subunit, issuer)
	amount := sdkmath.NewInt(1000)
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
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(1000)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)

	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)
}

// TestAssetFTSendingToNonWhitelistedSmartContractIsDenied verifies that this is not possible to send token to
// smart contract if it is not whitelisted.
func TestAssetFTSendingToNonWhitelistedSmartContractIsDenied(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issuer := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(issuer, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

	clientCtx := chain.ClientContext

	// Issue a fungible token which cannot be sent to the smart contract
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     6,
		InitialAmount: sdkmath.NewInt(1000),
		Description:   "ABC Description",
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_whitelisting,
		},
		BurnRate:           sdk.ZeroDec(),
		SendCommissionRate: sdk.ZeroDec(),
	}

	_, err := client.BroadcastTx(
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
	requireT.ErrorIs(err, assetfttypes.ErrWhitelistedLimitExceeded)
}

// TestAssetFTAttachingToNonWhitelistedSmartContractCallIsDenied verifies that this is not possible to attach
// token to smart contract call if contract is not whitelisted.
func TestAssetFTAttachingToNonWhitelistedSmartContractCallIsDenied(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issuer := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(issuer, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

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
			assetfttypes.Feature_whitelisting,
		},
		BurnRate:           sdk.ZeroDec(),
		SendCommissionRate: sdk.ZeroDec(),
	}

	_, err := client.BroadcastTx(
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
	requireT.ErrorContains(err, "whitelisted limit exceeded")
}

// TestAssetFTAttachingToNonWhitelistedSmartContractInstantiationIsDenied verifies that this is not possible
// to attach token to smart contract instantiation if contract is not whitelisted.
func TestAssetFTAttachingToNonWhitelistedSmartContractInstantiationIsDenied(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issuer := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(issuer, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

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
			assetfttypes.Feature_whitelisting,
		},
		BurnRate:           sdk.ZeroDec(),
		SendCommissionRate: sdk.ZeroDec(),
	}

	_, err := client.BroadcastTx(
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

	// This operation should fail due to coins being attached to it
	_, _, err = chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		issuer,
		moduleswasm.SimpleStateWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayload,
			Amount:     sdk.NewInt64Coin(denom, 100),
			Label:      "simple_state",
		},
	)
	requireT.ErrorContains(err, "whitelisted limit exceeded")
}

// TestBareToken checks none of the features will work if the flags are not set.
func TestBareToken(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgMint{},
			&assetfttypes.MsgBurn{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgFreeze{},
			&assetfttypes.MsgGloballyFreeze{},
			&assetfttypes.MsgSetWhitelistedLimit{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})
	chain.FundAccountWithOptions(ctx, t, recipient, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgBurn{},
		},
	})

	// Issue a bare token
	amount := sdkmath.NewInt(1000)
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
		Coin:   sdk.NewCoin(denom, sdkmath.NewInt(1000)),
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
		Coin:   sdk.NewCoin(denom, sdkmath.NewInt(10)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.NoError(err)

	// try to burn from non-issuer account (must fail)
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(10))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	burnMsg = &assetfttypes.MsgBurn{
		Sender: recipient.String(),
		Coin:   sdk.NewCoin(denom, sdkmath.NewInt(10)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.ErrorIs(err, assetfttypes.ErrFeatureDisabled)

	// try to whitelist
	whitelistMsg := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(1000)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	requireT.ErrorIs(err, assetfttypes.ErrFeatureDisabled)

	// try to freeze
	freezeMsg := &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(1000)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	requireT.ErrorIs(err, assetfttypes.ErrFeatureDisabled)

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
	requireT.ErrorIs(err, assetfttypes.ErrFeatureDisabled)
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

	chain.FundAccountWithOptions(ctx, t, granter, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&authztypes.MsgGrant{},
			&authztypes.MsgGrant{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	// mint and grant authorization
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        granter.String(),
		Symbol:        "symbol",
		Subunit:       "subunit",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(1000),
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
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	grantWhitelistMsg, err := authztypes.NewMsgGrant(
		granter,
		grantee,
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&assetfttypes.MsgSetWhitelistedLimit{})),
		lo.ToPtr(time.Now().Add(time.Minute)),
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
	requireT.Len(gransRes.Grants, 2)

	// try to whitelist and freeze using the authz
	msgFreeze := &assetfttypes.MsgFreeze{
		Sender:  granter.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(240)),
	}

	msgWhitelist := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  granter.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(921)),
	}

	execMsg := authztypes.NewMsgExec(grantee, []sdk.Msg{msgFreeze, msgWhitelist})
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

// TestAuthzMintAuthorizationLimit tests the authz MintLimitAuthorization msg works as expected.
func TestAuthzMintAuthorizationLimit(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	authzClient := authztypes.NewQueryClient(chain.ClientContext)

	granter := chain.GenAccount()
	grantee := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, granter, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&authztypes.MsgGrant{},
			&authztypes.MsgGrant{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	// mint and grant authorization
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        granter.String(),
		Symbol:        "symbol",
		Subunit:       "subunit",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(0),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_minting,
		},
	}
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, granter)
	grantMintMsg, err := authztypes.NewMsgGrant(
		granter,
		grantee,
		assetfttypes.NewMintAuthorization(sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1000)))),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantMintMsg, issueMsg)),
		grantMintMsg, issueMsg,
	)
	requireT.NoError(err)

	// assert granted
	gransRes, err := authzClient.Grants(ctx, &authztypes.QueryGrantsRequest{
		Granter: granter.String(),
		Grantee: grantee.String(),
	})
	requireT.NoError(err)
	requireT.Len(gransRes.Grants, 1)

	// try to mint using the authz
	msgMint := &assetfttypes.MsgMint{
		Sender: granter.String(),
		Coin:   sdk.NewCoin(denom, sdkmath.NewInt(501)),
	}

	execMsg := authztypes.NewMsgExec(grantee, []sdk.Msg{msgMint})
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

	supply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: denom})
	requireT.NoError(err)
	requireT.EqualValues("501", supply.Amount.Amount.String())

	gransRes, err = authzClient.Grants(ctx, &authztypes.QueryGrantsRequest{
		Granter: granter.String(),
		Grantee: grantee.String(),
	})
	requireT.NoError(err)
	requireT.Len(gransRes.Grants, 1)
	updatedGrant := assetfttypes.MintAuthorization{}
	chain.ClientContext.Codec().MustUnmarshal(gransRes.Grants[0].Authorization.Value, &updatedGrant)
	requireT.EqualValues("499", updatedGrant.MintLimit.AmountOf(denom).String())

	// try to mint exceeding limit
	msgMint = &assetfttypes.MsgMint{
		Sender: granter.String(),
		Coin:   sdk.NewCoin(denom, sdkmath.NewInt(500)),
	}

	execMsg = authztypes.NewMsgExec(grantee, []sdk.Msg{msgMint})
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
	requireT.Error(err)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// minting the entire limit should remove the grant
	msgMint = &assetfttypes.MsgMint{
		Sender: granter.String(),
		Coin:   sdk.NewCoin(denom, sdkmath.NewInt(499)),
	}

	execMsg = authztypes.NewMsgExec(grantee, []sdk.Msg{msgMint})
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
	gransRes, err = authzClient.Grants(ctx, &authztypes.QueryGrantsRequest{
		Granter: granter.String(),
		Grantee: grantee.String(),
	})
	requireT.NoError(err)
	requireT.Empty(gransRes.Grants)
}

// TestAuthzMintAuthorizationLimit_GrantFromNonIssuer tests the authz MintLimitAuthorization msg works as expected if
// the granter is non-issuer address.
func TestAuthzMintAuthorizationLimit_GrantFromNonIssuer(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	authzClient := authztypes.NewQueryClient(chain.ClientContext)

	issuer := chain.GenAccount()
	granter := chain.GenAccount()
	grantee := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	chain.FundAccountWithOptions(ctx, t, granter, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&authztypes.MsgGrant{},
			&authztypes.MsgGrant{},
		},
	})

	// issue and grant authorization
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "symbol",
		Subunit:       "subunit",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(0),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_minting,
		},
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)
	grantMintMsg, err := authztypes.NewMsgGrant(
		granter,
		grantee,
		assetfttypes.NewMintAuthorization(sdk.NewCoins(
			sdk.NewCoin(denom, sdk.NewInt(1000)),
		)),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantMintMsg)),
		grantMintMsg,
	)
	requireT.NoError(err)

	// assert granted
	gransRes, err := authzClient.Grants(ctx, &authztypes.QueryGrantsRequest{
		Granter: granter.String(),
		Grantee: grantee.String(),
	})
	requireT.NoError(err)
	requireT.Len(gransRes.Grants, 1)

	// try to mint using the authz
	msgMint := &assetfttypes.MsgMint{
		Sender: granter.String(),
		Coin:   sdk.NewCoin(denom, sdkmath.NewInt(501)),
	}

	execMsg := authztypes.NewMsgExec(grantee, []sdk.Msg{msgMint})
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
	requireT.Error(err)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)
}

// TestAuthzMintAuthorizationLimit_MultipleCoins tests the authz MintLimitAuthorization msg works as expected
// if there are multiple coins in the grant.
func TestAuthzMintAuthorizationLimit_MultipleCoins(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	authzClient := authztypes.NewQueryClient(chain.ClientContext)

	issuer := chain.GenAccount()
	grantee := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgIssue{},
			&authztypes.MsgGrant{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.Mul(sdk.NewInt(2)),
	})

	// issue and grant authorization
	issueMsg1 := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "symbolminting",
		Subunit:       "subunitminting",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(0),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_minting,
		},
	}

	issueMsg2 := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "symbol",
		Subunit:       "subunit",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(0),
		Features:      []assetfttypes.Feature{},
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg1, issueMsg2)),
		issueMsg1, issueMsg2,
	)
	requireT.NoError(err)

	denom1 := assetfttypes.BuildDenom(issueMsg1.Subunit, issuer)
	denom2 := assetfttypes.BuildDenom(issueMsg2.Subunit, issuer)
	grantMintMsg, err := authztypes.NewMsgGrant(
		issuer,
		grantee,
		assetfttypes.NewMintAuthorization(sdk.NewCoins(
			sdk.NewCoin(denom1, sdk.NewInt(1000)),
			sdk.NewCoin(denom2, sdk.NewInt(1000)),
		)),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantMintMsg)),
		grantMintMsg,
	)
	requireT.NoError(err)

	// assert granted
	gransRes, err := authzClient.Grants(ctx, &authztypes.QueryGrantsRequest{
		Granter: issuer.String(),
		Grantee: grantee.String(),
	})
	requireT.NoError(err)
	requireT.Len(gransRes.Grants, 1)

	// try to mint using the authz
	msgMint := &assetfttypes.MsgMint{
		Sender: issuer.String(),
		Coin:   sdk.NewCoin(denom1, sdkmath.NewInt(501)),
	}

	execMsg := authztypes.NewMsgExec(grantee, []sdk.Msg{msgMint})
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

	gransRes, err = authzClient.Grants(ctx, &authztypes.QueryGrantsRequest{
		Granter: issuer.String(),
		Grantee: grantee.String(),
	})
	requireT.NoError(err)
	requireT.Len(gransRes.Grants, 1)
	updatedGrant := assetfttypes.BurnAuthorization{}
	chain.ClientContext.Codec().MustUnmarshal(gransRes.Grants[0].Authorization.Value, &updatedGrant)
	requireT.EqualValues("499", updatedGrant.BurnLimit.AmountOf(denom1).String())
	requireT.EqualValues("1000", updatedGrant.BurnLimit.AmountOf(denom2).String())
}

// TestAuthzBurnAuthorizationLimit tests the authz BurnLimitAuthorization msg works as expected.
func TestAuthzBurnAuthorizationLimit(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	authzClient := authztypes.NewQueryClient(chain.ClientContext)

	granter := chain.GenAccount()
	grantee := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, granter, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&authztypes.MsgGrant{},
			&authztypes.MsgGrant{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	// grant authorization
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        granter.String(),
		Symbol:        "symbol",
		Subunit:       "subunit",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(10000),
		Features:      []assetfttypes.Feature{},
	}
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, granter)
	grantBurnMsg, err := authztypes.NewMsgGrant(
		granter,
		grantee,
		assetfttypes.NewBurnAuthorization(sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1000)))),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantBurnMsg, issueMsg)),
		grantBurnMsg, issueMsg,
	)
	requireT.NoError(err)

	// assert granted
	gransRes, err := authzClient.Grants(ctx, &authztypes.QueryGrantsRequest{
		Granter: granter.String(),
		Grantee: grantee.String(),
	})
	requireT.NoError(err)
	requireT.Len(gransRes.Grants, 1)

	// try to burn using the authz
	msgBurn := &assetfttypes.MsgBurn{
		Sender: granter.String(),
		Coin:   sdk.NewCoin(denom, sdkmath.NewInt(501)),
	}

	execMsg := authztypes.NewMsgExec(grantee, []sdk.Msg{msgBurn})
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

	supply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: denom})
	requireT.NoError(err)
	requireT.EqualValues("9499", supply.Amount.Amount.String())

	gransRes, err = authzClient.Grants(ctx, &authztypes.QueryGrantsRequest{
		Granter: granter.String(),
		Grantee: grantee.String(),
	})
	requireT.NoError(err)
	requireT.Len(gransRes.Grants, 1)
	updatedGrant := assetfttypes.BurnAuthorization{}
	chain.ClientContext.Codec().MustUnmarshal(gransRes.Grants[0].Authorization.Value, &updatedGrant)
	requireT.EqualValues("499", updatedGrant.BurnLimit.AmountOf(denom).String())

	// try to burn exceeding limit
	msgBurn = &assetfttypes.MsgBurn{
		Sender: granter.String(),
		Coin:   sdk.NewCoin(denom, sdkmath.NewInt(500)),
	}

	execMsg = authztypes.NewMsgExec(grantee, []sdk.Msg{msgBurn})
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
	requireT.Error(err)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// burning the entire limit should remove the grant
	msgBurn = &assetfttypes.MsgBurn{
		Sender: granter.String(),
		Coin:   sdk.NewCoin(denom, sdkmath.NewInt(499)),
	}

	execMsg = authztypes.NewMsgExec(grantee, []sdk.Msg{msgBurn})
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
	gransRes, err = authzClient.Grants(ctx, &authztypes.QueryGrantsRequest{
		Granter: granter.String(),
		Grantee: grantee.String(),
	})
	requireT.NoError(err)
	requireT.Empty(gransRes.Grants)

	supply, err = bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: denom})
	requireT.NoError(err)
	requireT.EqualValues("9000", supply.Amount.Amount.String())
}

// TestAuthzBurnAuthorizationLimit_GrantFromNonIssuer tests the authz BurnLimitAuthorization msg works as expected if
// the granter is non-issuer address.
func TestAuthzBurnAuthorizationLimit_GrantFromNonIssuer(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	authzClient := authztypes.NewQueryClient(chain.ClientContext)

	issuer := chain.GenAccount()
	granter := chain.GenAccount()
	grantee := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.Mul(sdk.NewInt(2)),
	})

	chain.FundAccountWithOptions(ctx, t, granter, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&authztypes.MsgGrant{},
			&authztypes.MsgGrant{},
		},
	})

	// issue and grant authorization
	issueWithBurningFeatureMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "symbolburning",
		Subunit:       "subunitburning",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(10000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_burning,
		},
	}

	issueWithoutBurningFeatureMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "symbol",
		Subunit:       "subunit",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(10000),
		Features:      []assetfttypes.Feature{},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueWithBurningFeatureMsg, issueWithoutBurningFeatureMsg)),
		issueWithBurningFeatureMsg, issueWithoutBurningFeatureMsg,
	)
	requireT.NoError(err)

	denomBurning := assetfttypes.BuildDenom(issueWithBurningFeatureMsg.Subunit, issuer)
	denomNoBurning := assetfttypes.BuildDenom(issueWithoutBurningFeatureMsg.Subunit, issuer)

	// send coins to granter
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   granter.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denomBurning, sdkmath.NewInt(1000)),
			sdk.NewCoin(denomNoBurning, sdkmath.NewInt(1000)),
		),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	// grant authz to burn
	grantMsg, err := authztypes.NewMsgGrant(
		granter,
		grantee,
		assetfttypes.NewBurnAuthorization(sdk.NewCoins(
			sdk.NewCoin(denomBurning, sdk.NewInt(1000)),
			sdk.NewCoin(denomNoBurning, sdk.NewInt(1000)),
		)),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantMsg)),
		grantMsg,
	)
	requireT.NoError(err)

	// assert granted
	gransRes, err := authzClient.Grants(ctx, &authztypes.QueryGrantsRequest{
		Granter: granter.String(),
		Grantee: grantee.String(),
	})
	requireT.NoError(err)
	requireT.Len(gransRes.Grants, 1)

	// try to burn using the authz when burning is enabled
	msgBurn := &assetfttypes.MsgBurn{
		Sender: granter.String(),
		Coin:   sdk.NewCoin(denomBurning, sdkmath.NewInt(501)),
	}

	execMsg := authztypes.NewMsgExec(grantee, []sdk.Msg{msgBurn})
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

	// try to burn using the authz when burning is not enabled
	msgBurn = &assetfttypes.MsgBurn{
		Sender: granter.String(),
		Coin:   sdk.NewCoin(denomNoBurning, sdkmath.NewInt(501)),
	}

	execMsg = authztypes.NewMsgExec(grantee, []sdk.Msg{msgBurn})
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
	requireT.Error(err)
	requireT.ErrorIs(err, assetfttypes.ErrFeatureDisabled)
}

// TestAssetFTBurnRate_OnMinting verifies both burn rate and send commission rate are not applied on received
// minted tokens.
func TestAssetFT_RatesAreNotApplied_OnMinting(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	issuer := chain.GenAccount()

	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgMint{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	// Issue a fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             issuer.String(),
		Symbol:             "ABC",
		Subunit:            "abc",
		Precision:          6,
		InitialAmount:      sdkmath.NewInt(1000),
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
	mintCoin := sdk.NewCoin(denom, sdkmath.NewInt(500))
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
	assertT.EqualValues(sdk.NewCoin(denom, sdkmath.NewInt(1500)).String(), balance.GetBalance().String())
}

// TestAssetFTBurnRate_OnBurning verifies that both burn rate and send commission rate are not applied when a
// token is burnt.
func TestAssetFTBurnRate_SendCommissionRate_OnBurning(t *testing.T) {
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
			&assetfttypes.MsgIssue{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	chain.FundAccountWithOptions(ctx, t, recipient, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgBurn{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	// Issue a fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             issuer.String(),
		Symbol:             "ABC",
		Subunit:            "abc",
		Precision:          6,
		InitialAmount:      sdkmath.NewInt(1000),
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
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(200))),
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
	burnCoin := sdk.NewCoin(denom, sdkmath.NewInt(100))

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
	recipientBalance, err := bankClient.Balance(
		ctx,
		&banktypes.QueryBalanceRequest{Address: recipient.String(), Denom: denom},
	)
	requireT.NoError(err)
	// verify issuer balance after burning was not affected by the send commission rate
	assertT.EqualValues(sdk.NewCoin(denom, sdkmath.NewInt(800)).String(), issuerBalance.GetBalance().String())
	// verify recipient balance after burning was not affected by the burn rate
	assertT.EqualValues(sdk.NewCoin(denom, sdkmath.NewInt(100)).String(), recipientBalance.GetBalance().String())

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

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgFreeze{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	chain.FundAccountWithOptions(ctx, t, recipient, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgBurn{},
			&assetfttypes.MsgBurn{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	// Issue a fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             issuer.String(),
		Symbol:             "ABC",
		Subunit:            "abc",
		Precision:          6,
		InitialAmount:      sdkmath.NewInt(1000),
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
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(550))),
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
		Coin:    sdk.NewCoin(denom, sdkmath.NewInt(300)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	requireT.NoError(err)

	// recipient burns tokens within allowed unfrozen limit
	burnCoin := sdk.NewCoin(denom, sdkmath.NewInt(200))

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

	recipientBalance, err := bankClient.Balance(
		ctx,
		&banktypes.QueryBalanceRequest{Address: recipient.String(), Denom: denom},
	)
	requireT.NoError(err)
	// verify recipient balance after burning
	assertT.EqualValues(sdk.NewCoin(denom, sdkmath.NewInt(350)).String(), recipientBalance.GetBalance().String())

	// recipient tries to burn more token than allowed (left from unfrozen balance). Tx should fail
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.Error(err)
	assertT.True(cosmoserrors.ErrInsufficientFunds.Is(err))
	// verify recipient balance did not change
	recipientBalance, err = bankClient.Balance(
		ctx,
		&banktypes.QueryBalanceRequest{Address: recipient.String(), Denom: denom},
	)
	requireT.NoError(err)
	assertT.EqualValues(sdk.NewCoin(denom, sdkmath.NewInt(350)).String(), recipientBalance.GetBalance().String())
}

// TestAssetFTFreeze_WithRates verifies freezing with both burn and send commission rates applied
// and when one of the rates goes outside unfrozen balance.
func TestAssetFTFreeze_WithRates(t *testing.T) {
	t.Parallel()

	//nolint:lll // we don't care about test cases.
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

			chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
				Messages: []sdk.Msg{
					&banktypes.MsgSend{},
					&assetfttypes.MsgIssue{},
					&assetfttypes.MsgFreeze{},
				},
				Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
			})

			chain.FundAccountWithOptions(ctx, t, recipient1, integration.BalancesOptions{
				Messages: []sdk.Msg{
					&banktypes.MsgSend{},
					&banktypes.MsgSend{},
				},
				Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
			})

			// Issue a fungible token
			issueMsg := &assetfttypes.MsgIssue{
				Issuer:             issuer.String(),
				Symbol:             "ABC",
				Subunit:            "abc",
				Precision:          6,
				InitialAmount:      sdkmath.NewInt(1000),
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
				Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(500))),
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
				Coin:    sdk.NewCoin(denom, sdkmath.NewInt(200)),
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
				Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100))),
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
				Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(100))),
			}

			_, err = client.BroadcastTx(
				ctx,
				chain.ClientContext.WithFromAddress(recipient1),
				chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
				sendMsg,
			)
			requireT.Error(err)
			assertT.True(cosmoserrors.ErrInsufficientFunds.Is(err))
			// verify balances did not change
			assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
				&issuer:     int64(tc.expectedCoinDistribution[0]),
				&recipient1: int64(tc.expectedCoinDistribution[1]),
				&recipient2: int64(tc.expectedCoinDistribution[2]),
			})
		})
	}
}

// TestAssetFTAminoMultisig tests that assetnf module works seamlessly with amino multisig.
func TestAssetFTAminoMultisig(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	multisigPublicKey, keyNamesSet, err := chain.GenMultisigAccount(2, 2)
	requireT.NoError(err)
	multisigAddress := sdk.AccAddress(multisigPublicKey.Address())
	signer1KeyName := keyNamesSet[0]
	signer2KeyName := keyNamesSet[1]

	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	chain.FundAccountWithOptions(ctx, t, multisigAddress, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgBurn{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	// Issue a fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             multisigAddress.String(),
		Symbol:             "ABC",
		Subunit:            "abc",
		Precision:          6,
		InitialAmount:      sdkmath.NewInt(1000),
		Description:        "ABC Description",
		Features:           []assetfttypes.Feature{assetfttypes.Feature_burning, assetfttypes.Feature_freezing},
		BurnRate:           sdk.NewDec(0),
		SendCommissionRate: sdk.NewDec(0),
	}

	_, err = chain.SignAndBroadcastMultisigTx(
		ctx,
		chain.ClientContext.WithFromAddress(multisigAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
		signer1KeyName, signer2KeyName)
	requireT.NoError(err)

	denom := assetfttypes.BuildDenom(issueMsg.Subunit, multisigAddress)

	burnMsg := &assetfttypes.MsgBurn{
		Sender: multisigAddress.String(),
		Coin:   sdk.NewCoin(denom, sdkmath.NewInt(300)),
	}
	_, err = chain.SignAndBroadcastMultisigTx(
		ctx,
		chain.ClientContext.WithFromAddress(multisigAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
		signer1KeyName, signer2KeyName)
	requireT.NoError(err)

	balanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: multisigAddress.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(700)).String(), balanceRes.Balance.String())
}

// TestAssetFTAminoMultisigWithAuthz tests that assetnf module works seamlessly with amino multisig and authz.
func TestAssetFTAminoMultisigWithAuthz(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	multisigPublicKeyGranter, keyNamesSet, err := chain.GenMultisigAccount(2, 2)
	requireT.NoError(err)
	multisigGranterAddress := sdk.AccAddress(multisigPublicKeyGranter.Address())
	granterSigner1KeyName := keyNamesSet[0]
	granterSigner2KeyName := keyNamesSet[1]

	multisigPublicKeyGrantee, keyNamesSet, err := chain.GenMultisigAccount(2, 2)
	requireT.NoError(err)
	multisigGranteeAddress := sdk.AccAddress(multisigPublicKeyGrantee.Address())
	granteeSigner1KeyName := keyNamesSet[0]
	granteeSigner2KeyName := keyNamesSet[1]

	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	grantMsg, err := authztypes.NewMsgGrant(
		multisigGranterAddress,
		multisigGranteeAddress,
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&assetfttypes.MsgIssue{})),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	chain.FundAccountWithOptions(ctx, t, multisigGranterAddress, integration.BalancesOptions{
		Messages: []sdk.Msg{
			grantMsg,
		},
		// the fee will be charged from the granter
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	_, err = chain.SignAndBroadcastMultisigTx(
		ctx,
		chain.ClientContext.WithFromAddress(multisigGranterAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantMsg)),
		grantMsg,
		granterSigner1KeyName, granterSigner2KeyName)
	requireT.NoError(err)

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             multisigGranterAddress.String(),
		Symbol:             "ABC",
		Subunit:            "abc",
		Precision:          6,
		InitialAmount:      sdkmath.NewInt(1000),
		Description:        "ABC Description",
		Features:           []assetfttypes.Feature{assetfttypes.Feature_burning, assetfttypes.Feature_freezing},
		BurnRate:           sdk.NewDec(0),
		SendCommissionRate: sdk.NewDec(0),
	}

	execMsg := authztypes.NewMsgExec(multisigGranteeAddress, []sdk.Msg{issueMsg})

	chain.FundAccountWithOptions(ctx, t, multisigGranteeAddress, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&execMsg,
		},
	})

	_, err = chain.SignAndBroadcastMultisigTx(
		ctx,
		chain.ClientContext.WithFromAddress(multisigGranteeAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(&execMsg)),
		&execMsg,
		granteeSigner1KeyName, granteeSigner2KeyName)
	requireT.NoError(err)

	denom := assetfttypes.BuildDenom(issueMsg.Subunit, multisigGranterAddress)
	balanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: multisigGranterAddress.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(1000)).String(), balanceRes.Balance.String())
}

// TestAssetFTSendCommissionAndBurnRateWithSmartContract verifies that burn rate and send commission are correctly
// accounted when funds are sent from/to smart contract.
func TestAssetFTSendCommissionAndBurnRateWithSmartContract(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issuer := chain.GenAccount()
	admin := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(issuer, chain.NewCoin(sdkmath.NewInt(5000000000))),
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

	clientCtx := chain.ClientContext
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)

	// Issue a fungible token with burn rate and send commission rate
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             issuer.String(),
		Symbol:             "ABC",
		Subunit:            "abc",
		Precision:          6,
		InitialAmount:      sdkmath.NewInt(1000),
		Description:        "ABC Description",
		Features:           []assetfttypes.Feature{},
		BurnRate:           sdk.MustNewDecFromStr("0.10"),
		SendCommissionRate: sdk.MustNewDecFromStr("0.20"),
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

	// send half of the amount to the second account
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   admin.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(500))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	// deploy smart contract and send ft tokens to it, burn rate and send commission should not apply,
	// because tokens are sent by the issuer.
	contractAddr, contractCodeID, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		issuer,
		moduleswasm.BankSendWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    moduleswasm.EmptyPayload,
			Amount:     sdk.NewInt64Coin(denom, 100),
			Label:      "bank_send",
		},
	)
	requireT.NoError(err)
	contract1 := sdk.MustAccAddressFromBech32(contractAddr)

	// verify amounts
	assertCoinDistribution(ctx, clientCtx, t, denom, map[*sdk.AccAddress]int64{
		&issuer:    400,
		&admin:     500,
		&contract1: 100,
	})

	// send additional coins to contract directly
	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   contractAddr,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(denom, 100)),
	}

	_, err = client.BroadcastTx(ctx, clientCtx.WithFromAddress(issuer), txf, sendMsg)
	requireT.NoError(err)

	// verify amounts
	assertCoinDistribution(ctx, clientCtx, t, denom, map[*sdk.AccAddress]int64{
		&issuer:    300,
		&admin:     500,
		&contract1: 200,
	})

	// send to smart contract from the second address, burn rate and send commission should apply
	sendMsg = &banktypes.MsgSend{
		FromAddress: admin.String(),
		ToAddress:   contractAddr,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(denom, 300)),
	}

	_, err = client.BroadcastTx(ctx, clientCtx.WithFromAddress(admin), txf, sendMsg)
	requireT.NoError(err)

	// verify amounts
	assertCoinDistribution(ctx, clientCtx, t, denom, map[*sdk.AccAddress]int64{
		&issuer:    360,
		&admin:     110,
		&contract1: 500,
	})

	// send from the smart contract to issuer, fees should not apply
	wasmBankSend := &wasmtypes.MsgExecuteContract{
		Sender:   issuer.String(),
		Contract: contractAddr,
		Msg: wasmtypes.RawContractMessage(
			moduleswasm.BankSendExecuteWithdrawRequest(sdk.NewInt64Coin(denom, 100), issuer),
		),
		Funds: sdk.Coins{},
	}
	_, err = client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(issuer),
		txf.WithGasAdjustment(1.5),
		wasmBankSend,
	)
	requireT.NoError(err)

	// verify amounts
	assertCoinDistribution(ctx, clientCtx, t, denom, map[*sdk.AccAddress]int64{
		&issuer:    460,
		&admin:     110,
		&contract1: 400,
	})

	// send from the smart contract to another account, fees should not apply again
	wasmBankSend = &wasmtypes.MsgExecuteContract{
		Sender:   issuer.String(),
		Contract: contractAddr,
		Msg: wasmtypes.RawContractMessage(
			moduleswasm.BankSendExecuteWithdrawRequest(sdk.NewInt64Coin(denom, 100), admin),
		),
		Funds: sdk.Coins{},
	}

	_, err = client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(issuer),
		txf.WithGasAdjustment(1.5),
		wasmBankSend,
	)
	requireT.NoError(err)

	// verify amounts
	assertCoinDistribution(ctx, clientCtx, t, denom, map[*sdk.AccAddress]int64{
		&issuer:    460,
		&admin:     210,
		&contract1: 300,
	})

	// instantiate contract again using non-issuer account, fees should apply.
	salt, err := chain.Wasm.GenerateSalt()
	requireT.NoError(err)
	contractAddr, err = chain.Wasm.InstantiateWASMContract2(
		ctx,
		txf,
		admin,
		salt,
		integration.InstantiateConfig{
			CodeID:     contractCodeID,
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    moduleswasm.EmptyPayload,
			Amount:     sdk.NewInt64Coin(denom, 100),
			Label:      "bank_send",
		},
	)
	requireT.NoError(err)
	contract2 := sdk.MustAccAddressFromBech32(contractAddr)

	// verify amounts
	assertCoinDistribution(ctx, clientCtx, t, denom, map[*sdk.AccAddress]int64{
		&issuer:    480,
		&admin:     80,
		&contract1: 300,
		&contract2: 100,
	})

	// send from one contract to another, fees should not apply
	wasmBankSend = &wasmtypes.MsgExecuteContract{
		Sender:   issuer.String(),
		Contract: contract1.String(),
		Msg: wasmtypes.RawContractMessage(
			moduleswasm.BankSendExecuteWithdrawRequest(sdk.NewInt64Coin(denom, 100), contract2),
		),
		Funds: sdk.Coins{},
	}

	_, err = client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(issuer),
		txf.WithGasAdjustment(1.5),
		wasmBankSend,
	)
	requireT.NoError(err)

	// verify amounts
	assertCoinDistribution(ctx, clientCtx, t, denom, map[*sdk.AccAddress]int64{
		&issuer:    480,
		&admin:     80,
		&contract1: 200,
		&contract2: 200,
	})
}

// TestAssetFTSendCommissionAndBurnRateWithSmartContractInstantiation verifies that burn rate and send commission are
// not accounted when smart contract sends token during instantiation.
func TestAssetFTSendCommissionAndBurnRateWithSmartContractInstantiation(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t, integration.NewFundedAccount(issuer, chain.NewCoin(sdkmath.NewInt(5000000000))))

	clientCtx := chain.ClientContext
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)

	// Issue a fungible token with burn rate and send commission rate
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             issuer.String(),
		Symbol:             "ABC",
		Subunit:            "abc",
		Precision:          6,
		InitialAmount:      sdkmath.NewInt(500),
		Description:        "ABC Description",
		Features:           []assetfttypes.Feature{},
		BurnRate:           sdk.MustNewDecFromStr("0.10"),
		SendCommissionRate: sdk.MustNewDecFromStr("0.20"),
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

	// send tokens from smart contract during instantiation - fees should not be charged.
	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		issuer,
		moduleswasm.BankSendWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    moduleswasm.BankSendWithdrawPayload(sdk.NewInt64Coin(denom, 40), recipient),
			Amount:     sdk.NewInt64Coin(denom, 100),
			Label:      "bank_send",
		},
	)
	requireT.NoError(err)
	contract := sdk.MustAccAddressFromBech32(contractAddr)

	// verify amounts
	assertCoinDistribution(ctx, clientCtx, t, denom, map[*sdk.AccAddress]int64{
		&issuer:    400,
		&contract:  60,
		&recipient: 40,
	})
}

// TestAssetFTSendingToSmartContractIsDenied verifies that this is not possible to send token to smart contract
// if issuer blocked this operation.
func TestAssetFTSendingToSmartContractIsDenied(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issuer := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(issuer, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

	clientCtx := chain.ClientContext

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
		},
		BurnRate:           sdk.ZeroDec(),
		SendCommissionRate: sdk.ZeroDec(),
	}

	_, err := client.BroadcastTx(
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
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

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
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)
}

// TestAssetFTAttachingToSmartContractIsDenied verifies that this is not possible to attach token to smart contract call
// if issuer blocked this operation.
func TestAssetFTAttachingToSmartContractCallIsDenied(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issuer := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(issuer, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

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
		},
		BurnRate:           sdk.ZeroDec(),
		SendCommissionRate: sdk.ZeroDec(),
	}

	_, err := client.BroadcastTx(
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
	requireT.ErrorContains(err, "unauthorized")
}

// TestAssetFTAttachingToSmartContractIsDenied verifies that this is not possible to attach token to smart contract
// instantiation if issuer blocked this operation.
func TestAssetFTAttachingToSmartContractInstantiationIsDenied(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issuer := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(issuer, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

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
		},
		BurnRate:           sdk.ZeroDec(),
		SendCommissionRate: sdk.ZeroDec(),
	}

	_, err := client.BroadcastTx(
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

	// This operation should fail due to coins being attached to it
	_, _, err = chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		issuer,
		moduleswasm.SimpleStateWASM,
		integration.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayload,
			Amount:     sdk.NewInt64Coin(denom, 100),
			Label:      "simple_state",
		},
	)
	requireT.ErrorContains(err, "unauthorized")
}

// TestAssetFTIssuingSmartContractIsAllowedToReceive verifies that issuing smart contract is allowed to
// receive coins even if sending them to smart contract is disabled.
func TestAssetFTIssuingSmartContractIsAllowedToSendAndReceive(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	admin := chain.GenAccount()
	recipient := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)
	chain.FundAccountWithOptions(ctx, t, recipient, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})

	txf := chain.TxFactory().
		WithSimulateAndExecute(true)

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
		},
		BurnRate:           sdk.ZeroDec().String(),
		SendCommissionRate: sdk.ZeroDec().String(),
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

// TestAssetFTMintingAndSendingOnBehalfOfIssuingSmartContractIsPossibleEvenIfSmartContractsAreBlocked verifies
// that it is possible to use authz to mint and send the token on behalf of the issuing smart contract if smart
// contracts are blocked.
func TestAssetFTMintingAndSendingOnBehalfOfIssuingSmartContractIsPossibleEvenIfSmartContractsAreBlocked(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	admin := chain.GenAccount()
	grantee := chain.GenAccount()
	recipient := chain.GenAccount()

	requireT := require.New(t)
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

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
	subunit := "uabc"
	denom := assetfttypes.BuildDenom(subunit, sdk.MustAccAddressFromBech32(contractAddr))
	execMsg := authztypes.NewMsgExec(grantee, []sdk.Msg{
		&assetfttypes.MsgIssue{
			Issuer:        contractAddr,
			Symbol:        "ABC",
			Subunit:       subunit,
			Precision:     6,
			Description:   "ABC Description",
			InitialAmount: sdkmath.NewInt(1000),
			Features: []assetfttypes.Feature{
				assetfttypes.Feature_minting,
				assetfttypes.Feature_block_smart_contracts,
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

// TestAssetFTSendingTokensFromRegularAccountBySmartContractUsingAuthZIsDenied verifies that it is
// forbidden to execute authz message to transfer coins by smart contract from regular account if
// smart contracts are blocked.
func TestAssetFTSendingTokensFromRegularAccountBySmartContractUsingAuthZIsDenied(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issuer := chain.GenAccount()
	admin := chain.GenAccount()
	granter := chain.GenAccount()
	recipient := chain.GenAccount()

	requireT := require.New(t)
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
			&authztypes.MsgGrant{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})
	chain.FundAccountWithOptions(ctx, t, granter, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&authztypes.MsgGrant{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})
	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(5000000000))),
	)

	// issue token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_block_smart_contracts,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

	// send half of the tokens to another account
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   granter.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(denom, 500)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
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
			Label:      "authz",
			Amount:     chain.QueryAssetFTParams(ctx, t).IssueFee,
		},
	)
	requireT.NoError(err)

	// grant authorization from issuer
	grantMsg, err := authztypes.NewMsgGrant(
		issuer,
		sdk.MustAccAddressFromBech32(contractAddr),
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantMsg)),
		grantMsg,
	)
	requireT.NoError(err)

	// grant authorization from granter
	grantMsg, err = authztypes.NewMsgGrant(
		granter,
		sdk.MustAccAddressFromBech32(contractAddr),
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantMsg)),
		grantMsg,
	)
	requireT.NoError(err)

	// send from the issuer - should work
	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(denom, 100)),
	}

	sendMsgAny, err := codectypes.NewAnyWithValue(sendMsg)
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		admin,
		contractAddr,
		moduleswasm.AuthZExecuteStargateRequest(&authztypes.MsgExec{
			Grantee: contractAddr,
			Msgs: []*codectypes.Any{
				sendMsgAny,
			},
		}),
		sdk.Coin{},
	)
	requireT.NoError(err)

	// send from the granter - should be rejected
	sendMsg = &banktypes.MsgSend{
		FromAddress: granter.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(denom, 100)),
	}

	sendMsgAny, err = codectypes.NewAnyWithValue(sendMsg)
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		admin,
		contractAddr,
		moduleswasm.AuthZExecuteStargateRequest(&authztypes.MsgExec{
			Grantee: contractAddr,
			Msgs: []*codectypes.Any{
				sendMsgAny,
			},
		}),
		sdk.Coin{},
	)
	requireT.ErrorContains(err, "unauthorized")

	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&issuer:    400,
		&granter:   500,
		&recipient: 100,
	})
}

func assertCoinDistribution(
	ctx context.Context,
	clientCtx client.Context,
	t *testing.T, denom string,
	dist map[*sdk.AccAddress]int64,
) {
	bankClient := banktypes.NewQueryClient(clientCtx)
	requireT := require.New(t)

	total := int64(0)
	for acc, expectedBalance := range dist {
		total += expectedBalance
		getBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{Address: acc.String(), Denom: denom})
		requireT.NoError(err)
		requireT.Equal(sdk.NewCoin(denom, sdkmath.NewInt(expectedBalance)).String(), getBalance.Balance.String())
	}

	supply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: denom})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdkmath.NewInt(total)).String(), supply.Amount.String())
}
