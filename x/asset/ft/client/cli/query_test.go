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
	"github.com/CoreumFoundation/coreum/x/asset/ft/client/cli"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

func TestQueryToken(t *testing.T) {
	requireT := require.New(t)

	testNetwork := network.New(t)

	// the denom must start from the letter
	symbol := "btc" + uuid.NewString()[:4]
	subunit := "sub" + symbol
	precision := "8"
	ctx := testNetwork.Validators[0].ClientCtx

	denom := issue(requireT, ctx, symbol, subunit, precision, testNetwork)

	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryTokenInfo(), []string{denom, "--output", "json"})
	requireT.NoError(err)

	var resp types.QueryTokenResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))

	requireT.Equal(types.FT{
		Denom:              denom,
		Issuer:             testNetwork.Validators[0].Address.String(),
		Symbol:             symbol,
		Subunit:            strings.ToLower(subunit),
		Precision:          8,
		Description:        "",
		Features:           []types.TokenFeature{},
		BurnRate:           sdk.NewDec(0),
		SendCommissionRate: sdk.NewDec(0),
	}, resp.Token)
}

func issue(requireT *require.Assertions, ctx client.Context, symbol, subunit, precision string, testNetwork *network.Network) string {
	args := []string{symbol, subunit, precision, "", ""}
	args = append(args, txValidator1Args(testNetwork)...)
	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssue(), args)
	requireT.NoError(err)

	var res sdk.TxResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &res))
	requireT.NotEmpty(res.TxHash)
	requireT.Equal(uint32(0), res.Code, "can't submit Issue tx for query", res)

	eventIssuedName := proto.MessageName(&types.EventTokenIssued{})
	for i := range res.Events {
		if res.Events[i].Type != eventIssuedName {
			continue
		}
		eventsTokenIssued, err := event.FindTypedEvents[*types.EventTokenIssued](res.Events)
		requireT.NoError(err)
		return eventsTokenIssued[0].Denom
	}
	requireT.Failf("event: %s not found in the issue response", eventIssuedName)

	return ""
}
