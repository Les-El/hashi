package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Les-El/chexum/internal/checkpoint"
	"github.com/Les-El/chexum/internal/testutil"
)

func TestHandleRunFailure_WithCleanupError(t *testing.T) {
	_, cleanup := testutil.TempDir(t)
	defer cleanup()

	cm := checkpoint.NewCleanupManager(false)
	// Set baseDir to a non-existent path to cause CleanupOnExit to fail
	cm.SetBaseDir("/tmp/path-that-definitely-does-not-exist-12345")

	stdout, stderr, err := testutil.CaptureOutput(func() {
		handleRunFailure(cm, "test-phase", fmt.Errorf("test-error"))
	})
	if err != nil {
		t.Fatalf("CaptureOutput failed: %v", err)
	}

	testutil.AssertContains(t, stdout, "Attempting cleanup after test-phase failure...")
	testutil.AssertContains(t, stderr, "Cleanup also failed:")
}

func TestMain_ExitCode_Coverage(t *testing.T) {
	// Mock osExit
	oldExit := osExit
	defer func() { osExit = oldExit }()
	var exitCode int
	osExit = func(code int) {
		exitCode = code
	}

	// Make run() fail by blocking major_checkpoint
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	testutil.CreateFile(t, tmpDir, "major_checkpoint", "blocker")

	main()

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestRun_AnalysisFailure(t *testing.T) {
	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// run() fails if analysis fails. Analysis fails if root path is invalid.
	// But run() uses "." as root path.
	// We can make run() fail by having no config package.
	os.MkdirAll(filepath.Join(tmpDir, "internal"), 0755)

	err := run()
	if err == nil {
		t.Error("expected run() to fail when analysis fails")
	}
}

func TestGenerateReports_AllErrorPaths(t *testing.T) {
	ctx := context.Background()
	cleanupMgr := checkpoint.NewCleanupManager(false)
	issues := []checkpoint.Issue{}
	flags := []checkpoint.FlagStatus{}

	t.Run("WorkspaceFailure", func(t *testing.T) {
		oldNewWorkspace := checkpoint.NewWorkspace
		defer func() { checkpoint.NewWorkspace = oldNewWorkspace }()
		checkpoint.NewWorkspace = func(useMem bool) (*checkpoint.Workspace, error) {
			return nil, fmt.Errorf("mock error")
		}

		_, err := generateReports(ctx, issues, flags, cleanupMgr)
		if err == nil || !strings.Contains(err.Error(), "creating workspace") {
			t.Errorf("expected workspace creation error, got: %v", err)
		}
	})
}

func TestRunAnalysis_WorkspaceFailure(t *testing.T) {
	ctx := context.Background()
	cleanupMgr := checkpoint.NewCleanupManager(false)
	engines := []checkpoint.AnalysisEngine{}

	t.Run("WorkspaceFailure", func(t *testing.T) {
		oldNewWorkspace := checkpoint.NewWorkspace
		defer func() { checkpoint.NewWorkspace = oldNewWorkspace }()
		count := 0
		checkpoint.NewWorkspace = func(useMem bool) (*checkpoint.Workspace, error) {
			count++
			if count == 2 {
				return nil, fmt.Errorf("mock error")
			}
			return oldNewWorkspace(useMem)
		}

		_, _, err := runAnalysis(ctx, engines, cleanupMgr, ".")
		if err == nil || !strings.Contains(err.Error(), "creating workspace") {
			t.Errorf("expected creating workspace error, got: %v", err)
		}
	})
}




func TestRunAnalysis_FlagSystemErrorPaths(t *testing.T) {
	ctx := context.Background()
	cleanupMgr := checkpoint.NewCleanupManager(false)
	engines := []checkpoint.AnalysisEngine{
		checkpoint.NewCodeAnalyzer(),
	}

	t.Run("NoConfigPackage", func(t *testing.T) {
		tmpDir, cleanup := testutil.TempDir(t)
		defer cleanup()
		
		// Create internal dir but NO config package
		os.MkdirAll(filepath.Join(tmpDir, "internal"), 0755)

		_, _, err := runAnalysis(ctx, engines, cleanupMgr, tmpDir)
		if err == nil {
			t.Error("expected error when config package is missing")
		}
	})
}