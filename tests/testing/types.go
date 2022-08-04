package testing

import (
	"context"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/client"
)

// T is an interface wrapper around *testing.T
type T interface {
	Errorf(format string, args ...interface{})
	FailNow()
}

// PrepareFunc defines function which is executed before environment is deployed
type PrepareFunc func(ctx context.Context) error

// RunFunc defines function which is responsible for running the test
type RunFunc func(ctx context.Context, t T)

// Chain holds network and client for the blockchain
type Chain struct {
	Network *app.Network
	Client  client.Client
}

// SingleChainSignature is the signature of test function accepting a chain
type SingleChainSignature func(chain Chain) (PrepareFunc, RunFunc)

// TestSet is a container for tests
type TestSet struct {
	SingleChain []SingleChainSignature
}
