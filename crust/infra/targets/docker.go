package targets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum-tools/pkg/parallel"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/crust/exec"
	"github.com/CoreumFoundation/coreum/crust/infra"
)

const (
	// AppHomeDir is the path inide container where application's home directory is mounted
	AppHomeDir = "/app"

	labelEnv = "com.coreum.crust.znet.env"
	labelApp = "com.coreum.crust.znet.app"
)

// FIXME (wojciech): Entire logic here could be easily implemented by using docker API instead of binary execution

// NewDocker creates new docker target
func NewDocker(config infra.Config, spec *infra.Spec) infra.Target {
	return &Docker{
		config: config,
		spec:   spec,
	}
}

// Docker is the target deploying apps to docker
type Docker struct {
	config infra.Config
	spec   *infra.Spec

	mu            sync.Mutex
	networkExists bool
}

// Stop stops running applications
func (d *Docker) Stop(ctx context.Context) error {
	dependencies := map[string][]chan struct{}{}
	readyChs := map[string]chan struct{}{}
	for appName, app := range d.spec.Apps {
		readyCh := make(chan struct{})
		readyChs[appName] = readyCh

		for _, depName := range app.Info().DependsOn {
			dependencies[depName] = append(dependencies[depName], readyCh)
		}
	}

	return forContainer(ctx, d.config.EnvName, func(ctx context.Context, info container) error {
		log := logger.Get(ctx).With(zap.String("id", info.ID), zap.String("name", info.Name),
			zap.String("appName", info.AppName))

		if _, exists := d.spec.Apps[info.AppName]; !exists {
			log.Info("Unexpected container found, deleting it")

			if err := removeContainer(ctx, info); err != nil {
				return err
			}

			log.Info("Container deleted")
			return nil
		}

		if deps := dependencies[info.AppName]; len(deps) > 0 {
			log.Info("Waiting for dependencies to be stopped")
			for _, depCh := range deps {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-depCh:
				}
			}
		}

		log.Info("Stopping container")

		if err := libexec.Exec(ctx, noStdout(exec.Docker("stop", "--time", "60", info.ID))); err != nil {
			return errors.Wrapf(err, "stopping container `%s` failed", info.Name)
		}

		log.Info("Container stopped")
		close(readyChs[info.AppName])
		return nil
	})
}

// Remove removes running applications
func (d *Docker) Remove(ctx context.Context) error {
	err := forContainer(ctx, d.config.EnvName, func(ctx context.Context, info container) error {
		log := logger.Get(ctx).With(zap.String("id", info.ID), zap.String("name", info.Name),
			zap.String("appName", info.AppName))
		log.Info("Deleting container")

		if err := removeContainer(ctx, info); err != nil {
			return err
		}

		log.Info("Container deleted")
		return nil
	})
	if err != nil {
		return err
	}
	return d.deleteNetwork(ctx, d.config.EnvName)
}

// Deploy deploys environment to docker target
func (d *Docker) Deploy(ctx context.Context, mode infra.Mode) error {
	return mode.Deploy(ctx, d, d.config, d.spec)
}

