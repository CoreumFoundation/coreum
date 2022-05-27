package zstress

import (
	"context"
	"errors"
	"math/big"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum-tools/pkg/parallel"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
)

// StressConfig contains config for benchmarking the blockchain
type StressConfig struct {
	// ChainID is the ID of the chain to connect to
	ChainID string

	// NodeAddress is the address of a cored node RPC endpoint, in the form of host:port, to connect to
	NodeAddress string

	// Accounts is the list of private keys used to send transactions during benchmark
	Accounts []cored.Secp256k1PrivateKey

	// NumOfAccounts is the number of accounts used to benchmark the node in parallel
	NumOfAccounts int

	// NumOfTransactions to send from each account
	NumOfTransactions int
}

// Stress runs a benchmark test
func Stress(ctx context.Context, config StressConfig) error {
	if config.NumOfAccounts <= 0 {
		return errors.New("number of accounts must be greater than 0")
	}
	if config.NumOfAccounts > len(config.Accounts) {
		return errors.New("number of accounts is greater than the number of provided private keys")
	}
	return parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		client := cored.NewClient(config.ChainID, config.NodeAddress)
		for i := 0; i < config.NumOfAccounts; i++ {
			fromPrivateKey := config.Accounts[i]
			toPrivateKeyIndex := i + 1
			spawn(fromPrivateKey.Address(), parallel.Continue, func(ctx context.Context) error {
				if toPrivateKeyIndex >= config.NumOfAccounts {
					toPrivateKeyIndex = 0
				}
				toPrivateKey := config.Accounts[toPrivateKeyIndex]

				fromWallet := cored.Wallet{Name: "sender", Key: fromPrivateKey}
				toWallet := cored.Wallet{Name: "receiver", Key: toPrivateKey}

				log := logger.Get(ctx)
				for i := 0; i < config.NumOfTransactions; i++ {
					if _, err := client.Broadcast(must.Bytes(client.TxBankSend(fromWallet, toWallet, cored.Balance{Amount: big.NewInt(1), Denom: "core"}))); err != nil {
						log.Error("Sending transaction failed", zap.Error(err))
					}
				}
				return nil
			})
		}
		return nil
	})
}
