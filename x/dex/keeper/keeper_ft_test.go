package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func TestKeeper_PlaceOrderWithBlockDEXFeature(t *testing.T) {
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
			assetfttypes.Feature_block_dex,
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
			assetfttypes.Feature_restrict_dex,
		},
		DEXRestrictions: &assetfttypes.DEXRestrictions{
			DenomsToTradeWith: []string{
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
		t, testApp.DEXKeeper.PlaceOrder(sdkCtx, orderReceiveDenom2), "is prohibited for receive denom denom2",
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
	require.NoError(t, testApp.AssetFTKeeper.UpdateDEXRestrictions(sdkCtx, issuer, denom, assetfttypes.DEXRestrictions{
		DenomsToTradeWith: nil,
	}))
	require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, orderReceiveDenom2))
}
