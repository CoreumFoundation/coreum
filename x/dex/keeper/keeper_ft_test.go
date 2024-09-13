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
