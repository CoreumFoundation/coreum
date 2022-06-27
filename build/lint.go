package build

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/pkg/errors"
)

// lint runs linters and check that git status is clean
func lint(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureAllRepos, goLint, lintNewLines, goModTidy, gitStatusClean)
	return nil
}

func lintNewLines() error {
	for _, repoPath := range repositories {
		err := filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if strings.HasPrefix(d.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return errors.WithStack(err)
			}
			if info.Mode()&0o111 != 0 {
				// skip executable files
				return nil
			}
			f, err := os.Open(path)
			if err != nil {
				return errors.WithStack(err)
			}
			defer f.Close()

			if _, err := f.Seek(-2, io.SeekEnd); err != nil {
				return errors.WithStack(err)
			}

			buf := make([]byte, 2)
			if _, err := f.Read(buf); err != nil {
				return errors.WithStack(err)
			}
			if buf[1] != '\n' {
				return errors.Errorf("no empty line at the end of file '%s'", path)
			}
			if buf[0] == '\n' {
				return errors.Errorf("many empty lines at the end of file '%s'", path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
