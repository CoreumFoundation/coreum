package znet

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	osexec "os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/ioc"
	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum-tools/pkg/parallel"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
	"github.com/CoreumFoundation/coreum/coreznet/infra/testing"
	"github.com/CoreumFoundation/coreum/coreznet/pkg/zstress"
	"github.com/CoreumFoundation/coreum/coreznet/tests"
)

var exe = must.String(filepath.EvalSymlinks(must.String(os.Executable())))

// Activate starts preconfigured shell environment
func Activate(ctx context.Context, configF *ConfigFactory) error {
	config := configF.Config()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.WithStack(err)
	}
	defer watcher.Close()

	if err := watcher.Add(config.HomeDir); err != nil {
		return errors.WithStack(err)
	}

	saveWrapper(config.WrapperDir, "start", "start")
	saveWrapper(config.WrapperDir, "stop", "stop")
	saveWrapper(config.WrapperDir, "remove", "remove")
	// `test` can't be used here because it is a reserved keyword in bash
	saveWrapper(config.WrapperDir, "tests", "test")
	saveWrapper(config.WrapperDir, "spec", "spec")
	saveWrapper(config.WrapperDir, "ping-pong", "ping-pong")
	saveWrapper(config.WrapperDir, "stress", "stress")
	saveLogsWrapper(config.WrapperDir, config.EnvName, "logs")

	shell, promptVar, err := shellConfig(config.EnvName)
	if err != nil {
		return err
	}
	shellCmd := osexec.Command(shell)
	shellCmd.Env = append(os.Environ(),
		"PATH="+config.WrapperDir+":"+os.Getenv("PATH"),
		"COREZNET_ENV="+configF.EnvName,
		"COREZNET_MODE="+configF.ModeName,
		"COREZNET_HOME="+configF.HomeDir,
		"COREZNET_TARGET="+configF.Target,
		"COREZNET_BIN_DIR="+configF.BinDir,
		"COREZNET_FILTERS="+strings.Join(configF.TestFilters, ","),
	)
	if promptVar != "" {
		shellCmd.Env = append(shellCmd.Env, promptVar)
	}
	shellCmd.Dir = config.HomeDir
	shellCmd.Stdin = os.Stdin

	return parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		spawn("session", parallel.Exit, func(ctx context.Context) error {
			err = libexec.Exec(ctx, shellCmd)
			if shellCmd.ProcessState != nil && shellCmd.ProcessState.ExitCode() != 0 {
				// shell returns non-exit code if command executed in the shell failed
				return nil
			}
			return err
		})
		spawn("fsnotify", parallel.Exit, func(ctx context.Context) error {
			defer func() {
				if shellCmd.Process != nil {
					// Shell exits only if SIGHUP is received. All the other signals are caught and passed to process
					// running inside the shell.
					_ = shellCmd.Process.Signal(syscall.SIGHUP)
				}
			}()
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case event := <-watcher.Events:
					// Rename is here because on some OSes removing is done by moving file to trash
					if event.Op&(fsnotify.Remove|fsnotify.Rename) == 0 {
						continue
					}
					if _, err := os.Stat(config.HomeDir); err != nil {
						if os.IsNotExist(err) {
							return nil
						}
						return errors.WithStack(err)
					}
				case err := <-watcher.Errors:
					return errors.WithStack(err)
				}
			}
		})
		return nil
	})
}

// Start starts environment
func Start(ctx context.Context, target infra.Target, mode infra.Mode) (retErr error) {
	return target.Deploy(ctx, mode)
}

