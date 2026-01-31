package checkpoint

import (
	"context"
	"os"
	"testing"
)

func TestNewCIEngine(t *testing.T) {
	engine := NewCIEngine(85.0)

	if engine.Name() != "CIEngine" {
		t.Errorf("expected CIEngine, got %s", engine.Name())
	}
}

func TestCIEngine_Name(t *testing.T) {
	engine := NewCIEngine(85.0)
	if engine.Name() != "CIEngine" {
		t.Errorf("expected CIEngine, got %s", engine.Name())
	}
}

func TestCIEngine_Analyze(t *testing.T) {
	ctx := context.Background()
	engine := NewCIEngine(85.0)

	t.Run("SkipAnalysis", func(t *testing.T) {
		os.Setenv("SKIP_CI_ANALYSIS", "true")
		defer os.Unsetenv("SKIP_CI_ANALYSIS")
		ws, _ := NewWorkspace(true)

		issues, err := engine.Analyze(ctx, ".", ws)
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected 0 issues, got %d", len(issues))
		}
	})

	t.Run("RealAnalysis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping in short mode")
		}
		ws, _ := NewWorkspace(true)
		// Run on internal/color which is fast and has no dependencies on checkpoint
		issues, err := engine.Analyze(ctx, "../color", ws)
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}
		t.Logf("Found %d issues", len(issues))
	})

	t.Run("FailingTests", func(t *testing.T) {
		ws, _ := NewWorkspace(true)
		// Point to a directory that doesn't exist to cause 'go test' to fail
		issues, err := engine.Analyze(ctx, "/tmp/non-existent-dir-for-test", ws)
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}
		foundFailure := false
		for _, issue := range issues {
			if issue.ID == "CI-TEST-FAILURE" {
				foundFailure = true
				break
			}
		}
		if !foundFailure {
			t.Error("expected CI-TEST-FAILURE issue")
		}
	})
}

func TestCIEngine_VerifyTestCompleteness(t *testing.T) {
	ctx := context.Background()
	engine := NewCIEngine(85.0)
	if !engine.VerifyTestCompleteness(ctx, ".") {
		t.Error("expected true")
	}
}

func TestCIEngine_CheckCoverage(t *testing.T) {
	engine := NewCIEngine(85.0)
	mockOutput := "github.com/Les-El/chexum/internal/config	coverage: 79.1% of statements"
	issues := engine.checkCoverage(mockOutput, "internal/config")
	if len(issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].ID != "CI-COVERAGE-LOW" {
		t.Errorf("expected CI-COVERAGE-LOW, got %s", issues[0].ID)
	}
}
