package cli_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/asset/nft/client/cli"
	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

func TestQueryClass(t *testing.T) {
	requireT := require.New(t)

	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	name := "class name"
	description := "class description"
	URI := "https://my-class-meta.invalid/1"
	URIHash := "content-hash"
	features := types.ClassFeature_burning.String() //nolint:nosnakecase // generated variable
	ctx := testNetwork.Validators[0].ClientCtx

	classID := issueClass(
		requireT, ctx,
		symbol, name, description, URI, URIHash, features,
		testNetwork,
	)

	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryClass(), []string{classID, "--output", "json"})
	requireT.NoError(err)

	var resp types.QueryClassResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))

	requireT.Equal(types.Class{
		Id:          classID,
		Issuer:      testNetwork.Validators[0].Address.String(),
		Name:        name,
		Symbol:      symbol,
		Description: description,
		URI:         URI,
		URIHash:     URIHash,
		Features: []types.ClassFeature{
			types.ClassFeature_burning, //nolint:nosnakecase // generated variable
		},
	}, resp.Class)
}

func issueClass(
	requireT *require.Assertions,
	ctx client.Context,
	symbol, name, description, url, urlHash, features string,
	testNetwork *network.Network,
) string {
	validator := testNetwork.Validators[0]
	args := []string{symbol, name, description, url, urlHash, fmt.Sprintf("--features=%s", features)}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssueClass(), args)
	requireT.NoError(err)

	return types.BuildClassID(symbol, validator.Address)
}
