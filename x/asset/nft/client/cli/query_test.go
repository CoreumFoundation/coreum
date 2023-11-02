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
	uri := "https://my-class-meta.invalid/1"
	uriHash := "content-hash"
	ctx := testNetwork.Validators[0].ClientCtx

	classID := issueClass(
		requireT, ctx,
		symbol, name, description, uri, uriHash,
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
		URI:         uri,
		URIHash:     uriHash,
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

func TestCmdQueryParams(t *testing.T) {
	requireT := require.New(t)

	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx

	var resp types.QueryParamsResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryParams(), []string{}, &resp))
	expectedMintFee := sdk.Coin{Denom: constant.DenomDev, Amount: sdkmath.NewInt(0)}
	requireT.Equal(expectedMintFee, resp.Params.MintFee)
}

//nolint:unparam // using constant values here will make this function less flexible.
func mint(
	requireT *require.Assertions,
	ctx client.Context,
	classID, nftID, uri, uriHash string,
	testNetwork *network.Network,
) {
	args := []string{
		classID, nftID,
		fmt.Sprintf("--%s=%s", cli.URIFlag, uri),
		fmt.Sprintf("--%s=%s", cli.URIHashFlag, uriHash),
	}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxMint(), args)
	requireT.NoError(err)
}

func issueClass(
	requireT *require.Assertions,
	ctx client.Context,
	symbol, name, description, uri, uriHash string, //nolint:unparam
	testNetwork *network.Network,
	royaltyRate string,
	features ...types.ClassFeature,
) string {
	featuresStringList := lo.Map(features, func(s types.ClassFeature, _ int) string {
		return s.String()
	})
	featuresString := strings.Join(featuresStringList, ",")
	validator := testNetwork.Validators[0]
	args := []string{
		symbol,
		name,
		description,
		fmt.Sprintf("--%s=%s", cli.FeaturesFlag, featuresString),
		fmt.Sprintf("--%s=%s", cli.URIFlag, uri),
		fmt.Sprintf("--%s=%s", cli.URIHashFlag, uriHash),
	}
	args = append(args, txValidator1Args(testNetwork)...)
	if royaltyRate != "" {
		args = append(args, fmt.Sprintf("--%s", cli.RoyaltyRateFlag), royaltyRate)
	}
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxIssueClass(), args)
	requireT.NoError(err)

	return types.BuildClassID(symbol, validator.Address)
}
