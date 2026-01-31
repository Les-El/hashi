package checkpoint

import (
	"context"
	"fmt"
	"sync"
)

// IssueCollector collects issues from multiple engines.
type IssueCollector struct {
	mu     sync.Mutex
	issues []Issue
}

// NewIssueCollector creates a new collector.
func NewIssueCollector() *IssueCollector {
	return &IssueCollector{
		issues: make([]Issue, 0),
	}
}

// Collect adds issues to the collector.
func (c *IssueCollector) Collect(issues []Issue) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.issues = append(c.issues, issues...)
}

// Issues returns all collected issues.
func (c *IssueCollector) Issues() []Issue {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]Issue(nil), c.issues...)
}

// Runner coordinates multiple analysis engines.
type Runner struct {
	engines        []AnalysisEngine
	collector      *IssueCollector
	cleanupManager *CleanupManager
}

// NewRunner creates a new runner.
func NewRunner(engines []AnalysisEngine, cleanupManager *CleanupManager) *Runner {
	return &Runner{
		engines:        engines,
		collector:      NewIssueCollector(),
		cleanupManager: cleanupManager,
	}
}

// Run executes all registered engines concurrently.
func (r *Runner) Run(ctx context.Context, path string) error {
	ws, err := NewWorkspace(false) // Use disk-based workspace for analysis
	if err != nil {
		return err
	}
	// Logic: Register workspace with CleanupManager for robust disposal
	if r.cleanupManager != nil {
		r.cleanupManager.RegisterWorkspace(ws)
	}
	defer ws.Cleanup()

	var wg sync.WaitGroup
	var errMu sync.Mutex
	var errors []error

	for _, engine := range r.engines {
		wg.Add(1)
		go func(eng AnalysisEngine) {
			defer wg.Done()
			issues, err := eng.Analyze(ctx, path, ws)
			if err != nil {
				errMu.Lock()
				errors = append(errors, fmt.Errorf("engine %s failed: %w", eng.Name(), err))
				errMu.Unlock()
				return
			}
			r.collector.Collect(issues)
		}(engine)
	}

	wg.Wait()

	if len(errors) > 0 {
		// Aggregate errors into a single error
		var combinedErr string
		for _, e := range errors {
			combinedErr += e.Error() + "; "
		}
		return fmt.Errorf("analysis engines encountered errors: %s", combinedErr)
	}

	return nil
}

// GetIssues returns the findings from the last run.
func (r *Runner) GetIssues() []Issue {
	return r.collector.Issues()
}
