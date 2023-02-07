package cli_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/samber/lo"
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
	ctx := testNetwork.Validators[0].ClientCtx

	classID := issueClass(
		requireT, ctx,
		symbol, name, description, URI, URIHash,
		testNetwork,
		"0.1",
		types.ClassFeature_burning,
		types.ClassFeature_disable_sending,
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
			types.ClassFeature_burning,
			types.ClassFeature_disable_sending,
		},
		RoyaltyRate: sdk.MustNewDecFromStr("0.1"),
	}, resp.Class)
}

func issueClass(
	requireT *require.Assertions,
	ctx client.Context,
	symbol, name, description, url, urlHash string,
	testNetwork *network.Network,
	royaltyRate string,
	features ...types.ClassFeature,
) string {
	featuresStringList := lo.Map(features, func(s types.ClassFeature, _ int) string {
		return s.String()
	})
	featuresString := strings.Join(featuresStringList, ",")
	validator := testNetwork.Validators[0]
	args := []string{symbol, name, description, url, urlHash, fmt.Sprintf("--features=%s", featuresString)}
	args = append(args, txValidator1Args(testNetwork)...)
	if royaltyRate != "" {
		args = append(args, "--royalty-rate", royaltyRate)
	}
	_, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssueClass(), args)
	requireT.NoError(err)

	return types.BuildClassID(symbol, validator.Address)
}

func mint(
	requireT *require.Assertions,
	ctx client.Context,
	classID, nftID, url, urlHash string,
	testNetwork *network.Network,
) {
	args := []string{classID, nftID, url, urlHash}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxMint(), args)
	requireT.NoError(err)
}
