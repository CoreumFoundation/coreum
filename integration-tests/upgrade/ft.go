package upgrade

import (
	"context"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

type ftTest struct {
	issuer                      sdk.AccAddress
	denomV0WithoutFeatures      string
	denomV0WithFeatures         string
	denomV0ForForbiddenUpgrades string
}

func (ft *ftTest) Before(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	ft.issuer = chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, ft.issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgUpgradeTokenV1{},
			&assetfttypes.MsgUpgradeTokenV1{},
			&assetfttypes.MsgUpgradeTokenV1{},
			&assetfttypes.MsgUpgradeTokenV1{},
			&assetfttypes.MsgUpgradeTokenV1{},
			&assetfttypes.MsgUpgradeTokenV1{},
			&assetfttypes.MsgUpgradeTokenV1{},
			&assetfttypes.MsgUpgradeTokenV1{},
			&assetfttypes.MsgUpgradeTokenV1{},
			&assetfttypes.MsgUpgradeTokenV1{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount.MulRaw(5),
	})

	ft.issueV0TokenWithoutFeatures(t)
	ft.issueV0TokenWithFeatures(t)
	ft.tryToUpgradeTokenFromV0ToV1BeforeUpgradingTheApp(t)
}

func (ft *ftTest) issueV0TokenWithoutFeatures(t *testing.T) {
	requireT := require.New(t)
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        ft.issuer.String(),
		Symbol:        "AAA",
		Subunit:       "uaaa",
		Precision:     6,
		Description:   "AAA Description",
		InitialAmount: sdk.NewInt(1000),
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(ft.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)
	ft.denomV0WithoutFeatures = assetfttypes.BuildDenom(issueMsg.Subunit, ft.issuer)
}

func (ft *ftTest) issueV0TokenWithFeatures(t *testing.T) {
	requireT := require.New(t)
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        ft.issuer.String(),
		Symbol:        "BBB",
		Subunit:       "ubbb",
		Precision:     6,
		Description:   "BBB Description",
		InitialAmount: sdk.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_minting,
			assetfttypes.Feature_freezing,
			assetfttypes.Feature_whitelisting,
			assetfttypes.Feature_burning,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(ft.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)
	ft.denomV0WithFeatures = assetfttypes.BuildDenom(issueMsg.Subunit, ft.issuer)
}

func (ft *ftTest) tryToUpgradeTokenFromV0ToV1BeforeUpgradingTheApp(t *testing.T) {
	requireT := require.New(t)
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        ft.issuer.String(),
		Symbol:        "CCC",
		Subunit:       "uccc",
		Precision:     6,
		Description:   "CCC Description",
		InitialAmount: sdk.NewInt(1000),
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(ft.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)
	ft.denomV0ForForbiddenUpgrades = assetfttypes.BuildDenom(issueMsg.Subunit, ft.issuer)

	// upgrading token before chain upgrade should not work
	upgradeMsg := &assetfttypes.MsgUpgradeTokenV1{
		Sender:     ft.issuer.String(),
		Denom:      ft.denomV0ForForbiddenUpgrades,
		IbcEnabled: true,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(ft.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(upgradeMsg)),
		upgradeMsg,
	)
	requireT.ErrorContains(err, "tx parse error")

	ftClient := assetfttypes.NewQueryClient(chain.ClientContext)
	resp, err := ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ft.denomV0ForForbiddenUpgrades,
	})
	requireT.NoError(err)
	requireT.EqualValues(0, resp.Token.Version)
	requireT.Len(resp.Token.Features, 0)
}

func (ft *ftTest) After(t *testing.T) {
	ft.tryToUpgradeV1TokenToEnableIBC(t)
	ft.tryToUpgradeV1TokenToDisableIBC(t)
	ft.tryToUpgradeV0ToV1ByNonIssuer(t)

	ft.upgradeFromV0ToV1ToDisableIBC(t)
	ft.upgradeFromV0ToV1ToEnableIBC(t)
	ft.tryToUpgradeV0ToV1AfterDecisionTimeout(t)
}

