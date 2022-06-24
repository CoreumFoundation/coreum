package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum-tools/pkg/run"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/crust/pkg/zstress"
)

const (
	defaultChainID       = "corestress"
	defaultNumOfAccounts = 100
)

func main() {
	run.Tool("zstress", nil, func(ctx context.Context) error {
		var stressConfig zstress.StressConfig
		var accountFile string
		var numOfAccounts int
		rootCmd := &cobra.Command{
			SilenceUsage:  true,
			SilenceErrors: true,
			Short:         "Run benchmark test",
			RunE: func(cmd *cobra.Command, args []string) error {
				if numOfAccounts <= 0 {
					return errors.New("number of accounts must be greater than 0")
				}

				keysRaw, err := ioutil.ReadFile(accountFile)
				if err != nil {
					return errors.WithStack(fmt.Errorf("reading account file failed: %w", err))
				}
				if err := json.Unmarshal(keysRaw, &stressConfig.Accounts); err != nil {
					return errors.WithStack(fmt.Errorf("parsing account file failed: %w", err))
				}

				if numOfAccounts > len(stressConfig.Accounts) {
					return errors.New("number of accounts is greater than the number of provided private keys")
				}
				stressConfig.Accounts = stressConfig.Accounts[:numOfAccounts]
				return zstress.Stress(ctx, stressConfig)
			},
		}
		logger.AddFlags(logger.ToolDefaultConfig, rootCmd.PersistentFlags())
		rootCmd.Flags().StringVar(&stressConfig.ChainID, "chain-id", defaultChainID, "ID of the chain to connect to")
		rootCmd.Flags().StringVar(&stressConfig.NodeAddress, "node-addr", "localhost:26657", "Address of a cored node RPC endpoint, in the form of host:port, to connect to")
		rootCmd.Flags().StringVar(&accountFile, "account-file", "", "Path to a JSON file containing private keys of accounts funded on blockchain")
		rootCmd.Flags().IntVar(&numOfAccounts, "accounts", defaultNumOfAccounts, "Number of accounts used to benchmark the node in parallel, must not be greater than the number of keys available in account file")
		rootCmd.Flags().IntVar(&stressConfig.NumOfTransactions, "transactions", 1000, "Number of transactions to send from each account")

		var generateConfig zstress.GenerateConfig
		generateCmd := &cobra.Command{
			Use:   "generate",
			Short: "Generates all the files required to deploy the blockchain used for benchmarking",
			RunE: func(cmd *cobra.Command, args []string) error {
				return zstress.Generate(generateConfig)
			},
		}
		generateCmd.Flags().StringVar(&generateConfig.ChainID, "chain-id", defaultChainID, "ID of the chain to generate")
		generateCmd.Flags().IntVar(&generateConfig.NumOfValidators, "validators", 32, "Number of validators present on the blockchain")
		generateCmd.Flags().IntVar(&generateConfig.NumOfSentryNodes, "sentry-nodes", 4, "Number of sentry nodes to generate config for")
		generateCmd.Flags().IntVar(&generateConfig.NumOfInstances, "instances", 8, "The maximum number of application instances used in the future during benchmarking")
		generateCmd.Flags().IntVar(&generateConfig.NumOfAccountsPerInstance, "accounts", defaultNumOfAccounts, "The maximum number of funded accounts per each instance used in the future during benchmarking")
		generateCmd.Flags().StringVar(&generateConfig.BinDirectory, "bin-dir", filepath.Dir(filepath.Dir(must.String(filepath.EvalSymlinks(must.String(os.Executable()))))), "Path to the directory where binaries exist")
		generateCmd.Flags().StringVar(&generateConfig.OutDirectory, "out-dir", must.String(os.Getwd()), "Path to the directory where generated files are stored")
		rootCmd.AddCommand(generateCmd)

		return rootCmd.Execute()
	})
}
