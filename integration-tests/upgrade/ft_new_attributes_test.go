//go:build integrationtests

package upgrade

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	"github.com/CoreumFoundation/coreum/v3/pkg/client"
	"github.com/CoreumFoundation/coreum/v3/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
)

type ftURIAttributesTest struct {
	token assetfttypes.Token
}

func (ftt *ftURIAttributesTest) Before(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	assetftClient := assetfttypes.NewQueryClient(chain.ClientContext)

	issuer := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdkmath.NewInt(1000),
		Features:      []assetfttypes.Feature{},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)
	tokenRes, err := assetftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: assetfttypes.BuildDenom(issueMsg.Subunit, issuer),
	})
	requireT.NoError(err)
	ftt.token = tokenRes.Token
}

func (ftt *ftURIAttributesTest) After(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	assetftClient := assetfttypes.NewQueryClient(chain.ClientContext)

	// check that prev token is available we can query it
	tokenRes, err := assetftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ftt.token.Denom,
	})
	requireT.NoError(err)
	require.Equal(t, ftt.token, tokenRes.Token)
}
