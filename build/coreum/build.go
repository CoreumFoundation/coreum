package coreum

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	coreumtools "github.com/CoreumFoundation/coreum/build/tools"
	"github.com/CoreumFoundation/crust/build/config"
	"github.com/CoreumFoundation/crust/build/docker"
	dockerbasic "github.com/CoreumFoundation/crust/build/docker/basic"
	"github.com/CoreumFoundation/crust/build/git"
	"github.com/CoreumFoundation/crust/build/golang"
	"github.com/CoreumFoundation/crust/build/lint"
	crusttools "github.com/CoreumFoundation/crust/build/tools"
	"github.com/CoreumFoundation/crust/build/types"
)

const (
	blockchainName    = "coreum"
	binaryName        = "cored"
	gaiaBinaryName    = "gaiad"
	hermesBinaryName  = "hermes"
	osmosisBinaryName = "osmosisd"
	repoPath          = "."

	binaryPath          = "bin/" + binaryName
	gaiaBinaryPath      = "bin/" + gaiaBinaryName
	hermesBinaryPath    = "bin/" + hermesBinaryName
	osmosisBinaryPath   = "bin/" + osmosisBinaryName
	integrationTestsDir = repoPath + "/integration-tests"

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
	ldFlags, err := coredVersionLDFlags(ctx, defaultBuildTags)
	if err != nil {
		return err
	}

	return golang.Build(ctx, deps, golang.BinaryBuildConfig{
		TargetPlatform: crusttools.TargetPlatformLocal,
		PackagePath:    "cmd/cored",
		BinOutputPath:  binaryPath,
		CGOEnabled:     true,
		Tags:           defaultBuildTags,
		LDFlags:        ldFlags,
	})
}

// BuildCoredInDocker builds cored in docker.
func BuildCoredInDocker(ctx context.Context, deps types.DepsFunc) error {
	return buildCoredInDocker(ctx, deps, crusttools.TargetPlatformLinuxLocalArchInDocker, []string{goCoverFlag})
}

// BuildGaiaDockerImage builds docker image of the gaia.
func BuildGaiaDockerImage(ctx context.Context, deps types.DepsFunc) error {
	if err := crusttools.Ensure(ctx, coreumtools.Gaia, crusttools.TargetPlatformLinuxAMD64InDocker); err != nil {
		return err
	}

	gaiaLocalPath := filepath.Join(
		"bin", ".cache", gaiaBinaryName, crusttools.TargetPlatformLinuxAMD64InDocker.String(),
	)
	if err := crusttools.CopyToolBinaries(
		coreumtools.Gaia,
		crusttools.TargetPlatformLinuxAMD64InDocker,
		gaiaLocalPath,
		gaiaBinaryPath,
	); err != nil {
		return err
	}

	dockerfile, err := dockerbasic.Execute(dockerbasic.Data{
		From:   docker.AlpineImage,
		Binary: gaiaBinaryPath,
	})
	if err != nil {
		return err
	}

	return docker.BuildImage(ctx, docker.BuildImageConfig{
		ContextDir:      gaiaLocalPath,
		ImageName:       gaiaBinaryName,
		TargetPlatforms: []crusttools.TargetPlatform{crusttools.TargetPlatformLinuxAMD64InDocker},
		Dockerfile:      dockerfile,
		Versions:        []string{config.ZNetVersion},
	})
}

// BuildHermesDockerImage builds docker image of the ibc relayer.
func BuildHermesDockerImage(ctx context.Context, deps types.DepsFunc) error {
	if err := crusttools.Ensure(ctx, coreumtools.Hermes, crusttools.TargetPlatformLinuxAMD64InDocker); err != nil {
		return err
	}

	hermesLocalPath := filepath.Join(
		"bin", ".cache", hermesBinaryName, crusttools.TargetPlatformLinuxAMD64InDocker.String(),
	)
	if err := crusttools.CopyToolBinaries(
		coreumtools.Hermes,
		crusttools.TargetPlatformLinuxAMD64InDocker,
		hermesLocalPath,
		hermesBinaryPath,
	); err != nil {
		return err
	}

	dockerfile, err := dockerbasic.Execute(dockerbasic.Data{
		From:   docker.UbuntuImage,
		Binary: hermesBinaryPath,
		Run:    "apt update && apt install curl jq -y",
	})
	if err != nil {
		return err
	}

	return docker.BuildImage(ctx, docker.BuildImageConfig{
		ContextDir:      hermesLocalPath,
		ImageName:       hermesBinaryName,
		TargetPlatforms: []crusttools.TargetPlatform{crusttools.TargetPlatformLinuxAMD64InDocker},
		Dockerfile:      dockerfile,
		Versions:        []string{config.ZNetVersion},
	})
}

