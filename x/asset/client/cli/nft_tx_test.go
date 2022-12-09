package cli_test

import (
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/asset/client/cli"
)

func TestCmdTxCreateNonFungibleTokenClass(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx

	args := []string{symbol, "class name", "class description", "https://my-class-meta.int/1", "35b326a2b3b605270c26185c38d2581e937b2eae0418b4964ef521efe79cdf34"}
	args = append(args, txValidator1Args(testNetwork)...)
	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxCreateNonFungibleTokenClass(), args)
	requireT.NoError(err)

	var res sdk.TxResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &res))
	requireT.NotEmpty(res.TxHash)
	requireT.Equal(uint32(0), res.Code, "can't submit CreateNonFungibleTokenClass tx", res)
}
