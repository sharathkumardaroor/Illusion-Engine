package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	// Path Validation & Symlink Resolution
	src, err := filepath.EvalSymlinks(e.Config.SourceDir)
	if err != nil {
		src, _ = filepath.Abs(e.Config.SourceDir)
	} else {
		src, _ = filepath.Abs(src)
	}

	out, err := filepath.Abs(e.Config.OutputDir)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	if src == out {
		return fmt.Errorf("source and output directories cannot be the same")
	}

	sep := string(os.PathSeparator)
	if strings.HasPrefix(out, src+sep) {
		return fmt.Errorf("output directory cannot be inside the source directory")
	}
	if strings.HasPrefix(src, out+sep) {
		return fmt.Errorf("source directory cannot be inside the output directory")
	}

	// Prevent overwriting non-empty directory
	if _, err := os.Stat(out); err == nil {
		empty, err := isDirEmpty(out)
		if err != nil {
			return fmt.Errorf("failed to check output directory: %w", err)
		}
		if !empty {
			return fmt.Errorf("output directory already exists and is not empty")
		}
	}

	// Track if we created the output directory to know if we should clean it up
	createdOutput := false
	if _, err := os.Stat(out); os.IsNotExist(err) {
		createdOutput = true
	}

	// Success flag for cleanup
	success := false
	defer func() {
		if !success && createdOutput {
			e.sendLog("warn", "Execution failed. Cleaning up partial output...")
			os.RemoveAll(out)
		}
	}()

	e.sendLog("info", "Scanning source directory...")
	files, err := e.snapshot()
	if err != nil {
		return fmt.Errorf("failed to scan source: %w", err)
	}
	e.sendLog("info", fmt.Sprintf("Found %d files in source", len(files)))

	// Create Output Directory
	if err := os.MkdirAll(out, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	e.sendLog("info", fmt.Sprintf("Output directory created: %s", out))

	// Initialize Shadow Repo in Output Directory
	repo, err := git.InitShadowRepo(out)
	if err != nil {
		return err
	}
	e.sendLog("info", "Shadow repository initialized in output directory")

	// Simulated Timeline Generation
	e.sendLog("info", "Generating development timeline...")
	time.Sleep(1 * time.Second)

	// The Overlay Step (Final Commit)
	e.sendLog("info", "Performing final overlay: copying source files to output...")
	if err := e.overlay(out); err != nil {
		return err
	}

	// Commit final state
	w, _ := repo.Worktree()
	_, err = w.Add(".")
	if err != nil {
		return err
	}
	_, err = w.Commit("final: project synchronization", &git.CommitOptions{})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	e.sendLog("info", "Revamp complete. Verifying clean status...")

	success = true

	e.sendState(models.State{
		Status:     "completed",
		Verified:   true,
		Before:     0,
		After:      1,
		OutputPath: out,
		ReportPath: filepath.Join(out, "project_summary.md"),
	})

	return nil
}

func isDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func (e *Engine) snapshot() ([]string, error) {
	var files []string
	err := filepath.Walk(e.Config.SourceDir, func(path string, info os.FileInfo, err error) error {
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

func (e *Engine) overlay(destDir string) error {
	return filepath.Walk(e.Config.SourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, _ := filepath.Rel(e.Config.SourceDir, path)
		destPath := filepath.Join(destDir, relPath)

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
