package build

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/pkg/errors"
)

// gitStatusClean checks that there are no uncommitted files
func gitStatusClean(ctx context.Context) error {
	buf := &bytes.Buffer{}
	cmd := exec.Command("git", "status", "-s")
	cmd.Stdout = buf
	if err := libexec.Exec(ctx, cmd); err != nil {
		return fmt.Errorf("git command failed: %w", err)
	}
	if buf.Len() > 0 {
		fmt.Println("git status:")
		fmt.Println(buf)
		return errors.New("git status is not empty")
	}
	return nil
}
