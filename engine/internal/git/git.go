package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

func InitShadowRepo(tempDir string) (*git.Repository, error) {
	repo, err := git.PlainInit(tempDir, false)
	if err != nil {
		return nil, fmt.Errorf("failed to init shadow repo: %w", err)
	}
	return repo, nil
}

type CommitOptions = git.CommitOptions

func ReplaceGitDir(targetDir, shadowGitDir string) error {
	targetGitDir := filepath.Join(targetDir, ".git")

	if _, err := os.Stat(targetGitDir); err == nil {
		backupDir := targetGitDir + ".bak"
		os.RemoveAll(backupDir)
		if err := os.Rename(targetGitDir, backupDir); err != nil {
			return fmt.Errorf("failed to backup .git: %w", err)
		}
	}

	if err := copyDir(shadowGitDir, targetGitDir); err != nil {
		return fmt.Errorf("failed to copy shadow .git: %w", err)
	}

	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetPath, data, info.Mode())
	})
}