func (ft *ftTest) tryToUpgradeV1TokenToEnableIBC(t *testing.T) {
	requireT := require.New(t)
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	// issuing token without IBC should succeed
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        ft.issuer.String(),
		Symbol:        "CDE",
		Subunit:       "ucde",
		Precision:     6,
		Description:   "CDE Description",
		InitialAmount: sdk.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_minting,
			assetfttypes.Feature_freezing,
			assetfttypes.Feature_whitelisting,
			assetfttypes.Feature_burning,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(ft.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	denomCDE := assetfttypes.BuildDenom(issueMsg.Subunit, ft.issuer)

	ftClient := assetfttypes.NewQueryClient(chain.ClientContext)
	resp, err := ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: denomCDE,
	})
	requireT.NoError(err)
	requireT.EqualValues(1, resp.Token.Version)
	requireT.Equal([]assetfttypes.Feature{
		assetfttypes.Feature_minting,
		assetfttypes.Feature_freezing,
		assetfttypes.Feature_whitelisting,
		assetfttypes.Feature_burning,
	}, resp.Token.Features)

	upgradeMsg := &assetfttypes.MsgUpgradeTokenV1{
		Sender:     ft.issuer.String(),
		Denom:      denomCDE,
		IbcEnabled: true,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(ft.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(upgradeMsg)),
		upgradeMsg,
	)
	requireT.ErrorContains(err, fmt.Sprintf("denom %s has been already upgraded to v1", denomCDE))

	resp, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: denomCDE,
	})
	requireT.NoError(err)
	requireT.EqualValues(1, resp.Token.Version)
	requireT.Equal([]assetfttypes.Feature{
		assetfttypes.Feature_minting,
		assetfttypes.Feature_freezing,
		assetfttypes.Feature_whitelisting,
		assetfttypes.Feature_burning,
	}, resp.Token.Features)
}

func (ft *ftTest) tryToUpgradeV1TokenToDisableIBC(t *testing.T) {
	requireT := require.New(t)
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	// issuing token with IBC should succeed after the upgrade
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        ft.issuer.String(),
		Symbol:        "XYZ",
		Subunit:       "uxyz",
		Precision:     6,
		Description:   "XYZ Description",
		InitialAmount: sdk.NewInt(1000),
		Features:      []assetfttypes.Feature{assetfttypes.Feature_ibc},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(ft.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	denomXYZ := assetfttypes.BuildDenom(issueMsg.Subunit, ft.issuer)

	ftClient := assetfttypes.NewQueryClient(chain.ClientContext)
	resp, err := ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: denomXYZ,
	})
	requireT.NoError(err)
	requireT.EqualValues(1, resp.Token.Version)
	requireT.Equal([]assetfttypes.Feature{assetfttypes.Feature_ibc}, resp.Token.Features)

	// upgrading v1 tokens should fail
	upgradeMsg := &assetfttypes.MsgUpgradeTokenV1{
		Sender:     ft.issuer.String(),
		Denom:      denomXYZ,
		IbcEnabled: false,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(ft.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(upgradeMsg)),
		upgradeMsg,
	)
	requireT.ErrorContains(err, fmt.Sprintf("denom %s has been already upgraded to v1", denomXYZ))

	resp, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: denomXYZ,
	})
	requireT.NoError(err)
	requireT.EqualValues(1, resp.Token.Version)
	requireT.Equal([]assetfttypes.Feature{assetfttypes.Feature_ibc}, resp.Token.Features)
}

func (ft *ftTest) tryToUpgradeV0ToV1ByNonIssuer(t *testing.T) {
	requireT := require.New(t)
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	// upgrading by the non-issuer should fail
	nonIssuer := chain.GenAccount()
	chain.FundAccountsWithOptions(ctx, t, nonIssuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgUpgradeTokenV1{},
			&assetfttypes.MsgUpgradeTokenV1{},
		},
	})
	upgradeMsg := &assetfttypes.MsgUpgradeTokenV1{
		Sender:     nonIssuer.String(),
		Denom:      ft.denomV0WithoutFeatures,
		IbcEnabled: true,
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(nonIssuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(upgradeMsg)),
		upgradeMsg,
	)
	requireT.ErrorContains(err, "unauthorized")

	ftClient := assetfttypes.NewQueryClient(chain.ClientContext)
	resp, err := ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ft.denomV0WithoutFeatures,
	})
	requireT.NoError(err)
	requireT.EqualValues(0, resp.Token.Version)
	requireT.Len(resp.Token.Features, 0)

	upgradeMsg = &assetfttypes.MsgUpgradeTokenV1{
		Sender:     nonIssuer.String(),
		Denom:      ft.denomV0WithFeatures,
		IbcEnabled: false,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(nonIssuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(upgradeMsg)),
		upgradeMsg,
	)
	requireT.ErrorContains(err, "unauthorized")

	resp, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ft.denomV0WithFeatures,
	})
	requireT.NoError(err)
	requireT.EqualValues(0, resp.Token.Version)
	requireT.Equal([]assetfttypes.Feature{
		assetfttypes.Feature_minting,
		assetfttypes.Feature_freezing,
		assetfttypes.Feature_whitelisting,
		assetfttypes.Feature_burning,
	}, resp.Token.Features)
}

