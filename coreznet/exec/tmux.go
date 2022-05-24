package exec

import (
	"io"
	"os/exec"
)

// TMux runs tmux command
func TMux(args ...string) *exec.Cmd {
	return toolCmd("tmux", args)
}

// TMuxNoOut runs tmux command with discarded outputs
func TMuxNoOut(args ...string) *exec.Cmd {
	cmd := TMux(args...)
	cmd.Stderr = io.Discard
	cmd.Stdout = io.Discard
	return cmd
}
