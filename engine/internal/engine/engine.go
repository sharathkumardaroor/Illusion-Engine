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
	rng    *rand.Rand
}

func New(cfg models.Config) *Engine {
	return &Engine{
		Config: cfg,
		AI:     ai.New(cfg.UseAI, cfg.BaseURL, cfg.APIKey, cfg.Model),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
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
	e.sendState(models.State{
		Status: "running",
		Before: models.CommitStats{Commits: scan.CommitCount},
	})

	// Success flag for cleanup
	success := false
	defer func() {
		if !success {
			e.sendLog("warn", "Execution failed. Cleaning up partial output...")
			os.RemoveAll(out)
		}
	}()

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
	allFiles, _ := e.snapshot(src)

	// Dependency Alignment: Identify manifest files
	manifestFiles := []string{}
	otherFiles := []string{}
	for _, f := range allFiles {
		name := filepath.Base(f)
		if name == "go.mod" || name == "package.json" || name == "pubspec.yaml" || name == "requirements.txt" || name == "pom.xml" {
			manifestFiles = append(manifestFiles, f)
		} else {
			otherFiles = append(otherFiles, f)
		}
	}

	e.rng.Shuffle(len(otherFiles), func(i, j int) { otherFiles[i], otherFiles[j] = otherFiles[j], otherFiles[i] })

	// Determine number of commits based on Cadence
	baseCount := 10 + e.rng.Intn(10)
	if scan.CommitCount > 0 {
		baseCount = scan.CommitCount
	}

	multiplier := 1.0
	switch e.Config.Cadence {
	case "Low":
		multiplier = 0.5
	case "High":
		multiplier = 2.0
	default:
		multiplier = 1.0
	}

	commitCount := int(float64(baseCount) * multiplier)
	if commitCount < 5 {
		commitCount = 5
	}

	timeStep := end.Sub(start) / time.Duration(commitCount)
	currentTime := start

	// Bootstrap Phase: Copy manifests if DepAlignment is on, OR copy all if ASTPhasing is off
	if e.Config.DepAlignment && len(manifestFiles) > 0 {
		e.sendLog("info", "Phase 1: Project bootstrapping and dependency alignment...")
		for _, f := range manifestFiles {
			rel, _ := filepath.Rel(src, f)
			e.copyFile(f, filepath.Join(out, rel))
			wt.Add(rel)
		}
		wt.Commit("chore: initialize project structure and dependencies", &git.CommitOptions{
			Author: &object.Signature{Name: "Chronos AI", Email: "revamp@chronos.local", When: currentTime},
		})
		currentTime = currentTime.Add(timeStep)
	}

	if !e.Config.ASTPhasing {
		e.sendLog("info", "Phase 1.5: Full project overlay (AST Phasing disabled)...")
		for _, f := range otherFiles {
			rel, _ := filepath.Rel(src, f)
			e.copyFile(f, filepath.Join(out, rel))
			wt.Add(rel)
		}
		wt.Commit("feat: complete project baseline", &git.CommitOptions{
			Author: &object.Signature{Name: "Chronos AI", Email: "revamp@chronos.local", When: currentTime},
		})
		currentTime = currentTime.Add(timeStep)
	}

	// Phase 2: Incremental Development
	for i := 0; i < commitCount; i++ {
		progress := float64(i+1) / float64(commitCount)

		if e.Config.ASTPhasing {
			filesToCopy := int(progress * float64(len(otherFiles)))
			if filesToCopy == 0 {
				filesToCopy = 1
			}
			for j := 0; j < filesToCopy; j++ {
				rel, _ := filepath.Rel(src, otherFiles[j])
				dest := filepath.Join(out, rel)
				e.copyFile(otherFiles[j], dest)

				// Human Error Injection (feat -> revert -> fix cycle simulation)
				if e.Config.HumanErrors && i == commitCount/3 && j == 0 {
					e.simulateHumanErrorCycle(wt, out, rel, currentTime)
					currentTime = currentTime.Add(timeStep / 2)
				}

				wt.Add(rel)
			}
		} else {
			// If AST Phasing is off, we still want to simulate activity on existing files
			numFiles := 1 + e.rng.Intn(3)
			for j := 0; j < numFiles; j++ {
				idx := e.rng.Intn(len(otherFiles))
				rel, _ := filepath.Rel(src, otherFiles[idx])
				// Simulate a small change
				dest := filepath.Join(out, rel)
				content, _ := os.ReadFile(dest)
				os.WriteFile(dest, append(content, []byte("\n// Periodic refactor\n")...), 0644)
				wt.Add(rel)
			}
		}

		ctx := fmt.Sprintf("Focus: %s", e.Config.FocusArea)
		if e.Config.StruggleArea != "" && e.rng.Float32() < 0.2 {
			ctx += fmt.Sprintf(". Addressing technical debt in: %s", e.Config.StruggleArea)
		}

		msg, _ := e.AI.GenerateCommitMessage("incremental update", ctx)

		_, err = wt.Commit(msg, &git.CommitOptions{
			Author: &object.Signature{Name: "Chronos AI", Email: "revamp@chronos.local", When: currentTime},
		})
		if err != nil {
			e.sendLog("warn", fmt.Sprintf("Commit %d skipped: %v", i+1, err))
		}

		currentTime = currentTime.Add(timeStep)

		// Phase 3: Simulated Branching and Merging
		if e.Config.Branches && i == commitCount/2 {
			e.simulateBranchWorkflow(repo, wt, out, currentTime)
			currentTime = currentTime.Add(timeStep)
		}

		e.sendLog("info", fmt.Sprintf("Synthesized commit %d/%d: %s", i+1, commitCount, msg))
	}

	// 4. Final Verification
	e.sendLog("info", "Generating project summary report...")
	reportContent := fmt.Sprintf("# Chronos Revamp Report\n\n- Source: %s\n- Output: %s\n- Total Commits: %d\n- Focus: %s\n- Struggle Area: %s\n\nGenerated by Chronos at %s", src, out, commitCount, e.Config.FocusArea, e.Config.StruggleArea, time.Now().Format(time.RFC1123))
	reportPath := filepath.Join(out, "project_summary.md")
	os.WriteFile(reportPath, []byte(reportContent), 0644)
	wt.Add("project_summary.md")
	wt.Commit("docs: add project summary report", &git.CommitOptions{
		Author: &object.Signature{Name: "Chronos AI", Email: "revamp@chronos.local", When: currentTime},
	})

	e.sendLog("info", "Performing final verification (git status)...")
	status, err := wt.Status()
	verified := err == nil && status.IsClean()

	success = true

	finalCommits, _ := repo.Log(&git.LogOptions{})
	finalCount := 0
	finalCommits.ForEach(func(c *object.Commit) error { finalCount++; return nil })

	e.sendState(models.State{
		Status:     "completed",
		Before:     models.CommitStats{Commits: scan.CommitCount},
		After:      models.CommitStats{Commits: finalCount},
		Verified:   verified,
		OutputPath: out,
		ReportPath: reportPath,
	})

	return nil
}

func (e *Engine) copyFile(src, dest string) error {
	os.MkdirAll(filepath.Dir(dest), 0755)
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, data, 0644)
}

