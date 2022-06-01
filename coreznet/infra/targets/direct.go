package targets

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	osexec "os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/coreznet/exec"
	"github.com/CoreumFoundation/coreum/coreznet/infra"
)

// NewDirect creates new direct target
func NewDirect(config infra.Config, spec *infra.Spec) infra.Target {
	return &Direct{
		config: config,
		spec:   spec,
	}
}

// Direct is the target deploying apps to os processes
type Direct struct {
	config infra.Config
	spec   *infra.Spec
}

// Deploy deploys environment to os processes
func (d *Direct) Deploy(ctx context.Context, mode infra.Mode) error {
	return mode.Deploy(ctx, d, d.config, d.spec)
}

// Stop stops running applications
func (d *Direct) Stop(ctx context.Context) error {
	if d.spec.PGID == 0 {
		return nil
	}
	procs, err := ioutil.ReadDir("/proc")
	if err != nil {
		return errors.WithStack(err)
	}
	reg := regexp.MustCompile("^[0-9]+$")
	var pids []int
	for _, procH := range procs {
		if !procH.IsDir() || !reg.MatchString(procH.Name()) {
			continue
		}
		statPath := "/proc/" + procH.Name() + "/stat"
		statRaw, err := ioutil.ReadFile(statPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return errors.WithStack(err)
		}
		properties := strings.Split(string(statRaw), " ")
		pgID, err := strconv.ParseInt(properties[4], 10, 32)
		if err != nil {
			return errors.WithStack(err)
		}
		if pgID != int64(d.spec.PGID) {
			continue
		}
		pID := int(must.Int64(strconv.ParseInt(procH.Name(), 10, 32)))
		if pID == os.Getpid() {
			continue
		}
		pids = append(pids, pID)
	}
	if len(pids) == 0 {
		return nil
	}
	return exec.Kill(ctx, pids)
}

// Remove removes running applications
func (d *Direct) Remove(ctx context.Context) error {
	return d.Stop(ctx)
}

// DeployBinary starts binary file inside os process
func (d *Direct) DeployBinary(ctx context.Context, app infra.Binary) (infra.DeploymentInfo, error) {
	binPath := app.BinPathFunc(runtime.GOOS)
	must.Any(os.Stat(binPath))
	cmd := osexec.Command("/bin/sh", "-ce", fmt.Sprintf(`exec %s >> "%s/%s.log" 2>&1`, osexec.Command(binPath, app.ArgsFunc(net.IPv4(127, 0, 0, 1), d.config.AppDir+"/"+app.Name)...).String(), d.config.LogDir, app.Name))
	if err := cmd.Start(); err != nil {
		return infra.DeploymentInfo{}, err
	}
	return infra.DeploymentInfo{IP: net.IPv4(127, 0, 0, 1)}, nil
}

// DeployContainer starts container
func (d *Direct) DeployContainer(ctx context.Context, app infra.Container) (infra.DeploymentInfo, error) {
	panic("not implemented yet")
}
