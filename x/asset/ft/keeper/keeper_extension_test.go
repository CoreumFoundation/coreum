package keeper_test

import (
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	testcontracts "github.com/CoreumFoundation/coreum/v4/x/asset/ft/keeper/test-contracts"
	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

func TestKeeper_Extension_Issue(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{
		Time:    time.Now(),
		AppHash: []byte("some-hash"),
	})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	codeID, _, err := testApp.WasmGovPermissionKeeper.Create(
		ctx, addr, testcontracts.AssetExtensionWasm, &wasmtypes.AllowEverybody,
	)
	requireT.NoError(err)

	settings := types.IssueSettings{
		Issuer:        addr,
		Symbol:        "ABC",
		Description:   "ABC Desc",
		Subunit:       "extensionabc",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(777),
		Features:      []types.Feature{types.Feature_extensions},
		ExtensionSettings: &types.ExtensionSettings{
			CodeId:           codeID,
			InstantiationMsg: []byte("{}"),
		},
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	requireT.Equal(types.BuildDenom(settings.Subunit, settings.Issuer), denom)

	gotToken, err := ftKeeper.GetToken(ctx, denom)
	requireT.NoError(err)
	requireT.Equal(types.Token{
		Denom:              denom,
		Issuer:             settings.Issuer.String(),
		Symbol:             settings.Symbol,
		Description:        settings.Description,
		Subunit:            strings.ToLower(settings.Subunit),
		Precision:          settings.Precision,
		Features:           []types.Feature{types.Feature_extensions},
		BurnRate:           sdk.NewDec(0),
		SendCommissionRate: sdk.NewDec(0),
		Version:            types.CurrentTokenVersion,
		URI:                settings.URI,
		URIHash:            settings.URIHash,
	}, gotToken)

	// check the metadata
	storedMetadata, found := bankKeeper.GetDenomMetaData(ctx, denom)
	requireT.True(found)
	requireT.Equal(banktypes.Metadata{
		Name:        settings.Symbol,
		Symbol:      settings.Symbol,
		Description: settings.Description,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    denom,
				Exponent: 0,
			},
			{
				Denom:    settings.Symbol,
				Exponent: settings.Precision,
			},
		},
		Base:    denom,
		Display: settings.Symbol,
		URI:     settings.URI,
		URIHash: settings.URIHash,
	}, storedMetadata)

	// check the account state
	issuedAssetBalance := bankKeeper.GetBalance(ctx, addr, denom)
	requireT.Equal(sdk.NewCoin(denom, settings.InitialAmount).String(), issuedAssetBalance.String())

	// send 1 coin will succeed
	receiver := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, settings.Issuer, receiver, sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1))))
	requireT.NoError(err)

	// send 7 coin will fail.
	// the POC contract is written as such that sending 7 will fail.
	// TODO replace with more meningful checks.
	err = bankKeeper.SendCoins(ctx, settings.Issuer, receiver, sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(7))))
	requireT.ErrorIs(err, types.ErrExtensionCallFailed)
}