// Stop stops environment
func Stop(ctx context.Context, target infra.Target, spec *infra.Spec) (retErr error) {
	defer func() {
		for _, app := range spec.Apps {
			app.SetInfo(infra.DeploymentInfo{Status: infra.AppStatusStopped})
		}
		if err := spec.Save(); retErr == nil {
			retErr = err
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
	return errors.WithStack(err)
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
func Spec(spec *infra.Spec) error {
	fmt.Println(spec)
	return nil
}

// PingPong connects to cored node and sends transactions back and forth from one account to another to generate
// transactions on the blockchain
func PingPong(ctx context.Context, mode infra.Mode) error {
	coredNode, err := coredNode(mode)
	if err != nil {
		return err
	}
	client := coredNode.Client()

	alice := cored.Wallet{Name: "alice", Key: cored.AlicePrivKey}
	bob := cored.Wallet{Name: "bob", Key: cored.BobPrivKey}
	charlie := cored.Wallet{Name: "charlie", Key: cored.CharliePrivKey}

	for {
		if err := sendTokens(ctx, client, alice, bob); err != nil {
			return err
		}
		if err := sendTokens(ctx, client, bob, charlie); err != nil {
			return err
		}
		if err := sendTokens(ctx, client, charlie, alice); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

// Stress runs benchmark implemented by `corezstress` on top of network deployed by `coreznet`
func Stress(ctx context.Context, mode infra.Mode) error {
	coredNode, err := coredNode(mode)
	if err != nil {
		return err
	}

	healthyCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if err := infra.WaitUntilHealthy(healthyCtx, coredNode); err != nil {
		return err
	}

	return zstress.Stress(ctx, zstress.StressConfig{
		ChainID:           coredNode.ChainID(),
		NodeAddress:       infra.JoinProtoIPPort("", coredNode.Info().FromHostIP, coredNode.Ports().RPC),
		Accounts:          cored.RandomWallets[:10],
		NumOfTransactions: 100,
	})
}

func coredNode(mode infra.Mode) (apps.Cored, error) {
	for _, app := range mode {
		if app.Type() == apps.CoredType && app.Info().Status == infra.AppStatusRunning {
			return app.(apps.Cored), nil
		}
	}
	return apps.Cored{}, errors.New("haven't found any running cored node")
}

func sendTokens(ctx context.Context, client cored.Client, from, to cored.Wallet) error {
	log := logger.Get(ctx)

	amount := cored.Balance{Amount: big.NewInt(1), Denom: "core"}
	txBytes, err := client.PrepareTxBankSend(from, to, amount)
	if err != nil {
		return err
	}
	txHash, err := client.Broadcast(ctx, txBytes)
	if err != nil {
		return err
	}

	log.Info("Sent tokens", zap.Stringer("from", from), zap.Stringer("to", to),
		zap.Stringer("amount", amount), zap.String("txHash", txHash))

	fromBalance, err := client.QueryBankBalances(ctx, from)
	if err != nil {
		return err
	}
	toBalance, err := client.QueryBankBalances(ctx, to)
	if err != nil {
		return err
	}

	log.Info("Current balance", zap.Stringer("wallet", from), zap.Stringer("balance", fromBalance["core"]))
	log.Info("Current balance", zap.Stringer("wallet", to), zap.Stringer("balance", toBalance["core"]))

	return nil
}

func saveWrapper(dir, file, command string) {
	must.OK(ioutil.WriteFile(dir+"/"+file, []byte(`#!/bin/bash
exec "`+exe+`" "`+command+`" "$@"
`), 0o700))
}

func saveLogsWrapper(dir, envName, file string) {
	must.OK(ioutil.WriteFile(dir+"/"+file, []byte(`#!/bin/bash
if [ "$1" == "" ]; then
  echo "Provide the name of application"
  exit 1
fi
exec docker logs -f "`+envName+`-$1"
`), 0o700))
}

var supportedShells = map[string]func(envName string) string{
	"bash": func(envName string) string {
		return "PS1=(" + envName + `) [\u@\h \W]\$ `
	},
	"zsh": func(envName string) string {
		return "PROMPT=(" + envName + `) [%n@%m %1~]%# `
	},
}

func shellConfig(envName string) (string, string, error) {
	shell := os.Getenv("SHELL")
	if _, exists := supportedShells[filepath.Base(shell)]; !exists {
		var shells []string
		switch runtime.GOOS {
		case "darwin":
			shells = []string{"zsh", "bash"}
		default:
			shells = []string{"bash", "zsh"}
		}
		for _, s := range shells {
			if shell2, err := osexec.LookPath(s); err == nil {
				shell = shell2
				break
			}
		}
	}
	if shell == "" {
		return "", "", errors.New("custom shell not defined and supported shell not found")
	}

	var promptVar string
	if promptVarFn, exists := supportedShells[filepath.Base(shell)]; exists {
		promptVar = promptVarFn(envName)
	}
	return shell, promptVar, nil
}
