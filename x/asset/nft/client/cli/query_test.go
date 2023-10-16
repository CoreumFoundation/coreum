package cli_test

import (
	"fmt"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v3/pkg/config/constant"
	coreumclitestutil "github.com/CoreumFoundation/coreum/v3/testutil/cli"
	"github.com/CoreumFoundation/coreum/v3/testutil/network"
	"github.com/CoreumFoundation/coreum/v3/x/asset/nft/client/cli"
	"github.com/CoreumFoundation/coreum/v3/x/asset/nft/types"
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

	var classRes types.QueryClassResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryClass(), []string{classID}, &classRes))

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
	var classesRes types.QueryClassesResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryClasses(),
		[]string{fmt.Sprintf("--%s", cli.IssuerFlag), testNetwork.Validators[0].Address.String(), "--output", "json"},
		&classesRes))
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
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxIssueClass(), args)
	requireT.NoError(err)

	classID := types.BuildClassID(symbol, validator.Address)
	args = []string{classID, "nft-1", "https://my-nft-meta.invalid/1", "9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9"}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxMint(), args)
	requireT.NoError(err)
}

func TestCmdTxBurn(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx

	args := []string{symbol, "class name", "class description", "https://my-class-meta.invalid/1", "content-hash"}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxIssueClass(), args)
	requireT.NoError(err)

	classID := types.BuildClassID(symbol, validator.Address)
	args = []string{classID, "nft-1", "https://my-nft-meta.invalid/1", "9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9"}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxMint(), args)
	requireT.NoError(err)

	args = []string{classID, "nft-1"}
	args = append(args, txValidator1Args(testNetwork)...)

	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxBurn(), args)
	requireT.NoError(err)

	var resp types.QueryBurntNFTResponse
	args = []string{classID, "nft-1", "--output", "json"}
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryBurnt(), args, &resp))

	requireT.True(resp.Burnt)

	args = []string{classID, "--output", "json"}
	var respList types.QueryBurntNFTsInClassResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryBurnt(), args, &respList))
	requireT.Len(respList.NftIds, 1)
}

func TestCmdQueryParams(t *testing.T) {
	requireT := require.New(t)

	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx

	var resp types.QueryParamsResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryParams(), []string{}, &resp))
	expectedMintFee := sdk.Coin{Denom: constant.DenomDev, Amount: sdkmath.NewInt(0)}
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
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxMint(), args)
	requireT.NoError(err)
}

func issueClass(
	requireT *require.Assertions,
	ctx client.Context,
	symbol, name, description, url, urlHash string, //nolint:unparam
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
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxIssueClass(), args)
	requireT.NoError(err)

	return types.BuildClassID(symbol, validator.Address)
}
