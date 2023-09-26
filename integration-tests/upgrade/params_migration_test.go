//go:build integrationtests

package upgrade

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypesproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	"github.com/CoreumFoundation/coreum/v3/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/v3/x/asset/nft/types"
	customparams "github.com/CoreumFoundation/coreum/v3/x/customparams/types"
	feemodeltypes "github.com/CoreumFoundation/coreum/v3/x/feemodel/types"
)

type paramsMigrationTest struct {
	assetftParamsBeforeMigration     assetfttypes.Params
	assetnftParamsBeforeMigration    assetnfttypes.Params
	feemodelParamsBeforeMigration    feemodeltypes.Params
	customparamsStaking              customparams.StakingParams
	consensusParamsBeforeMigration   tmtypes.ConsensusParams
	authParamsBeforeMigration        authtypes.Params
	bankParamsBeforeMigration        banktypes.Params
	stakingParamsBeforeMigration     stakingtypes.Params
	distrParamsBeforeMigration       distrtypes.Params
	slashingParamsBeforeMigration    slashingtypes.Params
	govParamsBeforeMigration         govtypesv1beta1.Params
	mintParamsBeforeMigration        minttypes.Params
	ibcTransferParamsBeforeMigration ibctransfertypes.Params
	wasmParamsBeforeMigration        wasmtypes.Params
}

func (pmt *paramsMigrationTest) Before(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	// assetft
	assetftResp, err := assetfttypes.NewQueryClient(chain.ClientContext).Params(ctx, &assetfttypes.QueryParamsRequest{})
	requireT.NoError(err)
	pmt.assetftParamsBeforeMigration = assetftResp.Params

	// assetnft
	assetnftResp, err := assetnfttypes.NewQueryClient(chain.ClientContext).Params(ctx, &assetnfttypes.QueryParamsRequest{})
	requireT.NoError(err)
	pmt.assetnftParamsBeforeMigration = assetnftResp.Params

	// feemodel
	feemodelResp, err := feemodeltypes.NewQueryClient(chain.ClientContext).Params(ctx, &feemodeltypes.QueryParamsRequest{})
	requireT.NoError(err)
	pmt.feemodelParamsBeforeMigration = feemodelResp.Params

	// customparams
	customparamsStakingResp, err := customparams.NewQueryClient(chain.ClientContext).StakingParams(ctx, &customparams.QueryStakingParamsRequest{})
	requireT.NoError(err)
	pmt.customparamsStaking = customparamsStakingResp.Params

	// crisis is skipped since it doesn't expose query in neither v45 nor v47.

	// consensus params are queried directly from the x/params module because the query is not implemented in v45.
	pmt.consensusParamsBeforeMigration = queryConsensusParams(ctx, t, chain)

	// gov
	govParams, err := chain.LegacyGovernance.QueryGovParams(ctx)
	requireT.NoError(err)
	pmt.govParamsBeforeMigration = govParams

	// auth
	authResp, err := authtypes.NewQueryClient(chain.ClientContext).Params(ctx, &authtypes.QueryParamsRequest{})
	requireT.NoError(err)
	pmt.authParamsBeforeMigration = authResp.Params

	// bank
	bankResp, err := banktypes.NewQueryClient(chain.ClientContext).Params(ctx, &banktypes.QueryParamsRequest{})
	requireT.NoError(err)
	pmt.bankParamsBeforeMigration = bankResp.Params

	// staking
	stakingResp, err := stakingtypes.NewQueryClient(chain.ClientContext).Params(ctx, &stakingtypes.QueryParamsRequest{})
	requireT.NoError(err)
	pmt.stakingParamsBeforeMigration = stakingResp.Params

	// distribution
	distrResp, err := distrtypes.NewQueryClient(chain.ClientContext).Params(ctx, &distrtypes.QueryParamsRequest{})
	requireT.NoError(err)
	pmt.distrParamsBeforeMigration = distrResp.Params

	// slashing
	slashingResp, err := slashingtypes.NewQueryClient(chain.ClientContext).Params(ctx, &slashingtypes.QueryParamsRequest{})
	requireT.NoError(err)
	pmt.slashingParamsBeforeMigration = slashingResp.Params

	// mint
	mintResp, err := minttypes.NewQueryClient(chain.ClientContext).Params(ctx, &minttypes.QueryParamsRequest{})
	requireT.NoError(err)
	pmt.mintParamsBeforeMigration = mintResp.Params

	// ibc/transfer
	ibcTransferResp, err := ibctransfertypes.NewQueryClient(chain.ClientContext).Params(ctx, &ibctransfertypes.QueryParamsRequest{})
	requireT.NoError(err)
	pmt.ibcTransferParamsBeforeMigration = *ibcTransferResp.Params

	// wasm
	wasmResp, err := wasmtypes.NewQueryClient(chain.ClientContext).Params(ctx, &wasmtypes.QueryParamsRequest{})
	requireT.NoError(err)
	pmt.wasmParamsBeforeMigration = wasmResp.Params
}

