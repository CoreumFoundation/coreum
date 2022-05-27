package main

import (
	"os"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum-tools/pkg/run"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/coreznet/cmd"
	"github.com/CoreumFoundation/coreum/coreznet/pkg/zstress"
)

func main() {
	cmd.ConfigureLoggerWithCLI(false)
	run.Tool("corezstress", nil, func() error {
		rootCmd := &cobra.Command{
			SilenceUsage: true,
			Short:        "Run benchmark test",
			RunE: func(cmd *cobra.Command, args []string) error {
				panic("not implemented")
			},
		}
		// dummy flag to match the one added by logger configurator
		rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Turns on verbose logging")

		var generateConfig zstress.GenerateConfig
		generateCmd := &cobra.Command{
			Use:   "generate",
			Short: "Generates all the files required to deploy the blockchain used for benchmarking",
			RunE: func(cmd *cobra.Command, args []string) error {
				return zstress.Generate(generateConfig)
			},
		}
		generateCmd.Flags().IntVar(&generateConfig.NumOfValidators, "validators", 16, "Number of validators present on the blockchain")
		generateCmd.Flags().IntVar(&generateConfig.NumOfInstances, "instances", 32, "The maximum number of application instances used in the future during benchmarking")
		generateCmd.Flags().IntVar(&generateConfig.NumOfAccountsPerInstance, "accounts", 1000, "The maximum number of funded accounts per each instance used in the future during benchmarking")
		generateCmd.Flags().StringVar(&generateConfig.OutDirectory, "out", must.String(os.Getwd()), "Path to the directory where generated files are stored")
		rootCmd.AddCommand(generateCmd)

		return rootCmd.Execute()
	})
}
