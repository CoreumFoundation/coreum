package asset

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
)

// TestIssueBasicFungibleToken checks that fungible token is issued.
func TestIssueBasicFungibleToken(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	clientCtx := chain.ClientContext

	assetClient := assettypes.NewQueryClient(clientCtx)
	bankClient := banktypes.NewQueryClient(clientCtx)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
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
		Issuer:      issuer.String(),
		Symbol:      "BTC",
		Description: "BTC Description",
		// the custom receiver
		Recipient:     recipient.String(),
		InitialAmount: sdk.NewInt(777),
	}

	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)

	require.NoError(t, err)
	evt := testing.FindTypedEvent(t, &assettypes.EventFungibleTokenIssued{}, res.Events)
	fungibleTokenIssuedEvt, ok := evt.(*assettypes.EventFungibleTokenIssued)
	require.True(t, ok)

	require.NoError(t, err)
	require.Equal(t, assettypes.EventFungibleTokenIssued{
		Denom:         assettypes.BuildFungibleTokenDenom(msg.Symbol, issuer),
		Issuer:        msg.Issuer,
		Symbol:        msg.Symbol,
		Description:   msg.Description,
		Recipient:     msg.Recipient,
		InitialAmount: msg.InitialAmount,
	}, *fungibleTokenIssuedEvt)

	denom := fungibleTokenIssuedEvt.Denom

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
	// check the recipient balance
	recipientBalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, msg.InitialAmount).String(), recipientBalanceRes.Balance.String())

	// check the issuer balance
	issuerBalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: issuer.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.True(issuerBalanceRes.Balance.Amount.IsZero())
}
