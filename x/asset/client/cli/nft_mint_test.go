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

	args := []string{symbol, "NFT-name", "My NFT class description.", "my-uri", "my-hash"}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxCreateNonFungibleTokenClass(), args)
	requireT.NoError(err)

	classID := types.BuildNonFungibleTokenClassID(symbol, validator.Address)
	args = []string{classID, "firstID", "my-uri", "my-hash"}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxMintNonFungibleToken(), args)
	requireT.NoError(err)
}
