package staking

import "github.com/CoreumFoundation/coreum/integration-tests/testing"

// SingleChainTests returns single chain tests of the staking module
func SingleChainTests() []testing.SingleChainSignature {
	return []testing.SingleChainSignature{
		TestDelegate,
		TestUndelegate,
		TestCreateValidator,
	}
}
