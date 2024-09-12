//go:build integrationtests

package upgrade

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
)

type gov struct {
	oldParams *govtypesv1.Params
}

func (g *gov) Before(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	govParams, err := chain.Governance.QueryGovParams(ctx)
	requireT.NoError(err)

	g.oldParams = govParams
}

func (g *gov) After(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	assertT := assert.New(t)

	govParams, err := chain.Governance.QueryGovParams(ctx)
	requireT.NoError(err)

	assertT.Equal(sdkmath.LegacyMustNewDecFromStr("0.5").String(), govParams.ProposalCancelRatio)
	assertT.Equal("", govParams.ProposalCancelDest)
	assertT.Equal(lo.ToPtr(time.Hour), govParams.ExpeditedVotingPeriod)
	assertT.Equal(sdkmath.LegacyMustNewDecFromStr("0.667").String(), govParams.ExpeditedThreshold)
	assertT.Equal(sdk.NewCoins(
		sdk.NewCoin(chain.ChainSettings.Denom, sdkmath.NewInt(4_000_000_000)),
	), govParams.ExpeditedMinDeposit)
	assertT.Equal(sdkmath.LegacyMustNewDecFromStr("0.01").String(), govParams.MinDepositRatio)

	assertT.NotEqual(g.oldParams.ProposalCancelRatio, govParams.ProposalCancelRatio)
	assertT.NotEqual(g.oldParams.ExpeditedVotingPeriod, govParams.ExpeditedVotingPeriod)
	assertT.NotEqual(g.oldParams.ExpeditedThreshold, govParams.ExpeditedThreshold)
	assertT.NotEqual(g.oldParams.ExpeditedMinDeposit, govParams.ExpeditedMinDeposit)
	assertT.NotEqual(g.oldParams.MinDepositRatio, govParams.MinDepositRatio)
}
