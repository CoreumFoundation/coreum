package coreum

import (
	"bufio"
	"context"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/crust/build/golang"
	"github.com/CoreumFoundation/crust/build/tools"
)

func generateProtoGo(ctx context.Context, deps build.DepsFunc) error {
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

	err = executeGoProtocCommand(ctx, deps, includeDirs, generateDirs)
	if err != nil {
		return err
	}

	return nil
}

// executeGoProtocCommand generates go code from proto files.
func executeGoProtocCommand(ctx context.Context, deps build.DepsFunc, includeDirs, generateDirs []string) error {
	deps(tools.EnsureProtoc, tools.EnsureProtocGenGRPCGateway, tools.EnsureProtocGenGoCosmos)

	outDir, err := os.MkdirTemp("", "")
	if err != nil {
		return errors.WithStack(err)
	}
	defer os.RemoveAll(outDir) //nolint:errcheck // we don't care

	if err := os.MkdirAll(outDir, 0o700); err != nil {
		return errors.WithStack(err)
	}

	args := []string{
		"--gocosmos_out=plugins=interfacetype+grpc,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:.",
		"--grpc-gateway_out=logtostderr=true:.",
		"--plugin", must.String(filepath.Abs("bin/protoc-gen-gocosmos")),
		"--plugin", must.String(filepath.Abs("bin/protoc-gen-grpc-gateway")),
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
		cmd.Dir = outDir
		if err := libexec.Exec(ctx, cmd); err != nil {
			return err
		}
	}

	modulePath, err := golang.ModuleName(repoPath)
	if err != nil {
		return err
	}
	rootDir := filepath.Join(outDir, modulePath)
	err = filepath.Walk(rootDir,
		func(srcPath string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			relPath, err := filepath.Rel(rootDir, srcPath)
			if err != nil {
				return errors.WithStack(err)
			}
			dstPath := filepath.Join(repoPath, relPath)
			return copyFile(srcPath, dstPath, 0o600)
		})
	if err != nil {
		return err
	}

	return nil
}

func goPackage(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer f.Close()

	rx := regexp.MustCompile("[^ \t]*option[ \t]+go_package[ \t]*=[ \t]?\"(.+?)(;.+?)?\";[ \t]?$")

	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		matches := rx.FindStringSubmatch(s.Text())
		if len(matches) < 2 {
			continue
		}
		return matches[1], nil
	}
	return "", nil
}

func copyFile(src, dst string, perm os.FileMode) error {
	fr, err := os.Open(src)
	if err != nil {
		return errors.WithStack(err)
	}
	defer fr.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return errors.WithStack(err)
	}

	fw, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return errors.WithStack(err)
	}
	defer fw.Close()

	if _, err = io.Copy(fw, fr); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
