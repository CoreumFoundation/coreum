package coreum

import (
	"context"

	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/CoreumFoundation/crust/build/config"
	"github.com/CoreumFoundation/crust/build/docker"
	"github.com/CoreumFoundation/crust/build/git"
	"github.com/CoreumFoundation/crust/build/tools"
)

// ReleaseCored releases cored binary for amd64 and arm64 to be published inside the release.
func ReleaseCored(ctx context.Context, deps build.DepsFunc) error {
	clean, _, err := git.StatusClean(ctx, repoPath)
	if err != nil {
		return err
	}
	if !clean {
		return errors.New("released commit contains uncommitted changes")
	}

	version, err := git.VersionFromTag(ctx, repoPath)
	if err != nil {
		return err
	}
	if version == "" {
		return errors.New("no version present on released commit")
	}

	if err := buildCoredClientInDocker(ctx, deps, tools.TargetPlatformDarwinAMD64InDocker); err != nil {
		return err
	}

	if err := buildCoredClientInDocker(ctx, deps, tools.TargetPlatformDarwinARM64InDocker); err != nil {
		return err
	}

	if err := buildCoredInDocker(ctx, deps, tools.TargetPlatformLinuxAMD64InDocker, []string{}); err != nil {
		return err
	}
	return buildCoredInDocker(ctx, deps, tools.TargetPlatformLinuxARM64InDocker, []string{})
}

// ReleaseCoredImage releases cored docker images for amd64 and arm64.
func ReleaseCoredImage(ctx context.Context, deps build.DepsFunc) error {
	deps(ReleaseCored)

	return buildCoredDockerImage(ctx, imageConfig{
		TargetPlatforms: []tools.TargetPlatform{
			tools.TargetPlatformLinuxAMD64InDocker,
			tools.TargetPlatformLinuxARM64InDocker,
		},
		Action:   docker.ActionPush,
		Username: config.DockerHubUsername,
	})
}
