package coreum

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/crust/build/git"
	"github.com/CoreumFoundation/crust/build/golang"
	"github.com/CoreumFoundation/crust/build/tools"
)

const (
	blockchainName = "coreum"
	binaryName     = "cored"
	repoName       = "coreum"
	repoPath       = "."
	binaryPath     = "bin/" + binaryName
	testsDir       = repoPath + "/integration-tests"
	testsBinDir    = "bin/.cache/integration-tests"

	cosmovisorBinaryPath = "bin/cosmovisor"
	goCoverFlag          = "-cover"
	binaryOutputFlag     = "-o"
	linkStaticallyLDFlag = "-ldflags=-extldflags=-static"
)

var (
	tagsLocal  = []string{"netgo", "ledger"}
	tagsDocker = append([]string{"muslc"}, tagsLocal...)
)

// BuildCored builds all the versions of cored binary.
func BuildCored(ctx context.Context, deps build.DepsFunc) error {
	deps(BuildCoredLocally, BuildCoredInDocker)
	return nil
}

// BuildCoredLocally builds cored locally.
func BuildCoredLocally(ctx context.Context, deps build.DepsFunc) error {
	versionFlags, err := coredVersionLDFlags(ctx, tagsLocal)
	if err != nil {
		return err
	}

	return golang.Build(ctx, deps, golang.BinaryBuildConfig{
		TargetPlatform: tools.TargetPlatformLocal,
		PackagePath:    filepath.Join(repoPath, "cmd/cored"),
		CGOEnabled:     true,
		Flags: []string{
			goCoverFlag,
			versionFlags,
			"-tags=" + strings.Join(tagsLocal, ","),
			binaryOutputFlag + "=" + binaryPath,
		},
	})
}

// BuildCoredInDocker builds cored in docker.
func BuildCoredInDocker(ctx context.Context, deps build.DepsFunc) error {
	return buildCoredInDocker(ctx, deps, tools.TargetPlatformLinuxLocalArchInDocker, []string{goCoverFlag})
}

func buildCoredInDocker(
	ctx context.Context,
	deps build.DepsFunc,
	targetPlatform tools.TargetPlatform,
	extraFlags []string,
) error {
	versionFlags, err := coredVersionLDFlags(ctx, tagsDocker)
	if err != nil {
		return err
	}

	if err := tools.Ensure(ctx, tools.LibWASMMuslC, targetPlatform); err != nil {
		return err
	}

	binOutputPath := filepath.Join("bin", ".cache", binaryName, targetPlatform.String(), "bin", binaryName)
	return golang.Build(ctx, deps, golang.BinaryBuildConfig{
		TargetPlatform: targetPlatform,
		PackagePath:    "../coreum/cmd/cored",
		CGOEnabled:     true,
		Flags: append(
			extraFlags,
			versionFlags,
			linkStaticallyLDFlag,
			"-tags="+strings.Join(tagsDocker, ","),
			binaryOutputFlag+"="+binOutputPath,
		),
	})
}

// buildCoredClientInDocker builds cored binary without the wasm VM and with CGO disabled. The result binary might be
// used for the CLI on target platform, but can't be used to run the node.
func buildCoredClientInDocker(ctx context.Context, deps build.DepsFunc, targetPlatform tools.TargetPlatform) error {
	versionFlags, err := coredVersionLDFlags(ctx, tagsDocker)
	if err != nil {
		return err
	}

	binOutputPath := filepath.Join(
		"bin",
		".cache",
		binaryName,
		targetPlatform.String(),
		"bin",
		fmt.Sprintf("%s-client", binaryName),
	)
	return golang.Build(ctx, deps, golang.BinaryBuildConfig{
		TargetPlatform: targetPlatform,
		PackagePath:    filepath.Join(repoPath, "cmd/cored"),
		CGOEnabled:     false,
		Flags: []string{
			versionFlags,
			linkStaticallyLDFlag,
			"-tags=" + strings.Join(tagsDocker, ","),
			binaryOutputFlag + "=" + binOutputPath,
		},
	})
}

// Tidy runs `go mod tidy` for coreum repo.
func Tidy(ctx context.Context, deps build.DepsFunc) error {
	return golang.Tidy(ctx, repoPath, deps)
}

// Lint lints coreum repo.
func Lint(ctx context.Context, deps build.DepsFunc) error {
	deps(Generate, CompileAllSmartContracts, formatProto, lintProto, breakingProto)
	return golang.Lint(ctx, repoPath, deps)
}

// Test run unit tests in coreum repo.
func Test(ctx context.Context, deps build.DepsFunc) error {
	return golang.Test(ctx, repoPath, deps)
}

func coredVersionLDFlags(ctx context.Context, buildTags []string) (string, error) {
	hash, err := git.DirtyHeadHash(ctx, repoPath)
	if err != nil {
		return "", err
	}

	version, err := git.VersionFromTag(ctx, repoPath)
	if err != nil {
		return "", err
	}
	if version == "" {
		version = hash
	}
	ps := map[string]string{
		"github.com/cosmos/cosmos-sdk/version.Name":    blockchainName,
		"github.com/cosmos/cosmos-sdk/version.AppName": binaryName,
		"github.com/cosmos/cosmos-sdk/version.Version": version,
		"github.com/cosmos/cosmos-sdk/version.Commit":  hash,
	}

	if len(buildTags) > 0 {
		ps["github.com/cosmos/cosmos-sdk/version.BuildTags"] = strings.Join(buildTags, ",")
	}

	var ldFlags []string
	for k, v := range ps {
		ldFlags = append(ldFlags, fmt.Sprintf("-X %s=%s", k, v))
	}

	return "-ldflags=" + strings.Join(ldFlags, " "), nil
}

func formatProto(ctx context.Context, deps build.DepsFunc) error {
	deps(tools.EnsureBuf)

	cmd := exec.Command(tools.Path("bin/buf", tools.TargetPlatformLocal), "format", "-w")
	cmd.Dir = filepath.Join(repoPath, "proto", "coreum")
	return libexec.Exec(ctx, cmd)
}
