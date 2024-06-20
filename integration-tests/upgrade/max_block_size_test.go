//go:build integrationtests

package upgrade

import (
	"testing"

	"github.com/cometbft/cometbft/proto/tendermint/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
)

type maxBlockSizeTest struct {
	params *types.ConsensusParams
}

func (mbst *maxBlockSizeTest) Before(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	consensusClient := consensustypes.NewQueryClient(chain.ClientContext)
	consensusParams, err := consensusClient.Params(ctx, &consensustypes.QueryParamsRequest{})
	requireT.NoError(err)
	requireT.EqualValues(22_020_096, consensusParams.Params.Block.MaxBytes)
	mbst.params = consensusParams.Params
	mbst.params.Block.MaxBytes = 6_291_456
}

func (mbst *maxBlockSizeTest) After(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	consensusClient := consensustypes.NewQueryClient(chain.ClientContext)
	consensusParams, err := consensusClient.Params(ctx, &consensustypes.QueryParamsRequest{})
	requireT.NoError(err)
	requireT.Equal(mbst.params, consensusParams.Params)
}
