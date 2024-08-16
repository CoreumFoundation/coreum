package cli_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	sdkmath "cosmossdk.io/math"
	nfttypes "cosmossdk.io/x/nft"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	coreumclitestutil "github.com/CoreumFoundation/coreum/v4/testutil/cli"
	"github.com/CoreumFoundation/coreum/v4/testutil/network"
	"github.com/CoreumFoundation/coreum/v4/x/asset/nft/client/cli"
	"github.com/CoreumFoundation/coreum/v4/x/asset/nft/types"
)

const nftID = "nft-1"

func TestCmdTxIssueClass(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx

	args := []string{
		symbol,
		"class name",
		"class description",
		fmt.Sprintf("--%s=%s", cli.FeaturesFlag, types.ClassFeature_burning.String()),
		fmt.Sprintf("--%s=%s", cli.URIFlag, "https://my-class-meta.invalid/1"),
		fmt.Sprintf("--%s=%s", cli.URIHashFlag, "content-hash"),
	}
	args = append(args, txValidator1Args(testNetwork)...)
	res, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxIssueClass(), args)
	requireT.NoError(err)

	requireT.NotEmpty(res.TxHash)
	requireT.Equal(uint32(0), res.Code, "can't submit IssueClass tx", res)
}

func TestCmdTxMint(t *testing.T) {
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx

	classID := issueClass(
		t, ctx,
		symbol,
		"class name",
		"class description",
		"https://my-class-meta.invalid/1", "content-hash",
		testNetwork, "",
	)

	mint(
		t,
		ctx,
		classID,
		nftID,
		"https://my-nft-meta.invalid/1",
		"9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9",
		"",
		testNetwork,
	)
}

func TestCmdTxBurn(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx

	classID := issueClass(
		t, ctx,
		symbol,
		"class name",
		"class description",
		"https://my-class-meta.invalid/1", "content-hash",
		testNetwork, "",
	)

	mint(
		t,
		ctx,
		classID,
		"nft-1",
		"https://my-nft-meta.invalid/1",
		"9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9",
		"",
		testNetwork,
	)

	args := []string{classID, "nft-1"}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxBurn(), args)
	requireT.NoError(err)

	var resp types.QueryBurntNFTResponse
	args = []string{classID, "nft-1", "--output", "json"}
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryBurnt(), args, &resp)

	requireT.True(resp.Burnt)

	args = []string{classID, "--output", "json"}
	var respList types.QueryBurntNFTsInClassResponse
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryBurnt(), args, &respList)
	requireT.Len(respList.NftIds, 1)
}

func TestCmdMintToRecipient(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx

	// create class
	classID := issueClass(
		t,
		ctx,
		symbol,
		"class name",
		"class description",
		"https://my-class-meta.invalid/1",
		"",
		testNetwork,
		"0.0",
		types.ClassFeature_freezing,
	)
	// mint nft
	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	args := []string{
		classID,
		nftID,
		fmt.Sprintf("--%s=%s", cli.RecipientFlag, recipient.String()),
		fmt.Sprintf("--%s=%s", cli.URIFlag, "https://my-class-meta.invalid/1"),
		fmt.Sprintf("--%s=%s", cli.URIHashFlag, "content-hash"),
	}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxMint(), args)
	requireT.NoError(err)

	// query recipient
	var resp nfttypes.QueryOwnerResponse
	coreumclitestutil.ExecRootQueryCmd(t, ctx, []string{nfttypes.ModuleName, "owner", classID, nftID}, &resp)
	requireT.Equal(recipient.String(), resp.Owner)
}

func TestCmdMintDataDynamic(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx

	classID := issueClass(
		t,
		ctx,
		symbol,
		"name",
		"description",
		"",
		"",
		testNetwork,
		"0.0",
	)

	items := []types.DataDynamicItem{
		{
			Editors: nil,
			Data:    []byte("my-data"),
		},
	}
	data, err := json.Marshal(items)
	require.NoError(t, err)

	dataFile := filepath.Join(t.TempDir(), "data")
	require.NoError(t, os.WriteFile(dataFile, data, 0o600))

	args := []string{
		classID,
		nftID,
		fmt.Sprintf("--%s=%s", cli.DataFileFlag, dataFile),
		fmt.Sprintf("--%s=%s", cli.DataTypeFlag, cli.DataTypeDynamic),
	}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxMint(), args)
	requireT.NoError(err)

	// query stored
	var nftRes nfttypes.QueryNFTResponse
	coreumclitestutil.ExecRootQueryCmd(t, ctx, []string{nfttypes.ModuleName, "nft", classID, nftID}, &nftRes)
	var gotDataDynamic types.DataDynamic
	decodeAnyDataFromAmino(t, ctx, nftRes.Nft.Data, &gotDataDynamic)
	requireT.NoError(gotDataDynamic.Unmarshal(nftRes.Nft.Data.Value))
	requireT.Equal(items, gotDataDynamic.Items)
}

