package cli_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/nft"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/cmd/cored/cosmoscmd"
	"github.com/CoreumFoundation/coreum/v4/pkg/config/constant"
	coreumclitestutil "github.com/CoreumFoundation/coreum/v4/testutil/cli"
	"github.com/CoreumFoundation/coreum/v4/testutil/network"
	"github.com/CoreumFoundation/coreum/v4/x/asset/nft/client/cli"
	"github.com/CoreumFoundation/coreum/v4/x/asset/nft/types"
)

func TestQueryClassAndNFT(t *testing.T) {
	requireT := require.New(t)

	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4] //nolint:goconst
	name := "class name"
	description := "class description"
	uri := "https://my-class-meta.invalid/1"
	uriHash := "content-hash"
	ctx := testNetwork.Validators[0].ClientCtx

	classID := issueClass(
		t, ctx,
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
		Data: &codectypes.Any{
			TypeUrl: "/coreum.asset.nft.v1.DataBytes",
			Value:   []byte{0xa, 0x2, 0x11, 0x12},
		},
		Features: []types.ClassFeature{
			types.ClassFeature_burning,
			types.ClassFeature_disable_sending,
		},
		RoyaltyRate: sdkmath.LegacyMustNewDecFromStr("0.1"),
	}

	requireT.Equal(expectedClass, classRes.Class)

	// classes
	var classesRes types.QueryClassesResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryClasses(),
		[]string{fmt.Sprintf("--%s", cli.IssuerFlag), testNetwork.Validators[0].Address.String(), "--output", "json"},
		&classesRes))
	requireT.Equal(expectedClass, classesRes.Classes[0])

	mint(
		t,
		ctx,
		classID,
		nftID,
		"https://my-nft-meta.invalid/1",
		"9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9",
		testNetwork,
	)

	expectedNFT := nft.NFT{
		ClassId: classID,
		Id:      nftID,
		Uri:     "https://my-nft-meta.invalid/1",
		UriHash: "9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9",
		Data: &codectypes.Any{
			TypeUrl: "/coreum.asset.nft.v1.DataBytes",
			Value:   []byte{0xa, 0x2, 0x21, 0x22},
		},
	}

	var nftRes nft.QueryNFTResponse
	//TODO(fix-auto-cli)
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cosmoscmd.NewRootCmd(), []string{classID, nftID}, &nftRes))

	requireT.Equal(&expectedNFT, nftRes.Nft)
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
	t *testing.T,
	ctx client.Context,
	classID, nftID, uri, uriHash string,
	testNetwork *network.Network,
) {
	data := []byte{0x21, 0x22}
	dataFile := filepath.Join(t.TempDir(), "data")
	require.NoError(t, os.WriteFile(dataFile, data, 0o600))

	args := []string{
		classID, nftID,
		fmt.Sprintf("--%s=%s", cli.URIFlag, uri),
		fmt.Sprintf("--%s=%s", cli.URIHashFlag, uriHash),
		fmt.Sprintf("--%s=%s", cli.DataFileFlag, dataFile),
	}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxMint(), args)
	require.NoError(t, err)
}

func issueClass(
	t *testing.T,
	ctx client.Context,
	symbol, name, description, uri, uriHash string,
	testNetwork *network.Network,
	royaltyRate string,
	features ...types.ClassFeature,
) string {
	data := []byte{0x11, 0x12}
	dataFile := filepath.Join(t.TempDir(), "data")
	require.NoError(t, os.WriteFile(dataFile, data, 0o600))

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
		fmt.Sprintf("--%s=%s", cli.DataFileFlag, dataFile),
	}
	args = append(args, txValidator1Args(testNetwork)...)
	if royaltyRate != "" {
		args = append(args, fmt.Sprintf("--%s", cli.RoyaltyRateFlag), royaltyRate)
	}
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxIssueClass(), args)
	require.NoError(t, err)

	return types.BuildClassID(symbol, validator.Address)
}
