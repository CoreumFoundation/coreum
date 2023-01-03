package modules

import (
	"testing"

	codetypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

func TestFeeGrant(t *testing.T) {
	t.Parallel()
	requireT := require.New(t)
	ctx, chain := integrationtests.NewTestingContext(t)

	granter := chain.GenAccount()
	grantee := chain.GenAccount()
	recipient := chain.GenAccount()

	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, granter, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   sdk.NewInt(1_000_000),
	}))
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, grantee, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   sdk.NewInt(10),
	}))
	basicAllowance, err := codetypes.NewAnyWithValue(&feegrant.BasicAllowance{
		SpendLimit: nil, // empty means no limit
		Expiration: nil, // empty means no limit
	})
	requireT.NoError(err)

	grantMsg := &feegrant.MsgGrantAllowance{
		Granter:   granter.String(),
		Grantee:   grantee.String(),
		Allowance: basicAllowance,
	}

	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantMsg)),
		grantMsg,
	)
	requireT.NoError(err)

	sendMsg := &banktypes.MsgSend{
		FromAddress: grantee.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdk.NewInt(1))),
	}

	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(grantee).WithFeeGranterAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
}
