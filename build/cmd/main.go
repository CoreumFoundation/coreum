package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/CoreumFoundation/coreum-tools/pkg/ioc"
	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum-tools/pkg/run"

	selfBuild "github.com/CoreumFoundation/coreum/build"
)

func main() {
	run.Tool("build", nil, func(ctx context.Context, c *ioc.Container) error {
		flags := logger.Flags(logger.ToolDefaultConfig, "build")
		if err := flags.Parse(os.Args[1:]); err != nil {
			return err
		}
		exec := build.NewIoCExecutor(selfBuild.Commands, c)
		if build.Autocomplete(exec) {
			return nil
		}

		changeWorkingDir()
		return build.Do(ctx, "coreum", flags.Args(), exec)
	})
}

// changeWorkingDir sets working dir to the root directory of repository
func changeWorkingDir() {
	must.OK(os.Chdir(filepath.Dir(filepath.Dir(must.String(filepath.EvalSymlinks(must.String(os.Executable())))))))
}