func (ft *ftTest) upgradeFromV0ToV1ToDisableIBC(t *testing.T) {
	requireT := require.New(t)
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	// upgrading with disabled IBC should take effect immediately
	upgradeMsg := &assetfttypes.MsgUpgradeTokenV1{
		Sender:     ft.issuer.String(),
		Denom:      ft.denomV0WithoutFeatures,
		IbcEnabled: false,
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(ft.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(upgradeMsg)),
		upgradeMsg,
	)
	requireT.NoError(err)

	ftClient := assetfttypes.NewQueryClient(chain.ClientContext)
	resp, err := ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ft.denomV0WithoutFeatures,
	})
	requireT.NoError(err)
	requireT.EqualValues(1, resp.Token.Version)
	requireT.Len(resp.Token.Features, 0)

	// upgrading second time should fail
	upgradeMsg.IbcEnabled = true
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(ft.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(upgradeMsg)),
		upgradeMsg,
	)
	requireT.ErrorContains(err, fmt.Sprintf("denom %s has been already upgraded to v1", ft.denomV0WithoutFeatures))

	resp, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ft.denomV0WithoutFeatures,
	})
	requireT.NoError(err)
	requireT.EqualValues(1, resp.Token.Version)
	requireT.Len(resp.Token.Features, 0)
}

func (ft *ftTest) upgradeFromV0ToV1ToEnableIBC(t *testing.T) {
	requireT := require.New(t)
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	// setting grace period to some small value
	const gracePeriod = 15 * time.Second
	chain.Governance.UpdateParams(ctx, t, "Propose changing TokenUpgradeGracePeriod in the assetft module",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(assetfttypes.ModuleName, string(assetfttypes.KeyTokenUpgradeGracePeriod), string(must.Bytes(tmjson.Marshal(gracePeriod)))),
		})

	ftClient := assetfttypes.NewQueryClient(chain.ClientContext)
	ftParams, err := ftClient.Params(ctx, &assetfttypes.QueryParamsRequest{})
	requireT.NoError(err)
	requireT.Equal(gracePeriod, ftParams.Params.TokenUpgradeGracePeriod)

	// upgrading with enabled IBC should take effect after delay
	upgradeMsg := &assetfttypes.MsgUpgradeTokenV1{
		Sender:     ft.issuer.String(),
		Denom:      ft.denomV0WithFeatures,
		IbcEnabled: true,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(ft.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(upgradeMsg)),
		upgradeMsg,
	)
	requireT.NoError(err)

	// ensure that token hasn't been upgraded yet
	resp, err := ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ft.denomV0WithFeatures,
	})
	requireT.NoError(err)
	requireT.EqualValues(0, resp.Token.Version)
	requireT.Equal([]assetfttypes.Feature{
		assetfttypes.Feature_minting,
		assetfttypes.Feature_freezing,
		assetfttypes.Feature_whitelisting,
		assetfttypes.Feature_burning,
	}, resp.Token.Features)

	// upgrading second time should fail
	upgradeMsg = &assetfttypes.MsgUpgradeTokenV1{
		Sender:     ft.issuer.String(),
		Denom:      ft.denomV0WithFeatures,
		IbcEnabled: false,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(ft.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(upgradeMsg)),
		upgradeMsg,
	)
	requireT.ErrorContains(err, fmt.Sprintf("token upgrade is already pending for denom %q", ft.denomV0WithFeatures))

	resp, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ft.denomV0WithFeatures,
	})
	requireT.NoError(err)
	requireT.EqualValues(0, resp.Token.Version)
	requireT.Equal([]assetfttypes.Feature{
		assetfttypes.Feature_minting,
		assetfttypes.Feature_freezing,
		assetfttypes.Feature_whitelisting,
		assetfttypes.Feature_burning,
	}, resp.Token.Features)

	select {
	case <-ctx.Done():
		return
	case <-time.After(gracePeriod + 2*time.Second):
	}

	// ibc should be enabled now
	resp, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ft.denomV0WithFeatures,
	})
	requireT.NoError(err)
	requireT.EqualValues(1, resp.Token.Version)
	requireT.Equal([]assetfttypes.Feature{
		assetfttypes.Feature_minting,
		assetfttypes.Feature_freezing,
		assetfttypes.Feature_whitelisting,
		assetfttypes.Feature_burning,
		assetfttypes.Feature_ibc,
	}, resp.Token.Features)

	// following upgrade should fail again
	upgradeMsg = &assetfttypes.MsgUpgradeTokenV1{
		Sender:     ft.issuer.String(),
		Denom:      ft.denomV0WithFeatures,
		IbcEnabled: false,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(ft.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(upgradeMsg)),
		upgradeMsg,
	)
	requireT.ErrorContains(err, fmt.Sprintf("denom %s has been already upgraded to v1", ft.denomV0WithFeatures))

	resp, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ft.denomV0WithFeatures,
	})
	requireT.NoError(err)
	requireT.EqualValues(1, resp.Token.Version)
	requireT.Equal([]assetfttypes.Feature{
		assetfttypes.Feature_minting,
		assetfttypes.Feature_freezing,
		assetfttypes.Feature_whitelisting,
		assetfttypes.Feature_burning,
		assetfttypes.Feature_ibc,
	}, resp.Token.Features)
}

