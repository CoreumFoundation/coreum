package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

//nolint:funlen // this is complex test scenario and breaking it down is not helpful
func TestKeeper_FreezeWhitelistMultiSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	issuer2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings1 := types.IssueSettings{
		Issuer:        issuer1,
		Symbol:        "DEF",
		Subunit:       "def",
		Description:   "DEF Desc",
		InitialAmount: sdk.NewInt(1000),
		Features:      []types.TokenFeature{types.TokenFeature_freeze}, //nolint:nosnakecase
	}

	settings2 := types.IssueSettings{
		Issuer:        issuer2,
		Symbol:        "DEF",
		Subunit:       "def",
		Description:   "DEF Desc",
		InitialAmount: sdk.NewInt(2000),
		Features:      []types.TokenFeature{types.TokenFeature_whitelist}, //nolint:nosnakecase
	}

	bondDenom := testApp.StakingKeeper.BondDenom(ctx)
	// fund with the native coin
	err := testApp.FundAccount(ctx, issuer1, sdk.NewCoins(sdk.NewCoin(bondDenom, sdk.NewInt(1000))))
	requireT.NoError(err)

	denom1, err := ftKeeper.Issue(ctx, settings1)
	requireT.NoError(err)

	denom2, err := ftKeeper.Issue(ctx, settings2)
	requireT.NoError(err)

	recipient1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// freeze denom1 partially on the recipient1
	err = ftKeeper.Freeze(ctx, issuer1, recipient1, sdk.NewCoin(denom1, sdk.NewInt(10)))
	requireT.NoError(err)

	// whitelist denom2 partially on the recipient2
	err = ftKeeper.SetWhitelistedBalance(ctx, issuer2, recipient2, sdk.NewCoin(denom2, sdk.NewInt(10)))
	requireT.NoError(err)

	// multi-send valid amount
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{
			{Address: issuer1.String(), Coins: sdk.NewCoins(
				sdk.NewCoin(denom1, sdk.NewInt(15)),
				sdk.NewCoin(bondDenom, sdk.NewInt(20)),
			)},
			{Address: issuer2.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom2, sdk.NewInt(10)))},
		},
		[]banktypes.Output{
			// the recipient1 has frozen balance so that amount can be received
			{Address: recipient1.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom1, sdk.NewInt(15)))},
			// the recipient2 has whitelisted balance so that is the max amount recipient2 can receive
			{Address: recipient2.String(), Coins: sdk.NewCoins(
				sdk.NewCoin(denom2, sdk.NewInt(10)),
				sdk.NewCoin(bondDenom, sdk.NewInt(20)),
			)},
		})
	requireT.NoError(err)

	balance := bankKeeper.GetBalance(ctx, recipient1, denom1)
	requireT.Equal(sdk.NewCoin(denom1, sdk.NewInt(15)).String(), balance.String())
	balance = bankKeeper.GetBalance(ctx, recipient2, denom2)
	requireT.Equal(sdk.NewCoin(denom2, sdk.NewInt(10)).String(), balance.String())
	balance = bankKeeper.GetBalance(ctx, recipient2, bondDenom)
	requireT.Equal(sdk.NewCoin(bondDenom, sdk.NewInt(20)).String(), balance.String())

	// multi-send invalid frozen amount
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{
			// we can't return 15 coins since 10 are frozen
			{Address: recipient1.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom1, sdk.NewInt(15)))},
			{Address: recipient2.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom2, sdk.NewInt(10)))},
		},
		[]banktypes.Output{
			{Address: issuer1.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom1, sdk.NewInt(15)))},
			{Address: issuer2.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom2, sdk.NewInt(10)))},
		})
	requireT.ErrorIs(sdkerrors.ErrInsufficientFunds, err)

	// multi-send invalid whitelisted amount
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{
			{Address: issuer1.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom1, sdk.NewInt(15)))},
			{Address: issuer2.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom2, sdk.NewInt(15)))},
		},
		[]banktypes.Output{
			{Address: recipient1.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom1, sdk.NewInt(15)))},
			// the recipient2 has whitelisted 10 so can't receive 15
			{Address: recipient2.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom2, sdk.NewInt(15)))},
		})
	requireT.ErrorIs(types.ErrWhitelistedLimitExceeded, err)
}
