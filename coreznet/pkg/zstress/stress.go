package zstress

import (
	"context"
	"fmt"
	"math/big"
	"runtime"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum-tools/pkg/pace"
	"github.com/CoreumFoundation/coreum-tools/pkg/parallel"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
	retry "github.com/avast/retry-go"
)

// StressConfig contains config for benchmarking the blockchain
type StressConfig struct {
	// ChainID is the ID of the chain to connect to
	ChainID string

	// NodeAddress is the address of a cored node RPC endpoint, in the form of host:port, to connect to
	NodeAddress string

	// Accounts is the list of private keys used to send transactions during benchmark
	Accounts []cored.Secp256k1PrivateKey

	// NumOfTransactions to send from each account
	NumOfTransactions int
}

type tx struct {
	AccountIndex int
	TxIndex      int
	From         cored.Wallet
	To           cored.Wallet
	TxBytes      []byte
}

// Stress runs a benchmark test
func Stress(ctx context.Context, config StressConfig) error {
	numOfAccounts := len(config.Accounts)
	log := logger.Get(ctx)
	client := cored.NewClient(config.ChainID, config.NodeAddress)

	startTs := time.Now()
	signedTxPace := pace.New(ctx, "signed tx", 10*time.Second, pace.ZapReporter(log))
	getAccountNumberSequencePace := pace.New(ctx, "sequence fetched", 10*time.Second, pace.ZapReporter(log))

	log.Info(
		"Preparing signed transactions...",
		zap.Int("num", numOfAccounts*config.NumOfTransactions),
	)

	var signedTxs [][][]byte
	var initialAccountSequences []uint64
	err := parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		queue := make(chan tx)
		results := make(chan tx)
		for i := 0; i < runtime.NumCPU(); i++ {
			spawn(fmt.Sprintf("signer-%d", i), parallel.Continue, func(ctx context.Context) error {
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case tx, ok := <-queue:
						if !ok {
							return nil
						}
						tx.TxBytes = must.Bytes(client.PrepareTxBankSend(tx.From, tx.To, cored.Balance{Amount: big.NewInt(1), Denom: "core"}))
						select {
						case <-ctx.Done():
							return ctx.Err()
						case results <- tx:
						}
					}
				}
			})
		}
		spawn("enqueue", parallel.Continue, func(ctx context.Context) error {
			defer func() {
				getAccountNumberSequencePace.Stop()
			}()

			initialAccountSequences = make([]uint64, 0, numOfAccounts)

			for i := 0; i < numOfAccounts; i++ {
				fromPrivateKey := config.Accounts[i]
				toPrivateKeyIndex := i + 1
				if toPrivateKeyIndex >= numOfAccounts {
					toPrivateKeyIndex = 0
				}
				toPrivateKey := config.Accounts[toPrivateKeyIndex]

				accNum, accSeq, err := getAccountNumberSequence(ctx, client, fromPrivateKey.Address())
				if err != nil {
					return errors.WithStack(fmt.Errorf("fetching account number and sequence failed: %w", err))
				}

				getAccountNumberSequencePace.Step(1)
				initialAccountSequences = append(initialAccountSequences, accSeq)

				tx := tx{
					AccountIndex: i,
					From:         cored.Wallet{Name: "sender", Key: fromPrivateKey, AccountNumber: accNum, AccountSequence: accSeq},
					To:           cored.Wallet{Name: "receiver", Key: toPrivateKey},
				}

				for j := 0; j < config.NumOfTransactions; j++ {
					tx.TxIndex = j
					select {
					case <-ctx.Done():
						return ctx.Err()
					case queue <- tx:
					}
					tx.From.AccountSequence++
				}
			}
			return nil
		})
		spawn("integrate", parallel.Exit, func(ctx context.Context) error {
			defer func() {
				signedTxPace.Stop()
			}()

			signedTxs = make([][][]byte, numOfAccounts)
			for i := 0; i < numOfAccounts; i++ {
				signedTxs[i] = make([][]byte, config.NumOfTransactions)
			}
			for i := 0; i < numOfAccounts; i++ {
				for j := 0; j < config.NumOfTransactions; j++ {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case result := <-results:
						signedTxs[result.AccountIndex][result.TxIndex] = result.TxBytes
						signedTxPace.Step(1)
					}
				}
			}
			return nil
		})
		return nil
	})
	if err != nil {
		return err
	}
	log.Info("Transactions prepared")

	broadcastTxPace := pace.New(ctx, "sent tx", 10*time.Second, pace.ZapReporter(log))

	log.Info("Broadcasting transactions...")
	err = parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		spawn("accounts", parallel.Exit, func(ctx context.Context) error {
			return parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
				for i, accountTxs := range signedTxs {
					accountTxs := accountTxs

					accountClient := cored.NewClient(config.ChainID, config.NodeAddress)

					spawn(fmt.Sprintf("account-%d", i), parallel.Continue, func(ctx context.Context) error {
						for txIndex := 0; txIndex < config.NumOfTransactions; {
							tx := accountTxs[txIndex]
							txHash, err := accountClient.Broadcast(ctx, tx)
							if err != nil {
								return err
							}

							broadcastTxPace.Step(1)
							log.Debug("Transaction broadcasted", zap.String("txHash", txHash))
							txIndex++
						}
						return nil
					})
				}
				return nil
			})
		})
		return nil
	})
	if err != nil {
		return err
	}

	broadcastTxPace.Stop()
	log.Info(
		"Benchmark finished",
		zap.Duration("dur", time.Since(startTs)),
	)

	return nil
}

func getAccountNumberSequence(ctx context.Context, client cored.Client, accountAddress string) (uint64, uint64, error) {
	log := logger.Get(ctx)

	var accNum, accSeq uint64

	err := retry.Do(func() error {
		var err error
		accNum, accSeq, err = client.GetNumberSequence(accountAddress)
		if err != nil {
			log.With(zap.Error(err)).Warn("error while GetNumberSequence")
			return errors.Wrap(err, "querying for account number and sequence failed")
		}

		return nil
	},
		retry.Context(ctx),
		retry.Attempts(10),
		retry.MaxDelay(5*time.Second),
	)
	if err != nil {
		return 0, 0, err
	}

	return accNum, accSeq, nil
}
