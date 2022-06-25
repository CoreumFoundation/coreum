package znet

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/CoreumFoundation/coreum-tools/pkg/ioc"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/crust/infra"
	"github.com/CoreumFoundation/coreum/crust/infra/apps"
	"github.com/CoreumFoundation/coreum/crust/infra/targets"
)

// IoC configures IoC container
func IoC(c *ioc.Container) {
	c.Singleton(NewCmdFactory)
	c.Singleton(infra.NewConfigFactory)
	c.Singleton(infra.NewSpec)
	c.Transient(NewConfig)
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

// NewConfig produces final config
func NewConfig(configF *infra.ConfigFactory, spec *infra.Spec) infra.Config {
	must.OK(os.MkdirAll(configF.HomeDir, 0o700))
	homeDir := must.String(filepath.Abs(must.String(filepath.EvalSymlinks(configF.HomeDir)))) + "/" + configF.EnvName
	if err := os.Mkdir(homeDir, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		panic(err)
	}

	config := infra.Config{
		EnvName:        configF.EnvName,
		ModeName:       spec.Mode,
		Target:         configF.Target,
		HomeDir:        homeDir,
		AppDir:         homeDir + "/app",
		WrapperDir:     homeDir + "/bin",
		BinDir:         must.String(filepath.Abs(must.String(filepath.EvalSymlinks(configF.BinDir)))),
		TestingMode:    configF.TestingMode,
		VerboseLogging: configF.VerboseLogging,
	}

	for _, v := range configF.TestFilters {
		config.TestFilters = append(config.TestFilters, regexp.MustCompile(v))
	}

	createDirs(config)

	return config
}

func createDirs(config infra.Config) {
	must.OK(os.MkdirAll(config.AppDir, 0o700))
	must.OK(os.MkdirAll(config.WrapperDir, 0o700))
}
