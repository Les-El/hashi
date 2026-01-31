package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Les-El/chexum/internal/checkpoint"
)

var (
	osExit = os.Exit
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		osExit(1)
	}
}

func run() error {
	ctx := context.Background()

	fmt.Println("Starting Major Checkpoint Analysis...")

	cleanup := checkpoint.NewCleanupManager(true)
	checkInitialResources(cleanup, 75.0)

	engines := registerEngines()

	issues, flags, err := runAnalysis(ctx, engines, cleanup, ".")
	if err != nil {
		handleRunFailure(cleanup, "analysis", err)
		return fmt.Errorf("running analysis: %w", err)
	}

	reports, err := generateReports(ctx, issues, flags, cleanup)
	if err != nil {
		handleRunFailure(cleanup, "report generation", err)
		return fmt.Errorf("generating reports: %w", err)
	}

	if err := saveReports(reports); err != nil {
		handleRunFailure(cleanup, "report generation", err)
		return fmt.Errorf("saving reports: %w", err)
	}

	fmt.Println("Analysis complete. Reports generated in major_checkpoint/ directory.")

	// Perform cleanup at the end
	if os.Getenv("CHEXUM_SKIP_CLEANUP") == "" {
		fmt.Println("Performing post-analysis cleanup...")
		if err := cleanup.CleanupOnExit(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Cleanup failed: %v\n", err)
		}
	}
	return nil
}

func checkInitialResources(cleanup *checkpoint.CleanupManager, threshold float64) {
	// Check storage usage before starting
	if needsCleanup, usage := cleanup.CheckStorageUsage(threshold); needsCleanup {
		fmt.Printf("Warning: Storage usage is %.1f%%. Consider running cleanup before analysis.\n", usage)
	}
}

func registerEngines() []checkpoint.AnalysisEngine {
	return []checkpoint.AnalysisEngine{
		checkpoint.NewCodeAnalyzer(),
		checkpoint.NewDependencyAnalyzer(),
		checkpoint.NewDocAuditor(),
		checkpoint.NewTestValidationEngine(85.0), // Consolidated testing
		checkpoint.NewStaticAnalysisEngine(),     // Consolidated quality and missing tests
		checkpoint.NewFlagSystem(),
		checkpoint.NewCIEngine(85.0),
	}
}

func handleRunFailure(cleanup *checkpoint.CleanupManager, phase string, err error) {
	fmt.Printf("Attempting cleanup after %s failure...\n", phase)
	if cleanupErr := cleanup.CleanupOnExit(); cleanupErr != nil {
		fmt.Fprintf(os.Stderr, "Cleanup also failed: %v\n", cleanupErr)
	}
}

func runAnalysis(ctx context.Context, engines []checkpoint.AnalysisEngine, cleanup *checkpoint.CleanupManager, rootPath string) ([]checkpoint.Issue, []checkpoint.FlagStatus, error) {
	runner := checkpoint.NewRunner(engines, cleanup)

	fmt.Println("Running comprehensive project analysis...")
	if err := runner.Run(ctx, rootPath); err != nil {
		return nil, nil, err
	}

	issues := runner.GetIssues()

	// Special handling for flags as they return FlagStatus
	ws, err := checkpoint.NewWorkspace(true)
	if err != nil {
		return nil, nil, fmt.Errorf("creating workspace: %w", err)
	}
	cleanup.RegisterWorkspace(ws)
	defer ws.Cleanup()

	flagSystem := checkpoint.NewFlagSystem()
	flags, err := flagSystem.CatalogFlags(ctx, rootPath, ws)
	if err != nil {
		return nil, nil, fmt.Errorf("cataloging flags: %w", err)
	}

	if flags, err = flagSystem.ClassifyImplementation(ctx, rootPath, ws, flags); err != nil {
		return nil, nil, fmt.Errorf("classifying flags: %w", err)
	}
	if flags, err = flagSystem.PerformCrossReferenceAnalysis(ctx, rootPath, ws, flags); err != nil {
		return nil, nil, fmt.Errorf("cross-referencing flags: %w", err)
	}
	if flags, err = flagSystem.DetectConflicts(ctx, ws, flags); err != nil {
		return nil, nil, fmt.Errorf("detecting flag conflicts: %w", err)
	}
	if flags, err = flagSystem.ValidateFunctionality(ctx, ws, flags); err != nil {
		return nil, nil, fmt.Errorf("validating flag functionality: %w", err)
	}

	return issues, flags, nil
}

