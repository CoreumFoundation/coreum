package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/CoreumFoundation/coreum-tools/pkg/ioc"
	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum-tools/pkg/run"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/coreznet"
	"github.com/CoreumFoundation/coreum/coreznet/infra"
)

func main() {
	run.Tool("coreznet", coreznet.IoC, func(c *ioc.Container, configF *coreznet.ConfigFactory, cmdF *coreznet.CmdFactory) error {
		rootCmd := &cobra.Command{
			SilenceUsage: true,
			Short:        "Creates preconfigured session for environment",
			RunE:         cmdF.Cmd(coreznet.Activate),
		}
		logger.AddFlags(logger.ToolDefaultConfig, rootCmd.PersistentFlags())
		rootCmd.PersistentFlags().StringVar(&configF.EnvName, "env", defaultString("COREZNET_ENV", "coreznet"), "Name of the environment to run in")
		rootCmd.PersistentFlags().StringVar(&configF.Target, "target", defaultString("COREZNET_TARGET", "tmux"), "Target of the deployment: "+strings.Join(c.Names((*infra.Target)(nil)), " | "))
		rootCmd.PersistentFlags().StringVar(&configF.HomeDir, "home", defaultString("COREZNET_HOME", must.String(os.UserCacheDir())+"/coreznet"), "Directory where all files created automatically by coreznet are stored")
		addFlags(rootCmd, configF)
		addModeFlag(rootCmd, c, configF)
		addFilterFlag(rootCmd, configF)

		startCmd := &cobra.Command{
			Use:   "start",
			Short: "Starts environment",
			RunE:  cmdF.Cmd(coreznet.Start),
		}
		addFlags(startCmd, configF)
		addModeFlag(startCmd, c, configF)
		rootCmd.AddCommand(startCmd)

		stopCmd := &cobra.Command{
			Use:   "stop",
			Short: "Stops environment",
			RunE:  cmdF.Cmd(coreznet.Stop),
		}
		rootCmd.AddCommand(stopCmd)

		removeCmd := &cobra.Command{
			Use:   "remove",
			Short: "Removes environment",
			RunE:  cmdF.Cmd(coreznet.Remove),
		}
		rootCmd.AddCommand(removeCmd)

		testCmd := &cobra.Command{
			Use:   "test",
			Short: "Runs integration tests",
			RunE:  cmdF.Cmd(coreznet.Test),
		}
		addFlags(testCmd, configF)
		addFilterFlag(testCmd, configF)
		rootCmd.AddCommand(testCmd)

		specCmd := &cobra.Command{
			Use:   "spec",
			Short: "Prints specification of running environment",
			RunE:  cmdF.Cmd(coreznet.Spec),
		}
		addModeFlag(specCmd, c, configF)
		rootCmd.AddCommand(specCmd)

		pingPongCmd := &cobra.Command{
			Use:   "ping-pong",
			Short: "Sends tokens back and forth to generate transactions",
			RunE:  cmdF.Cmd(coreznet.PingPong),
		}
		addModeFlag(pingPongCmd, c, configF)
		rootCmd.AddCommand(pingPongCmd)

		return rootCmd.Execute()
	})
}

func addFlags(cmd *cobra.Command, configF *coreznet.ConfigFactory) {
	cmd.Flags().StringVar(&configF.BinDir, "bin-dir", defaultString("COREZNET_BIN_DIR", filepath.Dir(must.String(filepath.EvalSymlinks(must.String(os.Executable()))))), "Path to directory where executables exist")
}

func addModeFlag(cmd *cobra.Command, c *ioc.Container, configF *coreznet.ConfigFactory) {
	cmd.Flags().StringVar(&configF.ModeName, "mode", defaultString("COREZNET_MODE", "dev"), "List of applications to deploy: "+strings.Join(c.Names((*infra.Mode)(nil)), " | "))
}

func addFilterFlag(cmd *cobra.Command, configF *coreznet.ConfigFactory) {
	cmd.Flags().StringArrayVar(&configF.TestFilters, "filter", defaultFilters("COREZNET_FILTERS"), "Regular expression used to filter tests to run")
}

func defaultString(env, def string) string {
	val := os.Getenv(env)
	if val == "" {
		val = def
	}
	return val
}

func defaultFilters(env string) []string {
	val := os.Getenv(env)
	if val == "" {
		return nil
	}
	return strings.Split(val, ",")
}
