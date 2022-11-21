package cli_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/testutil/event"
	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/asset/client/cli"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestQueryFungibleToken(t *testing.T) {
	requireT := require.New(t)
	networkCfg, err := config.NetworkByChainID(constant.ChainIDDev)
	requireT.NoError(err)
	app.ChosenNetwork = networkCfg

	testNetwork := network.New(t)

	// the denom must start from the letter
	symbol := "BTC" + uuid.NewString()[:4]
	ctx := testNetwork.Validators[0].ClientCtx

	denom := createFungibleToken(requireT, ctx, symbol, testNetwork)

	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryFungibleToken(), []string{denom, "--output", "json"})
	requireT.NoError(err)

	var resp types.QueryFungibleTokenResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))

	requireT.Equal(types.FungibleToken{
		Denom:       denom,
		Issuer:      testNetwork.Validators[0].Address.String(),
		Symbol:      symbol,
		Description: "",
		Features:    []types.FungibleTokenFeature{},
	}, resp.FungibleToken)
}

func createFungibleToken(requireT *require.Assertions, ctx client.Context, symbol string, testNetwork *network.Network) string {
	args := []string{symbol, "", "", ""}
	args = append(args, txValidator1Args(testNetwork)...)
	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssueFungibleToken(), args)
	requireT.NoError(err)

	var res sdk.TxResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &res))
	requireT.NotEmpty(res.TxHash)
	requireT.Equal(uint32(0), res.Code, "can't submit IssueFungibleToken tx for query", res)

	eventFungibleTokenIssuedName := proto.MessageName(&types.EventFungibleTokenIssued{})
	for i := range res.Events {
		if res.Events[i].Type != eventFungibleTokenIssuedName {
			continue
		}
		eventFungibleTokenIssued, err := event.FindTypedEvent[*types.EventFungibleTokenIssued](res.Events)
		requireT.NoError(err)
		return eventFungibleTokenIssued.Denom
	}
	requireT.Failf("event: %s not found in the create fungible token response", eventFungibleTokenIssuedName)

	return ""
}
