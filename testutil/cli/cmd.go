package cli

import (
	"fmt"
	"testing"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	tmcli "github.com/cometbft/cometbft/libs/cli"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v5/app"
	"github.com/CoreumFoundation/coreum/v5/testutil/network"
)

// ExecTxCmd is a func to execute tx cmd.
func ExecTxCmd(
	clientCtx client.Context,
	testNetwork *network.Network,
	cmd *cobra.Command,
	extraArgs []string,
) (sdk.TxResponse, error) {
	buf, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, extraArgs)
	if err != nil {
		return sdk.TxResponse{}, errors.Errorf("can't execute, %s, err:%s", cmd.Use, err)
	}

	var res sdk.TxResponse
	if err := clientCtx.Codec.UnmarshalJSON(buf.Bytes(), &res); err != nil {
		return sdk.TxResponse{}, errors.Errorf("can't decode response, %s, err:%s", buf.Bytes(), err)
	}

	if uint32(0) != res.Code {
		// we don't wrap error to sdkerrors.ABCIError since it won't contain the codes at that time
		return sdk.TxResponse{}, errors.Errorf("tx failed, response code %d, response, %s", res.Code, buf.Bytes())
	}

	// we use it to get tx with result from the next block
	res, err = clitestutil.GetTxResponse(testNetwork, clientCtx, res.TxHash)
	if err != nil {
		return sdk.TxResponse{}, errors.Errorf("can't get tx response, err: %s", err)
	}

	if uint32(0) != res.Code {
		return sdk.TxResponse{}, errors.Wrapf(sdkerrors.ABCIError(res.Codespace, res.Code, res.Logs.String()),
			"transaction '%s' failed, raw log:%s", res.TxHash, res.RawLog)
	}

	return res, nil
}

// ExecQueryCmd is a func to execute query cmd.
func ExecQueryCmd(t *testing.T, clientCtx client.Context, cmd *cobra.Command, extraArgs []string, msg proto.Message) {
	t.Helper()

	extraArgs = append(extraArgs, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	buf, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, extraArgs)
	require.NoError(t, err, fmt.Sprintf("can't execute, %s, err:%s", cmd.Use, err))

	err = clientCtx.Codec.UnmarshalJSON(buf.Bytes(), msg)
	require.NoError(t, err, fmt.Sprintf("can't decode response, %s, err:%s", buf.Bytes(), err))
}

// ExecRootQueryCmd is a func to execute query cmd from root.
func ExecRootQueryCmd(t *testing.T, clientCtx client.Context, args []string, msg proto.Message) {
	t.Helper()
	// prepend query
	args = append([]string{"query"}, args...)
	// append json
	args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	cmd := &cobra.Command{
		Use: "root",
	}
	tempApp := app.New(log.NewNopLogger(), dbm.NewMemDB(), nil, true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))

	autoCliOpts := tempApp.AutoCliOpts()
	autoCliOpts.ClientCtx = clientCtx
	require.NoError(t, autoCliOpts.EnhanceRootCommand(cmd))

	buf, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	require.NoError(t, err, fmt.Sprintf("failed to execute, %v", args))
	require.NoError(
		t,
		clientCtx.LegacyAmino.UnmarshalJSON(buf.Bytes(), msg),
		fmt.Sprintf("failed to decode response, %s", buf.Bytes()),
	)
}
