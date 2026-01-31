package checkpoint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCleanupManager(t *testing.T) {
	cleanup := NewCleanupManager(true)
	if cleanup == nil {
		t.Error("Expected non-nil cleanup manager")
	}
	if !cleanup.verbose {
		t.Error("Expected verbose mode to be enabled")
	}
}

func TestCheckStorageUsage(t *testing.T) {
	cleanup := NewCleanupManager(false)

	needsCleanup, usage := cleanup.CheckStorageUsage(100.0) // Set threshold to 100% so it shouldn't trigger
	if needsCleanup {
		t.Errorf("Expected needsCleanup to be false with 100%% threshold, got true (usage: %.1f%%)", usage)
	}

	if usage < 0 || usage > 100 {
		t.Errorf("Expected usage to be between 0-100%%, got %.1f%%", usage)
	}
}

func TestCleanupManager_FormatBytes(t *testing.T) {
	cleanup := NewCleanupManager(false)

	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, test := range tests {
		result := cleanup.formatBytes(test.bytes)
		if result != test.expected {
			t.Errorf("formatBytes(%d) = %s, expected %s", test.bytes, result, test.expected)
		}
	}
}

func TestCleanupManager_GetDirSize(t *testing.T) {
	cleanup := NewCleanupManager(false)

	// Create a temporary directory with some files
	tmpDir := t.TempDir()

	// Create test files
	testFile1 := filepath.Join(tmpDir, "test1.txt")
	testFile2 := filepath.Join(tmpDir, "test2.txt")

	content1 := "Hello, World!"
	content2 := "This is a test file."

	if err := os.WriteFile(testFile1, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := os.WriteFile(testFile2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	size, err := cleanup.getDirSize(tmpDir)
	if err != nil {
		t.Fatalf("getDirSize failed: %v", err)
	}

	expectedSize := int64(len(content1) + len(content2))
	if size != expectedSize {
		t.Errorf("Expected directory size %d, got %d", expectedSize, size)
	}
}

func TestCleanupManager_Workspaces(t *testing.T) {
	tmpDir := t.TempDir()
	cleanup := NewCleanupManager(false)
	cleanup.SetBaseDir(tmpDir)

	// Create a disk-based workspace for testing
	ws, err := NewWorkspace(false)
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	// Manually set Root to something within our temp dir for the test
	oldRoot := ws.Root
	defer os.RemoveAll(oldRoot) // Cleanup the actual system temp dir

	ws.Root = filepath.Join(tmpDir, "test-workspace")
	os.MkdirAll(ws.Root, 0755)
	os.WriteFile(filepath.Join(ws.Root, "test.tmp"), []byte("data"), 0644)

	cleanup.RegisterWorkspace(ws)
	if len(cleanup.workspaces) != 1 {
		t.Errorf("Expected 1 workspace, got %d", len(cleanup.workspaces))
	}

	result, err := cleanup.CleanupTemporaryFiles()
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	if result.DirsRemoved != 1 {
		t.Errorf("Expected 1 directory removed, got %d", result.DirsRemoved)
	}

	if _, err := os.Stat(ws.Root); !os.IsNotExist(err) {
		t.Error("Workspace root still exists after cleanup")
	}

	if len(cleanup.workspaces) != 0 {
		t.Error("Workspaces list not cleared after cleanup")
	}
}

func TestCleanupTemporaryFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cleanup test in short mode")
	}

	tmpDir := t.TempDir()
	cleanup := NewCleanupManager(false)
	cleanup.SetBaseDir(tmpDir)

	// Create some dummy files to clean
	os.WriteFile(filepath.Join(tmpDir, "test-1.tmp"), []byte("data"), 0644)

	result, err := cleanup.CleanupTemporaryFiles()
	if err != nil {
		t.Fatalf("CleanupTemporaryFiles failed: %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result")
	}

	if result.FilesRemoved != 1 {
		t.Errorf("Expected 1 file removed, got %d", result.FilesRemoved)
	}
}
func TestCleanupOnExit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	cleanup := NewCleanupManager(false)
	cleanup.SetBaseDir(tmpDir)

	// Test that cleanup can be called without error
	err := cleanup.CleanupOnExit()
	if err != nil {
		t.Errorf("CleanupOnExit failed: %v", err)
	}
}

