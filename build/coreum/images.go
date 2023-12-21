package coreum

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/CoreumFoundation/coreum/v4/pkg/config/constant"
	"github.com/CoreumFoundation/crust/build/config"
	"github.com/CoreumFoundation/crust/build/coreum/image"
	"github.com/CoreumFoundation/crust/build/docker"
	"github.com/CoreumFoundation/crust/build/tools"
)

type imageConfig struct {
	TargetPlatforms []tools.TargetPlatform
	Action          docker.Action
	Username        string
	Versions        []string
}

// BuildCoredDockerImage builds cored docker image.
func BuildCoredDockerImage(ctx context.Context, deps build.DepsFunc) error {
	deps(BuildCoredInDocker, ensureReleasedBinaries)

	return buildCoredDockerImage(ctx, imageConfig{
		TargetPlatforms: []tools.TargetPlatform{tools.TargetPlatformLinuxLocalArchInDocker},
		Action:          docker.ActionLoad,
		Versions:        []string{config.ZNetVersion},
	})
}

func buildCoredDockerImage(ctx context.Context, cfg imageConfig) error {
	for _, platform := range cfg.TargetPlatforms {
		if err := ensureCosmovisor(ctx, platform); err != nil {
			return err
		}
	}
	dockerfile, err := image.Execute(image.Data{
		From:             docker.AlpineImage,
		CoredBinary:      binaryPath,
		CosmovisorBinary: cosmovisorBinaryPath,
		Networks: []string{
			string(constant.ChainIDDev),
			string(constant.ChainIDTest),
		},
	})
	if err != nil {
		return err
	}

	return docker.BuildImage(ctx, docker.BuildImageConfig{
		RepoPath:        repoPath,
		ContextDir:      filepath.Join("bin", ".cache", binaryName),
		ImageName:       binaryName,
		TargetPlatforms: cfg.TargetPlatforms,
		Action:          cfg.Action,
		Versions:        cfg.Versions,
		Username:        cfg.Username,
		Dockerfile:      dockerfile,
	})
}

// ensureReleasedBinaries ensures that all previous cored versions are installed.
func ensureReleasedBinaries(ctx context.Context, deps build.DepsFunc) error {
	for _, binaryTool := range []tools.Name{
		tools.CoredV302,
		tools.CoredV300,
		tools.CoredV202,
		tools.CoredV100,
	} {
		if err := tools.Ensure(ctx, binaryTool, tools.TargetPlatformLinuxLocalArchInDocker); err != nil {
			return err
		}
		if err := tools.CopyToolBinaries(
			binaryTool,
			tools.TargetPlatformLinuxLocalArchInDocker,
			filepath.Join("bin", ".cache", binaryName, tools.TargetPlatformLinuxLocalArchInDocker.String()),
			fmt.Sprintf("bin/%s", binaryTool)); err != nil {
			return err
		}
	}

	return nil
}
