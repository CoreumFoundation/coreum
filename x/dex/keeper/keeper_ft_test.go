package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/docker/distribution/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	testcontracts "github.com/CoreumFoundation/coreum/v4/x/asset/ft/keeper/test-contracts"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func TestKeeper_PlaceOrderWithExtension(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	acc, _ := testApp.GenAccount(sdkCtx)
	issuer, _ := testApp.GenAccount(sdkCtx)

	// extension
	codeID, _, err := testApp.WasmPermissionedKeeper.Create(
		sdkCtx, issuer, testcontracts.AssetExtensionWasm, &wasmtypes.AllowEverybody,
	)
	require.NoError(t, err)
	settingsWithExtension := assetfttypes.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEFEXT",
		Subunit:       "defext",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 10),
		Features:      []assetfttypes.Feature{assetfttypes.Feature_extension},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
		},
	}
	denomWithExtension, err := testApp.AssetFTKeeper.Issue(sdkCtx, settingsWithExtension)
	require.NoError(t, err)

	order := types.Order{
		Creator:    acc.String(),
		Type:       types.ORDER_TYPE_LIMIT,
		ID:         uuid.Generate().String(),
		BaseDenom:  denomWithExtension,
		QuoteDenom: denom2,
		Price:      lo.ToPtr(types.MustNewPriceFromString("12e-1")),
		Quantity:   sdkmath.NewInt(10),
		Side:       types.SIDE_SELL,
		GoodTil: &types.GoodTil{
			GoodTilBlockHeight: 390,
		},
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	lockedBalance, err := order.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	require.NoError(t, testApp.BankKeeper.SendCoins(sdkCtx, issuer, acc, sdk.NewCoins(lockedBalance)))

	require.ErrorContains(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order), "not supported for the tokens with extensions")
}

func TestKeeper_PlaceOrderWithDEXBlockFeature(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	acc, _ := testApp.GenAccount(sdkCtx)
	issuer, _ := testApp.GenAccount(sdkCtx)

	settingsWithExtension := assetfttypes.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEFEXT",
		Subunit:       "defext",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 10),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_dex_block,
		},
	}
	denomWithExtension, err := testApp.AssetFTKeeper.Issue(sdkCtx, settingsWithExtension)
	require.NoError(t, err)

	order := types.Order{
		Creator:    acc.String(),
		Type:       types.ORDER_TYPE_LIMIT,
		ID:         uuid.Generate().String(),
		BaseDenom:  denomWithExtension,
		QuoteDenom: denom2,
		Price:      lo.ToPtr(types.MustNewPriceFromString("12e-1")),
		Quantity:   sdkmath.NewInt(10),
		Side:       types.SIDE_SELL,
		GoodTil: &types.GoodTil{
			GoodTilBlockHeight: 390,
		},
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	lockedBalance, err := order.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	require.NoError(t, testApp.BankKeeper.SendCoins(sdkCtx, issuer, acc, sdk.NewCoins(lockedBalance)))
	require.ErrorContains(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order), "locking coins for DEX disabled for")
}

