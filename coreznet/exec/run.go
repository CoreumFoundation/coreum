package exec

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/parallel"
)

func toolCmd(tool string, args []string) *exec.Cmd {
	verifyTool(tool)
	return exec.Command(tool, args...)
}

func verifyTool(tool string) {
	if _, err := exec.LookPath(tool); err != nil {
		panic(fmt.Errorf("%s is not available, please install it", tool))
	}
}

// Kill tries to terminate processes gracefully, after timeout it kills them
func Kill(ctx context.Context, pids []int) error {
	return parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		for _, pid := range pids {
			pid := pid
			spawn(fmt.Sprintf("%d", pid), parallel.Continue, func(ctx context.Context) error {
				return parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
					proc, err := os.FindProcess(pid)
					if err != nil {
						return err
					}
					spawn("waiter", parallel.Exit, func(ctx context.Context) error {
						_, _ = proc.Wait()
						return nil
					})
					spawn("killer", parallel.Continue, func(ctx context.Context) error {
						if err := proc.Signal(syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
							return err
						}
						select {
						case <-ctx.Done():
							return ctx.Err()
						case <-time.After(20 * time.Second):
						}
						if err := proc.Signal(syscall.SIGKILL); err != nil && !errors.Is(err, os.ErrProcessDone) {
							return err
						}
						return nil
					})
					return nil
				})
			})
		}
		return nil
	})
}
