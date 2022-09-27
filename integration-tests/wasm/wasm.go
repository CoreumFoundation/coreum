package wasm

import "github.com/CoreumFoundation/coreum/integration-tests/testing"

// SingleChainTests returns single chain tests of the wasm module
func SingleChainTests() []testing.SingleChainSignature {
	return []testing.SingleChainSignature{
		TestSimpleStateWasmContract,
		TestBankSendWasmContract,
	}
}
