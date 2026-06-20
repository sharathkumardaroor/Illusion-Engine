package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/illusion-engine/chronos/engine/internal/git"
	"github.com/illusion-engine/chronos/engine/internal/models"
)

type Engine struct {
	Config models.Config
}

func New(cfg models.Config) *Engine {
	return &Engine{Config: cfg}
}

func (e *Engine) Run() error {
	e.sendLog("info", fmt.Sprintf("Analyzing target directory: %s", e.Config.TargetDir))

	files, err := e.snapshot()
	if err != nil {
		return err
	}
	e.sendLog("info", fmt.Sprintf("Found %d files to deconstruct", len(files)))

	tempDir, err := os.MkdirTemp("", "chronos-shadow-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	repo, err := git.InitShadowRepo(tempDir)
	if err != nil {
		return err
	}
	e.sendLog("info", "Shadow repository initialized")

	time.Sleep(1 * time.Second)

	e.sendLog("info", "Performing final overlay...")
	if err := e.overlay(tempDir); err != nil {
		return err
	}

	w, _ := repo.Worktree()
	_, err = w.Add(".")
	if err != nil {
		return err
	}
	_, err = w.Commit("final: project synchronization", &git.CommitOptions{})

	e.sendLog("info", "Replacing project .git history...")
	shadowGitDir := filepath.Join(tempDir, ".git")
	if err := git.ReplaceGitDir(e.Config.TargetDir, shadowGitDir); err != nil {
		return err
	}

	e.sendState(models.State{
		Status:   "completed",
		Verified: true,
		After:    1,
	})

	return nil
}

func (e *Engine) snapshot() ([]string, error) {
	var files []string
	err := filepath.Walk(e.Config.TargetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

func (e *Engine) overlay(shadowDir string) error {
	return filepath.Walk(e.Config.TargetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, _ := filepath.Rel(e.Config.TargetDir, path)
		destPath := filepath.Join(shadowDir, relPath)

		os.MkdirAll(filepath.Dir(destPath), 0755)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(destPath, data, info.Mode())
	})
}

func (e *Engine) sendLog(level, msg string) {
	event := models.LogEvent{
		Type:    "log",
		Level:   level,
		Message: msg,
	}
	data, _ := json.Marshal(event)
	fmt.Println(string(data))
}

func (e *Engine) sendState(state models.State) {
	event := models.LogEvent{
		Type:    "state",
		Payload: state,
	}
	data, _ := json.Marshal(event)
	fmt.Println(string(data))
}
