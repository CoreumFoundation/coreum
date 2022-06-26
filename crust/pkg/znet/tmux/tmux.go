package tmux

import (
	"context"
	"os"

	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/crust/exec"
)

// Attach attaches to tmux session
func Attach(ctx context.Context, sessionName string) error {
	cmd := exec.TMux("attach-session", "-t", sessionName)
	cmd.Stdin = os.Stdin
	return libexec.Exec(ctx, cmd)
}

// ShowContainerLogs adds new window to tmux session presenting logs from docker container
func ShowContainerLogs(ctx context.Context, sessionName string, windowName string, container string) error {
	hasSession, err := sessionExists(ctx, sessionName)
	if err != nil {
		return err
	}

	cmd := []string{"docker", "logs", "-f", container}
	if hasSession {
		return libexec.Exec(ctx, exec.TMux(append([]string{"new-window", "-d", "-n", windowName, "-t", sessionName + ":"}, cmd...)...))
	}
	return libexec.Exec(ctx, exec.TMux(append([]string{"new-session", "-d", "-s", sessionName, "-n", windowName}, cmd...)...))
}

// Kill kills tmux session
func Kill(ctx context.Context, sessionName string) error {
	hasSession, err := sessionExists(ctx, sessionName)
	if err != nil {
		return err
	}
	if !hasSession {
		return nil
	}
	cmd := exec.TMux("kill-session", "-t", sessionName)
	cmd.Stdin = os.Stdin
	return libexec.Exec(ctx, cmd)
}

func sessionExists(ctx context.Context, sessionName string) (bool, error) {
	err := libexec.Exec(ctx, exec.TMuxNoOut("has-session", "-t", sessionName))
	if err != nil && errors.Is(err, ctx.Err()) {
		return false, err
	}
	return err == nil, nil
}
