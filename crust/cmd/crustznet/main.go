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

	"github.com/CoreumFoundation/coreum/crust/infra"
	"github.com/CoreumFoundation/coreum/crust/pkg/znet"
)

func main() {
	run.Tool("crustznet", znet.IoC, func(c *ioc.Container, configF *znet.ConfigFactory, cmdF *znet.CmdFactory) error {
		rootCmd := &cobra.Command{
			SilenceUsage:  true,
			SilenceErrors: true,
			Short:         "Creates preconfigured session for environment",
			RunE:          cmdF.Cmd(znet.Activate),
		}
		logger.AddFlags(logger.ToolDefaultConfig, rootCmd.PersistentFlags())
		rootCmd.PersistentFlags().StringVar(&configF.EnvName, "env", defaultString("CRUSTZNET_ENV", "crustznet"), "Name of the environment to run in")
		rootCmd.PersistentFlags().StringVar(&configF.Target, "target", defaultString("CRUSTZNET_TARGET", "docker"), "Target of the deployment: "+strings.Join(c.Names((*infra.Target)(nil)), " | "))
		rootCmd.PersistentFlags().StringVar(&configF.HomeDir, "home", defaultString("CRUSTZNET_HOME", must.String(os.UserCacheDir())+"/crustznet"), "Directory where all files created automatically by crustznet are stored")
		addFlags(rootCmd, configF)
		addModeFlag(rootCmd, c, configF)
		addFilterFlag(rootCmd, configF)

		startCmd := &cobra.Command{
			Use:   "start",
			Short: "Starts environment",
			RunE:  cmdF.Cmd(znet.Start),
		}
		addFlags(startCmd, configF)
		addModeFlag(startCmd, c, configF)
		rootCmd.AddCommand(startCmd)

		stopCmd := &cobra.Command{
			Use:   "stop",
			Short: "Stops environment",
			RunE:  cmdF.Cmd(znet.Stop),
		}
		rootCmd.AddCommand(stopCmd)

		removeCmd := &cobra.Command{
			Use:   "remove",
			Short: "Removes environment",
			RunE:  cmdF.Cmd(znet.Remove),
		}
		rootCmd.AddCommand(removeCmd)

		testCmd := &cobra.Command{
			Use:   "test",
			Short: "Runs integration tests",
			RunE:  cmdF.Cmd(znet.Test),
		}
		addFlags(testCmd, configF)
		addFilterFlag(testCmd, configF)
		rootCmd.AddCommand(testCmd)

		specCmd := &cobra.Command{
			Use:   "spec",
			Short: "Prints specification of running environment",
			RunE:  cmdF.Cmd(znet.Spec),
		}
		addModeFlag(specCmd, c, configF)
		rootCmd.AddCommand(specCmd)

		pingPongCmd := &cobra.Command{
			Use:   "ping-pong",
			Short: "Sends tokens back and forth to generate transactions",
			RunE:  cmdF.Cmd(znet.PingPong),
		}
		addModeFlag(pingPongCmd, c, configF)
		rootCmd.AddCommand(pingPongCmd)

		stressCmd := &cobra.Command{
			Use:   "stress",
			Short: "Runs the logic used by crustzstress to test benchmarking",
			RunE:  cmdF.Cmd(znet.Stress),
		}
		addModeFlag(stressCmd, c, configF)
		rootCmd.AddCommand(stressCmd)

		return rootCmd.Execute()
	})
}

func addFlags(cmd *cobra.Command, configF *znet.ConfigFactory) {
	cmd.Flags().StringVar(&configF.BinDir, "bin-dir", defaultString("CRUSTZNET_BIN_DIR", filepath.Dir(must.String(filepath.EvalSymlinks(must.String(os.Executable()))))), "Path to directory where executables exist")
}

func addModeFlag(cmd *cobra.Command, c *ioc.Container, configF *znet.ConfigFactory) {
	cmd.Flags().StringVar(&configF.ModeName, "mode", defaultString("CRUSTZNET_MODE", "dev"), "List of applications to deploy: "+strings.Join(c.Names((*infra.Mode)(nil)), " | "))
}

func addFilterFlag(cmd *cobra.Command, configF *znet.ConfigFactory) {
	cmd.Flags().StringArrayVar(&configF.TestFilters, "filter", defaultFilters("CRUSTZNET_FILTERS"), "Regular expression used to filter tests to run")
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
