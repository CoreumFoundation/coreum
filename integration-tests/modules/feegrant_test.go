//go:build integrationtests

package modules

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/feegrant"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v5/integration-tests"
	"github.com/CoreumFoundation/coreum/v5/pkg/client"
	"github.com/CoreumFoundation/coreum/v5/testutil/integration"
)

func TestFeeGrant(t *testing.T) {
	t.Parallel()
	requireT := require.New(t)
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	granter := chain.GenAccount()
	grantee := chain.GenAccount()
	recipient := chain.GenAccount()
	feegrantClient := feegrant.NewQueryClient(chain.ClientContext)
	tmQueryClient := cmtservice.NewServiceClient(chain.ClientContext)

	chain.FundAccountWithOptions(ctx, t, granter, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&feegrant.MsgGrantAllowance{},
			&feegrant.MsgGrantAllowance{},
			&feegrant.MsgRevokeAllowance{},
		},
		Amount: sdkmath.NewInt(200_000),
	})
	chain.FundAccountWithOptions(ctx, t, grantee, integration.BalancesOptions{
		Amount: sdkmath.NewInt(1),
	})

	basicAllowance, err := codectypes.NewAnyWithValue(&feegrant.BasicAllowance{
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
	requireT.EqualValues(chain.GasLimitByMsgs(grantMsg), res.GasUsed)

	blockRes, err := tmQueryClient.GetLatestBlock(ctx, &cmtservice.GetLatestBlockRequest{})
	requireT.NoError(err)

	expiringAllowance, err := codectypes.NewAnyWithValue(&feegrant.BasicAllowance{
		SpendLimit: nil, // empty means no limit
		Expiration: lo.ToPtr(blockRes.SdkBlock.Header.Time.Add(10 * time.Second)),
	})
	requireT.NoError(err)

	expiringGrantMsg := &feegrant.MsgGrantAllowance{
		Granter:   granter.String(),
		Grantee:   recipient.String(),
		Allowance: expiringAllowance,
	}

	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(expiringGrantMsg)),
		expiringGrantMsg,
	)
	requireT.NoError(err)
	requireT.EqualValues(chain.GasLimitByMsgs(expiringGrantMsg), res.GasUsed)

	allowancesRes, err := feegrantClient.AllowancesByGranter(ctx, &feegrant.QueryAllowancesByGranterRequest{
		Granter: granter.String(),
	})
	requireT.NoError(err)
	requireT.Len(allowancesRes.Allowances, 2)

	// await next 5 blocks
	requireT.NoError(client.AwaitNextBlocks(ctx, chain.ClientContext, 10))

	pruneAllowancesMsg := &feegrant.MsgPruneAllowances{
		Pruner: granter.String(),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(200_000),
		pruneAllowancesMsg,
	)
	requireT.NoError(err)

	allowancesRes, err = feegrantClient.AllowancesByGranter(ctx, &feegrant.QueryAllowancesByGranterRequest{
		Granter: granter.String(),
	})
	requireT.NoError(err)
	requireT.Len(allowancesRes.Allowances, 1)

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
	requireT.EqualValues(chain.GasLimitByMsgs(revokeMsg), res.GasUsed)

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
