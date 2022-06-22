package targets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum-tools/pkg/parallel"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/coreznet/exec"
	"github.com/CoreumFoundation/coreum/coreznet/infra"
)

const (
	// AppHomeDir is the path inide container where application's home directory is mounted
	AppHomeDir = "/app"

	labelEnv = "com.coreum.coreznet.env"
)

// FIXME (wojciech): Entire logic here could be easily implemented by using docker API instead of binary execution

// NewDocker creates new docker target
func NewDocker(config infra.Config, spec *infra.Spec) *Docker {
	return &Docker{
		config: config,
		spec:   spec,
	}
}

// Docker is the target deploying apps to docker
type Docker struct {
	config infra.Config
	spec   *infra.Spec
}

// Stop stops running applications
func (d *Docker) Stop(ctx context.Context) error {
	return forContainer(ctx, d.config.EnvName, func(ctx context.Context, info container) error {
		logger.Get(ctx).Info("Stopping container", zap.String("id", info.ID), zap.String("name", info.Name))
		stopCmd := exec.Docker("stop", "--time", "60", info.ID)
		stopCmd.Stdout = io.Discard
		return libexec.Exec(ctx, stopCmd)
	})
}

// Remove removes running applications
func (d *Docker) Remove(ctx context.Context) error {
	return forContainer(ctx, d.config.EnvName, func(ctx context.Context, info container) error {
		logger.Get(ctx).Info("Deleting container", zap.String("id", info.ID), zap.String("name", info.Name))

		cmds := []*osexec.Cmd{}
		if info.Running {
			// Everything will be removed, so we don't care about graceful shutdown
			killCmd := exec.Docker("kill", info.ID)
			killCmd.Stdout = io.Discard
			cmds = append(cmds, killCmd)
		}
		rmCmd := exec.Docker("rm", info.ID)
		rmCmd.Stdout = io.Discard
		cmds = append(cmds, rmCmd)
		return libexec.Exec(ctx, cmds...)
	})
}

// Deploy deploys environment to docker target
func (d *Docker) Deploy(ctx context.Context, mode infra.Mode) error {
	return mode.Deploy(ctx, d, d.config, d.spec)
}

// DeployBinary builds container image out of binary file and starts it in docker
func (d *Docker) DeployBinary(ctx context.Context, app infra.Binary) (infra.DeploymentInfo, error) {
	name := d.config.EnvName + "-" + app.Name

	log := logger.Get(ctx).With(zap.String("name", name))
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
			"--label", labelEnv + "=" + d.config.EnvName, "-v", appHomeDir + ":" + AppHomeDir, "-v",
			app.BinPath + ":" + internalBinPath}
		for _, port := range app.Ports {
			portStr := strconv.Itoa(port)
			runArgs = append(runArgs, "-p", ipLocalhost.String()+":"+portStr+":"+portStr+"/tcp")
		}
		runArgs = append(runArgs, app.DockerImage(), internalBinPath)
		runArgs = append(runArgs, app.ArgsFunc()...)

		startCmd = exec.Docker(runArgs...)
	}
	idBuf := &bytes.Buffer{}
	startCmd.Stdout = idBuf

	ipBuf := &bytes.Buffer{}
	ipCmd := exec.Docker("inspect", "-f", "{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}", name)
	ipCmd.Stdout = ipBuf

	if err := libexec.Exec(ctx, startCmd, ipCmd); err != nil {
		return infra.DeploymentInfo{}, err
	}

	log.Info("Container started", zap.String("id", strings.TrimSuffix(idBuf.String(), "\n")))

	// FromHostIP = ipLocalhost here means that application is available on host's localhost, not container's localhost
	return infra.DeploymentInfo{
		Container:       name,
		Status:          infra.AppStatusRunning,
		FromHostIP:      ipLocalhost,
		FromContainerIP: net.ParseIP(strings.TrimSuffix(ipBuf.String(), "\n")),
		Ports:           app.Ports,
	}, nil
}

// DeployContainer starts container in docker
func (d *Docker) DeployContainer(ctx context.Context, app infra.Container) (infra.DeploymentInfo, error) {
	name := d.config.EnvName + "-" + app.Name

	log := logger.Get(ctx).With(zap.String("name", name))
	log.Info("Starting container")

	id, err := containerExists(ctx, name)
	if err != nil {
		return infra.DeploymentInfo{}, err
	}

	var startCmd *osexec.Cmd
	if id != "" {
		startCmd = exec.Docker("start", id)
	} else {
		runArgs := []string{"run", "--name", name, "-d", "--label", labelEnv + "=" + d.config.EnvName}
		for _, port := range app.Ports {
			portStr := strconv.Itoa(port)
			runArgs = append(runArgs, "-p", ipLocalhost.String()+":"+portStr+":"+portStr+"/tcp")
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

	ipBuf := &bytes.Buffer{}
	ipCmd := exec.Docker("inspect", "-f", "{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}", name)
	ipCmd.Stdout = ipBuf
	if err := libexec.Exec(ctx, startCmd, ipCmd); err != nil {
		return infra.DeploymentInfo{}, err
	}

	log.Info("Container started", zap.String("id", strings.TrimSuffix(idBuf.String(), "\n")))

	// FromHostIP = ipLocalhost here means that application is available on host's localhost, not container's localhost
	return infra.DeploymentInfo{
		Container:       name,
		Status:          infra.AppStatusRunning,
		FromHostIP:      ipLocalhost,
		FromContainerIP: net.ParseIP(strings.TrimSuffix(ipBuf.String(), "\n")),
		Ports:           app.Ports,
	}, nil
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
					Running: cInfo.State.Running,
				})
			})
		}
		return nil
	})
}
