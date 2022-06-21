package targets

import (
	"bytes"
	"context"
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
	return forContainer(ctx, d.config.EnvName, func(ctx context.Context, id string) error {
		logger.Get(ctx).Info("Stopping container", zap.String("id", id))
		stopCmd := exec.Docker("stop", "--time", "60", id)
		stopCmd.Stdout = io.Discard
		return libexec.Exec(ctx, stopCmd)
	})
}

// Remove removes running applications
func (d *Docker) Remove(ctx context.Context) error {
	return forContainer(ctx, d.config.EnvName, func(ctx context.Context, id string) error {
		logger.Get(ctx).Info("Deleting container", zap.String("id", id))

		// FIXME (wojciech): convert to `kill` - it requires a check if container is running
		stopCmd := exec.Docker("stop", id)
		stopCmd.Stdout = io.Discard
		rmCmd := exec.Docker("rm", id)
		rmCmd.Stdout = io.Discard
		return libexec.Exec(ctx, stopCmd, rmCmd)
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

	err = osexec.Command("/bin/bash", "-ce",
		fmt.Sprintf("%s >> \"%s/%s.log\" 2>&1", exec.Docker("logs", "-f", name).String(),
			d.config.LogDir, app.Name)).Start()
	if err != nil {
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

	err = osexec.Command("/bin/bash", "-ce",
		fmt.Sprintf("%s >> \"%s/%s.log\" 2>&1", exec.Docker("logs", "-f", name).String(),
			d.config.LogDir, app.Name)).Start()
	if err != nil {
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

func forContainer(ctx context.Context, envName string, fn func(ctx context.Context, id string) error) error {
	buf := &bytes.Buffer{}
	listCmd := exec.Docker("ps", "-aq", "--no-trunc", "--filter", "label="+labelEnv+"="+envName)
	listCmd.Stdout = buf
	if err := libexec.Exec(ctx, listCmd); err != nil {
		return err
	}

	return parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		for _, cID := range strings.Split(buf.String(), "\n") {
			// last item is empty
			if cID == "" {
				break
			}
			cID := cID
			spawn("container."+cID, parallel.Continue, func(ctx context.Context) error {
				return fn(ctx, cID)
			})
		}
		return nil
	})
}
