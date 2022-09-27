package gov

import "github.com/CoreumFoundation/coreum/integration-tests/testing"

// SingleChainTests returns single chain tests of the gov module
func SingleChainTests() []testing.SingleChainSignature {
	return []testing.SingleChainSignature{
		TestProposalParamChange,
	}
}
