package cli

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	tmcli "github.com/cometbft/cometbft/libs/cli"
	"github.com/cosmos/cosmos-sdk/client"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/v2/testutil/network"
)

// ExecTxCmd is a func to execute tx cmd.
func ExecTxCmd(clientCtx client.Context, testNetwork *network.Network, cmd *cobra.Command, extraArgs []string) (sdk.TxResponse, error) {
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
func ExecQueryCmd(clientCtx client.Context, cmd *cobra.Command, extraArgs []string, msg proto.Message) error {
	extraArgs = append(extraArgs, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	buf, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, extraArgs)
	if err != nil {
		return errors.Errorf("can't execute, %s, err:%s", cmd.Use, err)
	}

	if err := clientCtx.Codec.UnmarshalJSON(buf.Bytes(), msg); err != nil {
		return errors.Errorf("can't decode response, %s, err:%s", buf.Bytes(), err)
	}

	return nil
}
