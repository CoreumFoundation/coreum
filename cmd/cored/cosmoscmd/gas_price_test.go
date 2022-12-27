package cosmoscmd

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmostx "github.com/cosmos/cosmos-sdk/types/tx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/CoreumFoundation/coreum/testutil/network"
)

func TestAutoGasPrices(t *testing.T) {
	testNetwork := network.New(t)
	ctx := testNetwork.Validators[0].ClientCtx

	testCases := []struct {
		name         string
		flags        []string
		feeAssertion func(t *testing.T, fee sdk.Coins)
		expectError  bool
	}{
		{
			name:  "no flags set",
			flags: []string{},
			feeAssertion: func(t *testing.T, fee sdk.Coins) {
				assert.False(t, fee.IsZero())
			},
		},
		{
			name:  "auto flag provided",
			flags: []string{"--gas-prices=auto"},
			feeAssertion: func(t *testing.T, fee sdk.Coins) {
				assert.False(t, fee.IsZero())
			},
		},
		{
			name:  "specific gas prices is provided",
			flags: []string{"--gas-prices=0.1ducore", "--gas=100000"},
			feeAssertion: func(t *testing.T, fee sdk.Coins) {
				assert.True(t, fee.IsEqual(sdk.NewCoins(sdk.NewCoin("ducore", sdk.NewInt(10000)))))
			},
		},
		{
			name:  "specific fees are provided",
			flags: []string{"--fees=12345ducore"},
			feeAssertion: func(t *testing.T, fee sdk.Coins) {
				assert.True(t, fee.IsEqual(sdk.NewCoins(sdk.NewCoin("ducore", sdk.NewInt(12345)))))
			},
		},
		{
			name:        "both gas prices and fees are provided",
			flags:       []string{"--fees=12345ducore", "--gas-prices=auto"},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assertT := assert.New(t)
			recipient := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
			args := append([]string{
				"send", testNetwork.Validators[0].Address.String(), recipient.String(), "100ducore",
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			}, tc.flags...)
			bankTx := bankcli.NewTxCmd()
			addQueryGasPriceToAllLeafs(bankTx)

			bufWriter, err := clitestutil.ExecTestCLICmd(ctx, bankTx, args)
			if tc.expectError {
				assertT.Error(err)
				return
			}
			assertT.NoError(err)

			txRes := sdk.TxResponse{}
			bts := bufWriter.Bytes()
			err = ctx.Codec.UnmarshalJSON(bts, &txRes)
			assertT.NoError(err)

			txQuery, err := authtx.QueryTx(ctx, txRes.TxHash)
			assertT.NoError(err)
			tx := txQuery.Tx.GetCachedValue().(*cosmostx.Tx)
			tc.feeAssertion(t, tx.GetFee())
		})
	}
}
