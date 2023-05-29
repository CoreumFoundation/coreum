package integrationtests

import "testing"

// SkipUnsafe skips the test if run-unsafe flag is set to false.
func SkipUnsafe(t *testing.T) {
	if !runUnsafe {
		t.SkipNow()
	}
}
