package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Les-El/chexum/internal/checkpoint"
	"github.com/Les-El/chexum/internal/testutil"
)

var binaryName = "checkpoint"

// TestMain runs after all tests and cleans up temporary files to prevent disk space issues
func TestMain(m *testing.M) {
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	// Build the binary for integration tests
	tmpDir, err := os.MkdirTemp("", "h-build-*")
	if err != nil {
		fmt.Printf("Failed to create temp dir for build: %v\n", err)
		os.Exit(1)
	}

	binaryPath := filepath.Join(tmpDir, binaryName)
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		fmt.Printf("Failed to build checkpoint: %v\nOutput: %s\n", err, string(output))
		os.RemoveAll(tmpDir)
		os.Exit(1)
	}

	// Add binary path to PATH
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)

	// Prevent tests from triggering global cleanup
	os.Setenv("CHEXUM_SKIP_CLEANUP", "true")
	defer os.Unsetenv("CHEXUM_SKIP_CLEANUP")

	code := m.Run()

	os.RemoveAll(tmpDir)
	os.Setenv("PATH", oldPath)
	cleanupTemporaryFiles()
	os.Exit(code)
}

func TestCLI_Checkpoint(t *testing.T) {
	// Skip CI analysis during integration test to save time and prevent recursion
	os.Setenv("SKIP_CI_ANALYSIS", "true")
	defer os.Unsetenv("SKIP_CI_ANALYSIS")

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Minimal environment
	testutil.CreateFile(t, tmpDir, "internal/config/config.go", "package config\n")
	testutil.CreateFile(t, tmpDir, "README.md", "# User Docs\n")
	testutil.CreateFile(t, tmpDir, "major_checkpoint/design.md", "# Design\n")
	testutil.GenerateMockGoFile(t, tmpDir, "main.go", false, false)

	cmd := exec.Command(binaryName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("checkpoint binary failed: %v\nOutput: %s", err, string(output))
	}

	if !strings.Contains(string(output), "Analysis complete.") {
		t.Errorf("expected output to contain 'Analysis complete.', got: %s", string(output))
	}
}

func TestMain_Error(t *testing.T) {
	// Create an environment that causes run() to fail
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Blocking major_checkpoint creation will cause run() to fail at saveReports
	testutil.CreateFile(t, tmpDir, "major_checkpoint", "blocker")

	cmd := exec.Command(binaryName)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("expected error from checkpoint binary, got nil")
	}

	if !strings.Contains(string(output), "Error:") {
		t.Errorf("expected output to contain 'Error:', got: %s", string(output))
	}
}

func TestMainDirect(t *testing.T) {
	// Skip CI analysis
	os.Setenv("SKIP_CI_ANALYSIS", "true")
	defer os.Unsetenv("SKIP_CI_ANALYSIS")

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Minimal environment
	testutil.CreateFile(t, tmpDir, "internal/config/config.go", "package config\n")
	testutil.CreateFile(t, tmpDir, "README.md", "# User Docs\n")
	testutil.CreateFile(t, tmpDir, "major_checkpoint/design.md", "# Design\n")
	testutil.GenerateMockGoFile(t, tmpDir, "main.go", false, false)

	// Capture output
	stdout, _, err := testutil.CaptureOutput(func() {
		main()
	})

	if err != nil {
		t.Fatalf("CaptureOutput failed: %v", err)
	}

	testutil.AssertContains(t, stdout, "Analysis complete.")
}

func cleanupTemporaryFiles() {
	// Only remove temporary files created by tests, not active Go build artifacts
	tmpDir := os.TempDir()
	tmpPatterns := []string{
		filepath.Join(tmpDir, "chexum-*"),
		filepath.Join(tmpDir, "checkpoint-*"),
		filepath.Join(tmpDir, "test-*"),
	}

	for _, pattern := range tmpPatterns {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			os.RemoveAll(match)
		}
	}
}

func TestRunAnalysis(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	setupRunAnalysisMockEnv(t, tmpDir)

	ctx := context.Background()
	engines := []checkpoint.AnalysisEngine{
		checkpoint.NewCodeAnalyzer(),
		checkpoint.NewDocAuditor(),
	}
	cleanupMgr := checkpoint.NewCleanupManager(false)

	t.Run("Normal", func(t *testing.T) {
		issues, _, err := runAnalysis(ctx, engines, cleanupMgr, tmpDir)
		if err != nil {
			t.Fatalf("runAnalysis failed: %v", err)
		}
		if len(issues) == 0 {
			t.Errorf("expected some issues, got 0")
		}
	})

	t.Run("NonExistentPath", func(t *testing.T) {
		_, _, err := runAnalysis(ctx, engines, cleanupMgr, "/tmp/non-existent-path-abc-123")
		if err == nil {
			t.Error("expected error for non-existent path")
		}
	})

	t.Run("EngineFailure", func(t *testing.T) {
		failEngines := []checkpoint.AnalysisEngine{&failingEngine{}}
		_, _, err := runAnalysis(ctx, failEngines, cleanupMgr, tmpDir)
		if err == nil {
			t.Error("expected error from failing engine, got nil")
		}
	})
}

