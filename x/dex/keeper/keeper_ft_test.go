package keeper_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/docker/distribution/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v5/testutil/event"
	"github.com/CoreumFoundation/coreum/v5/testutil/simapp"
	testcontracts "github.com/CoreumFoundation/coreum/v5/x/asset/ft/keeper/test-contracts"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

const (
	ExtensionOrderDataWASMAttribute = "order_data"
	IDDEXOrderSuffixTrigger         = "blocked"
)

var (
	AmountDEXExpectToSpendTrigger   = sdkmath.NewInt(103)
	AmountDEXExpectToReceiveTrigger = sdkmath.NewInt(104)
)

func TestKeeper_PlaceOrderWithExtension(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

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
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
		},
	}
	denomWithExtension, err := testApp.AssetFTKeeper.Issue(sdkCtx, settingsWithExtension)
	require.NoError(t, err)

	tests := []struct {
		name       string
		order      types.Order
		wantDEXErr bool
	}{
		{
			name: "sell_positive",
			order: types.Order{
				Creator: func() string {
					creator, _ := testApp.GenAccount(sdkCtx)
					return creator.String()
				}(),
				Type:        types.ORDER_TYPE_LIMIT,
				ID:          uuid.Generate().String(),
				BaseDenom:   denomWithExtension,
				QuoteDenom:  denom2,
				Price:       lo.ToPtr(types.MustNewPriceFromString("1")),
				Quantity:    sdkmath.NewInt(10),
				Side:        types.SIDE_SELL,
				TimeInForce: types.TIME_IN_FORCE_GTC,
			},
			wantDEXErr: false,
		},
		{
			name: "sell_dex_error",
			order: types.Order{
				Creator: func() string {
					creator, _ := testApp.GenAccount(sdkCtx)
					return creator.String()
				}(),
				Type:        types.ORDER_TYPE_LIMIT,
				ID:          uuid.Generate().String(),
				BaseDenom:   denomWithExtension,
				QuoteDenom:  denom2,
				Price:       lo.ToPtr(types.MustNewPriceFromString("1")),
				Quantity:    AmountDEXExpectToSpendTrigger,
				Side:        types.SIDE_SELL,
				TimeInForce: types.TIME_IN_FORCE_GTC,
			},
			wantDEXErr: true,
		},
		{
			name: "buy_positive",
			order: types.Order{
				Creator: func() string {
					creator, _ := testApp.GenAccount(sdkCtx)
					return creator.String()
				}(),
				Type:        types.ORDER_TYPE_LIMIT,
				ID:          uuid.Generate().String(),
				BaseDenom:   denom2,
				QuoteDenom:  denomWithExtension,
				Price:       lo.ToPtr(types.MustNewPriceFromString("1")),
				Quantity:    sdkmath.NewInt(10),
				Side:        types.SIDE_BUY,
				TimeInForce: types.TIME_IN_FORCE_GTC,
			},
			wantDEXErr: false,
		},
		{
			name: "buy_dex_error",
			order: types.Order{
				Creator: func() string {
					creator, _ := testApp.GenAccount(sdkCtx)
					return creator.String()
				}(),
				Type:        types.ORDER_TYPE_LIMIT,
				ID:          uuid.Generate().String(),
				BaseDenom:   denom2,
				QuoteDenom:  denomWithExtension,
				Price:       lo.ToPtr(types.MustNewPriceFromString("1")),
				Quantity:    AmountDEXExpectToReceiveTrigger,
				Side:        types.SIDE_BUY,
				TimeInForce: types.TIME_IN_FORCE_GTC,
			},
			wantDEXErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creator := sdk.MustAccAddressFromBech32(tt.order.Creator)
			lockedBalance, err := tt.order.ComputeLimitOrderLockedBalance()
			require.NoError(t, err)
			testApp.MintAndSendCoin(t, sdkCtx, creator, sdk.NewCoins(lockedBalance))
			fundOrderReserve(t, testApp, sdkCtx, creator)
			if !tt.wantDEXErr {
				sdkCtx = sdkCtx.WithEventManager(sdk.NewEventManager())
				require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, tt.order))

				// decode wasm events
				orderStr, err := event.FindStringEventAttribute(
					sdkCtx.EventManager().Events().ToABCIEvents(),
					wasmtypes.WasmModuleEventType,
					ExtensionOrderDataWASMAttribute,
				)
				require.NoError(t, err)

				extensionOrderData := assetfttypes.DEXOrder{}
				require.NoError(t, json.Unmarshal([]byte(orderStr), &extensionOrderData))

				order, err := testApp.DEXKeeper.GetOrderByAddressAndID(sdkCtx, creator, tt.order.ID)
				require.NoError(t, err)

				require.Equal(t, assetfttypes.DEXOrder{
					Creator:    sdk.MustAccAddressFromBech32(order.Creator),
					Type:       order.Type.String(),
					ID:         order.ID,
					Sequence:   order.Sequence,
					BaseDenom:  order.BaseDenom,
					QuoteDenom: order.QuoteDenom,
					Price:      lo.ToPtr(order.Price.String()),
					Quantity:   order.Quantity,
					Side:       order.Side.String(),
				}, extensionOrderData)
			} else {
				require.ErrorContains(
					t,
					testApp.DEXKeeper.PlaceOrder(simapp.CopyContextWithMultiStore(sdkCtx), tt.order),
					"wasm error: DEX order placement is failed",
				)
			}
		})
	}
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
	fundOrderReserve(t, testApp, sdkCtx, acc)
	errStr := fmt.Sprintf("usage of %s is not supported for DEX, the token has dex_block", denomWithExtension)
	require.ErrorContains(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order), errStr)

	// use the denomWithExtension as quote
	order = types.Order{
		Creator:    acc.String(),
		Type:       types.ORDER_TYPE_LIMIT,
		ID:         uuid.Generate().String(),
		BaseDenom:  denom2,
		QuoteDenom: denomWithExtension,
		Price:      lo.ToPtr(types.MustNewPriceFromString("12e-1")),
		Quantity:   sdkmath.NewInt(10),
		Side:       types.SIDE_SELL,
		GoodTil: &types.GoodTil{
			GoodTilBlockHeight: 390,
		},
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	lockedBalance, err = order.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, sdkCtx, acc, sdk.NewCoins(lockedBalance))
	fundOrderReserve(t, testApp, sdkCtx, acc)
	require.ErrorContains(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order), errStr)
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
	fundOrderReserve(t, testApp, sdkCtx, acc)
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
	fundOrderReserve(t, testApp, sdkCtx, acc)
	require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, orderReceiveDenom3))

	// now update settings to remove all limit and place orderReceiveDenom2
	require.NoError(t, testApp.AssetFTKeeper.UpdateDEXWhitelistedDenoms(sdkCtx, issuer, denom, nil))
	fundOrderReserve(t, testApp, sdkCtx, acc)
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
	fundOrderReserve(t, testApp, sdkCtx, acc)
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

	balance := testApp.BankKeeper.GetBalance(sdkCtx, acc, denomToStake)
	require.Equal(t, sdk.NewInt64Coin(denomToStake, 0).String(), balance.String())

	lockedBalance := testApp.AssetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, denomToStake)
	require.Equal(t, sdk.NewInt64Coin(denomToStake, 0).String(), lockedBalance.String())

	fundOrderReserve(t, testApp, sdkCtx, acc)

	require.Error(t, testApp.DEXKeeper.PlaceOrder(
		simapp.CopyContextWithMultiStore(sdkCtx), order), cosmoserrors.ErrInsufficientFunds,
	)
	require.NoError(t, testApp.FundAccount(sdkCtx, acc, sdk.NewCoins(orderLockedBalance)))

	require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))

	balance = testApp.BankKeeper.GetBalance(sdkCtx, acc, denomToStake)
	params, err := testApp.DEXKeeper.GetParams(sdkCtx)
	require.NoError(t, err)
	orderReserve := params.OrderReserve
	require.Equal(t, orderLockedBalance.Add(orderReserve).String(), balance.String())

	lockedBalance = testApp.AssetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, denomToStake)
	require.Equal(t, orderLockedBalance.Add(orderReserve).String(), lockedBalance.String())

	_, err = testApp.StakingKeeper.Delegate(
		simapp.CopyContextWithMultiStore(sdkCtx),
		acc,
		orderLockedBalance.Amount,
		stakingtypes.Unbonded,
		val[0],
		true,
	)
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
	fundOrderReserve(t, testApp, sdkCtx, validatorOwner2)
	require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))
	err = addValidator(sdkCtx, testApp.StakingKeeper, validatorOwner2, orderLockedBalance)
	require.ErrorContains(t, err, "does not have enough stake tokens to delegate")
}

