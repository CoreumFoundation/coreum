package build

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var tools = map[string]tool{
	"go": {
		Version: "1.18.2",
		Sources: sources{
			linuxAMD64: {
				URL:  "https://go.dev/dl/go1.18.2.linux-amd64.tar.gz",
				Hash: "sha256:e54bec97a1a5d230fc2f9ad0880fcbabb5888f30ed9666eca4a91c5a32e86cbc",
			},
			darwinAMD64: {
				URL:  "https://go.dev/dl/go1.18.2.darwin-amd64.tar.gz",
				Hash: "sha256:1f5f539ce0baa8b65f196ee219abf73a7d9cf558ba9128cc0fe4833da18b04f2",
			},
			darwinARM64: {
				URL:  "https://go.dev/dl/go1.18.2.darwin-arm64.tar.gz",
				Hash: "sha256:6c7df9a2405f09aa9bab55c93c9c4ce41d3e58127d626bc1825ba5d0a0045d5c",
			},
		},
		Binaries: []string{
			"go/bin/go",
			"go/bin/gofmt",
		},
	},
	"golangci": {
		Version: "1.46.2",
		Sources: sources{
			linuxAMD64: {
				URL:  "https://github.com/golangci/golangci-lint/releases/download/v1.46.2/golangci-lint-1.46.2-linux-amd64.tar.gz",
				Hash: "sha256:242cd4f2d6ac0556e315192e8555784d13da5d1874e51304711570769c4f2b9b",
				Binaries: []string{
					"golangci-lint-1.46.2-linux-amd64/golangci-lint",
				},
			},
			darwinAMD64: {
				URL:  "https://github.com/golangci/golangci-lint/releases/download/v1.46.2/golangci-lint-1.46.2-darwin-amd64.tar.gz",
				Hash: "sha256:658078aaaf7608693f37c4cf1380b2af418ab8b2d23fdb33e7e2d4339328590e",
				Binaries: []string{
					"golangci-lint-1.46.2-darwin-amd64/golangci-lint",
				},
			},
			darwinARM64: {
				URL:  "https://github.com/golangci/golangci-lint/releases/download/v1.46.2/golangci-lint-1.46.2-darwin-arm64.tar.gz",
				Hash: "sha256:81f9b4afd62ec5e612ef8bc3b1d612a88b56ff289874831845cdad394427385f",
				Binaries: []string{
					"golangci-lint-1.46.2-darwin-arm64/golangci-lint",
				},
			},
		},
	},
	"ignite": {
		Version: "v0.21.2",
		Sources: sources{
			linuxAMD64: {
				URL:  "https://github.com/ignite-hq/cli/releases/download/v0.21.2/ignite_0.21.2_linux_amd64.tar.gz",
				Hash: "sha256:c79e7119a0e14881336b92a5191cba861130c80a5a21bb0f0aa8f79c4c237204",
			},
			darwinAMD64: {
				URL:  "https://github.com/ignite-hq/cli/releases/download/v0.21.2/ignite_0.21.2_darwin_amd64.tar.gz",
				Hash: "sha256:b9570804d1cc7023b780f50aece91c3c57cbf7f877850946642101e313cb0ec6",
			},
			darwinARM64: {
				URL:  "https://github.com/ignite-hq/cli/releases/download/v0.21.2/ignite_0.21.2_darwin_arm64.tar.gz",
				Hash: "sha256:9455af04670b7a57d3d7320ae28fa40ee8dc331581f1029854e5697cd89e93d8",
			},
		},
		Binaries: []string{
			"ignite",
		},
	},
}

const binDir = "bin"

type platform struct {
	OS   string
	Arch string
}

func (p platform) String() string {
	return p.OS + "/" + p.Arch
}

var (
	linuxAMD64  = platform{OS: "linux", Arch: "amd64"}
	darwinAMD64 = platform{OS: "darwin", Arch: "amd64"}
	darwinARM64 = platform{OS: "darwin", Arch: "arm64"}
)

type tool struct {
	Version  string
	Sources  sources
	Binaries []string
}

type source struct {
	URL      string
	Hash     string
	Binaries []string
}

type sources map[platform]source

func installTools(deps build.DepsFunc) {
	toolFns := make([]interface{}, 0, len(tools))
	for tool := range tools {
		tool := tool
		toolFns = append(toolFns, func(ctx context.Context) error {
			return ensure(ctx, tool)
		})
	}
	deps(toolFns...)
}

func ensure(ctx context.Context, tool string) error {
	info, exists := tools[tool]
	if !exists {
		return errors.Errorf("tool %s is not defined", tool)
	}

	platform := platform{OS: runtime.GOOS, Arch: runtime.GOARCH}
	source, exists := info.Sources[platform]
	if !exists {
		panic(errors.Errorf("tool %s is not configured for platform %s", tool, platform))
	}

	toolDir := toolDir(tool)
	for _, bin := range combine(info.Binaries, source.Binaries) {
		srcPath, err := filepath.Abs(toolDir + "/" + bin)
		if err != nil {
			return install(ctx, tool, info)
		}

		binName := filepath.Base(bin)
		dstPath, err := filepath.Abs(binDir + "/" + binName)
		if err != nil {
			return install(ctx, tool, info)
		}

		realPath, err := filepath.EvalSymlinks(dstPath)
		if err != nil || realPath != srcPath {
			return install(ctx, tool, info)
		}

		binPath, err := exec.LookPath(binName)
		if err != nil || binPath != dstPath {
			return errors.Errorf("binary %s can't be resolved from PATH, add %s to your PATH",
				binName, must.String(filepath.Abs(binDir)))
		}
	}
	return nil
}