func setupRunAnalysisMockEnv(t *testing.T, tmpDir string) {
	testutil.CreateFile(t, tmpDir, "internal/config/config.go", "package config\n")
	testutil.CreateFile(t, tmpDir, "README.md", "# User Docs\n")
	testutil.CreateFile(t, tmpDir, "major_checkpoint/design.md", "# Design\n")
	testutil.GenerateMockGoFile(t, tmpDir, "main.go", true, true)
}

type failingEngine struct{}

func (f *failingEngine) Name() string { return "FailingEngine" }
func (f *failingEngine) Analyze(ctx context.Context, path string, ws *checkpoint.Workspace) ([]checkpoint.Issue, error) {
	return nil, fmt.Errorf("intentional failure")
}

func TestGenerateReports(t *testing.T) {
	ctx := context.Background()
	issues := []checkpoint.Issue{
		{
			Category:    checkpoint.CodeQuality,
			Severity:    checkpoint.Medium,
			Description: "Test issue",
			Location:    "test.go:10",
		},
	}
	flags := []checkpoint.FlagStatus{
		{
			Name: "test-flag",
		},
	}

	cleanup := checkpoint.NewCleanupManager(false)
	reports, err := generateReports(ctx, issues, flags, cleanup)
	if err != nil {
		t.Fatalf("generateReports failed: %v", err)
	}

	if reports.plan == "" {
		t.Error("expected remediation plan report, got empty string")
	}
	if reports.dashboard == "" {
		t.Error("expected dashboard report, got empty string")
	}
	if reports.jsonReport == "" {
		t.Error("expected JSON report, got empty string")
	}
}

// Reviewed: LONG-FUNCTION - Integration test with multiple file system assertions.
func TestSaveReports(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	reports := analysisReports{
		plan:       "plan content",
		dashboard:  "dashboard content",
		guide:      "guide content",
		flagReport: "flag content",
		jsonReport: "{}",
		csvReport:  "csv,content",
	}

	err := saveReports(reports)
	if err != nil {
		t.Fatalf("saveReports failed: %v", err)
	}

	// Since saveReports now creates a snapshot, we should check in the snapshots directory
	snapshotsDir := filepath.Join("major_checkpoint", "active", "snapshots")
	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		t.Fatalf("failed to read snapshots dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one snapshot")
	}

	snapshotPath := filepath.Join(snapshotsDir, entries[0].Name())
	expectedFiles := []string{
		"findings_remediation_plan.md",
		"findings_status_dashboard.md",
		"findings_onboarding_guide.md",
		"findings_flag_report.md",
		"findings_remediation_plan.json",
		"findings_remediation_plan.csv",
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(snapshotPath, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist, but it doesn't", fullPath)
		}
	}

	t.Run("MkdirFailure", func(t *testing.T) {
		os.RemoveAll(filepath.Join(tmpDir, "major_checkpoint"))
		// Block major_checkpoint directory creation
		blockedDir := filepath.Join(tmpDir, "major_checkpoint")
		os.WriteFile(blockedDir, []byte("I am a file"), 0644)

		err := saveReports(reports)
		if err == nil {
			t.Error("expected error when mkdir fails, got nil")
		}
	})

	t.Run("WriteFileFailure", func(t *testing.T) {
		os.RemoveAll(filepath.Join(tmpDir, "major_checkpoint"))
		latestDir := filepath.Join("major_checkpoint", "active", "latest")
		os.MkdirAll(latestDir, 0755)
		// Block one of the files with a directory
		blockedFile := filepath.Join(latestDir, "findings_remediation_plan.md")
		os.MkdirAll(blockedFile, 0755)

		err := saveReports(reports)
		if err == nil {
			t.Error("expected error when WriteFile fails, got nil")
		}
	})

	t.Run("CreateSnapshotFailure", func(t *testing.T) {
		os.RemoveAll(filepath.Join(tmpDir, "major_checkpoint"))
		// Block snapshot directory creation
		snapshotsDir := filepath.Join("major_checkpoint", "active", "snapshots")
		os.MkdirAll(filepath.Dir(snapshotsDir), 0755)
		// Block snapshots directory with a file
		os.WriteFile(snapshotsDir, []byte("blocker"), 0644)

		err := saveReports(reports)
		if err == nil {
			t.Error("expected error when CreateSnapshot fails, got nil")
		}
	})

	t.Run("ArchiveWarning", func(t *testing.T) {
		os.RemoveAll(filepath.Join(tmpDir, "major_checkpoint"))
		// Create 6 snapshots to trigger archival (maxActive is 5)
		for i := 0; i < 6; i++ {
			snapDir := filepath.Join("major_checkpoint", "active", "snapshots", fmt.Sprintf("snap%d", i))
			os.MkdirAll(snapDir, 0755)
		}

		// Block archival by creating a file where the archive dir should be
		archiveRoot := filepath.Join("major_checkpoint", "archive")
		os.MkdirAll(filepath.Dir(archiveRoot), 0755)
		os.WriteFile(archiveRoot, []byte("blocker"), 0644)

		stdout, _, err := testutil.CaptureOutput(func() {
			err := saveReports(reports)
			if err != nil {
				t.Fatalf("saveReports failed: %v", err)
			}
		})
		if err != nil {
			t.Fatalf("CaptureOutput failed: %v", err)
		}
		testutil.AssertContains(t, stdout, "Warning: Archival failed")
	})
}

