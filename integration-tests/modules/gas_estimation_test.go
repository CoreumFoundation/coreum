//go:build integrationtests && gasestimationtests

package modules

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	"github.com/CoreumFoundation/coreum/v3/pkg/client"
	"github.com/CoreumFoundation/coreum/v3/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
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