func TestCmdUpdateData(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx

	classID := issueClass(
		t,
		ctx,
		symbol,
		"name",
		"description",
		"",
		"",
		testNetwork,
		"0.0",
	)

	dataDynamicItems := []types.DataDynamicItem{
		{
			Editors: []types.DataEditor{
				types.DataEditor_admin,
				types.DataEditor_owner,
			},
			Data: []byte("initial-data"),
		},
	}
	dataDynamic, err := json.Marshal(dataDynamicItems)
	require.NoError(t, err)

	initialDataFile := filepath.Join(t.TempDir(), "initial-data")
	require.NoError(t, os.WriteFile(initialDataFile, dataDynamic, 0o600))

	args := []string{
		classID,
		nftID,
		fmt.Sprintf("--%s=%s", cli.DataFileFlag, initialDataFile),
		fmt.Sprintf("--%s=%s", cli.DataTypeFlag, cli.DataTypeDynamic),
	}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxMint(), args)
	requireT.NoError(err)

	// update the data

	dataDynamicIndexedItems := []types.DataDynamicIndexedItem{
		{
			Index: 0,
			Data:  []byte("new-data"),
		},
	}
	dataToUpdate, err := json.Marshal(dataDynamicIndexedItems)
	require.NoError(t, err)

	updateDataFile := filepath.Join(t.TempDir(), "update")
	require.NoError(t, os.WriteFile(updateDataFile, dataToUpdate, 0o600))

	args = []string{
		classID,
		nftID,
		fmt.Sprintf("--%s=%s", cli.DataFileFlag, updateDataFile),
	}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxUpdateData(), args)
	requireT.NoError(err)

	// query stored

	var nftRes nfttypes.QueryNFTResponse
	coreumclitestutil.ExecRootQueryCmd(t, ctx, []string{nfttypes.ModuleName, "nft", classID, nftID}, &nftRes)

	var gotDataDynamic types.DataDynamic
	decodeAnyDataFromAmino(t, ctx, nftRes.Nft.Data, &gotDataDynamic)
	requireT.NoError(gotDataDynamic.Unmarshal(nftRes.Nft.Data.Value))
	requireT.Equal(string(dataDynamicIndexedItems[0].Data), string(gotDataDynamic.Items[0].Data))
}

func TestCmdFreeze(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx

	// create class
	classID := issueClass(
		t,
		ctx,
		symbol,
		"class name",
		"class description",
		"https://my-class-meta.invalid/1",
		"",
		testNetwork,
		"0.0",
		types.ClassFeature_freezing,
	)
	// mint nft
	mint(
		t,
		ctx,
		classID,
		nftID,
		"https://my-nft-meta.invalid/1",
		"9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9",
		"",
		testNetwork,
	)

	// freeze
	args := []string{classID, nftID}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxFreeze(), args)
	requireT.NoError(err)

	// query frozen
	var frozenResp types.QueryFrozenResponse
	args = []string{classID, nftID}
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryFrozen(), args, &frozenResp)

	// unfreeze
	args = []string{classID, nftID}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxUnfreeze(), args)
	requireT.NoError(err)

	// query frozen
	args = []string{classID, nftID}
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryFrozen(), args, &frozenResp)
	requireT.False(frozenResp.Frozen)
}

