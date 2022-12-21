package cli_test

import (
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/asset/client/cli"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestCmdTxMintNonFungibleToken(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx

	args := []string{symbol, "class name", "class description", "https://my-class-meta.invalid/1", "content-hash"}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssueNonFungibleTokenClass(), args)
	requireT.NoError(err)

	classID := types.BuildNonFungibleTokenClassID(symbol, validator.Address)
	args = []string{classID, "nft-1", "https://my-nft-meta.invalid/1", "9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9"}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxMintNonFungibleToken(), args)
	requireT.NoError(err)
}
