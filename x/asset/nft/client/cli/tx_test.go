package cli_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	cosmoscli "github.com/cosmos/cosmos-sdk/x/nft/client/cli"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	coreumclitestutil "github.com/CoreumFoundation/coreum/v3/testutil/cli"
	"github.com/CoreumFoundation/coreum/v3/testutil/network"
	"github.com/CoreumFoundation/coreum/v3/x/asset/nft/client/cli"
	"github.com/CoreumFoundation/coreum/v3/x/asset/nft/types"
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
		"https://my-class-meta.invalid/1",
		"content-hash",
		fmt.Sprintf("--%s=%s", cli.FeaturesFlag, types.ClassFeature_burning.String()),
	}
	args = append(args, txValidator1Args(testNetwork)...)
	res, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxIssueClass(), args)
	requireT.NoError(err)

	requireT.NotEmpty(res.TxHash)
	requireT.Equal(uint32(0), res.Code, "can't submit IssueClass tx", res)
}

func TestCmdMintToRecipient(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx

	// create class
	classID := issueClass(
		requireT,
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
	args := []string{classID, nftID, "", "", "--recipient", recipient.String()}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxMint(), args)
	requireT.NoError(err)

	// query recipient

	var resp nft.QueryOwnerResponse
	args = []string{classID, nftID}
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cosmoscli.GetCmdQueryOwner(), args, &resp))
	requireT.Equal(recipient.String(), resp.Owner)
}

func TestCmdFreeze(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	symbol := "nft" + uuid.NewString()[:4]
	validator := testNetwork.Validators[0]
	ctx := validator.ClientCtx

	// create class
	classID := issueClass(
		requireT,
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
		requireT,
		ctx,
		classID,
		nftID,
		"https://my-nft-meta.invalid/1",
		"9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9",
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
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryFrozen(), args, &frozenResp))

	// unfreeze
	args = []string{classID, nftID}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxUnfreeze(), args)
	requireT.NoError(err)

	// query frozen
	args = []string{classID, nftID}
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryFrozen(), args, &frozenResp))
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
		requireT,
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
		requireT,
		ctx,
		classID,
		nftID,
		"https://my-nft-meta.invalid/1",
		"9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9",
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
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryWhitelisted(), args, &whitelistedResp))
	requireT.True(whitelistedResp.Whitelisted)

	// query with pagination
	var resPage types.QueryWhitelistedAccountsForNFTResponse
	args = []string{classID, nftID}
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryWhitelistedAccounts(), args, &resPage))
	requireT.ElementsMatch([]string{account.String()}, resPage.Accounts)

	// unwhitelist
	args = []string{classID, nftID, account.String()}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxUnwhitelist(), args)
	requireT.NoError(err)

	// query whitelisted
	args = []string{classID, nftID, account.String()}
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryWhitelisted(), args, &whitelistedResp))
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
		requireT,
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
		requireT,
		ctx,
		classID,
		nftID,
		"https://my-nft-meta.invalid/1",
		"9309e7e6e96150afbf181d308fe88343ab1cbec391b7717150a7fb217b4cf0a9",
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
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryWhitelisted(), args, &whitelistedResp))
	requireT.True(whitelistedResp.Whitelisted)

	// query with pagination
	var resPage types.QueryClassWhitelistedAccountsResponse
	args = []string{classID}
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryClassWhitelistedAccounts(), args, &resPage))
	requireT.ElementsMatch([]string{account.String()}, resPage.Accounts)

	// unwhitelist
	args = []string{classID, account.String()}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxClassUnwhitelist(), args)
	requireT.NoError(err)

	// query whitelisted
	args = []string{classID, nftID, account.String()}
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryWhitelisted(), args, &whitelistedResp))
	requireT.False(whitelistedResp.Whitelisted)
}

func txValidator1Args(testNetwork *network.Network) []string {
	return []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, testNetwork.Validators[0].Address.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(testNetwork.Config.BondDenom, sdkmath.NewInt(1000000))).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}
}
