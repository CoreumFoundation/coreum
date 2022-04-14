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
	"strings"

	"github.com/CoreumFoundation/coreum-build-tools/pkg/build"
	"github.com/CoreumFoundation/coreum-build-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-build-tools/pkg/must"
	"go.uber.org/zap"
)

const binDir = "bin"

type tool struct {
	Version  string
	URL      string
	Hash     string
	Binaries []string
}

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

	toolDir := toolDir(tool)
	for _, bin := range info.Binaries {
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
	ctx = logger.With(ctx, zap.String("name", name), zap.String("version", info.Version),
		zap.String("url", info.URL))
	log := logger.Get(ctx)
	log.Info("Installing tool")

	resp, err := http.DefaultClient.Do(must.HTTPRequest(http.NewRequestWithContext(ctx, http.MethodGet, info.URL, nil)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	hasher, expectedChecksum := hasher(info.Hash)
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

	if err := extract(info.URL, reader, toolDir); err != nil {
		return err
	}

	actualChecksum := fmt.Sprintf("%02x", hasher.Sum(nil))
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum does not match for tool %s, expected: %s, actual: %s, url: %s", name,
			expectedChecksum, actualChecksum, info.URL)
	}

	for _, bin := range info.Binaries {
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
