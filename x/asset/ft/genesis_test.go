package ft_test

import (
	"fmt"
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v4/x/asset/ft"
	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

func TestInitAndExportGenesis(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	testApp := simapp.New()

	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})
	ftKeeper := testApp.AssetFTKeeper
	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	// prepare the genesis data

	// token definitions
	var tokens []types.Token
	var pendingTokenUpgrades []types.PendingTokenUpgrade
	for i := uint32(0); i < 5; i++ {
		token := types.Token{
			Denom:       types.BuildDenom(fmt.Sprintf("abc%d", i), issuer),
			Issuer:      issuer.String(),
			Symbol:      fmt.Sprintf("ABC%d", i),
			Subunit:     fmt.Sprintf("abc%d", i),
			Precision:   uint32(rand.Int31n(19) + 1),
			Description: fmt.Sprintf("DESC%d", i),
			Features: []types.Feature{
				types.Feature_freezing,
				types.Feature_whitelisting,
			},
			BurnRate:           sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("0.%d", i)),
			SendCommissionRate: sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("0.%d", i+1)),
			Version:            i,
			URI:                fmt.Sprintf("https://my-class-meta.invalid/%d", i),
			URIHash:            fmt.Sprintf("content-hash%d", i),
		}
		// Globally freeze some Tokens.
		if i%2 == 0 {
			token.GloballyFrozen = true
		}
		tokens = append(tokens, token)
		requireT.NoError(ftKeeper.SetDenomMetadata(
			ctx,
			token.Denom,
			token.Symbol,
			token.Description,
			token.URI,
			token.URIHash,
			token.Precision))
		if i == 0 {
			pendingTokenUpgrades = append(pendingTokenUpgrades, types.PendingTokenUpgrade{
				Denom:   token.Denom,
				Version: types.CurrentTokenVersion,
			})
		}
	}

	// frozen balances
	var frozenBalances []types.Balance
	for i := 0; i < 5; i++ {
		addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		frozenBalances = append(frozenBalances,
			types.Balance{
				Address: addr.String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin(tokens[0].Denom, sdkmath.NewInt(rand.Int63())),
					sdk.NewCoin(tokens[1].Denom, sdkmath.NewInt(rand.Int63())),
				),
			})
	}

	// whitelisted balances
	var whitelistedBalances []types.Balance
	for i := 0; i < 4; i++ {
		addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		whitelistedBalances = append(whitelistedBalances,
			types.Balance{
				Address: addr.String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin(tokens[0].Denom, sdkmath.NewInt(rand.Int63())),
					sdk.NewCoin(tokens[1].Denom, sdkmath.NewInt(rand.Int63())),
				),
			})
	}

	// DEX locked balances
	var dexLockedBalances []types.Balance
	for i := 0; i < 8; i++ {
		addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		dexLockedBalances = append(dexLockedBalances,
			types.Balance{
				Address: addr.String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin(tokens[0].Denom, sdkmath.NewInt(rand.Int63())),
					sdk.NewCoin(tokens[1].Denom, sdkmath.NewInt(rand.Int63())),
				),
			})
	}

	// DEX settings
	var dexSettings []types.DEXSettingsWithDenom
	for i := 0; i < 4; i++ {
		dexSettings = append(dexSettings,
			types.DEXSettingsWithDenom{
				Denom: fmt.Sprintf("denom-%d", i),
				DEXSettings: types.DEXSettings{
					UnifiedRefAmount: sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("1.%d", i)),
				},
			})
	}

	// DEX restrictions
	var dexRestrictions []types.DEXRestrictionsWithDenom
	for i := 0; i < 4; i++ {
		dexRestrictions = append(dexRestrictions,
			types.DEXRestrictionsWithDenom{
				Denom: fmt.Sprintf("denom-%d", i),
				DEXRestrictions: types.DEXRestrictions{
					DenomsToTradeWith: []string{"denom1", "denom2", fmt.Sprintf("denomx1.%d", i)},
				},
			})
	}

	genState := types.GenesisState{
		Params:               types.DefaultParams(),
		Tokens:               tokens,
		FrozenBalances:       frozenBalances,
		WhitelistedBalances:  whitelistedBalances,
		PendingTokenUpgrades: pendingTokenUpgrades,
		DEXLockedBalances:    dexLockedBalances,
		DEXSettings:          dexSettings,
		DEXRestrictions:      dexRestrictions,
	}

	// init the keeper
	ft.InitGenesis(ctx, ftKeeper, genState)

	// assert the keeper state

	// params

	params := ftKeeper.GetParams(ctx)
	assertT.EqualValues(types.DefaultParams(), params)

	// token definitions
	for _, definition := range tokens {
		storedToken, err := ftKeeper.GetToken(ctx, definition.Denom)
		requireT.NoError(err)
		assertT.EqualValues(definition, storedToken)
	}

	// frozen balances
	for _, balance := range frozenBalances {
		address, err := sdk.AccAddressFromBech32(balance.Address)
		requireT.NoError(err)
		coins, _, err := ftKeeper.GetFrozenBalances(ctx, address, nil)
		requireT.NoError(err)
		assertT.EqualValues(balance.Coins.String(), coins.String())
	}

	// whitelisted balances
	for _, balance := range whitelistedBalances {
		address, err := sdk.AccAddressFromBech32(balance.Address)
		requireT.NoError(err)
		coins, _, err := ftKeeper.GetWhitelistedBalances(ctx, address, nil)
		requireT.NoError(err)
		assertT.EqualValues(balance.Coins.String(), coins.String())
	}

	// DEX locked balances
	for _, balance := range dexLockedBalances {
		address, err := sdk.AccAddressFromBech32(balance.Address)
		requireT.NoError(err)
		coins, _, err := ftKeeper.GetDEXLockedBalances(ctx, address, nil)
		requireT.NoError(err)
		assertT.EqualValues(balance.Coins.String(), coins.String())
	}

	// DEX locked balances
	for _, settings := range dexSettings {
		storedSettings, err := ftKeeper.GetDEXSettings(ctx, settings.Denom)
		requireT.NoError(err)
		assertT.EqualValues(settings.DEXSettings, storedSettings)
	}

	for _, restrictions := range dexRestrictions {
		storedSettings, err := ftKeeper.GetDEXRestrictions(ctx, restrictions.Denom)
		requireT.NoError(err)
		assertT.EqualValues(restrictions.DEXRestrictions, storedSettings)
	}

	// check that export is equal import
	exportedGenState := ft.ExportGenesis(ctx, ftKeeper)

	assertT.EqualValues(genState.Params, exportedGenState.Params)
	assertT.ElementsMatch(genState.Tokens, exportedGenState.Tokens)
	assertT.ElementsMatch(genState.PendingTokenUpgrades, exportedGenState.PendingTokenUpgrades)
	assertT.ElementsMatch(genState.FrozenBalances, exportedGenState.FrozenBalances)
	assertT.ElementsMatch(genState.WhitelistedBalances, exportedGenState.WhitelistedBalances)
	assertT.ElementsMatch(genState.DEXLockedBalances, exportedGenState.DEXLockedBalances)
}