// DeployBinary builds container image out of binary file and starts it in docker
func (d *Docker) DeployBinary(ctx context.Context, app infra.Binary) (infra.DeploymentInfo, error) {
	if err := d.ensureNetwork(ctx, d.config.EnvName); err != nil {
		return infra.DeploymentInfo{}, nil
	}

	name := d.config.EnvName + "-" + app.Name

	log := logger.Get(ctx).With(zap.String("name", name), zap.String("appName", app.Name))
	log.Info("Starting container")

	id, err := containerExists(ctx, name)
	if err != nil {
		return infra.DeploymentInfo{}, err
	}

	var startCmd *osexec.Cmd
	if id != "" {
		startCmd = exec.Docker("start", id)
	} else {
		appHomeDir := d.config.AppDir + "/" + app.Name
		must.Any(os.Stat(app.BinPath))
		internalBinPath := "/bin/" + filepath.Base(app.BinPath)

		runArgs := []string{"run", "--name", name, "-d", "--user", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
			"--label", labelEnv + "=" + d.config.EnvName, "--label", labelApp + "=" + app.Name,
			"-v", appHomeDir + ":" + AppHomeDir, "-v", app.BinPath + ":" + internalBinPath,
			"--network", d.config.EnvName}
		for _, port := range app.Ports {
			portStr := strconv.Itoa(port)
			runArgs = append(runArgs, "-p", "127.0.0.1:"+portStr+":"+portStr+"/tcp")
		}
		runArgs = append(runArgs, app.DockerImage(), internalBinPath)
		runArgs = append(runArgs, app.ArgsFunc()...)

		startCmd = exec.Docker(runArgs...)
	}
	idBuf := &bytes.Buffer{}
	startCmd.Stdout = idBuf

	if err := libexec.Exec(ctx, startCmd); err != nil {
		return infra.DeploymentInfo{}, err
	}

	log.Info("Container started", zap.String("id", strings.TrimSuffix(idBuf.String(), "\n")))

	// FromHostIP = ipLocalhost here means that application is available on host's localhost, not container's localhost
	return infra.DeploymentInfo{
		Container:         name,
		Status:            infra.AppStatusRunning,
		HostFromHost:      "localhost",
		HostFromContainer: name,
		Ports:             app.Ports,
	}, nil
}

// DeployContainer starts container in docker
func (d *Docker) DeployContainer(ctx context.Context, app infra.Container) (infra.DeploymentInfo, error) {
	if err := d.ensureNetwork(ctx, d.config.EnvName); err != nil {
		return infra.DeploymentInfo{}, nil
	}

	name := d.config.EnvName + "-" + app.Name

	log := logger.Get(ctx).With(zap.String("name", name), zap.String("appName", app.Name))
	log.Info("Starting container")

	id, err := containerExists(ctx, name)
	if err != nil {
		return infra.DeploymentInfo{}, err
	}

	var startCmd *osexec.Cmd
	if id != "" {
		startCmd = exec.Docker("start", id)
	} else {
		appHomeDir := d.config.AppDir + "/" + app.Name
		runArgs := []string{"run", "--name", name, "-d", "--label", labelEnv + "=" + d.config.EnvName,
			"--label", labelApp + "=" + app.Name, "-v", appHomeDir + ":" + AppHomeDir, "--network", d.config.EnvName}
		for _, port := range app.Ports {
			portStr := strconv.Itoa(port)
			runArgs = append(runArgs, "-p", "127.0.0.1:"+portStr+":"+portStr+"/tcp")
		}
		for _, env := range app.EnvVars {
			runArgs = append(runArgs, "-e", env.Name+"="+env.Value)
		}
		runArgs = append(runArgs, app.DockerImage())
		runArgs = append(runArgs, app.ArgsFunc()...)

		startCmd = exec.Docker(runArgs...)
	}
	idBuf := &bytes.Buffer{}
	startCmd.Stdout = idBuf

	if err := libexec.Exec(ctx, startCmd); err != nil {
		return infra.DeploymentInfo{}, err
	}

	log.Info("Container started", zap.String("id", strings.TrimSuffix(idBuf.String(), "\n")))

	// FromHostIP = ipLocalhost here means that application is available on host's localhost, not container's localhost
	return infra.DeploymentInfo{
		Container:         name,
		Status:            infra.AppStatusRunning,
		HostFromHost:      "localhost",
		HostFromContainer: name,
		Ports:             app.Ports,
	}, nil
}

func (d *Docker) ensureNetwork(ctx context.Context, network string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.networkExists {
		return nil
	}

	log := logger.Get(ctx).With(zap.String("network", network))

	var err error
	d.networkExists, err = networkExists(ctx, network)
	if err != nil {
		return err
	}
	if d.networkExists {
		log.Info("Docker network exists")
		return nil
	}

	log.Info("Creating docker network")

	if err := libexec.Exec(ctx, noStdout(exec.Docker("network", "create", network))); err != nil {
		return errors.Wrapf(err, "creating network '%s' failed", network)
	}

	d.networkExists = true
	log.Info("Docker network created")
	return nil
}

