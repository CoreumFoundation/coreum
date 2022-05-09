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
	"github.com/spf13/pflag"

	"github.com/CoreumFoundation/coreum/coreznet"
	"github.com/CoreumFoundation/coreum/coreznet/infra"
)

func main() {
	verbose := defaultBool("COREZNET_VERBOSE", false)
	if len(os.Args) > 1 {
		flags := pflag.NewFlagSet("verbose", pflag.ContinueOnError)
		flags.ParseErrorsWhitelist.UnknownFlags = true
		flags.BoolVarP(&verbose, "verbose", "v", verbose, "Turns on verbose logging")
		// Dummy flag to turn off printing usage of this flag set
		flags.BoolP("help", "h", false, "")

		_ = flags.Parse(os.Args[1:])
	}

	if !verbose {
		logger.VerboseOff()
	}

	run.Tool("coreznet", coreznet.IoC, func(c *ioc.Container, configF *coreznet.ConfigFactory, cmdF *coreznet.CmdFactory) error {
		rootCmd := &cobra.Command{
			SilenceUsage: true,
			Short:        "Creates preconfigured bash session for environment",
			RunE:         cmdF.Cmd(coreznet.Activate),
		}
		rootCmd.PersistentFlags().StringVar(&configF.EnvName, "env", defaultString("COREZNET_ENV", "coreznet"), "Name of the environment to run in")
		rootCmd.PersistentFlags().StringVar(&configF.Target, "target", defaultString("COREZNET_TARGET", "tmux"), "Target of the deployment: "+strings.Join(c.Names((*infra.Target)(nil)), " | "))
		rootCmd.PersistentFlags().StringVar(&configF.HomeDir, "home", defaultString("COREZNET_HOME", must.String(os.UserCacheDir())+"/coreznet"), "Directory where all files created automatically by coreznet are stored")
		rootCmd.PersistentFlags().BoolVarP(&configF.VerboseLogging, "verbose", "v", defaultBool("COREZNET_VERBOSE", false), "Turns on verbose logging")
		addFlags(rootCmd, configF)
		addSetFlag(rootCmd, c, configF)
		addFilterFlag(rootCmd, configF)

		startCmd := &cobra.Command{
			Use:   "start",
			Short: "Starts environment",
			RunE:  cmdF.Cmd(coreznet.Start),
		}
		addFlags(startCmd, configF)
		addSetFlag(startCmd, c, configF)
		rootCmd.AddCommand(startCmd)

		stopCmd := &cobra.Command{
			Use:   "stop",
			Short: "Stops environment",
			RunE:  cmdF.Cmd(coreznet.Stop),
		}
		rootCmd.AddCommand(stopCmd)

		destroyCmd := &cobra.Command{
			Use:   "destroy",
			Short: "Destroys environment",
			RunE:  cmdF.Cmd(coreznet.Destroy),
		}
		rootCmd.AddCommand(destroyCmd)

		testsCmd := &cobra.Command{
			Use:   "tests",
			Short: "Runs integration tests",
			RunE:  cmdF.Cmd(coreznet.Tests),
		}
		addFlags(testsCmd, configF)
		addFilterFlag(testsCmd, configF)
		rootCmd.AddCommand(testsCmd)

		specCmd := &cobra.Command{
			Use:   "spec",
			Short: "Prints specification of running environment",
			RunE:  cmdF.Cmd(coreznet.Spec),
		}
		addSetFlag(specCmd, c, configF)
		rootCmd.AddCommand(specCmd)

		return rootCmd.Execute()
	})
}

func addFlags(cmd *cobra.Command, configF *coreznet.ConfigFactory) {
	cmd.Flags().StringVar(&configF.BinDir, "bin-dir", defaultString("COREZNET_BIN_DIR", filepath.Dir(must.String(filepath.EvalSymlinks(must.String(os.Executable()))))), "Path to directory where executables exist")
	cmd.Flags().StringVar(&configF.Network, "network", defaultString("COREZNET_NETWORK", "127.1.0.0"), "Network where IPs for applications are taken from (related to 'tmux' and 'direct' targets only)")
}

func addSetFlag(cmd *cobra.Command, c *ioc.Container, configF *coreznet.ConfigFactory) {
	cmd.Flags().StringVar(&configF.SetName, "set", defaultString("COREZNET_SET", "dev"), "Application set to deploy: "+strings.Join(c.Names((*infra.Set)(nil)), " | "))
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

func defaultBool(env string, def bool) bool {
	switch os.Getenv(env) {
	case "1", "true", "True", "TRUE":
		return true
	case "0", "false", "False", "FALSE":
		return false
	default:
		return def
	}
}

func defaultFilters(env string) []string {
	val := os.Getenv(env)
	if val == "" {
		return nil
	}
	return strings.Split(val, ",")
}
