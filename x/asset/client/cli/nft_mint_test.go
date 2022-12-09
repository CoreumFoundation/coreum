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

	args := []string{symbol, "class name", "class description", "https://my-class-meta.int/1", "35b326a2b3b605270c26185c38d2581e937b2eae0418b4964ef521efe79cdf34"}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxCreateNonFungibleTokenClass(), args)
	requireT.NoError(err)

	classID := types.BuildNonFungibleTokenClassID(symbol, validator.Address)
	args = []string{classID, "nft-1", "https://my-nft-meta.int/1", "9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9"}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxMintNonFungibleToken(), args)
	requireT.NoError(err)
}
