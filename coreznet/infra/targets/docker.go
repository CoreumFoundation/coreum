package targets

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum-tools/pkg/parallel"

	"github.com/CoreumFoundation/coreum/coreznet/exec"
	"github.com/CoreumFoundation/coreum/coreznet/infra"
)

const (
	// AppHomeDir is the path inide container where application's home directory is mounted
	AppHomeDir = "/app"

	labelEnv          = "com.coreum.coreznet.env"
	binaryDockerImage = "alpine:3.16.0"
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
		return libexec.Exec(ctx, exec.Docker("stop", "--time", "60", id))
	})
}

// Remove removes running applications
func (d *Docker) Remove(ctx context.Context) error {
	return forContainer(ctx, d.config.EnvName, func(ctx context.Context, id string) error {
		return libexec.Exec(ctx,
			exec.Docker("stop", "--time", "60", id),
			exec.Docker("rm", id),
		)
	})
}

// Deploy deploys environment to docker target
func (d *Docker) Deploy(ctx context.Context, mode infra.Mode) error {
	return mode.Deploy(ctx, d, d.config, d.spec)
}

// DeployBinary builds container image out of binary file and starts it in docker
func (d *Docker) DeployBinary(ctx context.Context, app infra.Binary) (infra.DeploymentInfo, error) {
	name := d.config.EnvName + "-" + app.Name
	exists, err := containerExists(ctx, name)
	if err != nil {
		return infra.DeploymentInfo{}, err
	}

	var startCmd *osexec.Cmd
	if exists {
		startCmd = exec.Docker("start", name)
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
		runArgs = append(runArgs, binaryDockerImage, internalBinPath)
		runArgs = append(runArgs, app.ArgsFunc()...)

		startCmd = exec.Docker(runArgs...)
	}

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
	exists, err := containerExists(ctx, name)
	if err != nil {
		return infra.DeploymentInfo{}, err
	}

	var startCmd *osexec.Cmd
	if exists {
		startCmd = exec.Docker("start", name)
	} else {
		runArgs := []string{"run", "--name", name, "-d", "--label", labelEnv + "=" + d.config.EnvName}
		for _, port := range app.Ports {
			portStr := strconv.Itoa(port)
			runArgs = append(runArgs, "-p", ipLocalhost.String()+":"+portStr+":"+portStr+"/tcp")
		}
		for _, env := range app.EnvVars {
			runArgs = append(runArgs, "-e", env.Name+"="+env.Value)
		}
		runArgs = append(runArgs, app.Image+":"+app.Tag)
		runArgs = append(runArgs, app.ArgsFunc()...)

		startCmd = exec.Docker(runArgs...)
	}

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

	// FromHostIP = ipLocalhost here means that application is available on host's localhost, not container's localhost
	return infra.DeploymentInfo{
		Container:       name,
		Status:          infra.AppStatusRunning,
		FromHostIP:      ipLocalhost,
		FromContainerIP: net.ParseIP(strings.TrimSuffix(ipBuf.String(), "\n")),
		Ports:           app.Ports,
	}, nil
}

func containerExists(ctx context.Context, name string) (bool, error) {
	existsBuf := &bytes.Buffer{}
	existsCmd := exec.Docker("ps", "-aqf", "name="+name)
	existsCmd.Stdout = existsBuf
	if err := libexec.Exec(ctx, existsCmd); err != nil {
		return false, err
	}
	return existsBuf.Len() > 0, nil
}

func forContainer(ctx context.Context, envName string, fn func(ctx context.Context, id string) error) error {
	buf := &bytes.Buffer{}
	listCmd := exec.Docker("ps", "-aq", "--filter", "label="+labelEnv+"="+envName)
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
