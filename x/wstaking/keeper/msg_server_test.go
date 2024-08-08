package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	customparamstypes "github.com/CoreumFoundation/coreum/v4/x/customparams/types"
)

func Test_WrappedMsgCreateValidatorHandler(t *testing.T) {
	simApp := simapp.New()

	// set min delegation param to 10k
	ctx := simApp.NewContext(false)
	minSelfDelegation := sdkmath.NewInt(10_000)
	require.NoError(t, simApp.CustomParamsKeeper.SetStakingParams(ctx, customparamstypes.StakingParams{
		MinSelfDelegation: minSelfDelegation,
	}))
	require.NoError(t, simApp.FinalizeBlock())

	// create new account
	accountAddress, privateKey := simApp.GenAccount(ctx)
	require.NoError(t, simApp.FinalizeBlock())

	// fund account
	bondDenom, err := simApp.StakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	balance := sdk.NewCoins(sdk.NewCoin(bondDenom, sdkmath.NewInt(100_000_000_000)))
	require.NoError(t, simApp.FundAccount(ctx, accountAddress, balance))
	require.NoError(t, simApp.FinalizeBlock())

	// create validator
	description := stakingtypes.Description{Moniker: "moniker"}
	selfDelegation := sdk.NewCoin(bondDenom, sdkmath.NewInt(10_000_000))
	commission := stakingtypes.CommissionRates{
		Rate:          sdkmath.LegacyZeroDec(),
		MaxRate:       sdkmath.LegacyZeroDec(),
		MaxChangeRate: sdkmath.LegacyZeroDec(),
	}

	feeAmt := sdk.NewCoin(bondDenom, sdkmath.NewInt(1_000_000))
	gas := uint64(300_000)

	// try to create with insufficient min self delegation
	createValidatorMsg, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(accountAddress).String(),
		ed25519.GenPrivKey().PubKey(),
		selfDelegation,
		description,
		commission,
		sdkmath.OneInt(),
	)
	require.NoError(t, err)
	_, _, err = simApp.SendTx(ctx, feeAmt, gas, privateKey, createValidatorMsg)
	require.Error(t, err)

	// try to create with min self delegation
	createValidatorMsg, err = stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(accountAddress).String(),
		ed25519.GenPrivKey().PubKey(),
		selfDelegation,
		description,
		commission,
		minSelfDelegation,
	)
	require.NoError(t, err)
	_, _, err = simApp.SendTx(ctx, feeAmt, gas, privateKey, createValidatorMsg)
	require.NoError(t, err)

	require.NoError(t, simApp.FinalizeBlock())
}
