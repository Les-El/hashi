package checkpoint

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// DocAuditor implements DocumentationAuditor.
type DocAuditor struct {
	fset *token.FileSet
}

// NewDocAuditor initializes a new DocAuditor.
func NewDocAuditor() *DocAuditor {
	return &DocAuditor{
		fset: token.NewFileSet(),
	}
}

// Name returns the name of the analyzer.
func (d *DocAuditor) Name() string { return "DocAuditor" }

// Analyze executes the documentation audit logic.
func (d *DocAuditor) Analyze(ctx context.Context, path string, ws *Workspace) ([]Issue, error) {
	var allIssues []Issue

	docIssues, err := d.AuditGoDocumentation(ctx, path, ws)
	if err != nil {
		allIssues = append(allIssues, Issue{
			ID:          "AUDIT-ERROR",
			Category:    Documentation,
			Severity:    High,
			Title:       "Documentation audit failed",
			Description: fmt.Sprintf("Error during documentation audit: %v", err),
			Location:    path,
			Suggestion:  "Check file permissions and project structure.",
			Effort:      Small,
			Priority:    P1,
		})
	}
	allIssues = append(allIssues, docIssues...)

	exampleIssues, err := d.VerifyExamples(ctx, path, ws)
	if err != nil {
		allIssues = append(allIssues, Issue{
			ID:          "EXAMPLE-ERROR",
			Category:    Documentation,
			Severity:    High,
			Title:       "Example verification failed",
			Description: fmt.Sprintf("Error during example verification: %v", err),
			Location:    path,
			Suggestion:  "Ensure 'examples/' directory exists and is accessible.",
			Effort:      Small,
			Priority:    P1,
		})
	}
	allIssues = append(allIssues, exampleIssues...)

	readmeIssues, _ := d.ValidateREADME(ctx, path, ws)
	allIssues = append(allIssues, readmeIssues...)

	archIssues, _ := d.CheckArchitecturalDocs(ctx, path, ws)
	allIssues = append(allIssues, archIssues...)

	return allIssues, nil
}

// AuditGoDocumentation checks for missing documentation on exported functions.
func (d *DocAuditor) AuditGoDocumentation(ctx context.Context, rootPath string, ws *Workspace) ([]Issue, error) {
	var issues []Issue

	err := filepath.Walk(rootPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(filePath, ".go") || strings.HasSuffix(filePath, "_test.go") {
			return nil
		}

		fileIssues, err := d.auditFile(filePath)
		if err != nil {
			issues = append(issues, Issue{
				ID:          "PARSE-ERROR",
				Category:    Documentation,
				Severity:    High,
				Title:       "Failed to parse Go file for documentation",
				Description: fmt.Sprintf("Parser error in %s: %v", filePath, err),
				Location:    filePath,
				Suggestion:  "Ensure the file is valid Go code.",
				Effort:      Small,
				Priority:    P1,
			})
			return nil
		}
		issues = append(issues, fileIssues...)
		return nil
	})

	return issues, err
}

func (d *DocAuditor) auditFile(filePath string) ([]Issue, error) {
	var issues []Issue

	f, err := parser.ParseFile(d.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.IsExported() && x.Doc == nil {
				issues = append(issues, Issue{
					ID:          "MISSING-DOC",
					Category:    Documentation,
					Severity:    Medium,
					Title:       "Missing documentation for exported function",
					Description: fmt.Sprintf("Exported function '%s' has no documentation comment.", x.Name.Name),
					Location:    d.fset.Position(x.Pos()).String(),
					Suggestion:  "Add a documentation comment to the function.",
					Effort:      Small,
					Priority:    P2,
				})
			}
		case *ast.TypeSpec:
			if x.Name.IsExported() && x.Doc == nil {
				// Note: TypeSpec doc can be on the GenDecl
				// This is a simplification.
			}
		}
		return true
	})

	return issues, nil
}