func (d *Docker) deleteNetwork(ctx context.Context, network string) error {
	exists, err := networkExists(ctx, network)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	log := logger.Get(ctx).With(zap.String("network", network))
	log.Info("Deleting docker network")

	if err := libexec.Exec(ctx, noStdout(exec.Docker("network", "rm", network))); err != nil {
		return errors.Wrapf(err, "deleting network '%s' failed", network)
	}

	log.Info("Docker network deleted")
	return nil
}

func containerExists(ctx context.Context, name string) (string, error) {
	idBuf := &bytes.Buffer{}
	existsCmd := exec.Docker("ps", "-aq", "--no-trunc", "--filter", "name="+name)
	existsCmd.Stdout = idBuf
	if err := libexec.Exec(ctx, existsCmd); err != nil {
		return "", err
	}
	return strings.TrimSuffix(idBuf.String(), "\n"), nil
}

type container struct {
	ID      string
	Name    string
	AppName string
	Running bool
}

func forContainer(ctx context.Context, envName string, fn func(ctx context.Context, info container) error) error {
	listBuf := &bytes.Buffer{}
	listCmd := exec.Docker("ps", "-aq", "--no-trunc", "--filter", "label="+labelEnv+"="+envName)
	listCmd.Stdout = listBuf
	if err := libexec.Exec(ctx, listCmd); err != nil {
		return err
	}

	listStr := strings.TrimSuffix(listBuf.String(), "\n")
	if listStr == "" {
		return nil
	}

	inspectBuf := &bytes.Buffer{}
	inspectCmd := exec.Docker(append([]string{"inspect"}, strings.Split(listStr, "\n")...)...)
	inspectCmd.Stdout = inspectBuf

	if err := libexec.Exec(ctx, inspectCmd); err != nil {
		return err
	}

	var info []struct {
		ID    string `json:"Id"` // nolint:tagliatelle // `Id` is defined by docker
		Name  string
		State struct {
			Running bool
		}
		Config struct {
			Labels map[string]string
		}
	}

	if err := json.Unmarshal(inspectBuf.Bytes(), &info); err != nil {
		return errors.Wrap(err, "unmarshalling container properties failed")
	}

	return parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		for _, cInfo := range info {
			cInfo := cInfo
			spawn("container."+cInfo.ID, parallel.Continue, func(ctx context.Context) error {
				return fn(ctx, container{
					ID:      cInfo.ID,
					Name:    strings.TrimPrefix(cInfo.Name, "/"),
					AppName: cInfo.Config.Labels[labelApp],
					Running: cInfo.State.Running,
				})
			})
		}
		return nil
	})
}

func noStdout(cmd *osexec.Cmd) *osexec.Cmd {
	cmd.Stdout = io.Discard
	return cmd
}

func removeContainer(ctx context.Context, info container) error {
	cmds := []*osexec.Cmd{}
	if info.Running {
		// Everything will be removed, so we don't care about graceful shutdown
		cmds = append(cmds, noStdout(exec.Docker("kill", info.ID)))
	}
	if err := libexec.Exec(ctx, append(cmds, noStdout(exec.Docker("rm", info.ID)))...); err != nil {
		return errors.Wrapf(err, "deleting container `%s` failed", info.Name)
	}
	return nil
}

func networkExists(ctx context.Context, network string) (bool, error) {
	buf := &bytes.Buffer{}
	cmd := exec.Docker("network", "ls", "-q", "--no-trunc", "--filter", "name="+network)
	cmd.Stdout = buf
	if err := libexec.Exec(ctx, cmd); err != nil {
		return false, err
	}
	return strings.TrimSuffix(buf.String(), "\n") != "", nil
}
