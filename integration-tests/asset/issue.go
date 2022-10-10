package asset

import (
	"context"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
)

// TestIssueFungibleToken checks that fungible token is issued.
func TestIssueFungibleToken(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	chainContext := chain.ClientContext

	assetClient := assettypes.NewQueryClient(chainContext)
	bankClient := banktypes.NewQueryClient(chainContext)

	issuer := chain.RandomWallet()
	requireT.NoError(chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(
			issuer,
			chain.NewCoin(testing.ComputeNeededBalance(
				chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
				chain.GasLimitByMsgs(&assettypes.MsgIssueFungibleToken{}),
				1,
				sdk.NewInt(0),
			)),
		),
	))

	// Issue the new fungible token
	msg := &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "BTC",
		Description:   "BTC Description",
		Recipient:     issuer.String(),
		InitialAmount: sdk.Int{},
	}

	res, err := tx.BroadcastTx(
		ctx,
		chainContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)
	requireT.NoError(err)

	eventFungibleTokenIssuedName := proto.MessageName(&assettypes.EventFungibleTokenIssued{})
	denom, ok := client.FindEventAttribute(sdk.StringifyEvents(res.Events), eventFungibleTokenIssuedName, "denom")
	requireT.True(ok)
	// the typed events are decoded with the strings escape
	denom = strings.ReplaceAll(denom, "\"", "")

	// query for the token to check what is stored
	gotToken, err := assetClient.FungibleToken(ctx, &assettypes.QueryFungibleTokenRequest{
		Denom: denom,
	})
	requireT.NoError(err)

	requireT.Equal(assettypes.FungibleToken{
		Denom:       denom,
		Issuer:      msg.Issuer,
		Symbol:      msg.Symbol,
		Description: msg.Description,
	}, gotToken.FungibleToken)

	// query balance
	tokenBalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: issuer.String(),
		Denom:   denom,
	})
	requireT.NoError(err)

	requireT.Equal(sdk.NewCoin(denom, msg.InitialAmount).String(), tokenBalanceRes.Balance.String())
}