func TestKeeper_PlaceOrderWithBurnRate(t *testing.T) {
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
		BurnRate: sdkmath.LegacyMustNewDecFromStr("0.5"),
	}
	denomWithBurnRate, err := testApp.AssetFTKeeper.Issue(sdkCtx, settingsWithExtension)
	require.NoError(t, err)

	order := types.Order{
		Creator:    acc.String(),
		Type:       types.ORDER_TYPE_LIMIT,
		ID:         uuid.Generate().String(),
		BaseDenom:  denomWithBurnRate,
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
	fundOrderReserve(t, testApp, sdkCtx, acc)
	balanceBeforePlaceOrder := testApp.BankKeeper.GetBalance(sdkCtx, acc, denomWithBurnRate)
	require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))
	balanceAfterPlaceOrder := testApp.BankKeeper.GetBalance(sdkCtx, acc, denomWithBurnRate)
	require.Equal(t, balanceBeforePlaceOrder, balanceAfterPlaceOrder)
}

func TestKeeper_PlaceOrderWithCommissionRate(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	acc, _ := testApp.GenAccount(sdkCtx)
	issuer, _ := testApp.GenAccount(sdkCtx)

	settingsWithExtension := assetfttypes.IssueSettings{
		Issuer:             issuer,
		Symbol:             "DEFEXT",
		Subunit:            "defext",
		Precision:          6,
		InitialAmount:      sdkmath.NewIntWithDecimal(1, 10),
		SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("0.5"),
	}
	denomWithCommissionRate, err := testApp.AssetFTKeeper.Issue(sdkCtx, settingsWithExtension)
	require.NoError(t, err)

	order := types.Order{
		Creator:    acc.String(),
		Type:       types.ORDER_TYPE_LIMIT,
		ID:         uuid.Generate().String(),
		BaseDenom:  denomWithCommissionRate,
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
	balanceBeforePlaceOrder := testApp.BankKeeper.GetBalance(sdkCtx, acc, denomWithCommissionRate)
	fundOrderReserve(t, testApp, sdkCtx, acc)
	require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))
	balanceAfterPlaceOrder := testApp.BankKeeper.GetBalance(sdkCtx, acc, denomWithCommissionRate)
	require.Equal(t, balanceBeforePlaceOrder, balanceAfterPlaceOrder)
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
