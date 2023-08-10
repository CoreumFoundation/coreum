package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v2/testutil/simapp"
	customparamstypes "github.com/CoreumFoundation/coreum/v2/x/customparams/types"
)

func Test_WrappedMsgCreateValidatorHandler(t *testing.T) {
	simApp := simapp.New()

	// set min delegation param to 10k
	ctx := simApp.BeginNextBlock(time.Time{})
	minSelfDelegation := sdkmath.NewInt(10_000)
	simApp.CustomParamsKeeper.SetStakingParams(ctx, customparamstypes.StakingParams{
		MinSelfDelegation: minSelfDelegation,
	})
	simApp.EndBlockAndCommit(ctx)

	// create new account
	ctx = simApp.BeginNextBlock(time.Time{})
	accountAddress, privateKey := simApp.GenAccount(ctx)
	simApp.EndBlockAndCommit(ctx)

	// fund account
	ctx = simApp.BeginNextBlock(time.Time{})
	bondDenom := simApp.StakingKeeper.BondDenom(ctx)
	balance := sdk.NewCoins(sdk.NewCoin(bondDenom, sdkmath.NewInt(100_000_000_000)))
	require.NoError(t, simApp.FundAccount(ctx, accountAddress, balance))
	simApp.EndBlockAndCommit(ctx)

	// create validator
	ctx = simApp.BeginNextBlock(time.Time{})
	description := stakingtypes.Description{Moniker: "moniker"}
	selfDelegation := sdk.NewCoin(bondDenom, sdkmath.NewInt(10_000_000))
	commission := stakingtypes.CommissionRates{
		Rate:          sdk.ZeroDec(),
		MaxRate:       sdk.ZeroDec(),
		MaxChangeRate: sdk.ZeroDec(),
	}

	feeAmt := sdk.NewCoin(bondDenom, sdkmath.NewInt(1_000_000))
	gas := uint64(300_000)

	// try to create with insufficient min self delegation
	createValidatorMsg, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(accountAddress), ed25519.GenPrivKey().PubKey(), selfDelegation, description, commission, sdk.OneInt(),
	)
	require.NoError(t, err)
	_, _, err = simApp.SendTx(ctx, feeAmt, gas, privateKey, createValidatorMsg)
	require.Error(t, err)

	// try to create with min self delegation
	createValidatorMsg, err = stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(accountAddress), ed25519.GenPrivKey().PubKey(), selfDelegation, description, commission, minSelfDelegation,
	)
	require.NoError(t, err)
	_, _, err = simApp.SendTx(ctx, feeAmt, gas, privateKey, createValidatorMsg)
	require.NoError(t, err)

	simApp.EndBlockAndCommit(ctx)
}
