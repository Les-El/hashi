package checkpoint

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Reporter implements SynthesisEngine.
type Reporter struct {
	issues       []Issue
	flagStatuses []FlagStatus
}

// NewReporter initializes a new Reporter.
func NewReporter() *Reporter {
	return &Reporter{}
}

// Aggregate collects issues and flag statuses for reporting.
func (r *Reporter) Aggregate(issues []Issue, flagStatuses []FlagStatus) error {
	r.issues = issues
	r.flagStatuses = flagStatuses
	return nil
}

// GenerateJSONReport creates a machine-readable JSON representation of all findings.
func (r *Reporter) GenerateJSONReport() (string, error) {
	data := struct {
		Issues       []Issue      `json:"issues"`
		FlagStatuses []FlagStatus `json:"flag_statuses"`
	}{
		Issues:       r.issues,
		FlagStatuses: r.flagStatuses,
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GenerateCSVReport creates a machine-readable CSV representation of issues.
func (r *Reporter) GenerateCSVReport() (string, error) {
	var sb strings.Builder
	writer := csv.NewWriter(&sb)

	// Header
	header := []string{"Status", "ID", "Category", "Severity", "Priority", "Title", "Location", "Effort", "Description", "Suggestion"}
	if err := writer.Write(header); err != nil {
		return "", err
	}

	for _, issue := range r.issues {
		row := []string{
			string(issue.Status),
			issue.ID,
			string(issue.Category),
			string(issue.Severity),
			string(issue.Priority),
			issue.Title,
			issue.Location,
			string(issue.Effort),
			issue.Description,
			issue.Suggestion,
		}
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}

	writer.Flush()
	return sb.String(), nil
}

// GenerateRemediationPlan creates a detailed markdown plan to fix identified issues.
func (r *Reporter) GenerateRemediationPlan() (string, error) {
	var sb strings.Builder
	sb.WriteString("# Remediation Plan: Chexum Project Stabilization\n\n")
	sb.WriteString("## Overview\n\n")
	sb.WriteString(fmt.Sprintf("Analysis identified %d issues across %d categories.\n\n", len(r.issues), r.countCategories()))

	// Group issues by priority
	byPriority := make(map[Priority][]Issue)
	for _, issue := range r.issues {
		byPriority[issue.Priority] = append(byPriority[issue.Priority], issue)
	}

	priorities := []Priority{P0, P1, P2, P3}
	for _, p := range priorities {
		issues := byPriority[p]
		if len(issues) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("## Priority %s Tasks\n\n", strings.ToUpper(string(p))))
		for _, issue := range issues {
			status := issue.Status
			if status == "" {
				status = Pending
			}
			sb.WriteString(fmt.Sprintf("### [%s] %s\n", issue.ID, issue.Title))
			sb.WriteString(fmt.Sprintf("- **Status**: %s\n", status))
			sb.WriteString(fmt.Sprintf("- **Location**: `%s`\n", issue.Location))
			sb.WriteString(fmt.Sprintf("- **Severity**: %s\n", issue.Severity))
			sb.WriteString(fmt.Sprintf("- **Effort**: %s\n", issue.Effort))
			sb.WriteString(fmt.Sprintf("- **Description**: %s\n", issue.Description))
			sb.WriteString(fmt.Sprintf("- **Suggestion**: %s\n\n", issue.Suggestion))
		}
	}

	return sb.String(), nil
}

// GenerateStatusDashboard creates a high-level health dashboard.
func (r *Reporter) GenerateStatusDashboard() (string, error) {
	var sb strings.Builder
	sb.WriteString("# Project Health Dashboard\n\n")

	// Summary stats
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Issues**: %d\n", len(r.issues)))
	sb.WriteString(fmt.Sprintf("- **Critical/High Issues**: %d\n", r.countHighSeverity()))
	sb.WriteString(fmt.Sprintf("- **CLI Flag Implementation**: %d/%d flags fully implemented\n\n", r.countImplementedFlags(), len(r.flagStatuses)))

	// Issue Distribution
	sb.WriteString("## Issue Distribution\n\n")
	sb.WriteString("| Category | Count | Status |\n")
	sb.WriteString("|----------|-------|--------|\n")

	categories := []IssueCategory{CodeQuality, Documentation, Testing, Security, Performance, Usability}
	for _, cat := range categories {
		count := r.countByCategory(cat)
		status := "✅ Good"
		if count > 5 {
			status = "⚠️ Needs Attention"
		}
		if count > 10 {
			status = "❌ Critical"
		}
		sb.WriteString(fmt.Sprintf("| %s | %d | %s |\n", cat, count, status))
	}

	return sb.String(), nil
}

// GenerateOnboardingGuide creates a guide for new developers.
func (r *Reporter) GenerateOnboardingGuide() (string, error) {
	var sb strings.Builder
	sb.WriteString("# Developer Onboarding Guide\n\n")

	sb.WriteString("## Prerequisites\n\n")
	sb.WriteString("- Go 1.24 or higher\n")
	sb.WriteString("- Git\n")
	sb.WriteString("- Make (optional, but recommended)\n\n")

	sb.WriteString("## Getting Started\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("# Clone the repository\n")
	sb.WriteString("git clone https://github.com/Les-El/chexum.git\n\n")
	sb.WriteString("# Install dependencies\n")
	sb.WriteString("go mod download\n\n")
	sb.WriteString("# Run tests\n")
	sb.WriteString("go test ./...\n\n")
	sb.WriteString("# Build the project\n")
	sb.WriteString("go build -o chexum ./cmd/chexum\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## Project Architecture\n\n")
	sb.WriteString("- `cmd/chexum`: The main entry point and CLI command definitions.\n")
	sb.WriteString("- `internal/`: Private packages containing the core logic:\n")
	sb.WriteString("  - `checkpoint`: The major checkpoint analysis system.\n")
	sb.WriteString("  - `config`: Configuration parsing and flag management.\n")
	sb.WriteString("  - `hash`: Core hashing algorithms and file processing.\n")
	sb.WriteString("  - `conflict`: Flag conflict resolution logic.\n")
	sb.WriteString("- `docs/`: Comprehensive project documentation and ADRs.\n\n")

	sb.WriteString("## Coding Standards\n\n")
	sb.WriteString("1. **Testing**: All new features must include unit tests and, where appropriate, property-based tests.\n")
	sb.WriteString("2. **Documentation**: All exported functions and types must have descriptive Go documentation.\n")
	sb.WriteString("3. **Error Handling**: Use custom error types and avoid `panic` for expected error conditions.\n")
	sb.WriteString("4. **Formatting**: Always run `go fmt` before committing.\n")

	return sb.String(), nil
}

func (r *Reporter) countCategories() int {
	cats := make(map[IssueCategory]bool)
	for _, issue := range r.issues {
		cats[issue.Category] = true
	}
	return len(cats)
}

func (r *Reporter) countByCategory(cat IssueCategory) int {
	count := 0
	for _, issue := range r.issues {
		if issue.Category == cat {
			count++
		}
	}
	return count
}

func (r *Reporter) countHighSeverity() int {
	count := 0
	for _, issue := range r.issues {
		if issue.Severity == Critical || issue.Severity == High {
			count++
		}
	}
	return count
}

func (r *Reporter) countImplementedFlags() int {
	count := 0
	for _, flag := range r.flagStatuses {
		if flag.Status == FullyImplemented {
			count++
		}
	}
	return count
}

// SortIssues sorts the collected issues by priority and then severity.
func (r *Reporter) SortIssues() {
	sort.Slice(r.issues, func(i, j int) bool {
		// Sort by priority first
		pi := r.priorityValue(r.issues[i].Priority)
		pj := r.priorityValue(r.issues[j].Priority)
		if pi != pj {
			return pi < pj
		}
		// Then by severity
		si := r.severityValue(r.issues[i].Severity)
		sj := r.severityValue(r.issues[j].Severity)
		return si < sj
	})
}

func (r *Reporter) priorityValue(p Priority) int {
	switch p {
	case P0:
		return 0
	case P1:
		return 1
	case P2:
		return 2
	case P3:
		return 3
	default:
		return 4
	}
}

func (r *Reporter) severityValue(s Severity) int {
	switch s {
	case Critical:
		return 0
	case High:
		return 1
	case Medium:
		return 2
	case Low:
		return 3
	case Info:
		return 4
	default:
		return 5
	}
}
