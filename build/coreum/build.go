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

	binaryPath          = "bin/" + binaryName
	extendedBinaryPath  = "bin/" + extendedBinaryName
	integrationTestsDir = repoPath + "/integration-tests"
	cometBFTCommit      = "099b4104e5b00b3cedd2c06ca3b1270baad2f4e9"

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
		PackagePath:    "cmd/cored",
		BinOutputPath:  binaryPath,
		CGOEnabled:     true,
		Tags:           defaultBuildTags,
		LDFlags:        ldFlags,
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

	return git.RollbackChanges(ctx, "go.mod", "go.sum", "go.work.sum")
}

func buildCoredInDocker(
	ctx context.Context,
	deps types.DepsFunc,
	targetPlatform tools.TargetPlatform,
	extraFlags []string,
	binaryName string,
	mod string,
) error {
	if err := tools.Ensure(ctx, tools.LibWASM, targetPlatform); err != nil {
		return err
	}

	ldFlags := make([]string, 0)
	var cc string
	buildTags := defaultBuildTags
	envs := make([]string, 0)
	dockerVolumes := make([]string, 0)
	switch targetPlatform.OS {
	case tools.OSLinux:
		// use cc not installed on the image we use for the build
		if err := tools.Ensure(ctx, tools.MuslCC, targetPlatform); err != nil {
			return err
		}
		buildTags = append(buildTags, "muslc")
		ldFlags = append(ldFlags, "-extldflags '-static'")
		var (
			hostCCDirPath string
			// path inside hostCCDirPath to the CC
			ccRelativePath string

			wasmHostDirPath string
			// path to the wasm lib in the CC
			wasmCCLibRelativeLibPath string
		)
		switch targetPlatform {
		case tools.TargetPlatformLinuxAMD64InDocker:
			hostCCDirPath = filepath.Dir(
				filepath.Dir(tools.Path("bin/x86_64-linux-musl-gcc", targetPlatform)),
			)
			ccRelativePath = "/bin/x86_64-linux-musl-gcc"
			wasmHostDirPath = tools.Path("lib/libwasmvm_muslc.x86_64.a", targetPlatform)
			wasmCCLibRelativeLibPath = "/x86_64-linux-musl/lib/libwasmvm_muslc.x86_64.a"
		case tools.TargetPlatformLinuxARM64InDocker:
			hostCCDirPath = filepath.Dir(
				filepath.Dir(tools.Path("bin/aarch64-linux-musl-gcc", targetPlatform)),
			)
			ccRelativePath = "/bin/aarch64-linux-musl-gcc"
			wasmHostDirPath = tools.Path("lib/libwasmvm_muslc.aarch64.a", targetPlatform)
			wasmCCLibRelativeLibPath = "/aarch64-linux-musl/lib/libwasmvm_muslc.aarch64.a"
		default:
			return errors.Errorf("building is not possible for platform %s", targetPlatform)
		}
		const ccDockerDir = "/musl-gcc"
		dockerVolumes = append(
			dockerVolumes,
			fmt.Sprintf("%s:%s", hostCCDirPath, ccDockerDir),
			// put the libwasmvm to the lib folder of the compiler
			fmt.Sprintf("%s:%s", wasmHostDirPath, fmt.Sprintf("%s%s", ccDockerDir, wasmCCLibRelativeLibPath)),
		)
		cc = fmt.Sprintf("%s%s", ccDockerDir, ccRelativePath)
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

	versionLDFlags, err := coredVersionLDFlags(ctx, buildTags, mod)
	if err != nil {
		return err
	}
	ldFlags = append(ldFlags, versionLDFlags...)

	binOutputPath := filepath.Join("bin", ".cache", binaryName, targetPlatform.String(), "bin", binaryName)
	return golang.Build(ctx, deps, golang.BinaryBuildConfig{
		TargetPlatform: targetPlatform,
		PackagePath:    "cmd/cored",
		BinOutputPath:  binOutputPath,
		CGOEnabled:     true,
		Tags:           buildTags,
		LDFlags:        ldFlags,
		Flags:          extraFlags,
		Envs:           envs,
		DockerVolumes:  dockerVolumes,
	})
}

// Lint lints coreum repo.
func Lint(ctx context.Context, deps types.DepsFunc) error {
	deps(
		Generate,
		CompileAllSmartContracts,
		formatProto,
		lintProto,
		//breakingProto, // TODO: uncomment in next PR
	)
	return golang.Lint(ctx, deps)
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
