package testing

import (
	"context"

	"github.com/stretchr/testify/require"
)

// T is an interface representing test, accepted by `assert.*` and `require.*` packages
type T interface {
	require.TestingT
}

// SingleChainSignature is the signature of test function accepting a chain
type SingleChainSignature func(ctx context.Context, t T, chain Chain)

// TestSet is a container for tests
type TestSet struct {
	SingleChain []SingleChainSignature
}
