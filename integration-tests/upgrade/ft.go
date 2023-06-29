//go:build integrationtests

package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	modules "github.com/CoreumFoundation/coreum/integration-tests/modules/testdata/wasm"
	"github.com/CoreumFoundation/coreum/pkg/client"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

type ftMethod string

const (
	ftMethodUpgradeTokenV1 ftMethod = "upgrade_token_v1"
	gracePeriod                     = 15 * time.Second
)

//nolint:tagliatelle
type ibcEnabledBodyFTRequest struct {
	IbcEnabled bool `json:"ibc_enabled"`
}

// fungible token wasm models
//
//nolint:tagliatelle
type issueFTRequest struct {
	Symbol             string                 `json:"symbol"`
	Subunit            string                 `json:"subunit"`
	Precision          uint32                 `json:"precision"`
	InitialAmount      string                 `json:"initial_amount"`
	Description        string                 `json:"description"`
	Features           []assetfttypes.Feature `json:"features"`
	BurnRate           string                 `json:"burn_rate"`
	SendCommissionRate string                 `json:"send_commission_rate"`
}

type ftV1UpgradeTest struct {
	issuer                         sdk.AccAddress
	contractAddressWithFeatures    string
	contractAddressWithoutFeatures string
	denomV0WithoutFeatures         string
	denomV0WithFeatures            string
	denomV0ForForbiddenUpgrades    string
	denomV0WasmWithFeatures        string
	denomV0WasmWithoutFeatures     string
}

func (ft *ftV1UpgradeTest) Before(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	ft.issuer = chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, ft.issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgIssue{},
			&assetfttypes.MsgIssue{},
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
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount.MulRaw(8),
	})

	ft.issueV0TokenWithoutFeatures(t)
	ft.issueV0TokenWithFeatures(t)
	ft.issueV0TokenWithoutFeaturesWASM(t)
	ft.issueV0TokenWithFeaturesWASM(t)
	ft.tryToUpgradeTokenFromV0ToV1BeforeUpgradingTheApp(t)
}

func (ft *ftV1UpgradeTest) issueV0TokenWithoutFeatures(t *testing.T) {
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

func (ft *ftV1UpgradeTest) issueV0TokenWithFeatures(t *testing.T) {
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

func (ft *ftV1UpgradeTest) issueV0TokenWithFeaturesWASM(t *testing.T) {
	requireT := require.New(t)
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	// ********** Issuance **********

	burnRate := sdk.MustNewDecFromStr("0.0")
	sendCommissionRate := sdk.MustNewDecFromStr("0.0")

	issuanceAmount := sdk.NewInt(1000)
	issuanceReq := issueFTRequest{
		Symbol:        "symbol",
		Subunit:       "subunit",
		Precision:     6,
		InitialAmount: issuanceAmount.String(),
		Description:   "my wasm fungible token",
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_minting,
			assetfttypes.Feature_burning,
			assetfttypes.Feature_freezing,
			assetfttypes.Feature_whitelisting,
		},
		BurnRate:           burnRate.String(),
		SendCommissionRate: sendCommissionRate.String(),
	}
	issuerFTInstantiatePayload, err := json.Marshal(issuanceReq)
	requireT.NoError(err)

	// instantiate new contract
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)
	ft.contractAddressWithFeatures, _, err = chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		ft.issuer,
		modules.FTWASM,
		integrationtests.InstantiateConfig{
			// we add the initial amount to let the contract issue the token on behalf of it
			Amount:     getIssueFee(ctx, t, chain.ClientContext),
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    issuerFTInstantiatePayload,
			Label:      "fungible_token",
		},
	)
	requireT.NoError(err)

	ft.denomV0WasmWithFeatures = assetfttypes.BuildDenom(issuanceReq.Subunit, sdk.MustAccAddressFromBech32(ft.contractAddressWithFeatures))
}

func (ft *ftV1UpgradeTest) issueV0TokenWithoutFeaturesWASM(t *testing.T) {
	requireT := require.New(t)
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	// ********** Issuance **********

	burnRate := sdk.MustNewDecFromStr("0.0")
	sendCommissionRate := sdk.MustNewDecFromStr("0.0")

	issuanceAmount := sdk.NewInt(1000)
	issuanceReq := issueFTRequest{
		Symbol:             "symbol",
		Subunit:            "subunit",
		Precision:          6,
		InitialAmount:      issuanceAmount.String(),
		Description:        "my wasm fungible token",
		BurnRate:           burnRate.String(),
		SendCommissionRate: sendCommissionRate.String(),
	}
	issuerFTInstantiatePayload, err := json.Marshal(issuanceReq)
	requireT.NoError(err)

	// instantiate new contract
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)
	ft.contractAddressWithoutFeatures, _, err = chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		ft.issuer,
		modules.FTWASM,
		integrationtests.InstantiateConfig{
			// we add the initial amount to let the contract issue the token on behalf of it
			Amount:     getIssueFee(ctx, t, chain.ClientContext),
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    issuerFTInstantiatePayload,
			Label:      "fungible_token",
		},
	)
	requireT.NoError(err)

	ft.denomV0WasmWithoutFeatures = assetfttypes.BuildDenom(issuanceReq.Subunit, sdk.MustAccAddressFromBech32(ft.contractAddressWithoutFeatures))
}