// Reviewed: LONG-FUNCTION - Smoke test for the main execution logic.
func TestRun(t *testing.T) {
	// This is a smoke test for the run function logic.
	os.Setenv("SKIP_CI_ANALYSIS", "true")
	defer os.Unsetenv("SKIP_CI_ANALYSIS")

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Create necessary directory structure and files for FlagSystem and other engines
	testutil.CreateFile(t, tmpDir, "internal/config/config.go", "package config\n")
	testutil.CreateFile(t, tmpDir, "README.md", "# User Docs\n")
	testutil.CreateFile(t, tmpDir, "major_checkpoint/design.md", "# Design\n")
	testutil.GenerateMockGoFile(t, tmpDir, "main.go", false, false)

	// Capture output to verify it runs
	stdout, stderr, err := testutil.CaptureOutput(func() {
		if err := run(); err != nil {
			t.Errorf("run() failed: %v", err)
		}
	})

	if err != nil {
		t.Fatalf("CaptureOutput failed: %v", err)
	}

	testutil.AssertContains(t, stdout, "Starting Major Checkpoint Analysis...")
	testutil.AssertContains(t, stdout, "Analysis complete.")
	if stderr != "" {
		t.Errorf("expected empty stderr, got %q", stderr)
	}

	t.Run("WithCleanup", func(t *testing.T) {
		os.Setenv("CHEXUM_SKIP_CLEANUP", "")
		defer os.Setenv("CHEXUM_SKIP_CLEANUP", "true")

		stdout, _, _ := testutil.CaptureOutput(func() {
			run()
		})
		testutil.AssertContains(t, stdout, "Performing post-analysis cleanup...")
	})

	t.Run("Failure", func(t *testing.T) {
		// Block saveReports
		os.RemoveAll(filepath.Join(tmpDir, "major_checkpoint"))
		testutil.CreateFile(t, tmpDir, "major_checkpoint", "blocker")

		stdout, _, err := testutil.CaptureOutput(func() {
			if err := run(); err == nil {
				t.Error("expected error from run() when saveReports fails, got nil")
			} else {
				t.Logf("run() failed as expected: %v", err)
			}
		})
		if err != nil {
			t.Fatalf("CaptureOutput failed: %v", err)
		}
		// If it failed at analysis, it might be due to missing files in the mock env
		if !strings.Contains(stdout, "Attempting cleanup after report generation failure...") &&
			!strings.Contains(stdout, "Attempting cleanup after analysis failure...") {
			t.Errorf("expected stdout to contain a failure cleanup message, got: %s", stdout)
		}
	})
}

func TestHandleRunFailure(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cm := checkpoint.NewCleanupManager(false)
	cm.SetBaseDir(tmpDir)

	// Minimal smoke test for handleRunFailure
	handleRunFailure(cm, "test-phase", fmt.Errorf("test-error"))

	t.Run("CleanupFail", func(t *testing.T) {
		// Mock cleanup failure is hard because CleanupOnExit returns error but doesn't exit.
		// It writes to stderr.
	})
}

func TestCheckInitialResources(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	cm := checkpoint.NewCleanupManager(false)
	cm.SetBaseDir(tmpDir)

	// Normal case
	checkInitialResources(cm, 100.0)

	// Trigger warning
	stdout, _, _ := testutil.CaptureOutput(func() {
		checkInitialResources(cm, -1.0)
	})
	testutil.AssertContains(t, stdout, "Warning: Storage usage")
}

func TestRegisterEngines(t *testing.T) {
	engines := registerEngines()
	if len(engines) == 0 {
		t.Error("expected at least one engine")
	}
}
