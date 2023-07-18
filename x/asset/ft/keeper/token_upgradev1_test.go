package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/v2/app"
	"github.com/CoreumFoundation/coreum/v2/pkg/config"
	"github.com/CoreumFoundation/coreum/v2/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v2/x/asset/ft/types"
)

func TestTokenUpgradeV1(t *testing.T) {
	requireT := require.New(t)

	cdc := config.NewEncodingConfig(app.ModuleBasics).Codec
	testApp := simapp.New()
	ctxSDK := testApp.BaseApp.NewContext(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	delayKeeper := testApp.DelayKeeper

	params := ftKeeper.GetParams(ctxSDK)
	params.TokenUpgradeDecisionTimeout = time.Date(2023, 2, 13, 1, 2, 3, 0, time.UTC)
	ftKeeper.SetParams(ctxSDK, params)

	issuer1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	issuer2 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	denom1, err := ftKeeper.IssueVersioned(ctxSDK, types.IssueSettings{
		Issuer:        issuer1,
		Symbol:        "ABC",
		Description:   "ABC Desc",
		Subunit:       "abc",
		Precision:     8,
		InitialAmount: sdk.NewInt(777),
	}, 0)
	requireT.NoError(err)

	denom2, err := ftKeeper.IssueVersioned(ctxSDK, types.IssueSettings{
		Issuer:        issuer2,
		Symbol:        "XYZ",
		Description:   "XYZ Desc",
		Subunit:       "xyz",
		Precision:     8,
		InitialAmount: sdk.NewInt(888),
	}, 0)
	requireT.NoError(err)

	// upgrade requested after timeout should fail
	ctxSDK = ctxSDK.WithBlockTime(params.TokenUpgradeDecisionTimeout.Add(time.Second))
	requireT.Error(ftKeeper.AddDelayedTokenUpgradeV1(ctxSDK, issuer1, denom1, false))
	requireT.Error(ftKeeper.AddDelayedTokenUpgradeV1(ctxSDK, issuer1, denom1, true))

	ctxSDK = ctxSDK.WithBlockTime(params.TokenUpgradeDecisionTimeout)

	// call for non-existing denom fails
	requireT.Error(ftKeeper.AddDelayedTokenUpgradeV1(ctxSDK, issuer1, "denom", false))

	// call from non-issuer account fails
	requireT.Error(ftKeeper.AddDelayedTokenUpgradeV1(ctxSDK, issuer1, denom2, false))

	// first call succeeds
	requireT.NoError(ftKeeper.AddDelayedTokenUpgradeV1(ctxSDK, issuer1, denom1, false))

	// ibc is set to false so the change should be applied immediately
	token1, err := ftKeeper.GetToken(ctxSDK, denom1)
	requireT.NoError(err)
	requireT.Empty(token1.Features)
	requireT.EqualValues(1, token1.Version)

	tokenUpgradeStatuses := ftKeeper.GetTokenUpgradeStatuses(ctxSDK, denom1)
	requireT.Equal(&types.TokenUpgradeV1Status{
		IbcEnabled: false,
		StartTime:  ctxSDK.BlockTime(),
		EndTime:    ctxSDK.BlockTime(),
	}, tokenUpgradeStatuses.V1)

	// delay module should not contain delayed item
	delayedItems, err := delayKeeper.ExportDelayedItems(ctxSDK)
	requireT.NoError(err)
	requireT.Empty(delayedItems)

	// second call fails
	requireT.Error(ftKeeper.AddDelayedTokenUpgradeV1(ctxSDK, issuer1, denom1, false))

	// setting pending version should work
	requireT.NoError(ftKeeper.SetPendingVersion(ctxSDK, denom1, 2))

	// for second denom we turn IBC on
	requireT.NoError(ftKeeper.AddDelayedTokenUpgradeV1(ctxSDK, issuer2, denom2, true))

	// ibc is set to true so the change should be posponed and parameters should stay the same for now
	token2, err := ftKeeper.GetToken(ctxSDK, denom2)
	requireT.NoError(err)
	requireT.Empty(token2.Features)
	requireT.EqualValues(0, token2.Version)

	tokenUpgradeStatuses2 := ftKeeper.GetTokenUpgradeStatuses(ctxSDK, denom2)
	requireT.Equal(&types.TokenUpgradeV1Status{
		IbcEnabled: true,
		StartTime:  ctxSDK.BlockTime(),
		EndTime:    ctxSDK.BlockTime().Add(params.TokenUpgradeGracePeriod),
	}, tokenUpgradeStatuses2.V1)

	// delay module should contain delayed item
	delayedItems, err = delayKeeper.ExportDelayedItems(ctxSDK)
	requireT.NoError(err)
	requireT.Len(delayedItems, 1)
	requireT.Equal("assetft-upgrade-1-"+denom2, delayedItems[0].Id)
	requireT.Equal(ctxSDK.BlockTime().Add(params.TokenUpgradeGracePeriod), delayedItems[0].ExecutionTime)

	var delayedItem codec.ProtoMarshaler
	requireT.NoError(cdc.UnpackAny(delayedItems[0].Data, &delayedItem))

	requireT.Equal(denom2, delayedItem.(*types.DelayedTokenUpgradeV1).Denom)

	// next call fails
	requireT.Error(ftKeeper.AddDelayedTokenUpgradeV1(ctxSDK, issuer2, denom2, true))
	requireT.Error(ftKeeper.AddDelayedTokenUpgradeV1(ctxSDK, issuer2, denom2, false))

	// setting pending version should fail
	requireT.Error(ftKeeper.SetPendingVersion(ctxSDK, denom2, 1))

	// now let's execute the upgrade
	requireT.NoError(ftKeeper.UpgradeTokenToV1(ctxSDK, &types.DelayedTokenUpgradeV1{
		Denom: denom2,
	}))

	// token should be upgraded
	token2, err = ftKeeper.GetToken(ctxSDK, denom2)
	requireT.NoError(err)
	requireT.Len(token2.Features, 1)
	requireT.Equal(types.Feature_ibc, token2.Features[0])
	requireT.EqualValues(1, token2.Version)

	// next call fails
	requireT.Error(ftKeeper.AddDelayedTokenUpgradeV1(ctxSDK, issuer2, denom2, true))

	// setting pending version should work
	requireT.NoError(ftKeeper.SetPendingVersion(ctxSDK, denom2, 1))
}