func (ft *ftV1UpgradeTest) tryToUpgradeTokenFromV0ToV1BeforeUpgradingTheApp(t *testing.T) {
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

func (ft *ftV1UpgradeTest) After(t *testing.T) {
	ft.tryToUpgradeV1TokenToEnableIBC(t)
	ft.tryToUpgradeV1TokenToDisableIBC(t)
	ft.tryToUpgradeV0ToV1ByNonIssuer(t)

	ft.changeGracePeriod(t)

	ft.upgradeFromV0ToV1ToDisableIBC(t)
	ft.upgradeFromV0ToV1ToEnableIBC(t)
	ft.upgradeFromV0ToV1ToDisableIBCWASM(t)
	ft.upgradeFromV0ToV1ToEnableIBCWASM(t)
	ft.tryToUpgradeV0ToV1AfterDecisionTimeout(t)
}

func (ft *ftV1UpgradeTest) tryToUpgradeV1TokenToEnableIBC(t *testing.T) {
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

func (ft *ftV1UpgradeTest) tryToUpgradeV1TokenToDisableIBC(t *testing.T) {
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

func (ft *ftV1UpgradeTest) tryToUpgradeV0ToV1ByNonIssuer(t *testing.T) {
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

func (ft *ftV1UpgradeTest) upgradeFromV0ToV1ToDisableIBC(t *testing.T) {
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

//nolint:funlen // there are many tests
func (ft *ftV1UpgradeTest) upgradeFromV0ToV1ToEnableIBC(t *testing.T) {
	requireT := require.New(t)
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

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

func (ft *ftV1UpgradeTest) upgradeFromV0ToV1ToDisableIBCWASM(t *testing.T) {
	requireT := require.New(t)
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	txf := chain.TxFactory().WithSimulateAndExecute(true)

	// upgrading with disabled IBC should take effect immediately
	upgradePayload, err := json.Marshal(map[ftMethod]ibcEnabledBodyFTRequest{
		ftMethodUpgradeTokenV1: {
			IbcEnabled: false,
		},
	})
	requireT.NoError(err)
	ftClient := assetfttypes.NewQueryClient(chain.ClientContext)
	resp1, err := ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ft.denomV0WasmWithFeatures,
	})
	requireT.NoError(err)
	requireT.EqualValues(0, resp1.Token.Version)
	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, ft.issuer, ft.contractAddressWithFeatures, upgradePayload, sdk.Coin{})
	requireT.NoError(err)

	resp, err := ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ft.denomV0WasmWithFeatures,
	})
	requireT.NoError(err)
	requireT.EqualValues(1, resp.Token.Version)
	requireT.Equal([]assetfttypes.Feature{
		assetfttypes.Feature_minting,
		assetfttypes.Feature_burning,
		assetfttypes.Feature_freezing,
		assetfttypes.Feature_whitelisting,
	}, resp.Token.Features)

	// upgrading second time should fail
	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, ft.issuer, ft.contractAddressWithFeatures, upgradePayload, sdk.Coin{})
	requireT.ErrorContains(err, fmt.Sprintf("denom %s has been already upgraded to v1", ft.denomV0WasmWithFeatures))
}

func (ft *ftV1UpgradeTest) upgradeFromV0ToV1ToEnableIBCWASM(t *testing.T) {
	requireT := require.New(t)
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	txf := chain.TxFactory().WithSimulateAndExecute(true)

	// upgrading with enabled IBC should take effect after delay
	upgradePayload, err := json.Marshal(map[ftMethod]ibcEnabledBodyFTRequest{
		ftMethodUpgradeTokenV1: {
			IbcEnabled: true,
		},
	})
	requireT.NoError(err)
	ftClient := assetfttypes.NewQueryClient(chain.ClientContext)
	resp1, err := ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ft.denomV0WasmWithoutFeatures,
	})
	requireT.NoError(err)
	requireT.EqualValues(0, resp1.Token.Version)
	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, ft.issuer, ft.contractAddressWithoutFeatures, upgradePayload, sdk.Coin{})
	requireT.NoError(err)

	// ensure that token hasn't been upgraded yet
	resp, err := ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ft.denomV0WasmWithoutFeatures,
	})
	requireT.NoError(err)
	requireT.EqualValues(0, resp.Token.Version)

	// upgrading second time should fail
	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, ft.issuer, ft.contractAddressWithoutFeatures, upgradePayload, sdk.Coin{})
	requireT.ErrorContains(err, fmt.Sprintf("token upgrade is already pending for denom %q", ft.denomV0WasmWithoutFeatures))

	select {
	case <-ctx.Done():
		return
	case <-time.After(gracePeriod + 2*time.Second):
	}

	// token should be upgraded now
	resp, err = ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ft.denomV0WasmWithoutFeatures,
	})
	requireT.NoError(err)
	requireT.EqualValues(1, resp.Token.Version)
	requireT.Equal([]assetfttypes.Feature{
		assetfttypes.Feature_ibc,
	}, resp.Token.Features)

	// following upgrade should fail again
	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, ft.issuer, ft.contractAddressWithoutFeatures, upgradePayload, sdk.Coin{})
	requireT.ErrorContains(err, fmt.Sprintf("denom %s has been already upgraded to v1", ft.denomV0WasmWithoutFeatures))
}

