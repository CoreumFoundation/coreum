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
)

var defaultBuildTags = []string{"netgo", "ledger"}

// BuildCored builds all the versions of cored binary.
func BuildCored(ctx context.Context, deps types.DepsFunc) error {
	deps(BuildCoredLocally, BuildCoredInDocker)
	return nil
}

// BuildCoredLocally builds cored locally.
func BuildCoredLocally(ctx context.Context, deps types.DepsFunc) error {
	ldFlags, err := coredVersionLDFlags(ctx, defaultBuildTags, "")
	if err != nil {
		return err
	}

	return golang.Build(ctx, deps, golang.BinaryBuildConfig{
		TargetPlatform: tools.TargetPlatformLocal,
		PackagePath:    filepath.Join(repoPath, "cmd/cored"),
		BinOutputPath:  binaryPath,
		CGOEnabled:     true,
		Tags:           defaultBuildTags,
		Flags: []string{
			goCoverFlag,
		},
		LDFlags: ldFlags,
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

	if err := Tidy(ctx, deps); err != nil {
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
	ldFlags, err := coredVersionLDFlags(ctx, defaultBuildTags, mod)
	if err != nil {
		return err
	}
	ldFlags = append(ldFlags, "-linkmode external")

	if err := tools.Ensure(ctx, tools.LibWASM, targetPlatform); err != nil {
		return err
	}

	var cc string
	buildTags := defaultBuildTags
	envs := make([]string, 0)
	dockerVolumes := make([]string, 0)
	switch targetPlatform.OS {
	case tools.OSLinux:
		// osusergo tag is required for the cross compilation to awit warnings, https://pkg.go.dev/os/user
		buildTags = append(buildTags, "muslc", "osusergo")
		ldFlags = append(ldFlags, "-extldflags '-Wl,-z,muldefs -static -lm'")
		switch targetPlatform {
		case tools.TargetPlatformLinuxAMD64InDocker:
			cc = "x86_64-linux-gnu-gcc"
			wasmHostDirPath := tools.Path("lib/libwasmvm_muslc.x86_64.a", targetPlatform)
			dockerVolumes = append(
				dockerVolumes,
				fmt.Sprintf("%s:%s", wasmHostDirPath, "/usr/lib/x86_64-linux-gnu/libwasmvm_muslc.a"),
			)
		case tools.TargetPlatformLinuxARM64InDocker:
			cc = "aarch64-linux-gnu-gcc"
			wasmHostDirPath := tools.Path("lib/libwasmvm_muslc.aarch64.a", targetPlatform)
			dockerVolumes = append(
				dockerVolumes,
				fmt.Sprintf("%s:%s", wasmHostDirPath, "/usr/lib/aarch64-linux-gnu/libwasmvm_muslc.a"),
			)
		default:
			return errors.Errorf("building is not possible for platform %s", targetPlatform)
		}
	case tools.OSDarwin:
		buildTags = append(buildTags, "static_wasm")
		switch targetPlatform {
		case tools.TargetPlatformDarwinAMD64InDocker:
			cc = "o64-clang"
		case tools.TargetPlatformDarwinARM64InDocker:
			cc = "oa64-clang"
		default:
			return errors.Errorf("building is not possible for platform %s", targetPlatform)
		}
		wasmHostDirPath := tools.Path("lib/libwasmvmstatic_darwin.a", targetPlatform)
		dockerVolumes = append(dockerVolumes, fmt.Sprintf("%s:%s", wasmHostDirPath, "/lib/libwasmvmstatic_darwin.a"))
		envs = append(envs, "CGO_LDFLAGS=-L/lib")
	default:
		return errors.Errorf("building is not possible for platform %s", targetPlatform)
	}
	envs = append(envs, fmt.Sprintf("CC=%s", cc))

	binOutputPath := filepath.Join("bin", ".cache", binaryName, targetPlatform.String(), "bin", binaryName)
	return golang.Build(ctx, deps, golang.BinaryBuildConfig{
		TargetPlatform: targetPlatform,
		PackagePath:    filepath.Join(repoPath, "cmd/cored"),
		BinOutputPath:  binOutputPath,
		CGOEnabled:     true,
		Tags:           buildTags,
		LDFlags:        ldFlags,
		Flags:          extraFlags,
		Envs:           envs,
		DockerVolumes:  dockerVolumes,
	})
}

// Tidy runs `go mod tidy` for coreum repo.
func Tidy(ctx context.Context, deps types.DepsFunc) error {
	return golang.Tidy(ctx, repoPath, deps)
}

// Lint lints coreum repo.
func Lint(ctx context.Context, deps types.DepsFunc) error {
	deps(Generate, CompileAllSmartContracts, formatProto, lintProto, breakingProto)
	return golang.Lint(ctx, repoPath, deps)
}

// Test run unit tests in coreum repo.
func Test(ctx context.Context, deps types.DepsFunc) error {
	deps(CompileAllSmartContracts)

	return golang.Test(ctx, repoPath, deps)
}

// DownloadDependencies downloads go dependencies.
func DownloadDependencies(ctx context.Context, deps types.DepsFunc) error {
	return golang.DownloadDependencies(ctx, repoPath, deps)
}

func coredVersionLDFlags(ctx context.Context, buildTags []string, mod string) ([]string, error) {
	hash, err := git.DirtyHeadHash(ctx, repoPath)
	if err != nil {
		return nil, err
	}

	version, err := git.VersionFromTag(ctx, repoPath)
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
