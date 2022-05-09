package exec

import (
	"io"
	"os/exec"
)

var tmux string

func init() {
	for _, app := range []string{"tmux"} {
		if _, err := exec.LookPath(app); err == nil {
			tmux = app
			break
		}
	}
}

// TMux runs tmux command
func TMux(args ...string) *exec.Cmd {
	return exec.Command(tmux, args...)
}

// TMuxNoOut runs tmux command with discarded outputs
func TMuxNoOut(args ...string) *exec.Cmd {
	cmd := exec.Command(tmux, args...)
	cmd.Stderr = io.Discard
	cmd.Stdout = io.Discard
	return cmd
}
