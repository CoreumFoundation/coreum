package build

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// gitStatusClean checks that there are no uncommitted files
func gitStatusClean(ctx context.Context) error {
	for _, repoPath := range repositories {
		buf := &bytes.Buffer{}
		cmd := exec.Command("git", "status", "-s")
		cmd.Dir = repoPath
		cmd.Stdout = buf
		if err := libexec.Exec(ctx, cmd); err != nil {
			return errors.Wrap(err, "git command failed")
		}
		if buf.Len() > 0 {
			fmt.Println("git status:")
			fmt.Println(buf)
			return errors.Errorf("git status of repository '%s' is not empty", filepath.Base(repoPath))
		}
	}
	return nil
}

func ensureRepo(ctx context.Context, repoURL string) error {
	repoName := strings.TrimSuffix(filepath.Base(repoURL), ".git")
	info, err := os.Stat("../" + repoName)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Get(ctx).Info("Cloning repository", zap.String("name", repoName), zap.String("url", repoURL))
			cmd := exec.Command("git", "clone", repoURL)
			cmd.Dir = "../"
			if err := libexec.Exec(ctx, cmd); err != nil {
				return errors.Wrapf(err, "cloning repository `%s` failed", repoURL)
			}
			return nil
		}
		return errors.WithStack(err)
	}
	if !info.IsDir() {
		return errors.Errorf("path '%s' is not a directory, while repository is expected", repoURL)
	}
	return nil
}
