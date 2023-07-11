package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

// ExecTestCLICmd executes ClI command and returns an error.
func ExecTestCLICmd(clientCtx client.Context, cmd *cobra.Command, extraArgs []string) error {
	buf, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, extraArgs)
	if err != nil {
		return errors.WithStack(err)
	}

	var txResponse sdk.TxResponse

	if err := tmjson.Unmarshal(buf.Bytes(), &txResponse); err != nil {
		return errors.WithStack(err)
	}

	if txResponse.Code == 0 {
		return nil
	}

	return errors.Wrapf(sdkerrors.ABCIError(txResponse.Codespace, txResponse.Code, txResponse.Logs.String()),
		"transaction '%s' failed, raw log:%s", txResponse.TxHash, txResponse.RawLog)
}
