package checkpoint

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestNewIssueCollector(t *testing.T) {
	collector := NewIssueCollector()
	if collector == nil {
		t.Fatal("NewIssueCollector returned nil")
	}
	if len(collector.issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(collector.issues))
	}
}

func TestCollect(t *testing.T) {
	collector := NewIssueCollector()
	issues := []Issue{
		{ID: "1", Title: "Issue 1"},
	}
	collector.Collect(issues)
	if len(collector.issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(collector.issues))
	}
}

func TestIssues(t *testing.T) {
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

	// Verify it's a copy
	collected[0].ID = "modified"
	if collector.issues[0].ID == "modified" {
		t.Error("Issues() returned a slice that shares the same underlying array")
	}
}

type MockAnalysisEngine struct {
	name   string
	issues []Issue
	err    error
}

func (m *MockAnalysisEngine) Name() string { return m.name }
func (m *MockAnalysisEngine) Analyze(ctx context.Context, path string, ws *Workspace) ([]Issue, error) {
	return m.issues, m.err
}

func TestNewRunner(t *testing.T) {
	engines := []AnalysisEngine{&MockAnalysisEngine{name: "Mock"}}
	runner := NewRunner(engines)
	if runner == nil {
		t.Fatal("NewRunner returned nil")
	}
	if len(runner.engines) != 1 {
		t.Errorf("expected 1 engine, got %d", len(runner.engines))
	}
	if runner.collector == nil {
		t.Error("runner.collector is nil")
	}
}

func TestRun(t *testing.T) {
	engine1 := &MockAnalysisEngine{
		name:   "Engine1",
		issues: []Issue{{ID: "E1-1"}},
	}
	runner := NewRunner([]AnalysisEngine{engine1})
	err := runner.Run(context.Background(), ".")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if len(runner.collector.issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(runner.collector.issues))
	}
}

func TestRun_Error(t *testing.T) {
	engine1 := &MockAnalysisEngine{
		name: "ErrorEngine",
		err:  fmt.Errorf("boom"),
	}
	runner := NewRunner([]AnalysisEngine{engine1})
	err := runner.Run(context.Background(), ".")
	if err == nil {
		t.Error("expected error from Run, got nil")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("expected error to contain 'boom', got %v", err)
	}
}

func TestGetIssues(t *testing.T) {
	engine1 := &MockAnalysisEngine{
		name:   "Engine1",
		issues: []Issue{{ID: "E1-1"}},
	}
	runner := NewRunner([]AnalysisEngine{engine1})
	_ = runner.Run(context.Background(), ".")
	issues := runner.GetIssues()
	if len(issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(issues))
	}
}
