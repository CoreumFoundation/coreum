package cosmoscmd

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cosmos/cosmos-sdk/client/flags"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmostx "github.com/cosmos/cosmos-sdk/types/tx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/app"
	coreumclitestutil "github.com/CoreumFoundation/coreum/v6/testutil/cli"
	"github.com/CoreumFoundation/coreum/v6/testutil/network"
)

func TestAutoGasPrices(t *testing.T) {
	testNetwork := network.New(t)
	ctx := testNetwork.Validators[0].ClientCtx
	denom := testNetwork.Config.BondDenom

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
			name:  "specific gas prices are provided",
			flags: []string{"--gas-prices=0.1" + denom, "--gas=115000"},
			feeAssertion: func(t *testing.T, fee sdk.Coins) {
				assert.True(t, fee.Equal(sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(11500)))))
			},
		},
		{
			name:  "specific fees are provided",
			flags: []string{"--fees=12345" + denom},
			feeAssertion: func(t *testing.T, fee sdk.Coins) {
				assert.True(t, fee.Equal(sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(12345)))))
			},
		},
		{
			name:        "both gas prices and fees are provided",
			flags:       []string{"--fees=12345" + denom, "--gas-prices=auto"},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requireT := require.New(t)
			recipient := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
			args := append([]string{
				"send", testNetwork.Validators[0].Address.String(), recipient.String(), "100" + denom,
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			}, tc.flags...)
			bankTx := bankcli.NewTxCmd(addresscodec.NewBech32Codec(app.ChosenNetwork.Provider.GetAddressPrefix()))
			addQueryGasPriceToAllLeafs(bankTx)

			res, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, bankTx, args)
			if tc.expectError {
				requireT.Error(err)
				return
			}
			requireT.NoError(err)

			txQuery, err := authtx.QueryTx(ctx, res.TxHash)
			requireT.NoError(err)
			tx := txQuery.Tx.GetCachedValue().(*cosmostx.Tx)
			tc.feeAssertion(t, tx.GetFee())
		})
	}
}
