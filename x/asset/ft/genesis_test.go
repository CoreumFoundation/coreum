package ft_test

import (
	"fmt"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/ft"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

//nolint:funlen
func TestInitAndExportGenesis(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	testApp := simapp.New()

	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})
	ftKeeper := testApp.AssetFTKeeper
	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	// prepare the genesis data

	// token definitions
	var tokens []types.Token
	for i := 0; i < 5; i++ {
		token := types.Token{
			Denom:              types.BuildDenom(fmt.Sprintf("abc%d", i), issuer),
			Issuer:             issuer.String(),
			Symbol:             fmt.Sprintf("ABC%d", i),
			Subunit:            fmt.Sprintf("abc%d", i),
			Precision:          uint32(rand.Int31n(100)),
			BurnRate:           sdk.MustNewDecFromStr(fmt.Sprintf("0.%d", i)),
			SendCommissionRate: sdk.MustNewDecFromStr(fmt.Sprintf("0.%d", i+1)),
			Features: []types.Feature{
				types.Feature_freezing,
				types.Feature_whitelisting,
			},
		}
		// Globally freeze some Tokens.
		if i%2 == 0 {
			token.GloballyFrozen = true
		}
		tokens = append(tokens, token)
		requireT.NoError(ftKeeper.SetDenomMetadata(ctx, token.Denom, token.Symbol, token.Description, token.Precision))
	}

	// frozen balances
	var frozenBalances []types.Balance
	for i := 0; i < 5; i++ {
		addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		frozenBalances = append(frozenBalances,
			types.Balance{
				Address: addr.String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin(tokens[0].Denom, sdk.NewInt(rand.Int63())),
					sdk.NewCoin(tokens[1].Denom, sdk.NewInt(rand.Int63())),
				),
			})
	}

	// whitelisted balances
	var whitelistedBalances []types.Balance
	for i := 0; i < 5; i++ {
		addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		whitelistedBalances = append(whitelistedBalances,
			types.Balance{
				Address: addr.String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin(tokens[0].Denom, sdk.NewInt(rand.Int63())),
					sdk.NewCoin(tokens[1].Denom, sdk.NewInt(rand.Int63())),
				),
			})
	}

	genState := types.GenesisState{
		Params:              types.DefaultParams(),
		Tokens:              tokens,
		FrozenBalances:      frozenBalances,
		WhitelistedBalances: whitelistedBalances,
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

	// check that export is equal import
	exportedGenState := ft.ExportGenesis(ctx, ftKeeper)

	assertT.EqualValues(genState.Params, exportedGenState.Params)
	assertT.ElementsMatch(genState.Tokens, exportedGenState.Tokens)
	assertT.ElementsMatch(genState.FrozenBalances, exportedGenState.FrozenBalances)
	assertT.ElementsMatch(genState.WhitelistedBalances, exportedGenState.WhitelistedBalances)
}
