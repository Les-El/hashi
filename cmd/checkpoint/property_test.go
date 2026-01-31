package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Les-El/chexum/internal/checkpoint"
	"github.com/Les-El/chexum/internal/testutil"
)

// Property 1: Test Suite Comprehensive Validation
// **Validates: Requirements 1.1, 1.2, 1.3**
//
// Reviewed: LONG-FUNCTION - Property test with complex environment setup.
func TestProperty_CheckpointAnalysisConsistency(t *testing.T) {
	f := func(hasTodo, hasUnsafe bool) bool {
		tmpDir, cleanup := testutil.TempDir(t)
		defer cleanup()

		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)

		// Setup environment
		testutil.CreateFile(t, tmpDir, "internal/config/config.go", "package config\n")
		testutil.CreateFile(t, tmpDir, "README.md", "# User Docs\n")
		testutil.CreateFile(t, tmpDir, "major_checkpoint/design.md", "# Design\n")
		testutil.GenerateMockGoFile(t, tmpDir, "main.go", hasTodo, hasUnsafe)

		ctx := context.Background()
		engines := []checkpoint.AnalysisEngine{
			checkpoint.NewCodeAnalyzer(),
		}
		cleanupMgr := checkpoint.NewCleanupManager(false)
		issues, _, err := runAnalysis(ctx, engines, cleanupMgr, tmpDir)
		if err != nil {
			t.Logf("runAnalysis failed: %v", err)
			return false
		}

		expectedCount := 0
		if hasTodo {
			expectedCount++
		}
		if hasUnsafe {
			expectedCount++
		}

		// Filter issues to only include those from main.go
		var mainIssues []checkpoint.Issue
		for _, issue := range issues {
			if strings.Contains(issue.Location, "main.go") {
				mainIssues = append(mainIssues, issue)
			}
		}

		if len(mainIssues) < expectedCount {
			t.Logf("hasTodo=%v, hasUnsafe=%v, expectedCount=%d, got %d", hasTodo, hasUnsafe, expectedCount, len(mainIssues))
			for i, issue := range issues {
				t.Logf("Issue %d: ID=%s, Location=%s", i, issue.ID, issue.Location)
			}
		}

		return len(mainIssues) >= expectedCount
	}

	testutil.CheckProperty(t, f)
}

// Property 19: Dogfooding Validation
// **Validates: Requirements 7.5**
func TestProperty_Dogfooding(t *testing.T) {
	// Feature: checkpoint-remediation, Property 19: Dogfooding validation
	// This property verifies that the checkpoint system can analyze itself.
	// Since we are running in the project root, we can run analysis on "."

	oldWd, _ := os.Getwd()
	// Find project root by looking for go.mod
	root := oldWd
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			t.Fatalf("could not find project root (go.mod)")
		}
		root = parent
	}
	os.Chdir(root)
	defer os.Chdir(oldWd)

	ctx := context.Background()
	engines := []checkpoint.AnalysisEngine{
		checkpoint.NewCodeAnalyzer(),
		checkpoint.NewDependencyAnalyzer(),
		checkpoint.NewDocAuditor(),
		checkpoint.NewTestingBattery(),
		checkpoint.NewFlagSystem(),
		checkpoint.NewQualityEngine(),
	}
	cleanupMgr := checkpoint.NewCleanupManager(false)
	issues, _, err := runAnalysis(ctx, engines, cleanupMgr, root)
	if err != nil {
		t.Fatalf("Dogfooding analysis failed: %v", err)
	}

	// We expect the system to find some issues in itself (it's not perfect yet)
	if len(issues) == 0 {
		t.Log("Wow, no issues found in the project itself!")
	} else {
		t.Logf("Found %d issues during dogfooding analysis", len(issues))
	}
}
