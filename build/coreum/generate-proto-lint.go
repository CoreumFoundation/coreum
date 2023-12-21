package coreum

import (
	"context"
	_ "embed"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/crust/build/tools"
)

//go:embed proto-lint.json
var configLint []byte

func lintProto(ctx context.Context, deps build.DepsFunc) error {
	deps(Tidy)

	_, includeDirs, err := protoCDirectories(ctx, repoPath, deps)
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return errors.WithStack(err)
	}

	generateDirs := []string{
		filepath.Join(absPath, "proto"),
	}

	err = executeLintProtocCommand(ctx, deps, includeDirs, generateDirs)
	if err != nil {
		return err
	}

	return nil
}

func executeLintProtocCommand(ctx context.Context, deps build.DepsFunc, includeDirs, generateDirs []string) error {
	deps(tools.EnsureProtoc, tools.EnsureProtocGenBufLint)

	// Linting rule descriptions might be found here: https://buf.build/docs/lint/rules

	args := []string{
		"--buf-lint_out=.",
		fmt.Sprintf("--buf-lint_opt=%s", configLint),
		"--plugin", must.String(filepath.Abs("bin/protoc-gen-buf-lint")),
	}

	for _, path := range includeDirs {
		args = append(args, "--proto_path", path)
	}

	allProtoFiles, err := findAllProtoFiles(generateDirs)
	if err != nil {
		return err
	}
	packages := map[string][]string{}
	for _, pf := range allProtoFiles {
		pkg, err := goPackage(pf)
		if err != nil {
			return err
		}
		packages[pkg] = append(packages[pkg], pf)
	}

	for _, files := range packages {
		args := append([]string{}, args...)
		args = append(args, files...)
		cmd := exec.Command(tools.Path("bin/protoc", tools.TargetPlatformLocal), args...)
		if err := libexec.Exec(ctx, cmd); err != nil {
			return err
		}
	}

	return nil
}
