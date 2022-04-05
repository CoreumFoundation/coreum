package cmd

import (
	"dexapp/app"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "dexd",
		Short: "Coreum core app",
	}

	// Set config for prefixes
	app.SetConfig()

	return rootCmd
}
