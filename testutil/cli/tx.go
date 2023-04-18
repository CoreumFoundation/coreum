package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/testutil/network"
)

// ExecTxCmdAndWaitNextBlock is a func to execute tx cmd.
func ExecTxCmdAndWaitNextBlock(clientCtx client.Context, testNetwork *network.Network, cmd *cobra.Command, extraArgs []string) (sdk.TxResponse, error) {
	buf, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, extraArgs)
	if err != nil {
		return sdk.TxResponse{}, errors.Errorf("can't execute, %s, err:%s", cmd.Use, err)
	}

	var res sdk.TxResponse
	if err := clientCtx.Codec.UnmarshalJSON(buf.Bytes(), &res); err != nil {
		return sdk.TxResponse{}, errors.Errorf("can't decode response, %s, err:%s", buf.Bytes(), err)
	}

	if uint32(0) != res.Code {
		return sdk.TxResponse{}, errors.Errorf("tx failed, response code %d, response, %s", res.Code, buf.Bytes())
	}

	res, err = clitestutil.GetTxResponse(testNetwork, clientCtx, res.TxHash)
	if err != nil {
		return sdk.TxResponse{}, errors.Errorf("can't get tx response, err: %s", err)
	}

	return res, err
}

// ExecQueryCmd is a func to execute query cmd.
func ExecQueryCmd(clientCtx client.Context, cmd *cobra.Command, extraArgs []string, msg proto.Message) error {
	extraArgs = append(extraArgs, "--output", "json")
	buf, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, extraArgs)
	if err != nil {
		return errors.Errorf("can't execute, %s, err:%s", cmd.Use, err)
	}

	if err := clientCtx.Codec.UnmarshalJSON(buf.Bytes(), msg); err != nil {
		return errors.Errorf("can't decode response, %s, err:%s", buf.Bytes(), err)
	}

	return nil
}
