package cli_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/asset/client/cli"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestBurnRate(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)
	// the denom must start from the letter
	symbol := "abc"
	ctx := testNetwork.Validators[0].ClientCtx
	issuer := testNetwork.Validators[0].Address
	recipient1, err := createAccount(ctx)
	requireT.NoError(err)
	recipient2, err := createAccount(ctx)
	requireT.NoError(err)
	denom := types.BuildFungibleTokenDenom(symbol, issuer)

	// Issue token
	args := []string{symbol, testNetwork.Validators[0].Address.String(), "777", `"My Token"`, "--burn_rate", "0.5"}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssueFungibleToken(), args)
	requireT.NoError(err)

	// send coins from issuer to recipient (no burn applied)
	tokens := sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(400)),
		sdk.NewCoin("ducore", sdk.NewInt(1000_000)),
	)
	args = append([]string{issuer.String(), recipient1.String(), tokens.String(), "--output", "json"}, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, bankcli.NewSendTxCmd(), args)
	requireT.NoError(err)

	assertBalance(ctx, t, recipient1, sdk.NewCoin(denom, sdk.NewInt(400)))
	assertBalance(ctx, t, issuer, sdk.NewCoin(denom, sdk.NewInt(377)))
	assertTotalSupply(ctx, t, sdk.NewCoin(denom, sdk.NewInt(777)))

	// send coins from recipient to another address (burn applied)
	args = []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, recipient1.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(testNetwork.Config.BondDenom, sdk.NewInt(1000000))).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}
	token := sdk.NewCoin(denom, sdk.NewInt(100))
	args = append([]string{recipient1.String(), recipient2.String(), token.String(), "--output", "json"}, args...)
	_, err = clitestutil.ExecTestCLICmd(ctx, bankcli.NewSendTxCmd(), args)
	requireT.NoError(err)

	assertBalance(ctx, t, recipient2, token)
	assertBalance(ctx, t, recipient1, sdk.NewCoin(denom, sdk.NewInt(250)))
	assertTotalSupply(ctx, t, sdk.NewCoin(denom, sdk.NewInt(727)))
}

func assertBalance(ctx client.Context, t assert.TestingT, addr sdk.AccAddress, coin sdk.Coin) {
	assertT := assert.New(t)
	var balanceRsp banktypes.QueryAllBalancesResponse
	buf, err := clitestutil.ExecTestCLICmd(ctx, bankcli.GetBalancesCmd(), []string{addr.String(), "--output", "json"})
	assertT.NoError(err)
	assertT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &balanceRsp))
	assertT.Equal(coin.Amount.String(), balanceRsp.Balances.AmountOf(coin.Denom).String())
}

func assertTotalSupply(ctx client.Context, t assert.TestingT, coin sdk.Coin) {
	assertT := assert.New(t)
	var supplyRsp sdk.Coin
	buf, err := clitestutil.ExecTestCLICmd(ctx, bankcli.GetCmdQueryTotalSupply(), []string{"--denom", coin.Denom, "--output", "json"})
	assertT.NoError(err)
	assertT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &supplyRsp))
	assertT.Equal(coin.String(), supplyRsp.String())
}

func createAccount(ctx client.Context) (sdk.AccAddress, error) {
	hdPath := sdk.GetConfig().GetFullBIP44Path()
	keyInfo, _, err := ctx.Keyring.NewMnemonic(uuid.NewString(), keyring.English, hdPath, "", hd.Secp256k1)
	if err != nil {
		return nil, err
	}

	return keyInfo.GetAddress(), nil
}