func TestRegisterWorkspace(t *testing.T) {
	cleanup := NewCleanupManager(false)
	ws, _ := NewWorkspace(true) // memory workspace
	cleanup.RegisterWorkspace(ws)
	if len(cleanup.workspaces) != 1 {
		t.Errorf("Expected 1 workspace, got %d", len(cleanup.workspaces))
	}
	if cleanup.workspaces[0] != ws {
		t.Error("Registered workspace does not match")
	}
}

func TestCleanupManager_SetDryRun(t *testing.T) {
	cleanup := NewCleanupManager(false)
	cleanup.SetDryRun(true)
	if !cleanup.dryRun {
		t.Error("Expected dryRun to be true")
	}
	cleanup.SetDryRun(false)
	if cleanup.dryRun {
		t.Error("Expected dryRun to be false")
	}
}

func TestCleanupManager_SetBaseDir(t *testing.T) {
	cleanup := NewCleanupManager(false)
	cleanup.SetBaseDir("/tmp/custom")
	if cleanup.baseDir != "/tmp/custom" {
		t.Errorf("Expected baseDir to be /tmp/custom, got %s", cleanup.baseDir)
	}
}

func TestCleanupManager_AddCustomPattern(t *testing.T) {
	cleanup := NewCleanupManager(false)
	cleanup.AddCustomPattern("*.log", "Log files")
	found := false
	for _, p := range cleanup.patterns {
		if p.Pattern == "*.log" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Custom pattern *.log not found")
	}
}

func TestCleanupManager_LoadConfig(t *testing.T) {
	cleanup := NewCleanupManager(false)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "cleanup.toml")
	os.WriteFile(configPath, []byte(`[cleanup]
custom_patterns = [{pattern = "*.old", description = "Old files", enabled = true}]`), 0644)

	err := cleanup.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
}

func TestCleanupManager_ValidatePatterns(t *testing.T) {
	cleanup := NewCleanupManager(false)
	if err := cleanup.ValidatePatterns(); err != nil {
		t.Errorf("ValidatePatterns failed: %v", err)
	}

	cleanup.AddCustomPattern("[", "Invalid pattern")
	if err := cleanup.ValidatePatterns(); err == nil {
		t.Error("Expected error for invalid pattern '['")
	}
}

func TestCleanupManager_PreviewCleanup(t *testing.T) {
	cleanup := NewCleanupManager(false)
	_, err := cleanup.PreviewCleanup()
	if err != nil {
		t.Errorf("PreviewCleanup failed: %v", err)
	}
}

func TestCleanupManager_AdditionalMethods(t *testing.T) {
	cleanup := NewCleanupManager(false)

	t.Run("ExcludePatterns", func(t *testing.T) {
		tmpDir := t.TempDir()
		cleanup.SetBaseDir(tmpDir)

		os.WriteFile(filepath.Join(tmpDir, "test-1.tmp"), []byte("data"), 0644)
		os.WriteFile(filepath.Join(tmpDir, "keep-me.tmp"), []byte("data"), 0644)

		cleanup.config.ExcludePatterns = []string{"keep-me.tmp"}

		result, _ := cleanup.CleanupTemporaryFiles()
		if result.FilesRemoved != 1 {
			t.Errorf("Expected 1 file removed, got %d", result.FilesRemoved)
		}

		if _, err := os.Stat(filepath.Join(tmpDir, "keep-me.tmp")); os.IsNotExist(err) {
			t.Error("Excluded file was removed")
		}
	})
}