// BuildOsmosisDockerImage builds docker image of the osmosis.
func BuildOsmosisDockerImage(ctx context.Context, deps types.DepsFunc) error {
	if err := crusttools.Ensure(ctx, coreumtools.Osmosis, crusttools.TargetPlatformLinuxLocalArchInDocker); err != nil {
		return err
	}

	binaryLocalPath := filepath.Join(
		"bin", ".cache", osmosisBinaryName, crusttools.TargetPlatformLinuxLocalArchInDocker.String(),
	)
	if err := crusttools.CopyToolBinaries(
		coreumtools.Osmosis,
		crusttools.TargetPlatformLinuxLocalArchInDocker,
		binaryLocalPath,
		osmosisBinaryPath,
	); err != nil {
		return err
	}

	dockerfile, err := dockerbasic.Execute(dockerbasic.Data{
		From:   docker.AlpineImage,
		Binary: osmosisBinaryPath,
	})
	if err != nil {
		return err
	}

	return docker.BuildImage(ctx, docker.BuildImageConfig{
		ContextDir: binaryLocalPath,
		ImageName:  osmosisBinaryName,
		Dockerfile: dockerfile,
		Versions:   []string{config.ZNetVersion},
	})
}

func buildCoredInDocker(
	ctx context.Context,
	deps types.DepsFunc,
	targetPlatform crusttools.TargetPlatform,
	extraFlags []string,
) error {
	if err := crusttools.Ensure(ctx, coreumtools.LibWASM, targetPlatform); err != nil {
		return err
	}

	ldFlags := make([]string, 0)
	var cc string
	buildTags := defaultBuildTags
	envs := make([]string, 0)
	dockerVolumes := make([]string, 0)
	switch targetPlatform.OS {
	case crusttools.OSLinux:
		// use cc not installed on the image we use for the build
		if err := crusttools.Ensure(ctx, coreumtools.MuslCC, targetPlatform); err != nil {
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
		case crusttools.TargetPlatformLinuxAMD64InDocker:
			hostCCDirPath = filepath.Dir(
				filepath.Dir(crusttools.Path("bin/x86_64-linux-musl-gcc", targetPlatform)),
			)
			ccRelativePath = "/bin/x86_64-linux-musl-gcc"
			wasmHostDirPath = crusttools.Path("lib/libwasmvm_muslc.x86_64.a", targetPlatform)
			wasmCCLibRelativeLibPath = "/x86_64-linux-musl/lib/libwasmvm_muslc.x86_64.a"
		case crusttools.TargetPlatformLinuxARM64InDocker:
			hostCCDirPath = filepath.Dir(
				filepath.Dir(crusttools.Path("bin/aarch64-linux-musl-gcc", targetPlatform)),
			)
			ccRelativePath = "/bin/aarch64-linux-musl-gcc"
			wasmHostDirPath = crusttools.Path("lib/libwasmvm_muslc.aarch64.a", targetPlatform)
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
	case crusttools.OSDarwin:
		buildTags = append(buildTags, "static_wasm")
		switch targetPlatform {
		case crusttools.TargetPlatformDarwinAMD64InDocker:
			cc = "o64-clang"
		case crusttools.TargetPlatformDarwinARM64InDocker:
			cc = "oa64-clang"
		default:
			return errors.Errorf("building is not possible for platform %s", targetPlatform)
		}
		wasmHostDirPath := crusttools.Path("lib/libwasmvmstatic_darwin.a", targetPlatform)
		dockerVolumes = append(dockerVolumes, fmt.Sprintf("%s:%s", wasmHostDirPath, "/lib/libwasmvmstatic_darwin.a"))
		envs = append(envs, "CGO_LDFLAGS=-L/lib")
	default:
		return errors.Errorf("building is not possible for platform %s", targetPlatform)
	}
	envs = append(envs, "CC="+cc)

	versionLDFlags, err := coredVersionLDFlags(ctx, buildTags)
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
		// breakingProto, TODO(Restore breaking proto)
	)
	return lint.Lint(ctx, deps)
}

func coredVersionLDFlags(ctx context.Context, buildTags []string) ([]string, error) {
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
	deps(coreumtools.EnsureBuf)

	cmd := exec.Command(crusttools.Path("bin/buf", crusttools.TargetPlatformLocal), "format", "-w")
	cmd.Dir = filepath.Join(repoPath, "proto", "coreum")
	return libexec.Exec(ctx, cmd)
}
