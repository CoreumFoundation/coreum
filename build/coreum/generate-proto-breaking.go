package coreum

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/crust/build/git"
	"github.com/CoreumFoundation/crust/build/golang"
	"github.com/CoreumFoundation/crust/build/tools"
	"github.com/CoreumFoundation/crust/build/types"
)

//go:embed proto-breaking.tmpl.json
var configBreakingTmpl string //nolint:unused // temp

func breakingProto(ctx context.Context, deps types.DepsFunc) error { //nolint:unused,deadcode // temp
	deps(golang.Tidy, tools.EnsureProtoc, tools.EnsureProtocGenBufBreaking)

	masterDir, err := os.MkdirTemp("", "coreum-master-*")
	if err != nil {
		return errors.WithStack(err)
	}
	defer os.RemoveAll(masterDir) //nolint:errcheck // error doesn't matter

	if err := git.CloneLocalBranch(ctx, masterDir, repoPath, "crust/proto-breaking", "master"); err != nil {
		return err
	}
	if err := golang.DownloadDependencies(ctx, deps, masterDir); err != nil {
		return errors.WithStack(err)
	}

	_, masterIncludeDirs, err := protoCDirectories(ctx, masterDir, deps)
	if err != nil {
		return errors.WithStack(err)
	}

	masterIncludeArgs := make([]string, 0, 2*len(masterIncludeDirs))
	for _, path := range masterIncludeDirs {
		masterIncludeArgs = append(masterIncludeArgs, "--proto_path", path)
	}

	imageFile := filepath.Join(os.TempDir(), "coreum.binpb")
	if err := os.MkdirAll(filepath.Dir(imageFile), 0o700); err != nil {
		return errors.WithStack(err)
	}
	defer os.Remove(imageFile) //nolint:errcheck // error doesn't matter

	masterProtoFiles, err := findAllProtoFiles([]string{filepath.Join(masterDir, "proto")})
	if err != nil {
		return errors.WithStack(err)
	}

	cmdImage := exec.Command(tools.Path("bin/protoc", tools.TargetPlatformLocal),
		append(
			append([]string{"--include_imports", "--include_source_info", "-o", imageFile}, masterIncludeArgs...),
			masterProtoFiles...)...)

	if err := libexec.Exec(ctx, cmdImage); err != nil {
		return errors.WithStack(err)
	}

	_, includeDirs, err := protoCDirectories(ctx, repoPath, deps)
	if err != nil {
		return errors.WithStack(err)
	}

	includeArgs := make([]string, 0, 2*len(includeDirs))
	for _, path := range includeDirs {
		includeArgs = append(includeArgs, "--proto_path", path)
	}

	absRepoPath, err := filepath.Abs(repoPath)
	if err != nil {
		return errors.WithStack(err)
	}

	masterProtoFiles, err = findAllProtoFiles([]string{filepath.Join(absRepoPath, "proto")})
	if err != nil {
		return errors.WithStack(err)
	}

	configBuf := &bytes.Buffer{}
	must.OK(template.Must(template.New("config").Parse(configBreakingTmpl)).Execute(configBuf, struct {
		AgainstInput string
	}{
		AgainstInput: imageFile,
	}))

	args := []string{
		"--buf-breaking_out=.",
		fmt.Sprintf("--buf-breaking_opt=%s", configBuf),
		fmt.Sprintf("--plugin=%s", tools.Path("bin/protoc-gen-buf-breaking", tools.TargetPlatformLocal)),
	}

	args = append(args, includeArgs...)
	args = append(args, masterProtoFiles...)
	cmdBreaking := exec.Command(tools.Path("bin/protoc", tools.TargetPlatformLocal), args...)
	return libexec.Exec(ctx, cmdBreaking)
}
