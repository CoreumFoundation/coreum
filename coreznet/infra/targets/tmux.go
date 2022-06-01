package targets

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	osexec "os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/coreznet/exec"
	"github.com/CoreumFoundation/coreum/coreznet/infra"
)

// NewTMux creates new tmux target
func NewTMux(config infra.Config, spec *infra.Spec) infra.Target {
	return &TMux{
		config: config,
		spec:   spec,
	}
}

// TMux is the target deploying apps to tmux session
type TMux struct {
	config infra.Config
	spec   *infra.Spec

	mu sync.Mutex // to protect tmux session
}

// Stop stops running applications
func (t *TMux) Stop(ctx context.Context) error {
	return t.sessionKill(ctx)
}

// Remove removes running applications
func (t *TMux) Remove(ctx context.Context) error {
	return t.Stop(ctx)
}

// Deploy deploys environment to tmux target
func (t *TMux) Deploy(ctx context.Context, mode infra.Mode) error {
	if err := mode.Deploy(ctx, t, t.config, t.spec); err != nil {
		return err
	}
	if t.config.TestingMode {
		return nil
	}
	return t.sessionAttach(ctx)
}

// DeployBinary starts binary file inside tmux session
func (t *TMux) DeployBinary(ctx context.Context, app infra.Binary) (infra.DeploymentInfo, error) {
	binPath := app.BinPathFunc(runtime.GOOS)
	must.Any(os.Stat(binPath))
	if err := t.sessionAddApp(ctx, app.Name, append([]string{binPath}, app.ArgsFunc(net.IPv4(127, 0, 0, 1), t.config.AppDir+"/"+app.Name)...)...); err != nil {
		return infra.DeploymentInfo{}, err
	}
	return infra.DeploymentInfo{IP: net.IPv4(127, 0, 0, 1)}, nil
}

// DeployContainer starts container inside tmux session
func (t *TMux) DeployContainer(ctx context.Context, app infra.Container) (infra.DeploymentInfo, error) {
	panic("not implemented yet")
}

func (t *TMux) sessionAddApp(ctx context.Context, name string, args ...string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	hasSession, err := t.sessionExists(ctx)
	if err != nil {
		return err
	}
	cmd := []string{
		"/bin/sh", "-ce",
		fmt.Sprintf(`exec %s > >(tee -a "%s/%s.log") 2>&1`, osexec.Command("", args...).String(), t.config.LogDir, name),
	}
	if hasSession {
		return libexec.Exec(ctx, exec.TMux(append([]string{"new-window", "-d", "-n", name, "-t", t.config.EnvName + ":"}, cmd...)...))
	}
	return libexec.Exec(ctx, exec.TMux(append([]string{"new-session", "-d", "-s", t.config.EnvName, "-n", name}, cmd...)...))
}

func (t *TMux) sessionAttach(ctx context.Context) error {
	cmd := exec.TMux("attach-session", "-t", t.config.EnvName)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return libexec.Exec(ctx, cmd)
}

func (t *TMux) sessionKill(ctx context.Context) error {
	// When using just `tmux kill-session` tmux sends SIGHUP to process, but we need SIGTERM.
	// After sending it to all apps, session is terminated automatically.

	// FIXME (wojciech): Yaroslav reports that on macOS tmux session is still there

	t.mu.Lock()
	defer t.mu.Unlock()

	if hasSession, err := t.sessionExists(ctx); err != nil || !hasSession {
		return err
	}
	pids, err := t.sessionPIDs(ctx)
	if err != nil || len(pids) == 0 {
		return err
	}
	return exec.Kill(ctx, pids)
}

func (t *TMux) sessionPIDs(ctx context.Context) ([]int, error) {
	buf := &bytes.Buffer{}
	cmd := exec.TMux("list-windows", "-t", t.config.EnvName, "-F", "#{pane_pid}")
	cmd.Stdout = buf
	if err := libexec.Exec(ctx, cmd); err != nil {
		return nil, err
	}
	var pids []int
	for _, pidStr := range strings.Split(buf.String(), "\n") {
		if pidStr == "" {
			break
		}
		pids = append(pids, int(must.Int64(strconv.ParseInt(pidStr, 10, 32))))
	}
	return pids, nil
}

func (t *TMux) sessionExists(ctx context.Context) (bool, error) {
	err := libexec.Exec(ctx, exec.TMuxNoOut("has-session", "-t", t.config.EnvName))
	if err != nil && errors.Is(err, ctx.Err()) {
		return false, err
	}
	return err == nil, nil
}