func TestCmdWhitelist(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx
	account := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// create class
	classID := issueClass(
		t,
		ctx,
		symbol,
		"class name",
		"class description",
		"https://my-class-meta.invalid/1",
		"",
		testNetwork,
		"0",
		types.ClassFeature_whitelisting,
	)
	// mint nft
	mint(
		t,
		ctx,
		classID,
		nftID,
		"https://my-nft-meta.invalid/1",
		"9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9",
		"",
		testNetwork,
	)

	// whitelist
	args := []string{classID, nftID, account.String()}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxWhitelist(), args)
	requireT.NoError(err)

	// query whitelisted
	var whitelistedResp types.QueryWhitelistedResponse
	args = []string{classID, nftID, account.String()}
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryWhitelisted(), args, &whitelistedResp)
	requireT.True(whitelistedResp.Whitelisted)

	// query with pagination
	var resPage types.QueryWhitelistedAccountsForNFTResponse
	args = []string{classID, nftID}
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryWhitelistedAccounts(), args, &resPage)
	requireT.ElementsMatch([]string{account.String()}, resPage.Accounts)

	// unwhitelist
	args = []string{classID, nftID, account.String()}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxUnwhitelist(), args)
	requireT.NoError(err)

	// query whitelisted
	args = []string{classID, nftID, account.String()}
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryWhitelisted(), args, &whitelistedResp)
	requireT.False(whitelistedResp.Whitelisted)
}

func TestCmdClassWhitelist(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx
	account := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// create class
	classID := issueClass(
		t,
		ctx,
		symbol,
		"class name",
		"class description",
		"https://my-class-meta.invalid/1",
		"",
		testNetwork,
		"0",
		types.ClassFeature_whitelisting,
	)
	// mint nft
	nftID := "nft"
	mint(
		t,
		ctx,
		classID,
		nftID,
		"https://my-nft-meta.invalid/1",
		"9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9",
		"",
		testNetwork,
	)

	// whitelist
	args := []string{classID, account.String()}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxClassWhitelist(), args)
	requireT.NoError(err)

	// query whitelisted
	var whitelistedResp types.QueryWhitelistedResponse
	args = []string{classID, nftID, account.String()}
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryWhitelisted(), args, &whitelistedResp)
	requireT.True(whitelistedResp.Whitelisted)

	// query with pagination
	var resPage types.QueryClassWhitelistedAccountsResponse
	args = []string{classID}
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryClassWhitelistedAccounts(), args, &resPage)
	requireT.ElementsMatch([]string{account.String()}, resPage.Accounts)

	// unwhitelist
	args = []string{classID, account.String()}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxClassUnwhitelist(), args)
	requireT.NoError(err)

	// query whitelisted
	args = []string{classID, nftID, account.String()}
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryWhitelisted(), args, &whitelistedResp)
	requireT.False(whitelistedResp.Whitelisted)
}

func TestCmdClassFreeze(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx
	account := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// create class
	classID := issueClass(
		t,
		ctx,
		symbol,
		"class name",
		"class description",
		"https://my-class-meta.invalid/1",
		"",
		testNetwork,
		"0.0",
		types.ClassFeature_freezing,
	)
	// mint nft
	mint(
		t,
		ctx,
		classID,
		nftID,
		"https://my-nft-meta.invalid/1",
		"9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9",
		"",
		testNetwork,
	)

	// class-freeze
	args := []string{classID, account.String()}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxClassFreeze(), args)
	requireT.NoError(err)

	// query class frozen
	var classFrozenResp types.QueryFrozenResponse
	args = []string{classID, account.String()}
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryClassFrozen(), args, &classFrozenResp)
	requireT.True(classFrozenResp.Frozen)

	// query frozen
	var frozenResp types.QueryFrozenResponse
	args = []string{classID, nftID}
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryFrozen(), args, &frozenResp)
	requireT.False(frozenResp.Frozen)

	// query with pagination
	var resPage types.QueryClassFrozenAccountsResponse
	args = []string{classID}
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryClassFrozenAccounts(), args, &resPage)
	requireT.ElementsMatch([]string{account.String()}, resPage.Accounts)

	// unfreeze
	args = []string{classID, account.String()}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxClassUnfreeze(), args)
	requireT.NoError(err)

	// query class frozen
	args = []string{classID, account.String()}
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryClassFrozen(), args, &classFrozenResp)
	requireT.False(classFrozenResp.Frozen)
}

func txValidator1Args(testNetwork *network.Network) []string {
	return []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, testNetwork.Validators[0].Address.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(
			sdk.NewCoin(testNetwork.Config.BondDenom, sdkmath.NewInt(1000000)),
		).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}
}