func (pmt *paramsMigrationTest) After(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	assertT := assert.New(t)
	// assetft
	assetftResp, err := assetfttypes.NewQueryClient(chain.ClientContext).Params(ctx, &assetfttypes.QueryParamsRequest{})
	requireT.NoError(err)
	assertT.Equal(pmt.assetftParamsBeforeMigration, assetftResp.Params)

	// assetnft
	assetnftResp, err := assetnfttypes.NewQueryClient(chain.ClientContext).Params(ctx, &assetnfttypes.QueryParamsRequest{})
	requireT.NoError(err)
	assertT.Equal(pmt.assetnftParamsBeforeMigration, assetnftResp.Params)

	// feemodel
	feemodelResp, err := feemodeltypes.NewQueryClient(chain.ClientContext).Params(ctx, &feemodeltypes.QueryParamsRequest{})
	requireT.NoError(err)
	assertT.Equal(pmt.feemodelParamsBeforeMigration, feemodelResp.Params)

	// customparams
	customparamsStakingResp, err := customparams.NewQueryClient(chain.ClientContext).StakingParams(ctx, &customparams.QueryStakingParamsRequest{})
	requireT.NoError(err)
	assertT.Equal(pmt.customparamsStaking.MinSelfDelegation.String(), customparamsStakingResp.Params.MinSelfDelegation.String())
	assertT.Equal(pmt.customparamsStaking, customparamsStakingResp.Params)

	// crisis is skipped since it doesn't expose query in neither v45 nor v47.

	// consensus
	consensusResp, err := consensustypes.NewQueryClient(chain.ClientContext).Params(ctx, &consensustypes.QueryParamsRequest{})
	requireT.NoError(err)
	assertT.Equal(pmt.consensusParamsBeforeMigration, *consensusResp.Params)

	// gov
	govParams, err := chain.Governance.QueryGovParams(ctx)
	requireT.NoError(err)
	// Since gov params were fully restructured, we need to initialize new struct type with values from old one to compare.
	paramsBeforeMigration := &govtypesv1.Params{
		MinDeposit:                 pmt.govParamsBeforeMigration.DepositParams.MinDeposit,
		MaxDepositPeriod:           &pmt.govParamsBeforeMigration.DepositParams.MaxDepositPeriod,
		VotingPeriod:               &pmt.govParamsBeforeMigration.VotingParams.VotingPeriod,
		Quorum:                     pmt.govParamsBeforeMigration.TallyParams.Quorum.String(),
		Threshold:                  pmt.govParamsBeforeMigration.TallyParams.Threshold.String(),
		VetoThreshold:              pmt.govParamsBeforeMigration.TallyParams.VetoThreshold.String(),
		MinInitialDepositRatio:     sdk.ZeroDec().String(),
		BurnVoteQuorum:             false,
		BurnProposalDepositPrevote: false,
		BurnVoteVeto:               true,
	}
	assertT.Equal(paramsBeforeMigration, govParams)

	// auth
	authResp, err := authtypes.NewQueryClient(chain.ClientContext).Params(ctx, &authtypes.QueryParamsRequest{})
	requireT.NoError(err)
	assertT.Equal(pmt.authParamsBeforeMigration, authResp.Params)

	// bank
	bankResp, err := banktypes.NewQueryClient(chain.ClientContext).Params(ctx, &banktypes.QueryParamsRequest{})
	requireT.NoError(err)
	assertT.Equal(pmt.bankParamsBeforeMigration, bankResp.Params)

	// staking
	stakingResp, err := stakingtypes.NewQueryClient(chain.ClientContext).Params(ctx, &stakingtypes.QueryParamsRequest{})
	requireT.NoError(err)
	assertT.True(pmt.stakingParamsBeforeMigration.MinCommissionRate.IsNil()) // Not present in v45.
	// Override new field to compare the rest of the params.
	pmt.stakingParamsBeforeMigration.MinCommissionRate = sdk.ZeroDec()
	assertT.Equal(pmt.stakingParamsBeforeMigration, stakingResp.Params)

	// distribution
	distrResp, err := distrtypes.NewQueryClient(chain.ClientContext).Params(ctx, &distrtypes.QueryParamsRequest{})
	requireT.NoError(err)
	assertT.False(pmt.distrParamsBeforeMigration.BaseProposerReward.IsZero())  //nolint:staticcheck
	assertT.False(pmt.distrParamsBeforeMigration.BonusProposerReward.IsZero()) //nolint:staticcheck
	assertT.True(distrResp.Params.BaseProposerReward.IsZero())                 //nolint:staticcheck
	assertT.True(distrResp.Params.BonusProposerReward.IsZero())                //nolint:staticcheck
	// Override the deprecated fields to compare the rest of the params.
	pmt.distrParamsBeforeMigration.BaseProposerReward = sdk.ZeroDec()  //nolint:staticcheck
	pmt.distrParamsBeforeMigration.BonusProposerReward = sdk.ZeroDec() //nolint:staticcheck
	assertT.Equal(pmt.distrParamsBeforeMigration, distrResp.Params)

	// slashing
	slashingResp, err := slashingtypes.NewQueryClient(chain.ClientContext).Params(ctx, &slashingtypes.QueryParamsRequest{})
	requireT.NoError(err)
	assertT.Equal(pmt.slashingParamsBeforeMigration, slashingResp.Params)

	// mint
	mintResp, err := minttypes.NewQueryClient(chain.ClientContext).Params(ctx, &minttypes.QueryParamsRequest{})
	requireT.NoError(err)
	assertT.Equal(pmt.mintParamsBeforeMigration, mintResp.Params)

	// ibc/transfer
	ibcTransferResp, err := ibctransfertypes.NewQueryClient(chain.ClientContext).Params(ctx, &ibctransfertypes.QueryParamsRequest{})
	requireT.NoError(err)
	assertT.Equal(pmt.ibcTransferParamsBeforeMigration, *ibcTransferResp.Params)

	// wasm
	wasmResp, err := wasmtypes.NewQueryClient(chain.ClientContext).Params(ctx, &wasmtypes.QueryParamsRequest{})
	requireT.NoError(err)
	assertT.Equal(pmt.wasmParamsBeforeMigration, wasmResp.Params)
}

