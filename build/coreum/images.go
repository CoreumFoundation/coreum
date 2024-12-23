package coreum

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/CoreumFoundation/coreum/build/coreum/image"
	"github.com/CoreumFoundation/coreum/v5/pkg/config/constant"
	"github.com/CoreumFoundation/crust/build/config"
	"github.com/CoreumFoundation/crust/build/docker"
	"github.com/CoreumFoundation/crust/build/tools"
	"github.com/CoreumFoundation/crust/build/types"
)

type imageConfig struct {
	BinaryPath      string
	TargetPlatforms []tools.TargetPlatform
	Action          docker.Action
	Username        string
	Versions        []string
}

// BuildCoredDockerImage builds cored docker image.
func BuildCoredDockerImage(ctx context.Context, deps types.DepsFunc) error {
	deps(BuildCoredInDocker, ensureReleasedBinaries)

	return buildCoredDockerImage(ctx, imageConfig{
		BinaryPath:      binaryPath,
		TargetPlatforms: []tools.TargetPlatform{tools.TargetPlatformLinuxLocalArchInDocker},
		Action:          docker.ActionLoad,
		Versions:        []string{config.ZNetVersion},
	})
}

// BuildExtendedCoredDockerImage builds extended cored docker image.
func BuildExtendedCoredDockerImage(ctx context.Context, deps types.DepsFunc) error {
	deps(BuildExtendedCoredInDocker)

	return buildCoredDockerImage(ctx, imageConfig{
		BinaryPath:      extendedBinaryPath,
		TargetPlatforms: []tools.TargetPlatform{tools.TargetPlatformLinuxLocalArchInDocker},
		Action:          docker.ActionLoad,
		Versions:        []string{config.ZNetVersion},
	})
}

func buildCoredDockerImage(ctx context.Context, cfg imageConfig) error {
	binaryName := filepath.Base(cfg.BinaryPath)
	for _, platform := range cfg.TargetPlatforms {
		if err := ensureCosmovisorWithInstalledBinary(ctx, platform, binaryName); err != nil {
			return err
		}
	}
	dockerfile, err := image.Execute(image.Data{
		From:             docker.AlpineImage,
		CoredBinary:      cfg.BinaryPath,
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
func ensureReleasedBinaries(ctx context.Context, deps types.DepsFunc) error {
	const binaryTool = tools.CoredV401
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
	// copy the release binary for the local platform to use for the genesis generation
	if err := tools.Ensure(ctx, binaryTool, tools.TargetPlatformLocal); err != nil {
		return err
	}
	return tools.CopyToolBinaries(
		binaryTool,
		tools.TargetPlatformLocal,
		filepath.Join("bin", ".cache", binaryName, tools.TargetPlatformLocal.String()),
		fmt.Sprintf("bin/%s", binaryTool),
	)
}
