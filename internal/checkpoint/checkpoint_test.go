package checkpoint

import (
	"context"
	"strings"
	"testing"
)

type MockEngine struct {
	name   string
	issues []Issue
}

func (m *MockEngine) Name() string { return m.name }
func (m *MockEngine) Analyze(ctx context.Context, path string, ws *Workspace) ([]Issue, error) {
	return m.issues, nil
}

func TestIssueCollector(t *testing.T) {
	collector := NewIssueCollector()
	issues := []Issue{
		{ID: "1", Title: "Issue 1"},
		{ID: "2", Title: "Issue 2"},
	}
	collector.Collect(issues)

	collected := collector.Issues()
	if len(collected) != 2 {
		t.Errorf("expected 2 issues, got %d", len(collected))
	}
}

func TestDiscoverPackageByName(t *testing.T) {
	// Test known packages
	pkg, err := discoverPackageByName("../../", "config")
	if err != nil {
		t.Errorf("failed to discover config: %v", err)
	}
	if !strings.Contains(pkg, "internal/config") {
		t.Errorf("expected internal/config, got %s", pkg)
	}

	// Test non-existent package
	_, err = discoverPackageByName("../../", "non-existent")
	if err == nil {
		t.Error("expected error for non-existent package")
	}
}

// Reviewed: LONG-FUNCTION - Comprehensive system integration test.
func TestSystemIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	engines := []AnalysisEngine{
		NewCodeAnalyzer(),
		NewDependencyAnalyzer(),
		NewDocAuditor(),
		// NewTestingBattery(), // Skipping to avoid recursive go test calls
		NewFlagSystem(),
		NewQualityEngine(),
	}

	runner := NewRunner(engines)
	err := runner.Run(ctx, "../../")
	if err != nil {
		t.Fatalf("System run failed: %v", err)
	}

	issues := runner.GetIssues()
	if len(issues) == 0 {
		t.Log("No issues found, which is possible but unlikely in this repo.")
	}

	ws, _ := NewWorkspace(true)
	flagSystem := NewFlagSystem()
	flags, _ := flagSystem.CatalogFlags(ctx, "../../", ws)
	flags, _ = flagSystem.ClassifyImplementation(ctx, "../../", ws, flags)
	flags, _ = flagSystem.PerformCrossReferenceAnalysis(ctx, "../../", ws, flags)
	flags, _ = flagSystem.DetectConflicts(ctx, ws, flags)

	reporter := NewReporter()
	reporter.Aggregate(issues, flags)

	plan, err := reporter.GenerateRemediationPlan()
	if err != nil {
		t.Errorf("Failed to generate remediation plan: %v", err)
	}
	if plan == "" {
		t.Errorf("Remediation plan is empty")
	}

	dashboard, err := reporter.GenerateStatusDashboard()
	if err != nil {
		t.Errorf("Failed to generate status dashboard: %v", err)
	}
	if dashboard == "" {
		t.Errorf("Status dashboard is empty")
	}
}
