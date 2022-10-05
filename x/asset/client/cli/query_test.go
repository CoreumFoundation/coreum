package cli_test

import (
	"encoding/json"
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/asset/client/cli"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestQueryAsset(t *testing.T) {
	const id = "id1"

	requireT := require.New(t)
	networkCfg, err := config.NetworkByChainID(config.Devnet)
	requireT.NoError(err)
	app.ChosenNetwork = networkCfg

	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx
	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryAsset(), []string{id, "--output", "json"})
	requireT.NoError(err)

	var resp types.QueryAssetResponse
	requireT.NoError(json.Unmarshal(buf.Bytes(), &resp))
	requireT.Equal(id, resp.Asset.Id)
}
