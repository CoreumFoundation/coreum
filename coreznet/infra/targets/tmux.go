package targets

import (
	"context"
	"os"
	"sync"

	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/coreznet/exec"
	"github.com/CoreumFoundation/coreum/coreznet/infra"
)

// NewTMux creates new tmux target
func NewTMux(config infra.Config, spec *infra.Spec, docker *Docker) infra.Target {
	return &TMux{
		config: config,
		spec:   spec,
		docker: docker,
	}
}

// TMux is the target deploying apps to tmux session
type TMux struct {
	config infra.Config
	spec   *infra.Spec
	docker *Docker

	mu sync.Mutex // to protect tmux session
}

// Stop stops running applications
func (t *TMux) Stop(ctx context.Context) error {
	return t.docker.Stop(ctx)
}

// Remove removes running applications
func (t *TMux) Remove(ctx context.Context) error {
	return t.docker.Remove(ctx)
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
	info, err := t.docker.DeployBinary(ctx, app)
	if err != nil {
		return infra.DeploymentInfo{}, err
	}
	if err := t.sessionShowContainerLogs(ctx, app.Name, info.Container); err != nil {
		return infra.DeploymentInfo{}, err
	}
	return info, nil
}

// DeployContainer starts container inside tmux session
func (t *TMux) DeployContainer(ctx context.Context, app infra.Container) (infra.DeploymentInfo, error) {
	info, err := t.docker.DeployContainer(ctx, app)
	if err != nil {
		return infra.DeploymentInfo{}, err
	}
	if err := t.sessionShowContainerLogs(ctx, app.Name, info.Container); err != nil {
		return infra.DeploymentInfo{}, err
	}
	return info, nil
}

func (t *TMux) sessionShowContainerLogs(ctx context.Context, name string, container string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	hasSession, err := t.sessionExists(ctx)
	if err != nil {
		return err
	}

	cmd := []string{"docker", "logs", "-f", container}
	if hasSession {
		return libexec.Exec(ctx, exec.TMux(append([]string{"new-window", "-d", "-n", name, "-t", t.config.EnvName + ":"}, cmd...)...))
	}
	return libexec.Exec(ctx, exec.TMux(append([]string{"new-session", "-d", "-s", t.config.EnvName, "-n", name}, cmd...)...))
}

func (t *TMux) sessionAttach(ctx context.Context) error {
	cmd := exec.TMux("attach-session", "-t", t.config.EnvName)
	cmd.Stdin = os.Stdin
	return libexec.Exec(ctx, cmd)
}

func (t *TMux) sessionExists(ctx context.Context) (bool, error) {
	err := libexec.Exec(ctx, exec.TMuxNoOut("has-session", "-t", t.config.EnvName))
	if err != nil && errors.Is(err, ctx.Err()) {
		return false, err
	}
	return err == nil, nil
}
