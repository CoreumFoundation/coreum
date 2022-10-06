package cli_test

import (
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/asset/client/cli"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestQueryAsset(t *testing.T) {
	const id = uint64(1)

	requireT := require.New(t)
	networkCfg, err := config.NetworkByChainID(config.Devnet)
	requireT.NoError(err)
	app.ChosenNetwork = networkCfg

	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx

	createAsset(requireT, ctx, testNetwork)

	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryAsset(), []string{strconv.Itoa(int(id)), "--output", "json"})
	requireT.NoError(err)

	var resp types.QueryAssetResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))
	requireT.Equal(id, resp.Asset.Id)
}

func createAsset(requireT *require.Assertions, ctx client.Context, testNetwork *network.Network) {
	args := []string{testNetwork.Validators[0].Address.String(), "BTC", `"BTC Token"`, "6", "777"}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssueFTAsset(), args)
	requireT.NoError(err)
}
