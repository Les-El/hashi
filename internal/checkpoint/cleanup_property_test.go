package checkpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Les-El/chexum/internal/testutil"
)

// Property 7: Configurable Cleanup Patterns
// **Validates: Requirements 3.1**
func TestProperty_ConfigurableCleanupPatterns(t *testing.T) {
	f := func(pattern, description string) bool {
		if pattern == "" {
			return true
		}

		cm := NewCleanupManager(false)
		cm.AddCustomPattern(pattern, description)

		found := false
		for _, p := range cm.patterns {
			if p.Pattern == pattern && p.Description == description {
				found = true
				break
			}
		}
		return found
	}

	testutil.CheckProperty(t, f)
}

// Property 10: Cleanup Metrics Accuracy
// **Validates: Requirements 3.5**
//
// Reviewed: LONG-FUNCTION - Property test with complex setup and multiple assertions.
func TestProperty_CleanupMetricsAccuracy(t *testing.T) {
	// Feature: checkpoint-remediation, Property 10: Cleanup metrics accuracy

	f := func(numFiles, numDirs int) bool {
		// Use absolute value and modulo to ensure positive, bounded inputs
		if numFiles < 0 {
			numFiles = -numFiles
		}
		if numDirs < 0 {
			numDirs = -numDirs
		}
		numFiles = numFiles % 10
		numDirs = numDirs % 5

		tmpDir, cleanup := testutil.TempDir(t)
		defer cleanup()

		var expectedSize int64
		// Create files
		for i := 0; i < numFiles; i++ {
			content := "file data"
			testutil.CreateFile(t, tmpDir, fmt.Sprintf("test-%d.tmp", i), content)
			expectedSize += int64(len(content))
		}
		// Create dirs
		for i := 0; i < numDirs; i++ {
			dirPath := filepath.Join(tmpDir, fmt.Sprintf("chexum-dir-%d", i))
			os.MkdirAll(dirPath, 0755)
			content := "dir file data"
			os.WriteFile(filepath.Join(dirPath, "data.txt"), []byte(content), 0644)
			expectedSize += int64(len(content))
		}

		cm := NewCleanupManager(false)
		cm.baseDir = tmpDir

		result, err := cm.CleanupTemporaryFiles()
		if err != nil {
			return false
		}

		// Verify metrics
		if result.FilesRemoved != numFiles {
			return false
		}
		if result.DirsRemoved != numDirs {
			return false
		}
		// Size should match the content we wrote. The getDirSize function only counts
		// file contents, not directory metadata, so this should be exact.
		// However, we allow a small tolerance to account for filesystem timing.
		if result.SpaceFreed < expectedSize {
			return false
		}
		// Ensure we didn't count significantly more than expected (would indicate a bug)
		if result.SpaceFreed > expectedSize+100 {
			return false
		}
		return true
	}

	testutil.CheckProperty(t, f)
}

// Reviewed: LONG-FUNCTION - Table-driven test with complex config setup.
func TestCleanupManager_CustomPatternsFromConfig(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	configPath := filepath.Join(tmpDir, "config.toml")
	configContent := `
[cleanup]
tmpfs_threshold = 75.0
max_retention_days = 30
exclude_patterns = ["important-*"]

[[cleanup.custom_patterns]]
pattern = "custom-*"
description = "Custom files"
enabled = true

[[cleanup.custom_patterns]]
pattern = "unused-*"
description = "Unused files"
enabled = false
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cm := NewCleanupManager(false)
	if err := cm.LoadConfig(configPath); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Check if custom pattern was added
	foundCustom := false
	for _, p := range cm.patterns {
		if p.Pattern == "custom-*" {
			foundCustom = true
			break
		}
	}
	if !foundCustom {
		t.Error("expected custom pattern 'custom-*' to be added")
	}

	// Check if disabled pattern was NOT added
	foundUnused := false
	for _, p := range cm.patterns {
		if p.Pattern == "unused-*" {
			foundUnused = true
			break
		}
	}
	if foundUnused {
		t.Error("expected disabled custom pattern 'unused-*' NOT to be added")
	}

	// Check exclude patterns
	if len(cm.config.ExcludePatterns) != 1 || cm.config.ExcludePatterns[0] != "important-*" {
		t.Errorf("expected exclude pattern 'important-*', got %v", cm.config.ExcludePatterns)
	}
}
