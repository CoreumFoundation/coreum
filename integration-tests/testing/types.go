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

// Faucet defines an interface to fund testing accounts
type Faucet interface {
	FundAccounts(ctx context.Context, accountsToFund ...FundedAccount) error
}

// Chain holds network and client for the blockchain
type Chain struct {
	NetworkConfig app.NetworkConfig
	Client        client.Client
	Faucet        Faucet
}

// FundedAccount represents a requirement of a test to get some funds for an account
type FundedAccount struct {
	Wallet types.Wallet
	Amount types.Coin
}

// SingleChainSignature is the signature of test function accepting a chain
type SingleChainSignature func(ctx context.Context, t T, chain Chain)

// TestSet is a container for tests
type TestSet struct {
	SingleChain []SingleChainSignature
}
