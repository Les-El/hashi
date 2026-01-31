package checkpoint

import (
	"context"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

// discoverPackageFiles returns the absolute paths of all Go files in a package relative to root.
func discoverPackageFiles(root, pkgPath string) ([]string, error) {
	ctx := build.Default
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	pkg, err := ctx.ImportDir(filepath.Join(absRoot, pkgPath), build.ImportComment)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, f := range pkg.GoFiles {
		files = append(files, filepath.Join(absRoot, pkgPath, f))
	}
	return files, nil
}

// discoverCorePackages scans the internal directory to find all available packages.
func discoverCorePackages(root string) ([]string, error) {
	var packages []string
	internalDir := filepath.Join(root, "internal")

	if _, err := os.Stat(internalDir); os.IsNotExist(err) {
		return nil, nil
	}

	err := filepath.Walk(internalDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != internalDir {
			// Get relative path from root
			rel, err := filepath.Rel(root, path)
			if err == nil {
				packages = append(packages, rel)
			}
		}
		return nil
	})

	return packages, err
}

// discoverPackageByName finds a package path by its suffix (e.g., "config").
func discoverPackageByName(root, name string) (string, error) {
	pkgs, err := discoverCorePackages(root)
	if err != nil {
		return "", err
	}
	for _, pkg := range pkgs {
		if strings.HasSuffix(pkg, "/"+name) || pkg == "internal/"+name {
			return pkg, nil
		}
	}
	return "", fmt.Errorf("package %s not found", name)
}

// AnalysisEngine is the base interface for all analysis components.
type AnalysisEngine interface {
	Name() string
	Analyze(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
}

// AnalysisTask represents a single unit of analysis that can be executed by an engine.
type AnalysisTask func(ctx context.Context, path string, ws *Workspace) ([]Issue, error)

// BaseEngine provides common functionality for engines that run a series of tasks.
type BaseEngine struct {
	name  string
	tasks []AnalysisTask
}

// Name returns the engine name.
func (e *BaseEngine) Name() string { return e.name }

// RegisterTask adds a new task to the engine.
func (e *BaseEngine) RegisterTask(task AnalysisTask) {
	e.tasks = append(e.tasks, task)
}

// Analyze executes all registered tasks.
func (e *BaseEngine) Analyze(ctx context.Context, path string, ws *Workspace) ([]Issue, error) {
	var allIssues []Issue
	for _, task := range e.tasks {
		issues, err := task(ctx, path, ws)
		if err != nil {
			allIssues = append(allIssues, Issue{
				ID:          "ENGINE-TASK-FAILURE",
				Category:    CodeQuality,
				Severity:    Medium,
				Title:       fmt.Sprintf("Task failed in %s", e.name),
				Description: err.Error(),
				Location:    path,
				Suggestion:  "Check logs and environment settings.",
				Priority:    P2,
			})
			continue
		}
		allIssues = append(allIssues, issues...)
	}
	return allIssues, nil
}

// NewBaseEngine creates a new BaseEngine instance.
func NewBaseEngine(name string) *BaseEngine {
	return &BaseEngine{
		name:  name,
		tasks: make([]AnalysisTask, 0),
	}
}

// SynthesisEngine aggregates findings into actionable reports.
type SynthesisEngine interface {
	Aggregate(issues []Issue, flagStatuses []FlagStatus) error
	GenerateRemediationPlan() (string, error)
	GenerateStatusDashboard() (string, error)
	GenerateJSONReport() (string, error)
	GenerateCSVReport() (string, error)
}
