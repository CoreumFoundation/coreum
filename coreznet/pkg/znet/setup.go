package znet

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/CoreumFoundation/coreum-tools/pkg/ioc"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps"
	"github.com/CoreumFoundation/coreum/coreznet/infra/targets"
)

// IoC configures IoC container
func IoC(c *ioc.Container) {
	c.Singleton(NewCmdFactory)
	c.Singleton(NewConfigFactory)
	c.Singleton(infra.NewSpec)
	c.Transient(func(configF *ConfigFactory) infra.Config {
		return configF.Config()
	})
	c.Transient(apps.NewFactory)
	c.TransientNamed("dev", DevMode)
	c.TransientNamed("test", TestMode)
	c.Transient(func(c *ioc.Container, config infra.Config) infra.Mode {
		var mode infra.Mode
		c.ResolveNamed(config.ModeName, &mode)
		return mode
	})
	c.Transient(targets.NewDocker)
	c.TransientNamed("tmux", targets.NewTMux)
	c.TransientNamed("docker", func(docker *targets.Docker) infra.Target {
		return docker
	})
	c.Transient(func(c *ioc.Container, config infra.Config) infra.Target {
		var target infra.Target
		c.ResolveNamed(config.Target, &target)
		return target
	})
}

// NewCmdFactory returns new CmdFactory
func NewCmdFactory(c *ioc.Container) *CmdFactory {
	return &CmdFactory{
		c: c,
	}
}

// CmdFactory is a wrapper around cobra RunE
type CmdFactory struct {
	c *ioc.Container
}

// Cmd returns function compatible with RunE
func (f *CmdFactory) Cmd(cmdFunc interface{}) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var err error
		f.c.Call(cmdFunc, &err)
		return err
	}
}

// NewConfigFactory creates new ConfigFactory
func NewConfigFactory() *ConfigFactory {
	return &ConfigFactory{}
}

// ConfigFactory collects config from CLI and produces real config
type ConfigFactory struct {
	// EnvName is the name of created environment
	EnvName string

	// ModeName is the name of the mode
	ModeName string

	// Target is the deployment target
	Target string

	// HomeDir is the path where all the files are kept
	HomeDir string

	// BinDir is the path where all binaries are present
	BinDir string

	// TestingMode means we are in testing mode and deployment should not block execution
	TestingMode bool

	// TestFilters are regular expressions used to filter tests to run
	TestFilters []string

	// VerboseLogging turns on verbose logging
	VerboseLogging bool
}

// Config produces final config
func (cf *ConfigFactory) Config() infra.Config {
	must.OK(os.MkdirAll(cf.HomeDir, 0o700))
	homeDir := must.String(filepath.Abs(must.String(filepath.EvalSymlinks(cf.HomeDir)))) + "/" + cf.EnvName
	if err := os.Mkdir(homeDir, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		panic(err)
	}

	config := infra.Config{
		EnvName:        cf.EnvName,
		ModeName:       cf.ModeName,
		Target:         cf.Target,
		HomeDir:        homeDir,
		AppDir:         homeDir + "/app",
		WrapperDir:     homeDir + "/bin",
		BinDir:         must.String(filepath.Abs(must.String(filepath.EvalSymlinks(cf.BinDir)))),
		TestingMode:    cf.TestingMode,
		VerboseLogging: cf.VerboseLogging,
	}

	for _, v := range cf.TestFilters {
		config.TestFilters = append(config.TestFilters, regexp.MustCompile(v))
	}

	createDirs(config)

	return config
}

func createDirs(config infra.Config) {
	must.OK(os.MkdirAll(config.AppDir, 0o700))
	must.OK(os.MkdirAll(config.WrapperDir, 0o700))
}
