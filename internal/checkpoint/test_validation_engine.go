package checkpoint

import (
	"context"
	"fmt"
	"strings"
)

// TestValidationEngine handles all execution-based testing and coverage monitoring.
type TestValidationEngine struct {
	*BaseEngine
	monitor *CoverageMonitor
}

// NewTestValidationEngine creates a new TestValidationEngine.
func NewTestValidationEngine(threshold float64) *TestValidationEngine {
	e := &TestValidationEngine{
		BaseEngine: NewBaseEngine("TestValidationEngine"),
		monitor:    NewCoverageMonitor(threshold),
	}

	// Register core tasks
	e.RegisterTask(e.RunStandardTests)
	e.RegisterTask(e.CheckRaceConditions)
	e.RegisterTask(e.AnalyzePerformance)

	return e
}

// RunStandardTests runs the test suite with coverage.
func (e *TestValidationEngine) RunStandardTests(ctx context.Context, path string, ws *Workspace) ([]Issue, error) {
	testPath := "./..."
	if path != "." && path != "" {
		testPath = path + "/..."
	}

	// Logic: Consolidated execution from CIEngine
	cmd, err := safeCommand(ctx, "go", "test", "-cover", testPath)
	if err != nil {
		return []Issue{{
			ID:          "TEST-FAILURE",
			Category:    Testing,
			Severity:    Critical,
			Title:       "Test suite failed",
			Description: fmt.Sprintf("Standard test suite failed to execute successfully in %s.", path),
			Location:    path,
			Suggestion:  "Fix the failing tests.",
			Priority:    P0,
		}}, nil
	}

	output, _ := cmd.CombinedOutput()
	outStr := string(output)

	var issues []Issue
	// Coverage analysis
	coverage, _ := e.monitor.ParseCoverageOutput(outStr)
	if len(coverage) > 0 {
		failures, ok := e.monitor.ValidateThreshold(coverage)
		if !ok {
			for _, failure := range failures {
				issues = append(issues, Issue{
					ID:          "LOW-COVERAGE",
					Category:    Testing,
					Severity:    High,
					Title:       "Test coverage below threshold",
					Description: failure,
					Location:    path,
					Suggestion:  "Add more tests to reach the target coverage.",
					Priority:    P1,
				})
			}
		}
	}

	return issues, nil
}

// CheckRaceConditions runs tests with the race detector.
func (e *TestValidationEngine) CheckRaceConditions(ctx context.Context, path string, ws *Workspace) ([]Issue, error) {
	testPath := "./..."
	if path != "." && path != "" {
		testPath = path + "/..."
	}

	// Gap: Implementation of Race Detection
	cmd, err := safeCommand(ctx, "go", "test", "-race", "-short", testPath)
	if err != nil {
		return nil, nil // Error already handled by standard tests or tool issue
	}

	output, _ := cmd.CombinedOutput()
	outStr := string(output)

	if strings.Contains(outStr, "DATA RACE") {
		return []Issue{{
			ID:          "DATA-RACE-DETECTED",
			Category:    Security, // Race conditions are security risks
			Severity:    Critical,
			Title:       "Data race detected during test execution",
			Description: "The Go race detector identified a concurrent data access issue.",
			Location:    path,
			Suggestion:  "Use synchronization primitives (mutexes, channels) to protect shared state.",
			Priority:    P0,
		}}, nil
	}

	return nil, nil
}

// AnalyzePerformance runs benchmarks and detects regressions.
func (e *TestValidationEngine) AnalyzePerformance(ctx context.Context, path string, ws *Workspace) ([]Issue, error) {
	// Gap: Placeholder for Performance Regression Testing
	// In a real implementation, this would run benchmarks and compare against a baseline.
	return nil, nil
}