func (ft *ftV1UpgradeTest) tryToUpgradeV0ToV1AfterDecisionTimeout(t *testing.T) {
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

func (ft *ftV1UpgradeTest) changeGracePeriod(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	chain.Governance.UpdateParams(ctx, t, "Propose changing TokenUpgradeGracePeriod in the assetft module",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(assetfttypes.ModuleName, string(assetfttypes.KeyTokenUpgradeGracePeriod), string(must.Bytes(tmjson.Marshal(gracePeriod)))),
		})
}

type ftFeaturesTest struct {
	denom string
}

func (ft *ftFeaturesTest) Before(t *testing.T) {
	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	issuer := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "AAA",
		Subunit:       "uaaa",
		Precision:     6,
		Description:   "AAA Description",
		InitialAmount: sdk.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_minting,
			assetfttypes.Feature_burning,
			assetfttypes.Feature_ibc, // should be removed by the migration
			assetfttypes.Feature_freezing,
			1000,                              // should be removed by the migration
			assetfttypes.Feature_whitelisting, // should be removed by the migration
			assetfttypes.Feature_minting,      // should be removed by the migration
			assetfttypes.Feature_burning,      // should be removed by the migration
			assetfttypes.Feature_ibc,          // should be removed by the migration
			assetfttypes.Feature_freezing,     // should be removed by the migration
			2000,                              // should be removed by the migration
			1000,                              // should be removed by the migration
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)
	ft.denom = assetfttypes.BuildDenom(issueMsg.Subunit, issuer)
}

func (ft *ftFeaturesTest) After(t *testing.T) {
	ft.verifyTokenIsFixed(t)
	ft.tryCreatingTokenWithInvalidFeature(t)
	ft.tryCreatingTokenWithDuplicatedFeature(t)
	ft.createValidToken(t)
}

func (ft *ftFeaturesTest) verifyTokenIsFixed(t *testing.T) {
	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	ftClient := assetfttypes.NewQueryClient(chain.ClientContext)
	resp, err := ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: ft.denom,
	})
	requireT.NoError(err)

	requireT.Equal([]assetfttypes.Feature{
		assetfttypes.Feature_minting,
		assetfttypes.Feature_burning,
		assetfttypes.Feature_freezing,
		assetfttypes.Feature_whitelisting,
	}, resp.Token.Features)
}

func (ft *ftFeaturesTest) tryCreatingTokenWithInvalidFeature(t *testing.T) {
	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	issuer := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "AAA",
		Subunit:       "uaaa",
		Precision:     6,
		Description:   "AAA Description",
		InitialAmount: sdk.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_minting,
			assetfttypes.Feature_burning,
			assetfttypes.Feature_ibc,
			assetfttypes.Feature_freezing,
			100,
			assetfttypes.Feature_whitelisting,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.ErrorContains(err, "invalid input")
}

func (ft *ftFeaturesTest) tryCreatingTokenWithDuplicatedFeature(t *testing.T) {
	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	issuer := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "AAA",
		Subunit:       "uaaa",
		Precision:     6,
		Description:   "AAA Description",
		InitialAmount: sdk.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_minting,
			assetfttypes.Feature_burning,
			assetfttypes.Feature_ibc,
			assetfttypes.Feature_freezing,
			assetfttypes.Feature_whitelisting,
			assetfttypes.Feature_ibc,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.ErrorContains(err, "invalid input")
}

func (ft *ftFeaturesTest) createValidToken(t *testing.T) {
	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	issuer := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
		},
		Amount: getIssueFee(ctx, t, chain.ClientContext).Amount,
	})

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "AAA",
		Subunit:       "uaaa",
		Precision:     6,
		Description:   "AAA Description",
		InitialAmount: sdk.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_minting,
			assetfttypes.Feature_burning,
			assetfttypes.Feature_ibc,
			assetfttypes.Feature_freezing,
			assetfttypes.Feature_whitelisting,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)
}

func getIssueFee(ctx context.Context, t *testing.T, clientCtx client.Context) sdk.Coin {
	queryClient := assetfttypes.NewQueryClient(clientCtx)
	resp, err := queryClient.Params(ctx, &assetfttypes.QueryParamsRequest{})
	require.NoError(t, err)

	return resp.Params.IssueFee
}
