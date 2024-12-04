//go:build integrationtests

package upgrade

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v5/integration-tests"
	moduleswasm "github.com/CoreumFoundation/coreum/v5/integration-tests/contracts/modules"
	"github.com/CoreumFoundation/coreum/v5/pkg/client"
	"github.com/CoreumFoundation/coreum/v5/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
)

type assetft struct {
	token assetfttypes.Token
}

func (a *assetft) Before(t *testing.T) {
	t.Logf("Checking asset FT before upgrade")

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	assetftClient := assetfttypes.NewQueryClient(chain.ClientContext)

	issuer := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.
			Add(sdkmath.NewInt(2_000_000)), // extension issuance
	})

	codeID, err := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactoryAuto(), issuer, moduleswasm.AssetFTExtensionLegacyWASM,
	)
	requireT.NoError(err)

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(100_000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_whitelisting,
			// extension but no IBC
			assetfttypes.Feature_extension,
		},
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId: codeID,
			Label:  "legacy-issuance",
		},
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactoryAuto(),
		issueMsg,
	)
	requireT.NoError(err)

	denom := assetfttypes.BuildDenom(issueMsg.Subunit, issuer)
	tokenRes, err := assetftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: denom,
	})
	requireT.NoError(err)

	a.token = tokenRes.Token
}

func (a *assetft) After(t *testing.T) {
	t.Logf("Checking asset FT after upgrade")

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	assetftClient := assetfttypes.NewQueryClient(chain.ClientContext)

	tokenRes, err := assetftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: a.token.Denom,
	})
	requireT.NoError(err)

	expectedToken := a.token
	expectedToken.Features = append(expectedToken.Features, assetfttypes.Feature_ibc)

	require.Equal(t, expectedToken, tokenRes.Token)
}