func querySubspaceParamsValue(ctx context.Context, t *testing.T, chain integration.CoreumChain, subspace, key string) string {
	res, err := paramstypesproposal.NewQueryClient(chain.ClientContext).Params(
		ctx,
		&paramstypesproposal.QueryParamsRequest{Subspace: subspace, Key: key},
	)
	require.NoError(t, err)
	return res.Param.Value
}

func queryConsensusParams(ctx context.Context, t *testing.T, chain integration.CoreumChain) tmtypes.ConsensusParams {
	requireT := require.New(t)

	blockParamsJSONStr := querySubspaceParamsValue(ctx, t, chain, "baseapp", "BlockParams")
	evidenceParamsJSONStr := querySubspaceParamsValue(ctx, t, chain, "baseapp", "EvidenceParams")
	validatorParamsJSONStr := querySubspaceParamsValue(ctx, t, chain, "baseapp", "ValidatorParams")

	//nolint:tagliatelle
	blockParams := struct {
		MaxBytes string `json:"max_bytes"`
		MaxGas   string `json:"max_gas"`
	}{}
	requireT.NoError(json.Unmarshal([]byte(blockParamsJSONStr), &blockParams))

	//nolint:tagliatelle
	evidenceParams := struct {
		MaxAgeNumBlocks string `json:"max_age_num_blocks"`
		MaxAgeDuration  string `json:"max_age_duration"`
		MaxBytes        string `json:"max_bytes"`
	}{}
	requireT.NoError(json.Unmarshal([]byte(evidenceParamsJSONStr), &evidenceParams))

	//nolint:tagliatelle
	validatorParams := struct {
		PubKeyTypes []string `json:"pub_key_types"`
	}{}
	requireT.NoError(json.Unmarshal([]byte(validatorParamsJSONStr), &validatorParams))

	blockMaxBytes, err := strconv.ParseInt(blockParams.MaxBytes, 10, 64)
	requireT.NoError(err)
	blockMaxGas, err := strconv.ParseInt(blockParams.MaxGas, 10, 64)
	requireT.NoError(err)
	evidenceMaxAgeNumBlocks, err := strconv.ParseInt(evidenceParams.MaxAgeNumBlocks, 10, 64)
	requireT.NoError(err)
	evidenceMaxAgeDuration, err := time.ParseDuration(evidenceParams.MaxAgeDuration + "ns") // the value returned is in nanoseconds.
	requireT.NoError(err)
	evidenceMaxBytes, err := strconv.ParseInt(evidenceParams.MaxBytes, 10, 64)
	requireT.NoError(err)

	return tmtypes.ConsensusParams{
		Block: &tmtypes.BlockParams{
			MaxBytes: blockMaxBytes,
			MaxGas:   blockMaxGas,
		},
		Evidence: &tmtypes.EvidenceParams{
			MaxAgeNumBlocks: evidenceMaxAgeNumBlocks,
			MaxAgeDuration:  evidenceMaxAgeDuration,
			MaxBytes:        evidenceMaxBytes,
		},
		Validator: &tmtypes.ValidatorParams{
			PubKeyTypes: validatorParams.PubKeyTypes,
		},
		Version: nil, // not present neither in v45 nor v47.
	}
}
