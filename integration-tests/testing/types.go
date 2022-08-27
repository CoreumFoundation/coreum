package testing

import (
	"context"

	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// T is an interface representing test, accepted by `assert.*` and `require.*` packages
type T interface {
	require.TestingT
}

// RunFunc defines function which is responsible for running the test
type RunFunc = func(ctx context.Context, t T)

// Chain holds network and client for the blockchain
type Chain struct {
	NetworkConfig app.NetworkConfig
	Client        client.Client
}

// Prerequisites reresent requirements of a test which must be met before it can be started
type Prerequisites struct {
	FundedAccounts []FundedAccount
}

// FundedAccount represents a requirement of a test to get some funds for an account
type FundedAccount struct {
	Wallet types.Wallet
	Amount types.Coin
}

// SingleChainSignature is the signature of test function accepting a chain
type SingleChainSignature func(chain Chain) (Prerequisites, RunFunc, error)

// TestSet is a container for tests
type TestSet struct {
	SingleChain []SingleChainSignature
}
