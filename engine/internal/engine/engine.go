package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/illusion-engine/chronos/engine/internal/ai"
	"github.com/illusion-engine/chronos/engine/internal/models"
)

type Engine struct {
	Config models.Config
	AI     *ai.Client
}

func New(cfg models.Config) *Engine {
	return &Engine{
		Config: cfg,
		AI:     ai.New(cfg.UseAI, cfg.BaseURL, cfg.APIKey, cfg.Model),
	}
}

func (e *Engine) Scan(path string) (*models.ScanResult, error) {
	result := &models.ScanResult{}

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				result.IsGit = true
				return filepath.SkipDir
			}
			result.FolderCount++
		} else {
			result.FileCount++
			result.SizeMB += float64(info.Size()) / (1024 * 1024)
		}
		return nil
	})

	if result.IsGit {
		repo, err := git.PlainOpen(path)
		if err == nil {
			iter, err := repo.Log(&git.LogOptions{})
			if err == nil {
				count := 0
				var first, latest time.Time
				iter.ForEach(func(c *object.Commit) error {
					if latest.IsZero() {
						latest = c.Author.When
					}
					first = c.Author.When
					count++
					return nil
				})
				result.CommitCount = count
				result.FirstCommit = first.Format("2006-01-02")
				result.LatestCommit = latest.Format("2006-01-02")
			}
			branches, err := repo.Branches()
			if err == nil {
				bCount := 0
				branches.ForEach(func(reference *plumbing.Reference) error {
					bCount++
					return nil
				})
				result.BranchCount = bCount
			}
		}
	}

	return result, err
}

func (e *Engine) Run() error {
	// 1. Setup & Validation
	src, out, err := e.validatePaths()
	if err != nil {
		return err
	}

	scan, _ := e.Scan(src)
	e.sendState(models.State{Status: "running", Before: scan.CommitCount})

	// 2. Initialize Output Repo
	repo, err := git.PlainInit(out, false)
	if err != nil {
		return fmt.Errorf("failed to init shadow repo: %w", err)
	}
	wt, _ := repo.Worktree()

	// 3. Chronos Logic: Generate Incremental History
	start, _ := time.Parse(time.RFC3339, e.Config.StartDate)
	if start.IsZero() {
		start = time.Now().AddDate(0, -3, 0)
	}
	end, _ := time.Parse(time.RFC3339, e.Config.EndDate)
	if end.IsZero() {
		end = time.Now()
	}

	e.sendLog("info", fmt.Sprintf("Simulating history from %s to %s", start.Format("2006-01-02"), end.Format("2006-01-02")))

	// Gather files for phasing
	allFiles, _ := e.snapshot()
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allFiles), func(i, j int) { allFiles[i], allFiles[j] = allFiles[j], allFiles[i] })

	// Determine number of commits
	commitCount := 10 + rand.Intn(20)
	if scan.CommitCount > 0 {
		commitCount = int(float64(scan.CommitCount) * 1.2) // Grow it a bit for "revamp"
	}

	timeStep := end.Sub(start) / time.Duration(commitCount)
	currentTime := start

	for i := 0; i < commitCount; i++ {
		// Progressively add files (AST Phasing simulation)
		progress := float64(i+1) / float64(commitCount)
		filesToAdd := int(progress * float64(len(allFiles)))
		if filesToAdd == 0 {
			filesToAdd = 1
		}

		for j := 0; j < filesToAdd; j++ {
			relPath, _ := filepath.Rel(src, allFiles[j])
			destPath := filepath.Join(out, relPath)
			os.MkdirAll(filepath.Dir(destPath), 0755)

			content, _ := os.ReadFile(allFiles[j])

			// Human Error Injection (Random Typos or Comments)
			if e.Config.HumanErrors && rand.Float32() < 0.1 {
				content = append(content, []byte("\n// TODO: Fix this later - temporary hack\n")...)
			}

			os.WriteFile(destPath, content, 0644)
			wt.Add(relPath)
		}

		// Generate Narrative
		msg, _ := e.AI.GenerateCommitMessage("various changes", e.Config.FocusArea)

		_, err = wt.Commit(msg, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Chronos AI",
				Email: "revamp@chronos.local",
				When:  currentTime,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to commit at step %d: %w", i, err)
		}

		currentTime = currentTime.Add(timeStep)
		e.sendLog("info", fmt.Sprintf("Synthesized commit %d/%d: %s", i+1, commitCount, msg))

		// Simulated Branching
		if e.Config.Branches && i == commitCount/2 {
			e.sendLog("info", "Simulating feature branch workflow...")
			head, _ := repo.Head()
			err = wt.Checkout(&git.CheckoutOptions{
				Branch: plumbing.NewBranchReferenceName("feature/architecture-refactor"),
				Create: true,
			})
			// Add some feature branch files
			os.WriteFile(filepath.Join(out, "ARCH.md"), []byte("# Architecture Notes\nSimulated refactor."), 0644)
			wt.Add("ARCH.md")
			wt.Commit("docs: initial architecture draft", &git.CommitOptions{
				Author: &object.Signature{Name: "Chronos AI", Email: "revamp@chronos.local", When: currentTime},
			})
			// Merge back
			wt.Checkout(&git.CheckoutOptions{Branch: head.Name()})
			// (Simplified merge for simulation)
			os.WriteFile(filepath.Join(out, "ARCH.md"), []byte("# Architecture Notes\nSimulated refactor."), 0644)
			wt.Add("ARCH.md")
			wt.Commit("Merge branch 'feature/architecture-refactor'", &git.CommitOptions{
				Author: &object.Signature{Name: "Chronos AI", Email: "revamp@chronos.local", When: currentTime},
			})
		}
	}

	// 4. Verification & Reporting
	e.sendLog("info", "Generating project summary report...")
	reportContent := fmt.Sprintf("# Chronos Revamp Report\n\n- Source: %s\n- Output: %s\n- Total Commits: %d\n- Focus: %s\n\nGenerated by Chronos at %s", src, out, commitCount, e.Config.FocusArea, time.Now().Format(time.RFC1123))
	reportPath := filepath.Join(out, "project_summary.md")
	os.WriteFile(reportPath, []byte(reportContent), 0644)

	e.sendLog("info", "Performing final verification...")
	_, err = repo.Log(&git.LogOptions{})
	verified := err == nil

	e.sendState(models.State{
		Status:     "completed",
		Verified:   verified,
		Before:     scan.CommitCount,
		After:      commitCount,
		OutputPath: out,
		ReportPath: reportPath,
	})

	return nil
}

func (e *Engine) validatePaths() (string, string, error) {
	src, err := filepath.EvalSymlinks(e.Config.SourceDir)
	if err != nil {
		src, _ = filepath.Abs(e.Config.SourceDir)
	} else {
		src, _ = filepath.Abs(src)
	}

	out, err := filepath.Abs(e.Config.OutputDir)
	if err != nil {
		return "", "", fmt.Errorf("invalid output path: %w", err)
	}

	if src == out {
		return "", "", fmt.Errorf("source and output directories cannot be the same")
	}

	sep := string(os.PathSeparator)
	if strings.HasPrefix(out, src+sep) {
		return "", "", fmt.Errorf("output directory cannot be inside the source directory")
	}

	// Prevent overwriting non-empty directory
	if _, err := os.Stat(out); err == nil {
		empty, _ := isDirEmpty(out)
		if !empty {
			return "", "", fmt.Errorf("output directory already exists and is not empty")
		}
	}

	return src, out, nil
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
