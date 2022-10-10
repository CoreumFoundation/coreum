package cli_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/asset/client/cli"
)

func TestIssueAsset(t *testing.T) {
	const name = "name"

	requireT := require.New(t)
	networkCfg, err := config.NetworkByChainID(config.Devnet)
	requireT.NoError(err)
	app.ChosenNetwork = networkCfg

	testNetwork := network.New(t)

	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx

	args := []string{name}
	args = append(args, txValidator1Args(testNetwork)...)
	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssueAsset(), args)
	requireT.NoError(err)

	var res sdk.TxResponse
	requireT.NoError(tmjson.Unmarshal(buf.Bytes(), &res))
	requireT.NotEmpty(res.TxHash)
	requireT.Equal(uint32(0), res.Code)
}

func txValidator1Args(testNetwork *network.Network) []string {
	return []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, testNetwork.Validators[0].Address.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(testNetwork.Config.BondDenom, sdk.NewInt(1000000))).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}
}