func (e *Engine) simulateHumanErrorCycle(wt *git.Worktree, outDir, relPath string, t time.Time) {
	e.sendLog("info", fmt.Sprintf("Simulating human error cycle for %s", relPath))

	dest := filepath.Join(outDir, relPath)
	content, _ := os.ReadFile(dest)

	// 1. Buggy commit
	buggyContent := append(content, []byte("\n// TODO: Temporary hack to fix race condition\nfunc workaround() { /* Logic pending */ }\n")...)
	os.WriteFile(dest, buggyContent, 0644)
	wt.Add(relPath)
	wt.Commit(fmt.Sprintf("feat: initial implementation of %s", filepath.Base(relPath)), &git.CommitOptions{
		Author: &object.Signature{Name: "Chronos AI", Email: "revamp@chronos.local", When: t},
	})

	// 2. Revert (manual rollback commit)
	os.WriteFile(dest, content, 0644)
	wt.Add(relPath)
	wt.Commit(fmt.Sprintf("revert: \"feat: initial implementation of %s\"", filepath.Base(relPath)), &git.CommitOptions{
		Author: &object.Signature{Name: "Chronos AI", Email: "revamp@chronos.local", When: t.Add(1 * time.Minute)},
	})

	// 3. Fix (original content + proper comment)
	fixedContent := append(content, []byte("\n// Corrected implementation after architectural review\n")...)
	os.WriteFile(dest, fixedContent, 0644)
	wt.Add(relPath)
}

func (e *Engine) simulateBranchWorkflow(repo *git.Repository, wt *git.Worktree, outDir string, t time.Time) {
	e.sendLog("info", "Simulating feature branch and actual merge...")
	head, _ := repo.Head()
	mainHash := head.Hash()
	branchName := "feature/architecture-refactor"

	wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchName),
		Create: true,
	})

	// Feature commit
	os.WriteFile(filepath.Join(outDir, "ARCH.md"), []byte("# Architecture Notes\nVerified refactor plan."), 0644)
	wt.Add("ARCH.md")
	featureCommit, _ := wt.Commit("docs: refine architecture specification", &git.CommitOptions{
		Author: &object.Signature{Name: "Chronos AI", Email: "revamp@chronos.local", When: t},
	})

	// Merge back to main with dual parents
	wt.Checkout(&git.CheckoutOptions{Branch: head.Name()})

	os.WriteFile(filepath.Join(outDir, "ARCH.md"), []byte("# Architecture Notes\nVerified refactor plan."), 0644)
	wt.Add("ARCH.md")
	wt.Commit(fmt.Sprintf("Merge branch '%s'", branchName), &git.CommitOptions{
		Author:  &object.Signature{Name: "Chronos AI", Email: "revamp@chronos.local", When: t.Add(2 * time.Minute)},
		Parents: []plumbing.Hash{mainHash, featureCommit},
	})
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
	if strings.HasPrefix(src, out+sep) {
		return "", "", fmt.Errorf("source directory cannot be inside the output directory")
	}

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

func (e *Engine) snapshot(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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
