//go:build integrationtests

package modules

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	codetypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v2/integration-tests"
	"github.com/CoreumFoundation/coreum/v2/pkg/client"
)

func TestFeeGrant(t *testing.T) {
	t.Parallel()
	requireT := require.New(t)
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	granter := chain.GenAccount()
	grantee := chain.GenAccount()
	recipient := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, granter, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&feegrant.MsgGrantAllowance{},
			&feegrant.MsgRevokeAllowance{},
		},
	})
	chain.FundAccountWithOptions(ctx, t, grantee, integrationtests.BalancesOptions{
		Amount: sdkmath.NewInt(1),
	})
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

	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantMsg)),
		grantMsg,
	)
	requireT.NoError(err)
	requireT.EqualValues(res.GasUsed, chain.GasLimitByMsgs(grantMsg))
	sendMsg := &banktypes.MsgSend{
		FromAddress: grantee.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(grantee).WithFeeGranterAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	revokeMsg := &feegrant.MsgRevokeAllowance{
		Granter: granter.String(),
		Grantee: grantee.String(),
	}

	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(revokeMsg)),
		revokeMsg,
	)
	requireT.NoError(err)
	requireT.EqualValues(res.GasUsed, chain.GasLimitByMsgs(revokeMsg))

	sendMsg = &banktypes.MsgSend{
		FromAddress: grantee.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1))),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(grantee).WithFeeGranterAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.Error(err)
	requireT.True(cosmoserrors.ErrNotFound.Is(err))
}
