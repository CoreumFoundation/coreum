package coreznet

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/ioc"
	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"

	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps"
	"github.com/CoreumFoundation/coreum/coreznet/infra/testing"
	"github.com/CoreumFoundation/coreum/coreznet/tests"
)

// Activate starts preconfigured bash environment
func Activate(ctx context.Context, configF *ConfigFactory) error {
	config := configF.Config()

	exe := must.String(filepath.EvalSymlinks(must.String(os.Executable())))

	must.OK(ioutil.WriteFile(config.WrapperDir+"/start", []byte(fmt.Sprintf("#!/bin/bash\nexec %s start \"$@\"", exe)), 0o700))
	must.OK(ioutil.WriteFile(config.WrapperDir+"/stop", []byte(fmt.Sprintf("#!/bin/bash\nexec %s stop \"$@\"", exe)), 0o700))
	must.OK(ioutil.WriteFile(config.WrapperDir+"/remove", []byte(fmt.Sprintf("#!/bin/bash\nexec %s remove \"$@\"", exe)), 0o700))
	// `test` can't be used here because it is a reserved keyword in bash
	must.OK(ioutil.WriteFile(config.WrapperDir+"/tests", []byte(fmt.Sprintf("#!/bin/bash\nexec %s test \"$@\"", exe)), 0o700))
	must.OK(ioutil.WriteFile(config.WrapperDir+"/spec", []byte(fmt.Sprintf("#!/bin/bash\nexec %s spec \"$@\"", exe)), 0o700))
	must.OK(ioutil.WriteFile(config.WrapperDir+"/logs", []byte(fmt.Sprintf(`#!/bin/bash
if [ "$1" == "" ]; then
  echo "Provide the name of application"
  exit 1
fi
exec tail -f -n +0 "%s/$1.log"
`, config.LogDir)), 0o700))

	bash := osexec.Command("bash")
	bash.Env = append(os.Environ(),
		fmt.Sprintf("PS1=%s", "("+configF.EnvName+`) [\u@\h \W]\$ `),
		fmt.Sprintf("PATH=%s", config.WrapperDir+":/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin"),
		fmt.Sprintf("COREZNET_ENV=%s", configF.EnvName),
		fmt.Sprintf("COREZNET_MODE=%s", configF.ModeName),
		fmt.Sprintf("COREZNET_HOME=%s", configF.HomeDir),
		fmt.Sprintf("COREZNET_TARGET=%s", configF.Target),
		fmt.Sprintf("COREZNET_BIN_DIR=%s", configF.BinDir),
		fmt.Sprintf("COREZNET_FILTERS=%s", strings.Join(configF.TestFilters, ",")),
		fmt.Sprintf("COREZNET_VERBOSE=%t", configF.VerboseLogging),
	)
	bash.Dir = config.LogDir
	bash.Stdin = os.Stdin
	err := libexec.Exec(ctx, bash)
	if bash.ProcessState != nil && bash.ProcessState.ExitCode() != 0 {
		// bash returns non-exit code if command executed in the shell failed
		return nil
	}
	return err
}

// Start starts environment
func Start(ctx context.Context, target infra.Target, mode infra.Mode) (retErr error) {
	return target.Deploy(ctx, mode)
}

// Stop stops environment
func Stop(ctx context.Context, target infra.Target, spec *infra.Spec) (retErr error) {
	defer func() {
		spec.PGID = 0
		for _, app := range spec.Apps {
			if app.Status() == infra.AppStatusRunning {
				app.SetStatus(infra.AppStatusStopped)
			}
			if err := spec.Save(); retErr == nil {
				retErr = err
			}
		}
	}()
	return target.Stop(ctx)
}

// Remove removes environment
func Remove(ctx context.Context, config infra.Config, target infra.Target) (retErr error) {
	if err := target.Remove(ctx); err != nil {
		return err
	}

	// It may happen that some files are flushed to disk even after processes are terminated
	// so let's try to delete dir a few times
	var err error
	for i := 0; i < 3; i++ {
		if err = os.RemoveAll(config.HomeDir); err == nil || errors.Is(err, os.ErrNotExist) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return err
}

// Test runs integration tests
func Test(c *ioc.Container, configF *ConfigFactory) error {
	configF.TestingMode = true
	configF.ModeName = "test"
	var err error
	c.Call(func(ctx context.Context, config infra.Config, target infra.Target, appF *apps.Factory, spec *infra.Spec) (retErr error) {
		defer func() {
			if err := spec.Save(); retErr == nil {
				retErr = err
			}
		}()

		env, tests := tests.Tests(appF)
		return testing.Run(ctx, target, env, tests, config.TestFilters)
	}, &err)
	return err
}

// Spec print specification of running environment
func Spec(spec *infra.Spec, _ infra.Mode) error {
	fmt.Println(spec)
	return nil
}
