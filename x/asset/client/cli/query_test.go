package cli_test

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/testutil/event"
	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/asset/client/cli"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestQueryFungibleToken(t *testing.T) {
	requireT := require.New(t)

	testNetwork := network.New(t)

	// the denom must start from the letter
	symbol := "btc" + uuid.NewString()[:4]
	subunit := "sub" + symbol
	precision := "8"
	ctx := testNetwork.Validators[0].ClientCtx

	denom := createFungibleToken(requireT, ctx, symbol, subunit, precision, testNetwork)

	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryFungibleToken(), []string{denom, "--output", "json"})
	requireT.NoError(err)

	var resp types.QueryFungibleTokenResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))

	requireT.Equal(types.FungibleToken{
		Denom:       denom,
		Issuer:      testNetwork.Validators[0].Address.String(),
		Symbol:      symbol,
		Subunit:     strings.ToLower(subunit),
		Precision:   8,
		Description: "",
		Features:    []types.FungibleTokenFeature{},
		BurnRate:    sdk.NewDec(0),
	}, resp.FungibleToken)
}

func createFungibleToken(requireT *require.Assertions, ctx client.Context, symbol, subunit, precision string, testNetwork *network.Network) string {
	args := []string{symbol, subunit, precision, "", ""}
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
		eventsFungibleTokenIssued, err := event.FindTypedEvents[*types.EventFungibleTokenIssued](res.Events)
		requireT.NoError(err)
		return eventsFungibleTokenIssued[0].Denom
	}
	requireT.Failf("event: %s not found in the create fungible token response", eventFungibleTokenIssuedName)

	return ""
}
