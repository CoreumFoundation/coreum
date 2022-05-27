package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum-tools/pkg/run"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/coreznet/cmd"
	"github.com/CoreumFoundation/coreum/coreznet/pkg/zstress"
)

const (
	defaultChainID       = "corestress"
	defaultNumOfAccounts = 1000
)

func main() {
	cmd.ConfigureLoggerWithCLI(false)
	run.Tool("corezstress", nil, func(ctx context.Context) error {
		var stressConfig zstress.StressConfig
		var accountFile string
		rootCmd := &cobra.Command{
			SilenceUsage: true,
			Short:        "Run benchmark test",
			RunE: func(cmd *cobra.Command, args []string) error {
				keysRaw, err := ioutil.ReadFile(accountFile)
				if err != nil {
					return fmt.Errorf("reading account file failed: %w", err)
				}

				if err := json.Unmarshal(keysRaw, &stressConfig.Accounts); err != nil {
					return fmt.Errorf("parsing account file failed: %w", err)
				}
				return zstress.Stress(ctx, stressConfig)
			},
		}
		// dummy flag to match the one added by logger configurator
		rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Turns on verbose logging")
		rootCmd.Flags().StringVar(&stressConfig.ChainID, "chain-id", defaultChainID, "ID of the chain to connect to")
		rootCmd.Flags().StringVar(&stressConfig.NodeAddress, "node-addr", "localhost:26657", "Address of a cored node RPC endpoint, in the form of host:port, to connect to")
		rootCmd.Flags().StringVar(&accountFile, "account-file", "", "Path to a JSON file containing private keys of accounts funded on blockchain")
		rootCmd.Flags().IntVar(&stressConfig.NumOfAccounts, "accounts", defaultNumOfAccounts, "Number of accounts used to benchmark the node in parallel, must not be greater than the number of keys available in account file")
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
		generateCmd.Flags().IntVar(&generateConfig.NumOfValidators, "validators", 16, "Number of validators present on the blockchain")
		generateCmd.Flags().IntVar(&generateConfig.NumOfInstances, "instances", 32, "The maximum number of application instances used in the future during benchmarking")
		generateCmd.Flags().IntVar(&generateConfig.NumOfAccountsPerInstance, "accounts", defaultNumOfAccounts, "The maximum number of funded accounts per each instance used in the future during benchmarking")
		generateCmd.Flags().StringVar(&generateConfig.OutDirectory, "out", must.String(os.Getwd()), "Path to the directory where generated files are stored")
		rootCmd.AddCommand(generateCmd)

		return rootCmd.Execute()
	})
}