// ValidateREADME ensures the README exists and meets quality standards.
func (d *DocAuditor) ValidateREADME(ctx context.Context, rootPath string, ws *Workspace) ([]Issue, error) {
	var issues []Issue
	readmePath := filepath.Join(rootPath, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		return []Issue{{
			ID:          "MISSING-README",
			Category:    Documentation,
			Severity:    Critical,
			Title:       "Missing README.md",
			Description: "The project is missing a README.md file in the root directory.",
			Location:    readmePath,
			Suggestion:  "Create a README.md file to provide project overview and instructions.",
			Effort:      Small,
			Priority:    P1,
		}}, nil
	}

	content := string(data)
	requiredSections := []string{"Overview", "Installation", "Usage", "License"}
	for _, section := range requiredSections {
		if !strings.Contains(content, "# "+section) && !strings.Contains(content, "## "+section) {
			issues = append(issues, Issue{
				ID:          "README-MISSING-SECTION",
				Category:    Documentation,
				Severity:    Medium,
				Title:       fmt.Sprintf("README missing '%s' section", section),
				Description: fmt.Sprintf("The README.md is missing a header for the '%s' section.", section),
				Location:    readmePath,
				Suggestion:  fmt.Sprintf("Add a '# %s' or '## %s' section to the README.md.", section, section),
				Effort:      Small,
				Priority:    P2,
			})
		}
	}

	return issues, nil
}

// CheckArchitecturalDocs verifies the presence of architectural documentation.
func (d *DocAuditor) CheckArchitecturalDocs(ctx context.Context, rootPath string, ws *Workspace) ([]Issue, error) {
	var issues []Issue
	adrDir := filepath.Join(rootPath, "docs/adr")
	files, err := os.ReadDir(adrDir)
	if err != nil || len(files) == 0 {
		issues = append(issues, Issue{
			ID:          "MISSING-ADRS",
			Category:    Documentation,
			Severity:    Low,
			Title:       "Missing Architectural Decision Records (ADRs)",
			Description: "No ADRs were found in 'docs/adr/'. ADRs are essential for tracking long-term architectural decisions.",
			Location:    adrDir,
			Suggestion:  "Start documenting major architectural decisions using the ADR format.",
			Effort:      MediumEffort,
			Priority:    P3,
		})
	}

	designDir := filepath.Join(rootPath, "docs/design")
	designFiles, err := os.ReadDir(designDir)
	if err != nil || len(designFiles) == 0 {
		issues = append(issues, Issue{
			ID:          "MISSING-DESIGN-DOCS",
			Category:    Documentation,
			Severity:    Medium,
			Title:       "Missing High-Level Design Documents",
			Description: "No design documents were found in 'docs/design/'.",
			Location:    designDir,
			Suggestion:  "Create high-level design documents for major system components.",
			Effort:      MediumEffort,
			Priority:    P2,
		})
	}

	return issues, nil
}

// VerifyExamples checks that example files exist and are valid.
func (d *DocAuditor) VerifyExamples(ctx context.Context, rootPath string, ws *Workspace) ([]Issue, error) {
	exampleDir := filepath.Join(rootPath, "examples")
	if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
		return nil, nil // No examples to verify
	}

	var issues []Issue
	err := filepath.Walk(exampleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}

		platformIssues := d.checkExampleCompilation(ctx, path)
		issues = append(issues, platformIssues...)
		return nil
	})

	return issues, err
}

func (d *DocAuditor) checkExampleCompilation(ctx context.Context, path string) []Issue {
	var issues []Issue
	platforms := []struct{ OS, Arch string }{
		{"linux", "amd64"},
		{"windows", "amd64"},
		{"darwin", "amd64"},
	}

	for _, p := range platforms {
		cmd, err := safeCommand(ctx, "go", "build", "-o", os.DevNull, path)
		if err != nil {
			continue
		}
		cmd.Env = append(os.Environ(), "GOOS="+p.OS, "GOARCH="+p.Arch)
		if output, err := cmd.CombinedOutput(); err != nil {
			issues = append(issues, Issue{
				ID:          fmt.Sprintf("BROKEN-EXAMPLE-%s", strings.ToUpper(p.OS)),
				Category:    Documentation,
				Severity:    High,
				Title:       fmt.Sprintf("Example file fails to compile for %s", p.OS),
				Description: fmt.Sprintf("Example file '%s' failed compilation for %s/%s:\n%s", path, p.OS, p.Arch, string(output)),
				Location:    path,
				Suggestion:  fmt.Sprintf("Fix the platform-specific compilation errors for %s.", p.OS),
				Effort:      Small,
				Priority:    P1,
			})
		}
	}
	return issues
}
