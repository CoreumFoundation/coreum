package targets

import (
	"context"
	"fmt"
	"net"
	"os"
	osexec "os/exec"
	"runtime"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"

	"github.com/CoreumFoundation/coreum/coreznet/exec"
	"github.com/CoreumFoundation/coreum/coreznet/infra"
)

// NewDirect creates new direct target
func NewDirect(config infra.Config, spec *infra.Spec, docker *Docker) infra.Target {
	return &Direct{
		config: config,
		spec:   spec,
		docker: docker,
	}
}

// Direct is the target deploying apps to os processes
type Direct struct {
	config infra.Config
	spec   *infra.Spec
	docker *Docker
}

// BindIP returns the IP application should bind to inside the target
func (d *Direct) BindIP() net.IP {
	return ipLocalhost
}

// Deploy deploys environment to os processes
func (d *Direct) Deploy(ctx context.Context, mode infra.Mode) error {
	return mode.Deploy(ctx, d, d.config, d.spec)
}

// Stop stops running applications
func (d *Direct) Stop(ctx context.Context) error {
	if err := d.stopProcesses(ctx); err != nil {
		return err
	}
	return d.docker.Stop(ctx)
}

// Remove removes running applications
func (d *Direct) Remove(ctx context.Context) error {
	if err := d.stopProcesses(ctx); err != nil {
		return err
	}
	return d.docker.Remove(ctx)
}

// DeployBinary starts binary file inside os process
func (d *Direct) DeployBinary(ctx context.Context, app infra.Binary) (infra.DeploymentInfo, error) {
	binPath := app.BinPathFunc(runtime.GOOS)
	must.Any(os.Stat(binPath))
	cmd := osexec.Command("/bin/bash", "-ce", fmt.Sprintf(`exec %s >> "%s/%s.log" 2>&1`, osexec.Command(binPath, app.ArgsFunc(ipLocalhost, d.config.AppDir+"/"+app.Name, hostIPResolver{})...).String(), d.config.LogDir, app.Name))
	if err := cmd.Start(); err != nil {
		return infra.DeploymentInfo{}, err
	}
	return infra.DeploymentInfo{
		Status:          infra.AppStatusRunning,
		ProcessID:       cmd.Process.Pid,
		FromHostIP:      ipLocalhost,
		FromContainerIP: ipLocalhost,
		Ports:           app.Ports}, nil
}

// DeployContainer starts container
func (d *Direct) DeployContainer(ctx context.Context, app infra.Container) (infra.DeploymentInfo, error) {
	return d.docker.DeployContainer(ctx, app)
}

func (d *Direct) stopProcesses(ctx context.Context) error {
	pIDs := make([]int, 0, len(d.spec.Apps))
	for _, app := range d.spec.Apps {
		pID := app.Info().ProcessID
		if pID == 0 {
			continue
		}
		pIDs = append(pIDs, pID)
	}
	if len(pIDs) == 0 {
		return nil
	}
	return exec.Kill(ctx, pIDs)
}
