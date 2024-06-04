package coreum

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/crust/build/git"
	"github.com/CoreumFoundation/crust/build/golang"
	"github.com/CoreumFoundation/crust/build/tools"
	"github.com/CoreumFoundation/crust/build/types"
)

const (
	blockchainName     = "coreum"
	binaryName         = "cored"
	extendedBinaryName = "cored-ext"
	repoPath           = "."
	binaryPath         = "bin/" + binaryName
	extendedBinaryPath = "bin/" + extendedBinaryName
	testsDir           = repoPath + "/integration-tests"
	cometBFTCommit     = "2644973fb58663f435aac0c6bdf9502fe78798a0"

	cosmovisorBinaryPath = "bin/cosmovisor"
	goCoverFlag          = "-cover"
	tagsFlag             = "-tags"
	ldFlagsFlag          = "-ldflags"
	linkStaticallyValue  = "-extldflags=-static"
)

var (
	tagsLocal  = []string{"netgo", "ledger"}
	tagsDocker = append([]string{"muslc"}, tagsLocal...)
)

// BuildCored builds all the versions of cored binary.
func BuildCored(ctx context.Context, deps types.DepsFunc) error {
	deps(BuildCoredLocally, BuildCoredInDocker)
	return nil
}

// BuildCoredLocally builds cored locally.
func BuildCoredLocally(ctx context.Context, deps types.DepsFunc) error {
	versionFlags, err := coredVersionLDFlags(ctx, tagsLocal, "")
	if err != nil {
		return err
	}

	return golang.Build(ctx, deps, golang.BinaryBuildConfig{
		TargetPlatform: tools.TargetPlatformLocal,
		PackagePath:    "cmd/cored",
		BinOutputPath:  binaryPath,
		CGOEnabled:     true,
		Flags: []string{
			goCoverFlag,
			convertToLdFlags(versionFlags),
			tagsFlag + "=" + strings.Join(tagsLocal, ","),
		},
	})
}

// BuildCoredInDocker builds cored in docker.
func BuildCoredInDocker(ctx context.Context, deps types.DepsFunc) error {
	return buildCoredInDocker(ctx, deps, tools.TargetPlatformLinuxLocalArchInDocker, []string{goCoverFlag},
		binaryName, "")
}

// BuildExtendedCoredInDocker builds extended cored in docker.
func BuildExtendedCoredInDocker(ctx context.Context, deps types.DepsFunc) error {
	f, err := os.OpenFile("go.mod", os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	_, err = f.WriteString("replace github.com/cometbft/cometbft => github.com/CoreumFoundation/cometbft " +
		cometBFTCommit)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := golang.Tidy(ctx, deps); err != nil {
		return err
	}

	err = buildCoredInDocker(ctx, deps, tools.TargetPlatformLinuxLocalArchInDocker, []string{goCoverFlag},
		extendedBinaryName, "ext")
	if err != nil {
		return err
	}

	return git.RollbackChanges(ctx, repoPath, "go.mod", "go.sum", "go.work.sum")
}

func buildCoredInDocker(
	ctx context.Context,
	deps types.DepsFunc,
	targetPlatform tools.TargetPlatform,
	extraFlags []string,
	binaryName string,
	mod string,
) error {
	versionFlags, err := coredVersionLDFlags(ctx, tagsDocker, mod)
	if err != nil {
		return err
	}

	if err := tools.Ensure(ctx, tools.LibWASMMuslC, targetPlatform); err != nil {
		return err
	}

	binOutputPath := filepath.Join("bin", ".cache", binaryName, targetPlatform.String(), "bin", binaryName)
	return golang.Build(ctx, deps, golang.BinaryBuildConfig{
		TargetPlatform: targetPlatform,
		PackagePath:    "cmd/cored",
		BinOutputPath:  binOutputPath,
		CGOEnabled:     true,
		Flags: append(
			extraFlags,
			convertToLdFlags(append(versionFlags, linkStaticallyValue)),
			tagsFlag+"="+strings.Join(tagsDocker, ","),
		),
	})
}

// buildCoredClientInDocker builds cored binary without the wasm VM and with CGO disabled. The result binary might be
// used for the CLI on target platform, but can't be used to run the node.
func buildCoredClientInDocker(ctx context.Context, deps types.DepsFunc, targetPlatform tools.TargetPlatform) error {
	versionFlags, err := coredVersionLDFlags(ctx, tagsDocker, "")
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
		PackagePath:    "cmd/cored",
		BinOutputPath:  binOutputPath,
		CGOEnabled:     false,
		Flags: []string{
			convertToLdFlags(append(versionFlags, linkStaticallyValue)),
			tagsFlag + "=" + strings.Join(tagsDocker, ","),
		},
	})
}

// Lint lints coreum repo.
func Lint(ctx context.Context, deps types.DepsFunc) error {
	deps(Generate, CompileAllSmartContracts, formatProto, lintProto, breakingProto)
	return golang.Lint(ctx, deps)
}

// Test run unit tests in coreum repo.
func Test(ctx context.Context, deps types.DepsFunc) error {
	deps(CompileAllSmartContracts)

	return golang.Test(ctx, deps)
}

// DownloadDependencies downloads go dependencies.
func DownloadDependencies(ctx context.Context, deps types.DepsFunc) error {
	return golang.DownloadDependencies(ctx, deps, repoPath)
}

func coredVersionLDFlags(ctx context.Context, buildTags []string, mod string) ([]string, error) {
	hash, err := git.DirtyHeadHash(ctx)
	if err != nil {
		return nil, err
	}

	version, err := git.VersionFromTag(ctx)
	if err != nil {
		return nil, err
	}
	if version == "" {
		version = hash
	}
	if mod != "" {
		version += "+" + mod
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

	var values []string
	for k, v := range ps {
		values = append(values, fmt.Sprintf("-X %s=%s", k, v))
	}

	return values, nil
}

func formatProto(ctx context.Context, deps types.DepsFunc) error {
	deps(tools.EnsureBuf)

	cmd := exec.Command(tools.Path("bin/buf", tools.TargetPlatformLocal), "format", "-w")
	cmd.Dir = filepath.Join(repoPath, "proto", "coreum")
	return libexec.Exec(ctx, cmd)
}

func convertToLdFlags(values []string) string {
	return "-" + ldFlagsFlag + "=" + strings.Join(values, " ")
}
