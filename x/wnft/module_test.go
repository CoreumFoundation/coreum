package wnft

import (
	"testing"

	"github.com/stretchr/testify/require"

	nft "github.com/CoreumFoundation/coreum/v3/x/nft/module"
)

// TestOriginalNFTModuleConsensusVersion tests the original nft module has not increased its consensus version
// if this tests fails, it means that we need to register the new migration handlers of the original nft module.
func TestNFTModuleConsensusVersion(t *testing.T) {
	nftModule := nft.AppModule{}
	require.EqualValues(t, 1, nftModule.ConsensusVersion())
}
