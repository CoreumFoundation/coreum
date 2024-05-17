//go:build integrationtests && gasestimationtests

package modules

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
	"github.com/CoreumFoundation/coreum/v4/pkg/client"
	"github.com/CoreumFoundation/coreum/v4/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

// TestBankSendEstimation is used to estimate gas required by each additional token present in bank send message.
// It executes transactions sending from 1 to 201 tokens in single message and reports gas estimated by each of them.
// Then you may copy the results to a spreadsheet and calculate the gas required by each transfer.
// Spreadsheet example might be found here: https://docs.google.com/spreadsheets/d/1qoKa8udTPYdS_-ofJ8xNbnZFDh-gqGb4n0_OgcTHUOA/edit?usp=sharing
// Keep in mind that to estimate the gas you need to move bank send message to nondeterministic section inside deterministic gas config.
func TestBankSendEstimation(t *testing.T) {
	const (
		nTokens = 101
		step    = 20
	)

	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issuer := chain.GenAccount()
	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	tokens := make([]string, 0, nTokens)
	deterministicMsgs := make([]sdk.Msg, 0, 3*nTokens)
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient1.String(),
	}
	initialAmount := sdkmath.NewInt(100_000_000_000)
	for i := 0; i < nTokens; i++ {
		subunit := fmt.Sprintf("tok%d", i)
		denom := assetfttypes.BuildDenom(subunit, issuer)
		deterministicMsgs = append(deterministicMsgs, &assetfttypes.MsgIssue{
			Issuer:        issuer.String(),
			Symbol:        fmt.Sprintf("TOK%d", i),
			Subunit:       fmt.Sprintf("tok%d", i),
			Precision:     1,
			Description:   fmt.Sprintf("TOK%d", i),
			InitialAmount: initialAmount,
			Features: []assetfttypes.Feature{
				assetfttypes.Feature_minting,
				assetfttypes.Feature_burning,
				assetfttypes.Feature_freezing,
				assetfttypes.Feature_whitelisting,
				assetfttypes.Feature_ibc,
			},
			BurnRate:           sdkmath.LegacyMustNewDecFromStr("0.01"),
			SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("0.01"),
		}, &assetfttypes.MsgSetWhitelistedLimit{
			Sender:  issuer.String(),
			Account: recipient1.String(),
			Coin:    sdk.NewInt64Coin(denom, 1_000_000_000_000),
		}, &assetfttypes.MsgSetWhitelistedLimit{
			Sender:  issuer.String(),
			Account: recipient2.String(),
			Coin:    sdk.NewInt64Coin(denom, 1_000_000_000_000),
		})

		tokens = append(tokens, assetfttypes.BuildDenom(subunit, issuer))
		sendMsg.Amount = sendMsg.Amount.Add(sdk.NewCoin(denom, initialAmount))
	}

	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: deterministicMsgs,
		Amount:   chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.MulRaw(nTokens).AddRaw(1_000_000_000),
	})

	chain.FundAccountWithOptions(ctx, t, recipient1, integration.BalancesOptions{
		Amount: sdk.NewIntFromUint64(1_000_000_000),
	})

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(deterministicMsgs...)+5_000_000),
		append(append([]sdk.Msg{}, deterministicMsgs...), sendMsg)...,
	)
	requireT.NoError(err)

	for n := 1; n <= nTokens; n += step {
		sendMsg := &banktypes.MsgSend{
			FromAddress: recipient1.String(),
			ToAddress:   recipient2.String(),
		}

		for i := 0; i < n; i++ {
			sendMsg.Amount = sendMsg.Amount.Add(sdk.NewCoin(tokens[i], sdkmath.NewInt(1_0000_000)))
		}

		txRes, err := client.BroadcastTx(
			ctx,
			chain.ClientContext.WithFromAddress(recipient1),
			chain.TxFactory().WithGas(50_000_000),
			sendMsg,
		)
		requireT.NoError(err)

		fmt.Printf("%d\t%d\n", n, txRes.GasUsed)
	}
}

// TestAuthzEstimation it estimates gas overhead required by authz message execution.
// It executes regular message first. Then the same message is executed using authz. By subtracting those values
// we know what the overhead of authz is.
// To get correct results, both authz and bank send must be temporarily configured as non-deterministic messages,
// to get real results.
func TestAuthzEstimation(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	granter := chain.GenAccount()
	grantee := chain.GenAccount()
	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	chain.Faucet.FundAccounts(ctx, t,
		integration.FundedAccount{
			Address: granter,
			Amount:  chain.NewCoin(sdk.NewInt(50000000)),
		},
		integration.FundedAccount{
			Address: grantee,
			Amount:  chain.NewCoin(sdk.NewInt(50000000)),
		},
	)

	// grant the authorization
	grantMsg, err := authztypes.NewMsgGrant(
		granter,
		grantee,
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

	// execute regular message
	amountToSend := sdkmath.NewInt(2_000)
	// we don't use the gas multiplier here intentionally
	txf := chain.TxFactory().WithSimulateAndExecute(true)
	resRegular, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		txf,
		&banktypes.MsgSend{
			FromAddress: granter.String(),
			ToAddress:   recipient1.String(),
			Amount:      sdk.NewCoins(chain.NewCoin(amountToSend)),
		},
	)
	requireT.NoError(err)

	// execute authz message
	execMsg := authztypes.NewMsgExec(grantee, []sdk.Msg{
		&banktypes.MsgSend{
			FromAddress: granter.String(),
			ToAddress:   recipient2.String(),
			Amount:      sdk.NewCoins(chain.NewCoin(amountToSend)),
		},
	})

	resAuthZ, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(grantee),
		txf,
		&execMsg,
	)
	requireT.NoError(err)

	fmt.Printf("Authz gas overhead: %d\n", resAuthZ.GasUsed-resRegular.GasUsed)
}