func install(ctx context.Context, name string, info tool) (retErr error) {
	platform := platform{OS: runtime.GOOS, Arch: runtime.GOARCH}
	source, exists := info.Sources[platform]
	if !exists {
		panic(errors.Errorf("tool %s is not configured for platform %s", name, platform))
	}
	ctx = logger.With(ctx, zap.String("name", name), zap.String("version", info.Version),
		zap.String("url", source.URL))
	log := logger.Get(ctx)
	log.Info("Installing tool")

	resp, err := http.DefaultClient.Do(must.HTTPRequest(http.NewRequestWithContext(ctx, http.MethodGet, source.URL, nil)))
	if err != nil {
		return errors.WithStack(err)
	}
	defer resp.Body.Close()

	hasher, expectedChecksum := hasher(source.Hash)
	reader := io.TeeReader(resp.Body, hasher)
	toolDir := toolDir(name)
	if err := os.RemoveAll(toolDir); err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	if err := os.MkdirAll(toolDir, 0o700); err != nil {
		panic(err)
	}
	defer func() {
		if retErr != nil {
			must.OK(os.RemoveAll(toolDir))
		}
	}()

	if err := extract(source.URL, reader, toolDir); err != nil {
		return err
	}

	actualChecksum := fmt.Sprintf("%02x", hasher.Sum(nil))
	if actualChecksum != expectedChecksum {
		return errors.Errorf("checksum does not match for tool %s, expected: %s, actual: %s, url: %s", name,
			expectedChecksum, actualChecksum, source.URL)
	}

	for _, bin := range combine(info.Binaries, source.Binaries) {
		srcPath := toolDir + "/" + bin
		dstPath := binDir + "/" + filepath.Base(bin)
		if err := os.Remove(dstPath); err != nil && !os.IsNotExist(err) {
			panic(err)
		}
		must.OK(os.Symlink(srcPath, dstPath))
		must.Any(filepath.EvalSymlinks(dstPath))
	}

	log.Info("Tool installed")
	return nil
}

func hasher(hashStr string) (hash.Hash, string) {
	parts := strings.SplitN(hashStr, ":", 2)
	if len(parts) != 2 {
		panic(errors.Errorf("incorrect checksum format: %s", hashStr))
	}
	hashAlgorithm := parts[0]
	checksum := parts[1]

	var hasher hash.Hash
	switch hashAlgorithm {
	case "sha256":
		hasher = sha256.New()
	default:
		panic(errors.Errorf("unsupported hashing algorithm: %s", hashAlgorithm))
	}

	return hasher, strings.ToLower(checksum)
}

func extract(url string, reader io.Reader, path string) error {
	switch {
	case strings.HasSuffix(url, ".tar.gz"):
		var err error
		reader, err = gzip.NewReader(reader)
		if err != nil {
			return errors.WithStack(err)
		}
		return untar(reader, path)
	default:
		panic(errors.Errorf("unsupported compression algorithm for url: %s", url))
	}
}

func untar(reader io.Reader, path string) error {
	tr := tar.NewReader(reader)
	for {
		header, err := tr.Next()
		switch {
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return errors.WithStack(err)
		case header == nil:
			continue
		}
		header.Name = path + "/" + header.Name

		// We take mode from header.FileInfo().Mode(), not from header.Mode because they may be in different formats (meaning of bits may be different).
		// header.FileInfo().Mode() returns compatible value.
		mode := header.FileInfo().Mode()

		switch {
		case header.Typeflag == tar.TypeDir:
			if err := os.MkdirAll(header.Name, mode); err != nil && !os.IsExist(err) {
				return errors.WithStack(err)
			}
		case header.Typeflag == tar.TypeReg:
			if err := ensureDir(header.Name); err != nil {
				return err
			}
			f, err := os.OpenFile(header.Name, os.O_CREATE|os.O_WRONLY, mode)
			if err != nil {
				return errors.WithStack(err)
			}
			_, err = io.Copy(f, tr)
			_ = f.Close()
			if err != nil {
				return errors.WithStack(err)
			}
		case header.Typeflag == tar.TypeSymlink:
			if err := ensureDir(header.Name); err != nil {
				return err
			}
			if err := os.Symlink(header.Linkname, header.Name); err != nil {
				return errors.WithStack(err)
			}
		case header.Typeflag == tar.TypeLink:
			if err := ensureDir(header.Name); err != nil {
				return err
			}
			// linked file may not exist yet, so let's create it - i will be overwritten later
			f, err := os.OpenFile(header.Linkname, os.O_CREATE|os.O_EXCL, mode)
			if err != nil {
				if !os.IsExist(err) {
					return errors.WithStack(err)
				}
			} else {
				_ = f.Close()
			}
			if err := os.Link(header.Linkname, header.Name); err != nil {
				return errors.WithStack(err)
			}
		default:
			return errors.Errorf("unsupported file type: %d", header.Typeflag)
		}
	}
}

func toolDir(name string) string {
	info, exists := tools[name]
	if !exists {
		panic(errors.Errorf("tool %s is not defined", name))
	}

	return must.String(os.UserCacheDir()) + "/coreum/" + name + "-" + info.Version
}

func ensureDir(file string) error {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); !os.IsExist(err) {
		return errors.WithStack(err)
	}
	return nil
}

func combine(a1 []string, a2 []string) []string {
	return append(append([]string{}, a1...), a2...)
}