func TestKeeper_PlaceOrderWithRestrictDEXFeature(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	acc, _ := testApp.GenAccount(sdkCtx)
	issuer, _ := testApp.GenAccount(sdkCtx)

	issuanceSettings := assetfttypes.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEFEXT",
		Subunit:       "defext",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 10),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_dex_whitelisted_denoms,
		},
		DEXSettings: &assetfttypes.DEXSettings{
			WhitelistedDenoms: []string{
				denom3,
			},
		},
	}
	denom, err := testApp.AssetFTKeeper.Issue(sdkCtx, issuanceSettings)
	require.NoError(t, err)

	orderReceiveDenom2 := types.Order{
		Creator:    acc.String(),
		Type:       types.ORDER_TYPE_LIMIT,
		ID:         uuid.Generate().String(),
		BaseDenom:  denom,
		QuoteDenom: denom2, // the denom2 is not allowed
		Price:      lo.ToPtr(types.MustNewPriceFromString("12e-1")),
		Quantity:   sdkmath.NewInt(10),
		Side:       types.SIDE_SELL,
		GoodTil: &types.GoodTil{
			GoodTilBlockHeight: 390,
		},
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	lockedBalance, err := orderReceiveDenom2.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	require.NoError(t, testApp.BankKeeper.SendCoins(sdkCtx, issuer, acc, sdk.NewCoins(lockedBalance)))
	require.ErrorContains(
		t, testApp.DEXKeeper.PlaceOrder(sdkCtx, orderReceiveDenom2), "denom denom2 not whitelisted",
	)

	orderReceiveDenom3 := types.Order{
		Creator:    acc.String(),
		Type:       types.ORDER_TYPE_LIMIT,
		ID:         uuid.Generate().String(),
		BaseDenom:  denom,
		QuoteDenom: denom3, // the denom3 is allowed
		Price:      lo.ToPtr(types.MustNewPriceFromString("7e-4")),
		Quantity:   sdkmath.NewInt(10),
		Side:       types.SIDE_SELL,
		GoodTil: &types.GoodTil{
			GoodTilBlockHeight: 390,
		},
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	lockedBalance, err = orderReceiveDenom2.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	require.NoError(t, testApp.BankKeeper.SendCoins(sdkCtx, issuer, acc, sdk.NewCoins(lockedBalance)))
	require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, orderReceiveDenom3))

	// now update settings to remove all limit and place orderReceiveDenom2
	require.NoError(t, testApp.AssetFTKeeper.UpdateDEXWhitelistedDenoms(sdkCtx, issuer, denom, nil))
	require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, orderReceiveDenom2))
}

func TestKeeper_PlaceOrderWithBurning(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	acc, _ := testApp.GenAccount(sdkCtx)
	issuer, _ := testApp.GenAccount(sdkCtx)

	settingsWithExtension := assetfttypes.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEFEXT",
		Subunit:       "defext",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 10),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_burning,
		},
	}
	denomWithBurn, err := testApp.AssetFTKeeper.Issue(sdkCtx, settingsWithExtension)
	require.NoError(t, err)

	order := types.Order{
		Creator:    acc.String(),
		Type:       types.ORDER_TYPE_LIMIT,
		ID:         uuid.Generate().String(),
		BaseDenom:  denomWithBurn,
		QuoteDenom: denom2,
		Price:      lo.ToPtr(types.MustNewPriceFromString("12e-1")),
		Quantity:   sdkmath.NewInt(10),
		Side:       types.SIDE_SELL,
		GoodTil: &types.GoodTil{
			GoodTilBlockHeight: 390,
		},
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	lockedBalance, err := order.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	require.NoError(t, testApp.BankKeeper.SendCoins(sdkCtx, issuer, acc, sdk.NewCoins(lockedBalance)))
	require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))
	require.ErrorContains(t, testApp.AssetFTKeeper.Burn(sdkCtx, acc, lockedBalance), "coins are not spendable")
}

