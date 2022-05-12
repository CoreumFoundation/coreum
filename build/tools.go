package build

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"errors"
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
	"go.uber.org/zap"
)

var tools = map[string]tool{
	"go": {
		Version: "1.18.1",
		Sources: sources{
			linuxAMD64: {
				URL:  "https://go.dev/dl/go1.18.1.linux-amd64.tar.gz",
				Hash: "sha256:b3b815f47ababac13810fc6021eb73d65478e0b2db4b09d348eefad9581a2334",
			},
			darwinAMD64: {
				URL:  "https://go.dev/dl/go1.18.1.darwin-amd64.tar.gz",
				Hash: "sha256:63e5035312a9906c98032d9c73d036b6ce54f8632b194228bd08fe3b9fe4ab01",
			},
			darwinARM64: {
				URL:  "https://go.dev/dl/go1.18.1.darwin-arm64.tar.gz",
				Hash: "sha256:6d5641a06edba8cd6d425fb0adad06bad80e2afe0fa91b4aa0e5aed1bc78f58e",
			},
		},
		Binaries: []string{
			"go/bin/go",
			"go/bin/gofmt",
		},
	},
	"golangci": {
		Version: "1.45.2",
		Sources: sources{
			linuxAMD64: {
				URL:  "https://github.com/golangci/golangci-lint/releases/download/v1.45.2/golangci-lint-1.45.2-linux-amd64.tar.gz",
				Hash: "sha256:595ad6c6dade4c064351bc309f411703e457f8ffbb7a1806b3d8ee713333427f",
				Binaries: []string{
					"golangci-lint-1.45.2-linux-amd64/golangci-lint",
				},
			},
			darwinAMD64: {
				URL:  "https://github.com/golangci/golangci-lint/releases/download/v1.45.2/golangci-lint-1.45.2-darwin-amd64.tar.gz",
				Hash: "sha256:995e509e895ca6a64ffc7395ac884d5961bdec98423cb896b17f345a9b4a19cf",
				Binaries: []string{
					"golangci-lint-1.45.2-darwin-amd64/golangci-lint",
				},
			},
			darwinARM64: {
				URL:  "https://github.com/golangci/golangci-lint/releases/download/v1.45.2/golangci-lint-1.45.2-darwin-arm64.tar.gz",
				Hash: "sha256:c2b9669decc1b638cf2ee9060571af4e255f6dfcbb225c293e3a7ee4bb2c7217",
				Binaries: []string{
					"golangci-lint-1.45.2-darwin-arm64/golangci-lint",
				},
			},
		},
	},
	"ignite": {
		Version: "v0.20.4",
		Sources: sources{
			linuxAMD64: {
				URL:  "https://github.com/ignite-hq/cli/releases/download/v0.20.4/ignite_0.20.4_linux_amd64.tar.gz",
				Hash: "sha256:6291e0e3571cfc81caa691932024519cabade44c061d4214f5f4090badb06ab2",
			},
			darwinAMD64: {
				URL:  "https://github.com/ignite-hq/cli/releases/download/v0.20.4/ignite_0.20.4_darwin_amd64.tar.gz",
				Hash: "sha256:2e9366168de8b8dbf743ec0de21c93430eca79c76d947c6de4d7c728c757f05e",
			},
			darwinARM64: {
				URL:  "https://github.com/ignite-hq/cli/releases/download/v0.20.4/ignite_0.20.4_darwin_arm64.tar.gz",
				Hash: "sha256:9543862fc1399dc1a4d40ca511af6bf8743dff5c79e2fa774632bdbe2196b779",
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
		return fmt.Errorf("tool %s is not defined", tool)
	}

	platform := platform{OS: runtime.GOOS, Arch: runtime.GOARCH}
	source, exists := info.Sources[platform]
	if !exists {
		panic(fmt.Errorf("tool %s is not configured for platform %s", tool, platform))
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
			return fmt.Errorf("binary %s can't be resolved from PATH, add %s to your PATH",
				binName, must.String(filepath.Abs(binDir)))
		}
	}
	return nil
}

func install(ctx context.Context, name string, info tool) (retErr error) {
	platform := platform{OS: runtime.GOOS, Arch: runtime.GOARCH}
	source, exists := info.Sources[platform]
	if !exists {
		panic(fmt.Errorf("tool %s is not configured for platform %s", name, platform))
	}
	ctx = logger.With(ctx, zap.String("name", name), zap.String("version", info.Version),
		zap.String("url", source.URL))
	log := logger.Get(ctx)
	log.Info("Installing tool")

	resp, err := http.DefaultClient.Do(must.HTTPRequest(http.NewRequestWithContext(ctx, http.MethodGet, source.URL, nil)))
	if err != nil {
		return err
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
		return fmt.Errorf("checksum does not match for tool %s, expected: %s, actual: %s, url: %s", name,
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
		panic(fmt.Errorf("incorrect checksum format: %s", hashStr))
	}
	hashAlgorithm := parts[0]
	checksum := parts[1]

	var hasher hash.Hash
	switch hashAlgorithm {
	case "sha256":
		hasher = sha256.New()
	default:
		panic(fmt.Errorf("unsupported hashing algorithm: %s", hashAlgorithm))
	}

	return hasher, strings.ToLower(checksum)
}

func extract(url string, reader io.Reader, path string) error {
	switch {
	case strings.HasSuffix(url, ".tar.gz"):
		var err error
		reader, err = gzip.NewReader(reader)
		if err != nil {
			return err
		}
		return untar(reader, path)
	default:
		panic(fmt.Errorf("unsupported compression algorithm for url: %s", url))
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
			return err
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
				return err
			}
		case header.Typeflag == tar.TypeReg:
			if err := ensureDir(header.Name); err != nil {
				return err
			}
			f, err := os.OpenFile(header.Name, os.O_CREATE|os.O_WRONLY, mode)
			if err != nil {
				return err
			}
			_, err = io.Copy(f, tr)
			_ = f.Close()
			if err != nil {
				return err
			}
		case header.Typeflag == tar.TypeSymlink:
			if err := ensureDir(header.Name); err != nil {
				return err
			}
			if err := os.Symlink(header.Linkname, header.Name); err != nil {
				return err
			}
		case header.Typeflag == tar.TypeLink:
			if err := ensureDir(header.Name); err != nil {
				return err
			}
			// linked file may not exist yet, so let's create it - i will be overwritten later
			f, err := os.OpenFile(header.Linkname, os.O_CREATE|os.O_EXCL, mode)
			if err != nil {
				if !os.IsExist(err) {
					return err
				}
			} else {
				_ = f.Close()
			}
			if err := os.Link(header.Linkname, header.Name); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported file type: %d", header.Typeflag)
		}
	}
}

func toolDir(name string) string {
	info, exists := tools[name]
	if !exists {
		panic(fmt.Errorf("tool %s is not defined", name))
	}

	return must.String(os.UserCacheDir()) + "/coreum/" + name + "-" + info.Version
}

func ensureDir(file string) error {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); !os.IsExist(err) {
		return err
	}
	return nil
}

func combine(a1 []string, a2 []string) []string {
	return append(append([]string{}, a1...), a2...)
}