type analysisReports struct {
	plan       string
	dashboard  string
	guide      string
	flagReport string
	jsonReport string
	csvReport  string
}

func generateReports(ctx context.Context, issues []checkpoint.Issue, flags []checkpoint.FlagStatus, cleanup *checkpoint.CleanupManager) (analysisReports, error) {
	reporter := checkpoint.NewReporter()
	reporter.Aggregate(issues, flags)
	reporter.SortIssues()

	var err error
	var plan, dashboard, guide, jsonReport, csvReport, flagReport string

	if plan, err = reporter.GenerateRemediationPlan(); err != nil {
		return analysisReports{}, fmt.Errorf("generating remediation plan: %w", err)
	}
	if dashboard, err = reporter.GenerateStatusDashboard(); err != nil {
		return analysisReports{}, fmt.Errorf("generating status dashboard: %w", err)
	}
	if guide, err = reporter.GenerateOnboardingGuide(); err != nil {
		return analysisReports{}, fmt.Errorf("generating onboarding guide: %w", err)
	}
	if jsonReport, err = reporter.GenerateJSONReport(); err != nil {
		return analysisReports{}, fmt.Errorf("generating JSON report: %w", err)
	}
	if csvReport, err = reporter.GenerateCSVReport(); err != nil {
		return analysisReports{}, fmt.Errorf("generating CSV report: %w", err)
	}

	ws, err := checkpoint.NewWorkspace(true)
	if err != nil {
		return analysisReports{}, fmt.Errorf("creating workspace for flag report: %w", err)
	}
	cleanup.RegisterWorkspace(ws)
	defer ws.Cleanup()

	flagSystem := checkpoint.NewFlagSystem()
	if flagReport, err = flagSystem.GenerateStatusReport(ctx, ws, flags); err != nil {
		return analysisReports{}, fmt.Errorf("generating flag status report: %w", err)
	}

	return analysisReports{
		plan:       plan,
		dashboard:  dashboard,
		guide:      guide,
		flagReport: flagReport,
		jsonReport: jsonReport,
		csvReport:  csvReport,
	}, nil
}

func saveReports(r analysisReports) error {
	const rootDir = "major_checkpoint"
	latestDir := filepath.Join(rootDir, "active", "latest")
	if err := os.MkdirAll(latestDir, 0755); err != nil {
		return err
	}

	files := map[string]string{
		filepath.Join(latestDir, "findings_remediation_plan.md"):   r.plan,
		filepath.Join(latestDir, "findings_status_dashboard.md"):   r.dashboard,
		filepath.Join(latestDir, "findings_onboarding_guide.md"):   r.guide,
		filepath.Join(latestDir, "findings_flag_report.md"):        r.flagReport,
		filepath.Join(latestDir, "findings_remediation_plan.json"): r.jsonReport,
		filepath.Join(latestDir, "findings_remediation_plan.csv"):  r.csvReport,
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	fmt.Println("Organizing checkpoint artifacts...")
	organizer := checkpoint.NewOrganizer(rootDir)
	if err := organizer.CreateSnapshot(""); err != nil {
		return fmt.Errorf("creating snapshot: %w", err)
	}
	if err := organizer.ArchiveOldSnapshots(5); err != nil {
		fmt.Printf("Warning: Archival failed: %v\n", err)
	}

	return nil
}
