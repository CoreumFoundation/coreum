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

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	codeID, _, err := testApp.WasmGovPermissionKeeper.Create(
		ctx, issuer, testcontracts.AssetExtensionWasm, &wasmtypes.AllowEverybody,
	)
	requireT.NoError(err)

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Description:   "ABC Desc",
		Subunit:       "extensionabc",
		Precision:     8,
		InitialAmount: sdkmath.NewInt(777),
		Features:      []types.Feature{types.Feature_extension},
		ExtensionSettings: &types.ExtensionIssueSettings{
			CodeId: codeID,
		},
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	requireT.Equal(types.BuildDenom(settings.Subunit, settings.Issuer), denom)

	gotToken, err := ftKeeper.GetToken(ctx, denom)
	requireT.NoError(err)
	requireT.EqualValues(gotToken.Denom, denom)
	requireT.EqualValues(gotToken.Issuer, settings.Issuer.String())
	requireT.EqualValues(gotToken.Symbol, settings.Symbol)
	requireT.EqualValues(gotToken.Description, settings.Description)
	requireT.EqualValues(gotToken.Subunit, strings.ToLower(settings.Subunit))
	requireT.EqualValues(gotToken.Precision, settings.Precision)
	requireT.EqualValues(gotToken.Features, []types.Feature{types.Feature_extension})
	requireT.EqualValues(gotToken.BurnRate, sdk.NewDec(0))
	requireT.EqualValues(gotToken.SendCommissionRate, sdk.NewDec(0))
	requireT.EqualValues(gotToken.Version, types.CurrentTokenVersion)
	requireT.EqualValues(gotToken.URI, settings.URI)
	requireT.EqualValues(gotToken.URIHash, settings.URIHash)
	requireT.EqualValues(66, len(gotToken.ExtensionCWAddress))

	// check the account state
	issuedAssetBalance := bankKeeper.GetBalance(ctx, issuer, denom)
	requireT.Equal(sdk.NewCoin(denom, settings.InitialAmount).String(), issuedAssetBalance.String())

	// send 1 coin will succeed
	receiver := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, settings.Issuer, receiver, sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(2))))
	requireT.NoError(err)
	balance := bankKeeper.GetBalance(ctx, receiver, denom)
	requireT.EqualValues("2", balance.Amount.String())

	// send 7 coin will fail.
	// the POC contract is written as such that sending 7 will fail.
	// TODO replace with more meningful checks.
	err = bankKeeper.SendCoins(ctx, settings.Issuer, receiver, sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(7))))
	requireT.ErrorIs(err, types.ErrExtensionCallFailed)
	balance = bankKeeper.GetBalance(ctx, receiver, denom)
	requireT.EqualValues("2", balance.Amount.String())
}
