package coreum

import (
	"context"
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	coreumtools "github.com/CoreumFoundation/coreum/build/tools"
	crusttools "github.com/CoreumFoundation/crust/build/tools"
	"github.com/CoreumFoundation/crust/build/types"
)

//go:embed typos.toml
var configTyposLint []byte

func lintTypos(ctx context.Context, deps types.DepsFunc) error {
	return executeLintTyposCommand(ctx, deps, repoPath)
}

func executeLintTyposCommand(ctx context.Context, deps types.DepsFunc, includePath string) error {
	deps(coreumtools.EnsureTypos)

	typosConfigPath := filepath.Join("bin", ".cache", "typos.toml")
	if err := os.WriteFile(typosConfigPath, configTyposLint, 0o644); err != nil {
		return errors.Wrap(err, "failed to write typos config file")
	}

	cmd := exec.Command(crusttools.Path("bin/typos", crusttools.TargetPlatformLocal),
		"--config",
		typosConfigPath,
		includePath)
	return libexec.Exec(ctx, cmd)
}
