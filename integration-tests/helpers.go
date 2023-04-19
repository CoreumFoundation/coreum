package integrationtests

import "testing"

// SkipUnsafe will skip the tests that are not safe to run against a real running chain.
// unsafe tests can only be run against a locally running chain since they modify parameters
// of the chain.
func SkipUnsafe(t *testing.T) {
	if !cfg.RunUnsafe {
		t.SkipNow()
	}
}
