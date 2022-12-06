package wstaking

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/stretchr/testify/require"
)

// TestAppModuleOriginalStakingModule_GetConsensusVersion checks that the wrapped module still uses the save consensus version.
func TestAppModuleOriginalStakingModule_GetConsensusVersion(t *testing.T) {
	stakingModule := staking.NewAppModule(&codec.AminoCodec{}, stakingkeeper.Keeper{}, authkeeper.AccountKeeper{}, bankkeeper.BaseKeeper{})
	require.Equal(t, uint64(2), stakingModule.ConsensusVersion())
}
