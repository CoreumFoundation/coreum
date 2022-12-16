package cli_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/asset/client/cli"
)

func TestIssueFungibleToken(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	// the denom must start from the letter
	symbol := "BTC" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx
	subunit := "wsatoshi"
	precision := "8"

	args := []string{symbol, subunit, precision, "777", `"My Token"`}
	args = append(args, txValidator1Args(testNetwork)...)
	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssueFungibleToken(), args)
	requireT.NoError(err)

	var res sdk.TxResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &res))
	requireT.NotEmpty(res.TxHash)
	requireT.Equal(uint32(0), res.Code, "can't submit IssueFungibleToken tx", res)
}

func txValidator1Args(testNetwork *network.Network) []string {
	return []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, testNetwork.Validators[0].Address.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(testNetwork.Config.BondDenom, sdk.NewInt(1000000))).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}
}
