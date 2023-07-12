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

	"github.com/CoreumFoundation/coreum/v2/pkg/config/constant"
	coreumclitestutil "github.com/CoreumFoundation/coreum/v2/testutil/cli"
	"github.com/CoreumFoundation/coreum/v2/testutil/network"
	"github.com/CoreumFoundation/coreum/v2/x/asset/nft/client/cli"
	"github.com/CoreumFoundation/coreum/v2/x/asset/nft/types"
)

func TestQueryClassAndClasses(t *testing.T) {
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

	var classRes types.QueryClassResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &classRes))

	expectedClass := types.Class{
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
	}

	requireT.Equal(expectedClass, classRes.Class)

	// classes
	buf, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryClasses(),
		[]string{fmt.Sprintf("--%s", cli.IssuerFlag), testNetwork.Validators[0].Address.String(), "--output", "json"},
	)
	requireT.NoError(err)

	var classesRes types.QueryClassesResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &classesRes))

	requireT.Equal(expectedClass, classesRes.Classes[0])
}

func TestCmdTxMint(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx

	args := []string{symbol, "class name", "class description", "https://my-class-meta.invalid/1", "content-hash"}
	args = append(args, txValidator1Args(testNetwork)...)
	requireT.NoError(coreumclitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssueClass(), args))

	classID := types.BuildClassID(symbol, validator.Address)
	args = []string{classID, "nft-1", "https://my-nft-meta.invalid/1", "9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9"}
	args = append(args, txValidator1Args(testNetwork)...)
	requireT.NoError(coreumclitestutil.ExecTestCLICmd(ctx, cli.CmdTxMint(), args))
}

func TestCmdTxBurn(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx

	args := []string{symbol, "class name", "class description", "https://my-class-meta.invalid/1", "content-hash"}
	args = append(args, txValidator1Args(testNetwork)...)
	requireT.NoError(coreumclitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssueClass(), args))

	classID := types.BuildClassID(symbol, validator.Address)
	args = []string{classID, "nft-1", "https://my-nft-meta.invalid/1", "9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9"}
	args = append(args, txValidator1Args(testNetwork)...)
	requireT.NoError(coreumclitestutil.ExecTestCLICmd(ctx, cli.CmdTxMint(), args))

	args = []string{classID, "nft-1"}
	args = append(args, txValidator1Args(testNetwork)...)
	requireT.NoError(coreumclitestutil.ExecTestCLICmd(ctx, cli.CmdTxBurn(), args))

	args = []string{classID, "nft-1", "--output", "json"}
	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryBurnt(), args)
	requireT.NoError(err)

	var resp types.QueryBurntNFTResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))
	requireT.True(resp.Burnt)

	args = []string{classID, "--output", "json"}
	buf, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryBurnt(), args)
	requireT.NoError(err)

	var respList types.QueryBurntNFTsInClassResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &respList))
	requireT.Len(respList.NftIds, 1)
}

func TestCmdQueryParams(t *testing.T) {
	requireT := require.New(t)

	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx

	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryParams(), []string{"--output", "json"})
	requireT.NoError(err)

	var resp types.QueryParamsResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))

	expectedMintFee := sdk.Coin{Denom: constant.DenomDev, Amount: sdk.NewInt(0)}
	requireT.Equal(expectedMintFee, resp.Params.MintFee)
}

func mint(
	requireT *require.Assertions,
	ctx client.Context,
	classID, nftID, url, urlHash string,
	testNetwork *network.Network,
) {
	args := []string{classID, nftID, url, urlHash}
	args = append(args, txValidator1Args(testNetwork)...)
	requireT.NoError(coreumclitestutil.ExecTestCLICmd(ctx, cli.CmdTxMint(), args))
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
	args := []string{symbol, name, description, url, urlHash, fmt.Sprintf("--%s=%s", cli.FeaturesFlag, featuresString)}
	args = append(args, txValidator1Args(testNetwork)...)
	if royaltyRate != "" {
		args = append(args, fmt.Sprintf("--%s", cli.RoyaltyRateFlag), royaltyRate)
	}
	requireT.NoError(coreumclitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssueClass(), args))

	return types.BuildClassID(symbol, validator.Address)
}
