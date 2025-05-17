package v4_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/testutil/simapp"
	testcontracts "github.com/CoreumFoundation/coreum/v6/x/asset/ft/keeper/test-contracts"
	v4 "github.com/CoreumFoundation/coreum/v6/x/asset/ft/migrations/v4"
	"github.com/CoreumFoundation/coreum/v6/x/asset/ft/types"
)

func TestMigrateDefinitions(t *testing.T) {
	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	tests := []struct {
		name     string
		settings types.IssueSettings
	}{
		{
			name: "no_ibs_no_extension",
			settings: types.IssueSettings{
				Issuer:        issuer,
				Symbol:        "DEF",
				Subunit:       "def",
				Precision:     1,
				InitialAmount: sdkmath.NewInt(1000),
				Features: []types.Feature{
					types.Feature_whitelisting,
				},
			},
		},
		{
			name: "ibs_no_extension",
			settings: types.IssueSettings{
				Issuer:        issuer,
				Symbol:        "DEF",
				Subunit:       "def",
				Precision:     1,
				InitialAmount: sdkmath.NewInt(1000),
				Features: []types.Feature{
					types.Feature_ibc,
				},
			},
		},
		{
			name: "extension_ibc",
			settings: types.IssueSettings{
				Issuer:        issuer,
				Symbol:        "DEF",
				Subunit:       "def",
				Precision:     1,
				InitialAmount: sdkmath.NewInt(1000),
				Features: []types.Feature{
					types.Feature_ibc,
					types.Feature_extension,
				},
			},
		},
		{
			name: "extension_no_ibc",
			settings: types.IssueSettings{
				Issuer:        issuer,
				Symbol:        "DEF",
				Subunit:       "def",
				Precision:     1,
				InitialAmount: sdkmath.NewInt(1000),
				Features: []types.Feature{
					types.Feature_whitelisting,
					types.Feature_extension,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requireT := require.New(t)

			testApp := simapp.New()
			ctx := testApp.NewContextLegacy(false, tmproto.Header{
				Time:    time.Now(),
				AppHash: []byte("some-hash"),
			})

			keeper := testApp.AssetFTKeeper

			settings := tt.settings
			if lo.Contains(settings.Features, types.Feature_extension) {
				codeID, _, err := testApp.WasmPermissionedKeeper.Create(
					ctx, issuer, testcontracts.AssetExtensionWasm, &wasmtypes.AllowEverybody,
				)
				requireT.NoError(err)
				settings.ExtensionSettings = &types.ExtensionIssueSettings{
					CodeId: codeID,
				}
			}

			denom, err := keeper.Issue(ctx, settings)
			requireT.NoError(err)
			def, err := keeper.GetDefinition(ctx, denom)
			requireT.NoError(err)
			requireT.NoError(keeper.SetDefinition(ctx, issuer, settings.Subunit, def))

			requireT.NoError(v4.MigrateDefinitions(ctx, keeper))

			def, err = keeper.GetDefinition(ctx, denom)
			requireT.NoError(err)
			if lo.Contains(settings.Features, types.Feature_extension) {
				requireT.Contains(def.Features, types.Feature_ibc)
				requireT.Contains(def.Features, types.Feature_dex_unified_ref_amount_change)
			}
		})
	}
}