func TestKeeper_PlaceOrderWithStaking(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	acc, _ := testApp.GenAccount(sdkCtx)
	validatorOwner, _ := testApp.GenAccount(sdkCtx)
	validatorOwner2, _ := testApp.GenAccount(sdkCtx)

	denomToStake := sdk.DefaultBondDenom

	require.NoError(t, testApp.FundAccount(sdkCtx, validatorOwner, sdk.NewCoins(sdk.NewInt64Coin(denomToStake, 10))))
	err := addValidator(sdkCtx, testApp.StakingKeeper, validatorOwner, sdk.NewInt64Coin(denomToStake, 10))
	require.NoError(t, err)
	val, err := testApp.StakingKeeper.GetValidators(sdkCtx, 1)
	require.NoError(t, err)

	order := types.Order{
		Creator:    acc.String(),
		Type:       types.ORDER_TYPE_LIMIT,
		ID:         uuid.Generate().String(),
		BaseDenom:  denomToStake,
		QuoteDenom: denom2,
		Price:      lo.ToPtr(types.MustNewPriceFromString("12e-1")),
		Quantity:   sdkmath.NewInt(10),
		Side:       types.SIDE_SELL,
		GoodTil: &types.GoodTil{
			GoodTilBlockHeight: 390,
		},
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	orderLockedBalance, err := order.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	require.NoError(t, testApp.FundAccount(sdkCtx, acc, sdk.NewCoins(orderLockedBalance)))

	_, err = testApp.StakingKeeper.Delegate(sdkCtx, acc, orderLockedBalance.Amount, stakingtypes.Unbonded, val[0], true)
	require.NoError(t, err)

	res, err := testApp.BankKeeper.Balance(sdkCtx, &banktypes.QueryBalanceRequest{
		Address: acc.String(),
		Denom:   denomToStake,
	})
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin(denomToStake, 0).String(), res.Balance.String())

	lockedBalance := testApp.AssetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, denomToStake)
	require.Equal(t, sdk.NewInt64Coin(denomToStake, 0).String(), lockedBalance.String())

	require.Error(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order), cosmoserrors.ErrInsufficientFunds)

	require.NoError(t, testApp.FundAccount(sdkCtx, acc, sdk.NewCoins(orderLockedBalance)))
	require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))

	res, err = testApp.BankKeeper.Balance(sdkCtx, &banktypes.QueryBalanceRequest{
		Address: acc.String(),
		Denom:   denomToStake,
	})
	require.NoError(t, err)
	require.Equal(t, orderLockedBalance.String(), res.Balance.String())

	lockedBalance = testApp.AssetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, denomToStake)
	require.Equal(t, orderLockedBalance.String(), lockedBalance.String())

	_, err = testApp.StakingKeeper.Delegate(sdkCtx, acc, orderLockedBalance.Amount, stakingtypes.Unbonded, val[0], true)
	require.Error(t, err, cosmoserrors.ErrInsufficientFunds)

	order = types.Order{
		Creator:    validatorOwner2.String(),
		Type:       types.ORDER_TYPE_LIMIT,
		ID:         uuid.Generate().String(),
		BaseDenom:  denomToStake,
		QuoteDenom: denom2,
		Price:      lo.ToPtr(types.MustNewPriceFromString("12e-1")),
		Quantity:   sdkmath.NewInt(10),
		Side:       types.SIDE_SELL,
		GoodTil: &types.GoodTil{
			GoodTilBlockHeight: 390,
		},
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	orderLockedBalance, err = order.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	require.NoError(t, testApp.FundAccount(sdkCtx, validatorOwner2, sdk.NewCoins(orderLockedBalance)))
	require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))
	err = addValidator(sdkCtx, testApp.StakingKeeper, validatorOwner2, orderLockedBalance)
	require.ErrorContains(t, err, "does not have enough stake tokens to delegate")
}

func addValidator(ctx sdk.Context, stakingKeeper *stakingkeeper.Keeper, owner sdk.AccAddress, value sdk.Coin) error {
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	valAddr := sdk.ValAddress(owner)

	pkAny, err := codectypes.NewAnyWithValue(pubKey)
	if err != nil {
		return err
	}
	msg := &stakingtypes.MsgCreateValidator{
		Description: stakingtypes.Description{
			Moniker: "Validator power",
		},
		Commission: stakingtypes.CommissionRates{
			Rate:          sdkmath.LegacyMustNewDecFromStr("0.1"),
			MaxRate:       sdkmath.LegacyMustNewDecFromStr("0.2"),
			MaxChangeRate: sdkmath.LegacyMustNewDecFromStr("0.01"),
		},
		MinSelfDelegation: sdkmath.OneInt(),
		DelegatorAddress:  owner.String(),
		ValidatorAddress:  valAddr.String(),
		Pubkey:            pkAny,
		Value:             value,
	}
	_, err = stakingkeeper.NewMsgServerImpl(stakingKeeper).CreateValidator(ctx, msg)
	return err
}
