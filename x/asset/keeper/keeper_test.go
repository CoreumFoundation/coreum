package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestKeeper_IssueFTAsset(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	assetKeeper := testApp.AssetKeeper
	bankKeeper := testApp.BankKeeper

	definition := genValidFTDefinition()
	// issue new fully valida asset with the precision
	id, err := assetKeeper.IssueAsset(ctx, definition)
	requireT.NoError(err)
	requireT.Equal(uint64(1), id)

	denomName := fmt.Sprintf("%s%s%d", types.ModuleName, definition.Code, id)
	denomBaseName := fmt.Sprintf("b%s", denomName)

	// check that we stored the right asset
	expectedDefinition := definition
	expectedDefinition.Ft.DenomName = denomName
	expectedDefinition.Ft.DenomBaseName = denomBaseName
	asset, err := assetKeeper.GetAsset(ctx, id)
	requireT.NoError(err)
	requireT.Equal(types.Asset{
		Id:         id,
		Definition: &expectedDefinition,
	}, asset)

	// check bank state

	// check the metadata
	storedMetadata, found := bankKeeper.GetDenomMetaData(ctx, denomBaseName)
	requireT.True(found)
	requireT.Equal(banktypes.Metadata{
		Name:        denomName,
		Symbol:      denomName,
		Description: definition.Description,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    denomBaseName,
				Exponent: uint32(0),
			},
			{
				Denom:    denomName,
				Exponent: definition.Ft.Precision,
			},
		},
		Base:    denomBaseName,
		Display: denomName,
	}, storedMetadata)

	// check the account state
	issuedAssetBalance := bankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(definition.Recipient), denomBaseName)
	expectedBalance := definition.Ft.InitialAmount.Mul(sdk.NewIntWithDecimal(1, int(definition.Ft.Precision)))
	requireT.Equal(sdk.NewCoin(denomBaseName, expectedBalance).String(), issuedAssetBalance.String())
}

func TestKeeper_IssueZeroPrecisionFTAsset(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	assetKeeper := testApp.AssetKeeper
	bankKeeper := testApp.BankKeeper

	definition := genValidFTDefinition()
	definition.Ft.Precision = 0

	// issue new fully valida asset with the precision
	id, err := assetKeeper.IssueAsset(ctx, definition)
	requireT.NoError(err)
	requireT.Equal(uint64(1), id)

	denomName := fmt.Sprintf("%s%s%d", types.ModuleName, definition.Code, id)
	denomBaseName := denomName

	// check that we stored the right asset
	expectedDefinition := definition
	expectedDefinition.Ft.DenomName = denomName
	expectedDefinition.Ft.DenomBaseName = denomBaseName
	asset, err := assetKeeper.GetAsset(ctx, id)
	requireT.NoError(err)
	requireT.Equal(types.Asset{
		Id:         id,
		Definition: &expectedDefinition,
	}, asset)

	// check bank state

	// check the metadata
	storedMetadata, found := bankKeeper.GetDenomMetaData(ctx, denomBaseName)
	requireT.True(found)
	requireT.Equal(banktypes.Metadata{
		Name:        denomName,
		Symbol:      denomName,
		Description: definition.Description,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    denomBaseName,
				Exponent: uint32(0),
			},
		},
		Base:    denomBaseName,
		Display: denomName,
	}, storedMetadata)

	// check the account state
	issuedAssetBalance := bankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(definition.Recipient), denomBaseName)
	requireT.Equal(sdk.NewCoin(denomBaseName, definition.Ft.InitialAmount).String(), issuedAssetBalance.String())
}

func genValidFTDefinition() types.AssetDefinition {
	return types.AssetDefinition{
		Recipient:   sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address()).String(),
		Type:        types.AssetType_FT, //nolint:nosnakecase // protogen
		Code:        "BTC",
		Description: "BTC Description",
		Ft: &types.FTCustomDefinition{
			Precision:     6,
			InitialAmount: sdk.NewInt(777),
		},
	}
}
