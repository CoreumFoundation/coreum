package main

import (
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

		generateCmd := &cobra.Command{
			Use:   "generate",
			Short: "Generates all the files required to deploy the blockchain used for benchmarking",
			RunE: func(cmd *cobra.Command, args []string) error {
				return zstress.Generate()
			},
		}
		rootCmd.AddCommand(generateCmd)

		return rootCmd.Execute()
	})
}
