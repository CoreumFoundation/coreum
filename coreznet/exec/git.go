package exec

import (
	"os/exec"
)

// Git runs git command
func Git(args ...string) *exec.Cmd {
	return exec.Command("git", args...)
}
