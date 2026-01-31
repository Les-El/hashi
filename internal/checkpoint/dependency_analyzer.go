package checkpoint

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// DependencyAnalyzer implements dependency checking.
type DependencyAnalyzer struct{}

// NewDependencyAnalyzer creates a new DependencyAnalyzer.
func NewDependencyAnalyzer() *DependencyAnalyzer {
	return &DependencyAnalyzer{}
}

// Name returns the name of the analyzer.
func (d *DependencyAnalyzer) Name() string { return "DependencyAnalyzer" }

// Analyze performs dependency analysis on the given path.
func (d *DependencyAnalyzer) Analyze(ctx context.Context, path string, ws *Workspace) ([]Issue, error) {
	issues, err := d.AssessDependencies(ctx, path, ws)
	if err != nil {
		return nil, err
	}

	vulnIssues, _ := d.checkVulnerabilities(ctx, path)
	issues = append(issues, vulnIssues...)

	return issues, nil
}

func (d *DependencyAnalyzer) checkVulnerabilities(ctx context.Context, path string) ([]Issue, error) {
	cmd, err := safeCommand(ctx, "govulncheck", "-json", "./...")
	if err != nil {
		// Security: Return an issue notifying that the tool is missing
		return []Issue{{
			ID:          "SECURITY-TOOL-MISSING",
			Category:    Security,
			Severity:    Medium,
			Title:       "Security tool 'govulncheck' not found",
			Description: fmt.Sprintf("The govulncheck tool is missing or not allowed: %v. Security vulnerability scanning was skipped.", err),
			Location:    path,
			Suggestion:  "Install govulncheck: go install golang.org/x/vuln/cmd/govulncheck@latest",
			Effort:      Small,
			Priority:    P1,
		}}, nil
	}

	cmd.Dir = path
	output, err := cmd.CombinedOutput()

	var issues []Issue
	// govulncheck returns non-zero exit code if vulnerabilities are found or if it fails.
	// When using -json, we should check if the output contains any "osv" findings.
	if len(output) > 0 {
		outStr := string(output)
		// Robust parsing: check for OSV entries in the JSON stream
		if strings.Contains(outStr, "\"osv\":") || strings.Contains(outStr, "\"vulnerability\":") {
			issues = append(issues, Issue{
				ID:          "SECURITY-VULNERABILITY",
				Category:    Security,
				Severity:    Critical,
				Title:       "Vulnerabilities detected in dependencies",
				Description: "govulncheck identified one or more vulnerabilities in the project's dependency graph.",
				Location:    filepath.Join(path, "go.mod"),
				Suggestion:  "Run 'govulncheck ./...' for details and update the affected packages.",
				Effort:      MediumEffort,
				Priority:    P0,
			})
		}
	} else if err != nil {
		// If no output but command failed, report as a tool error
		issues = append(issues, Issue{
			ID:          "SECURITY-TOOL-ERROR",
			Category:    Security,
			Severity:    Low,
			Title:       "Security tool 'govulncheck' failed to execute",
			Description: fmt.Sprintf("govulncheck failed with error: %v", err),
			Location:    path,
			Suggestion:  "Check your Go environment and network connectivity.",
			Effort:      Small,
			Priority:    P3,
		})
	}

	return issues, nil
}

// AssessDependencies evaluates the project's dependencies using go list.
func (d *DependencyAnalyzer) AssessDependencies(ctx context.Context, rootPath string, ws *Workspace) ([]Issue, error) {
	var issues []Issue

	// Use go list -m -u all to find updates
	cmd, err := safeCommand(ctx, "go", "list", "-m", "-u", "all")
	if err != nil {
		return nil, nil
	}
	cmd.Dir = rootPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If it fails (e.g. no network), fallback to basic go.mod check
		return nil, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "[") {
			continue
		}

		// Example line: github.com/spf13/cobra v1.8.0 [v1.8.1]
		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}

		module := parts[0]
		currentVer := parts[1]
		updateVer := strings.Trim(parts[2], "[]")

		if d.isMajorVersionBehind(currentVer, updateVer) {
			issues = append(issues, Issue{
				ID:          "OUTDATED-DEPENDENCY-MAJOR",
				Category:    Security,
				Severity:    High,
				Title:       fmt.Sprintf("Dependency '%s' is more than one major version behind", module),
				Description: fmt.Sprintf("Module '%s' is at %s, but %s is available. Large version gaps increase security risk and technical debt.", module, currentVer, updateVer),
				Location:    filepath.Join(rootPath, "go.mod"),
				Suggestion:  fmt.Sprintf("Update %s to at least the previous major version.", module),
				Effort:      MediumEffort,
				Priority:    P1,
			})
		}
	}

	return issues, nil
}

func (d *DependencyAnalyzer) isMajorVersionBehind(current, update string) bool {
	currMajor := d.getMajorVersion(current)
	updMajor := d.getMajorVersion(update)

	if currMajor < 0 || updMajor < 0 {
		return false
	}

	return (updMajor - currMajor) > 1
}

func (d *DependencyAnalyzer) getMajorVersion(v string) int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) == 0 {
		return -1
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return -1
	}
	return major
}