func (ft *ftTest) tryToUpgradeV0ToV1AfterDecisionTimeout(t *testing.T) {
	requireT := require.New(t)
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	// setting decision timeout to sth in the past
	decisionTimeout := time.Now().UTC().Add(-time.Hour)
	chain.Governance.UpdateParams(ctx, t, "Propose changing TokenUpgradeDecisionTimeout in the assetft module",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(assetfttypes.ModuleName, string(assetfttypes.KeyTokenUpgradeDecisionTimeout), string(must.Bytes(tmjson.Marshal(decisionTimeout)))),
		})

	ftClient := assetfttypes.NewQueryClient(chain.ClientContext)
	ftParams, err := ftClient.Params(ctx, &assetfttypes.QueryParamsRequest{})
	requireT.NoError(err)
	requireT.Equal(decisionTimeout, ftParams.Params.TokenUpgradeDecisionTimeout)

	// upgrade after timeout should fail
	upgradeMsg := &assetfttypes.MsgUpgradeTokenV1{
		Sender:     ft.issuer.String(),
		Denom:      ft.denomV0ForForbiddenUpgrades,
		IbcEnabled: false,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(ft.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(upgradeMsg)),
		upgradeMsg,
	)
	requireT.ErrorContains(err, "it is no longer possible to upgrade the token")

	resp, err := ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ft.denomV0ForForbiddenUpgrades,
	})
	requireT.NoError(err)
	requireT.EqualValues(0, resp.Token.Version)
	requireT.Len(resp.Token.Features, 0)

	upgradeMsg.IbcEnabled = true
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(ft.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(upgradeMsg)),
		upgradeMsg,
	)
	requireT.ErrorContains(err, "it is no longer possible to upgrade the token")

	resp, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ft.denomV0ForForbiddenUpgrades,
	})
	requireT.NoError(err)
	requireT.EqualValues(0, resp.Token.Version)
	requireT.Len(resp.Token.Features, 0)
}

func getIssueFee(ctx context.Context, t *testing.T, clientCtx client.Context) sdk.Coin {
	queryClient := assetfttypes.NewQueryClient(clientCtx)
	resp, err := queryClient.Params(ctx, &assetfttypes.QueryParamsRequest{})
	require.NoError(t, err)

	return resp.Params.IssueFee
}
