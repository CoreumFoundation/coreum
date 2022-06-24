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
	run.Tool("crustznet", znet.IoC, func(c *ioc.Container, configF *infra.ConfigFactory, cmdF *znet.CmdFactory) error {
		rootCmd := &cobra.Command{
			SilenceUsage:  true,
			SilenceErrors: true,
			Short:         "Creates preconfigured session for environment",
			RunE:          cmdF.Cmd(znet.Activate),
		}
		logger.AddFlags(logger.ToolDefaultConfig, rootCmd.PersistentFlags())
		rootCmd.PersistentFlags().StringVar(&configF.EnvName, "env", defaultString("CRUSTZNET_ENV", "crustznet"), "Name of the environment to run in")
		rootCmd.PersistentFlags().StringVar(&configF.HomeDir, "home", defaultString("CRUSTZNET_HOME", must.String(os.UserCacheDir())+"/crustznet"), "Directory where all files created automatically by crustznet are stored")
		addBinDirFlag(rootCmd, configF)
		addTargetFlag(rootCmd, c, configF)
		addModeFlag(rootCmd, c, configF)
		addFilterFlag(rootCmd, configF)

		startCmd := &cobra.Command{
			Use:   "start",
			Short: "Starts environment",
			RunE:  cmdF.Cmd(znet.Start),
		}
		addBinDirFlag(startCmd, configF)
		addTargetFlag(startCmd, c, configF)
		addModeFlag(startCmd, c, configF)
		rootCmd.AddCommand(startCmd)

		rootCmd.AddCommand(&cobra.Command{
			Use:   "stop",
			Short: "Stops environment",
			RunE:  cmdF.Cmd(znet.Stop),
		})

		rootCmd.AddCommand(&cobra.Command{
			Use:   "remove",
			Short: "Removes environment",
			RunE:  cmdF.Cmd(znet.Remove),
		})

		testCmd := &cobra.Command{
			Use:   "test",
			Short: "Runs integration tests",
			RunE:  cmdF.Cmd(znet.Test),
		}
		addBinDirFlag(testCmd, configF)
		addTargetFlag(testCmd, c, configF)
		addFilterFlag(testCmd, configF)
		rootCmd.AddCommand(testCmd)

		rootCmd.AddCommand(&cobra.Command{
			Use:   "spec",
			Short: "Prints specification of running environment",
			RunE:  cmdF.Cmd(znet.Spec),
		})

		rootCmd.AddCommand(&cobra.Command{
			Use:   "ping-pong",
			Short: "Sends tokens back and forth to generate transactions",
			RunE:  cmdF.Cmd(znet.PingPong),
		})

		rootCmd.AddCommand(&cobra.Command{
			Use:   "stress",
			Short: "Runs the logic used by crustzstress to test benchmarking",
			RunE:  cmdF.Cmd(znet.Stress),
		})

		return rootCmd.Execute()
	})
}

func addBinDirFlag(cmd *cobra.Command, configF *infra.ConfigFactory) {
	cmd.Flags().StringVar(&configF.BinDir, "bin-dir", defaultString("CRUSTZNET_BIN_DIR", filepath.Dir(must.String(filepath.EvalSymlinks(must.String(os.Executable()))))), "Path to directory where executables exist")
}

func addTargetFlag(cmd *cobra.Command, c *ioc.Container, configF *infra.ConfigFactory) {
	cmd.Flags().StringVar(&configF.Target, "target", defaultString("CRUSTZNET_TARGET", "docker"), "Target of the deployment: "+strings.Join(c.Names((*infra.Target)(nil)), " | "))
}

func addModeFlag(cmd *cobra.Command, c *ioc.Container, configF *infra.ConfigFactory) {
	cmd.Flags().StringVar(&configF.ModeName, "mode", defaultString("CRUSTZNET_MODE", "dev"), "List of applications to deploy: "+strings.Join(c.Names((*infra.Mode)(nil)), " | "))
}

func addFilterFlag(cmd *cobra.Command, configF *infra.ConfigFactory) {
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
