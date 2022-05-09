package exec

import (
	"os/exec"
)

// Docker runs docker command
func Docker(args ...string) *exec.Cmd {
	return exec.Command("docker", args...)
}
