package main

import (
	"context"
	"os"
	"path/filepath"

	me "github.com/CoreumFoundation/coreum/build"
	"github.com/CoreumFoundation/coreum/lib/must"
	"github.com/outofforest/build"
	"github.com/outofforest/ioc/v2"
	"github.com/outofforest/run"
)

func main() {
	run.Tool("build", nil, func(ctx context.Context, c *ioc.Container) error {
		exec := build.NewIoCExecutor(me.Commands, c)
		if build.Autocomplete(exec) {
			return nil
		}

		changeWorkingDir()
		return build.Do(ctx, "coreum", exec)
	})
}

// changeWorkingDir sets working dir to the location where executed file exists
func changeWorkingDir() {
	must.OK(os.Chdir(filepath.Dir(filepath.Dir(must.String(filepath.EvalSymlinks(must.String(os.Executable())))))))
}
