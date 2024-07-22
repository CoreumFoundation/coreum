package wstaking

import (
	"testing"

	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/require"
)

// TestAppModuleOriginalStakingModule_GetConsensusVersion checks that the wrapped module still uses the save
// consensus version.
func TestAppModuleOriginalStakingModule_GetConsensusVersion(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}).Codec
	stakingModule := staking.NewAppModule(
		cdc, nil, authkeeper.AccountKeeper{}, bankkeeper.BaseKeeper{}, nil,
	)
	require.Equal(t, uint64(4), stakingModule.ConsensusVersion())
}
