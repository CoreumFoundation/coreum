package exec

import (
	"os/exec"

	"github.com/pkg/errors"
)

func toolCmd(tool string, args []string) *exec.Cmd {
	verifyTool(tool)
	return exec.Command(tool, args...)
}

func verifyTool(tool string) {
	if _, err := exec.LookPath(tool); err != nil {
		panic(errors.Errorf("%s is not available, please install it", tool))
	}
}
