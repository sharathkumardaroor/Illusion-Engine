package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/illusion-engine/chronos/engine/internal/models"
)

func TestPathValidation(t *testing.T) {
	tmp, _ := os.MkdirTemp("", "chronos-test-*")
	defer os.RemoveAll(tmp)

	src := filepath.Join(tmp, "src")
	os.Mkdir(src, 0755)

	t.Run("Same Path Failure", func(t *testing.T) {
		eng := New(models.Config{SourceDir: src, OutputDir: src})
		_, _, err := eng.validatePaths()
		if err == nil || err.Error() != "source and output directories cannot be the same" {
			t.Errorf("Expected same path error, got %v", err)
		}
	})

	t.Run("Inside Source Failure", func(t *testing.T) {
		out := filepath.Join(src, "out")
		eng := New(models.Config{SourceDir: src, OutputDir: out})
		_, _, err := eng.validatePaths()
		if err == nil || err.Error() != "output directory cannot be inside the source directory" {
			t.Errorf("Expected inside source error, got %v", err)
		}
	})

	t.Run("Valid Paths", func(t *testing.T) {
		out := filepath.Join(tmp, "out")
		eng := New(models.Config{SourceDir: src, OutputDir: out})
		vSrc, vOut, err := eng.validatePaths()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if vSrc == "" || vOut == "" {
			t.Error("Paths should not be empty")
		}
	})
}

func TestScan(t *testing.T) {
	tmp, _ := os.MkdirTemp("", "chronos-scan-*")
	defer os.RemoveAll(tmp)

	os.WriteFile(filepath.Join(tmp, "file1.txt"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmp, "dir1"), 0755)

	eng := New(models.Config{})
	res, err := eng.Scan(tmp)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if res.FileCount != 1 {
		t.Errorf("Expected 1 file, got %d", res.FileCount)
	}
	// filepath.Walk includes the root path itself as a directory
	if res.FolderCount != 2 {
		t.Errorf("Expected 2 folders (root + dir1), got %d", res.FolderCount)
	}
}
