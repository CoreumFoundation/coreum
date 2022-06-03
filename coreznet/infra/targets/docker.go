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

	"github.com/CoreumFoundation/coreum/coreznet/exec"
	"github.com/CoreumFoundation/coreum/coreznet/infra"
)

const labelEnv = "com.coreum.coreznet.env"

const dockerTemplate = `FROM alpine
COPY . .
ENTRYPOINT ["%s"]
`

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
}

// Stop stops running applications
func (d *Docker) Stop(ctx context.Context) error {
	buf := &bytes.Buffer{}
	listCmd := exec.Docker("ps", "-q", "--filter", "label="+labelEnv+"="+d.config.EnvName)
	listCmd.Stdout = buf
	if err := libexec.Exec(ctx, listCmd); err != nil {
		return err
	}

	var commands []*osexec.Cmd
	for _, cID := range strings.Split(buf.String(), "\n") {
		// last item is empty
		if cID == "" {
			break
		}
		commands = append(commands, exec.Docker("stop", "--time", "60", cID))
	}
	// FIXME (wojtek): parallelize this
	return libexec.Exec(ctx, commands...)
}

// Remove removes running applications
func (d *Docker) Remove(ctx context.Context) error {
	if err := d.dropContainers(ctx); err != nil {
		return err
	}
	return d.dropImages(ctx)
}

// Deploy deploys environment to docker target
func (d *Docker) Deploy(ctx context.Context, mode infra.Mode) error {
	return mode.Deploy(ctx, d, d.config, d.spec)
}

// DeployBinary builds container image out of binary file and starts it in docker
func (d *Docker) DeployBinary(ctx context.Context, app infra.Binary) (infra.DeploymentInfo, error) {
	contextDir := d.config.AppDir + "/" + app.Name
	contextBinDir := contextDir + "/bin/"
	must.OK(os.MkdirAll(contextBinDir, 0o700))

	binPath := app.BinPathFunc("linux")
	must.Any(os.Stat(binPath))
	if err := os.Link(binPath, contextBinDir+"/"+filepath.Base(binPath)); err != nil && !os.IsExist(err) {
		panic(err)
	}

	image := d.config.EnvName + "/" + app.Name + ":latest"
	name := d.config.EnvName + "-" + app.Name
	existsBuf := &bytes.Buffer{}
	existsCmd := exec.Docker("ps", "-aqf", "name="+name)
	existsCmd.Stdout = existsBuf
	if err := libexec.Exec(ctx, existsCmd); err != nil {
		return infra.DeploymentInfo{}, err
	}

	var commands []*osexec.Cmd
	if existsBuf.String() != "" {
		commands = append(commands, exec.Docker("start", name))
	} else {
		runArgs := []string{"run", "--name", name, "-d", "--label", labelEnv + "=" + d.config.EnvName}
		for _, port := range app.Ports {
			portStr := strconv.Itoa(port)
			runArgs = append(runArgs, "-p", ipLocalhost.String()+":"+portStr+":"+portStr+"/tcp")
		}
		runArgs = append(runArgs, image)
		runArgs = append(runArgs, app.ArgsFunc(net.IPv4zero, "/", containerIPResolver{})...)

		buildCmd := exec.Docker("build", "--tag", image, "--label", labelEnv+"="+d.config.EnvName, "-f-", contextDir)
		buildCmd.Stdin = bytes.NewBufferString(fmt.Sprintf(dockerTemplate, filepath.Base(binPath)))
		commands = append(commands, buildCmd, exec.Docker(runArgs...))
	}

	ipBuf := &bytes.Buffer{}
	ipCmd := exec.Docker("inspect", "-f", "{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}", name)
	ipCmd.Stdout = ipBuf
	commands = append(commands, ipCmd)
	err := libexec.Exec(ctx, commands...)
	if err != nil {
		return infra.DeploymentInfo{}, err
	}

	err = osexec.Command("/bin/bash", "-ce",
		fmt.Sprintf("%s >> \"%s/%s.log\" 2>&1", exec.Docker("logs", "-f", name).String(),
			d.config.LogDir, app.Name)).Start()
	if err != nil {
		return infra.DeploymentInfo{}, err
	}

	// FromHostIP = ipLocalhost here means that application is available on host's localhost, not container's localhost
	return infra.DeploymentInfo{FromHostIP: ipLocalhost, FromContainerIP: net.ParseIP(strings.TrimSuffix(ipBuf.String(), "\n"))}, nil
}

// DeployContainer starts container in docker
func (d *Docker) DeployContainer(ctx context.Context, app infra.Container) (infra.DeploymentInfo, error) {
	panic("not implemented yet")
}

func (d *Docker) dropContainers(ctx context.Context) error {
	buf := &bytes.Buffer{}
	listCmd := exec.Docker("ps", "-q", "-a", "--filter", "label="+labelEnv+"="+d.config.EnvName)
	listCmd.Stdout = buf
	if err := libexec.Exec(ctx, listCmd); err != nil {
		return err
	}

	var commands []*osexec.Cmd
	for _, cID := range strings.Split(buf.String(), "\n") {
		// last item is empty
		if cID == "" {
			break
		}
		commands = append(commands, exec.Docker("stop", "--time", "60", cID))
		commands = append(commands, exec.Docker("rm", cID))
	}
	// FIXME (wojtek): parallelize this
	return libexec.Exec(ctx, commands...)
}

func (d *Docker) dropImages(ctx context.Context) error {
	buf := &bytes.Buffer{}
	listCmd := exec.Docker("images", "-q", "--filter", "label="+labelEnv+"="+d.config.EnvName)
	listCmd.Stdout = buf
	if err := libexec.Exec(ctx, listCmd); err != nil {
		return err
	}

	var commands []*osexec.Cmd
	for _, imageID := range strings.Split(buf.String(), "\n") {
		// last item is empty
		if imageID == "" {
			break
		}
		commands = append(commands, exec.Docker("rmi", "-f", imageID))
	}
	// FIXME (wojtek): parallelize this
	return libexec.Exec(ctx, commands...)
}
