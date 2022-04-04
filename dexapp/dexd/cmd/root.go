package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "dexd",
		Short: "Coreum core app",
	}

	return rootCmd
}
