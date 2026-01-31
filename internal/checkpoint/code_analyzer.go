package checkpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// CodeAnalyzer implements CodebaseAnalyzer.
type CodeAnalyzer struct {
	fset *token.FileSet
}

// NewCodeAnalyzer creates a new CodeAnalyzer.
func NewCodeAnalyzer() *CodeAnalyzer {
	return &CodeAnalyzer{
		fset: token.NewFileSet(),
	}
}

// Name returns the name of the analyzer.
func (c *CodeAnalyzer) Name() string { return "CodeAnalyzer" }

// Analyze implements AnalysisEngine.
func (c *CodeAnalyzer) Analyze(ctx context.Context, path string, ws *Workspace) ([]Issue, error) {
	var issues []Issue

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
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

		fileIssues, err := c.analyzeFile(filePath)
		if err != nil {
			issues = append(issues, Issue{
				ID:          "PARSE-ERROR",
				Category:    CodeQuality,
				Severity:    High,
				Title:       "Failed to parse Go file",
				Description: fmt.Sprintf("Parser error in %s: %v", filePath, err),
				Location:    filePath,
				Suggestion:  "Ensure the file is valid Go code and matches the configured Go version.",
				Effort:      Small,
				Priority:    P1,
			})
			return nil
		}
		issues = append(issues, fileIssues...)
		return nil
	})

	// Run gosec if available
	gosecIssues, _ := c.runGosec(ctx, path)
	issues = append(issues, gosecIssues...)

	return issues, err
}

type gosecOutput struct {
	Issues []gosecIssue `json:"Issues"`
}

type gosecIssue struct {
	Severity   string `json:"severity"`
	Confidence string `json:"confidence"`
	RuleID     string `json:"rule_id"`
	Details    string `json:"details"`
	File       string `json:"file"`
	Line       string `json:"line"`
	Column     string `json:"column"`
}

func (c *CodeAnalyzer) runGosec(ctx context.Context, path string) ([]Issue, error) {
	cmd, err := safeCommand(ctx, "gosec", "-fmt=json", "./...")
	if err != nil {
		return nil, nil // skip if gosec not found or not allowed
	}

	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		// gosec returns non-zero if issues found, so we check output even if error
		if len(output) == 0 {
			return nil, fmt.Errorf("gosec failed: %w", err)
		}
	}

	if len(output) == 0 {
		return nil, nil
	}

	var outputData gosecOutput
	if err := json.Unmarshal(output, &outputData); err != nil {
		return nil, fmt.Errorf("failed to parse gosec output: %w", err)
	}

	var issues []Issue
	for _, gi := range outputData.Issues {
		issues = append(issues, Issue{
			ID:          gi.RuleID,
			Category:    Security,
			Severity:    mapGosecSeverity(gi.Severity),
			Title:       gi.Details,
			Description: fmt.Sprintf("gosec identified a security risk: %s (Confidence: %s)", gi.Details, gi.Confidence),
			Location:    fmt.Sprintf("%s:%s", gi.File, gi.Line),
			Suggestion:  "Review the gosec documentation for this rule and apply the recommended fix.",
			Effort:      MediumEffort,
			Priority:    mapGosecPriority(gi.Severity),
			Status:      Pending,
		})
	}

	return issues, nil
}

func mapGosecSeverity(s string) Severity {
	switch strings.ToUpper(s) {
	case "HIGH":
		return High
	case "MEDIUM":
		return Medium
	case "LOW":
		return Low
	default:
		return Info
	}
}

func mapGosecPriority(s string) Priority {
	switch strings.ToUpper(s) {
	case "HIGH":
		return P1
	case "MEDIUM":
		return P2
	case "LOW":
		return P3
	default:
		return P3
	}
}

func (c *CodeAnalyzer) analyzeFile(filePath string) ([]Issue, error) {
	var issues []Issue

	f, err := parser.ParseFile(c.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	issues = append(issues, c.scanTechnicalDebt(f)...)
	issues = append(issues, c.scanSecurityIssues(f)...)

	return issues, nil
}

func (c *CodeAnalyzer) scanTechnicalDebt(f *ast.File) []Issue {
	var issues []Issue
	// Scan for technical debt markers (such as debt tags)
	for _, commentGroup := range f.Comments {
		for _, comment := range commentGroup.List {
			text := comment.Text
			if strings.Contains(text, "TODO") || strings.Contains(text, "FIXME") {

				issues = append(issues, Issue{
					ID:          "TECH-DEBT",
					Category:    CodeQuality,
					Severity:    Info,
					Title:       "Technical Debt identified",
					Description: fmt.Sprintf("Found TODO/FIXME in comment: %s", text),
					Location:    c.fset.Position(comment.Pos()).String(),
					Suggestion:  "Address the TODO or FIXME comment.",
					Effort:      Small,
					Priority:    P3,
				})
			}
		}
	}
	return issues
}

func (c *CodeAnalyzer) scanSecurityIssues(f *ast.File) []Issue {
	var issues []Issue

	reviewed := c.extractReviewedIssues(f)
	importMap := c.extractImports(f)

	ast.Inspect(f, func(n ast.Node) bool {
		// Check for unsafe import (even if aliased)
		if imp, ok := n.(*ast.ImportSpec); ok {
			if issue := c.checkUnsafeImport(imp, reviewed); issue != nil {
				issues = append(issues, *issue)
			}
		}

		// Check for dangerous function calls
		if call, ok := n.(*ast.CallExpr); ok {
			issues = append(issues, c.checkDangerousCall(call, importMap, reviewed)...)
		}
		return true
	})
	return issues
}

func (c *CodeAnalyzer) extractReviewedIssues(f *ast.File) map[string]bool {
	reviewed := make(map[string]bool)
	for _, cg := range f.Comments {
		for _, comment := range cg.List {
			if strings.Contains(comment.Text, "Reviewed: ") {
				parts := strings.Split(comment.Text, "Reviewed: ")
				if len(parts) > 1 {
					issueID := strings.Fields(parts[1])[0]
					reviewed[issueID] = true
				}
			}
		}
	}
	return reviewed
}

func (c *CodeAnalyzer) extractImports(f *ast.File) map[string]string {
	importMap := make(map[string]string)
	for _, imp := range f.Imports {
		pkgPath := strings.Trim(imp.Path.Value, "\"")
		pkgName := filepath.Base(pkgPath)
		if imp.Name != nil {
			pkgName = imp.Name.Name
		}
		importMap[pkgName] = pkgPath
	}
	return importMap
}

func (c *CodeAnalyzer) checkUnsafeImport(imp *ast.ImportSpec, reviewed map[string]bool) *Issue {
	if imp.Path != nil && imp.Path.Value == "\"unsafe\"" {
		if !reviewed["SECURITY-UNSAFE"] {
			return &Issue{
				ID:          "SECURITY-UNSAFE",
				Category:    Security,
				Severity:    Medium,
				Title:       "Usage of 'unsafe' package",
				Description: "The 'unsafe' package is used in this file. Unsafe pointer manipulation can lead to memory corruption.",
				Location:    c.fset.Position(imp.Pos()).String(),
				Suggestion:  "Verify if 'unsafe' is absolutely necessary. Prefer safe Go alternatives.",
				Effort:      MediumEffort,
				Priority:    P2,
				Status:      Pending,
			}
		}
	}
	return nil
}

func (c *CodeAnalyzer) checkDangerousCall(call *ast.CallExpr, importMap map[string]string, reviewed map[string]bool) []Issue {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	pkgAlias := ""
	if x, ok := sel.X.(*ast.Ident); ok {
		pkgAlias = x.Name
	}

	realPkg := importMap[pkgAlias]
	funcName := sel.Sel.Name

	var issues []Issue
	issues = append(issues, c.checkCommandExecution(call, realPkg, funcName, reviewed)...)
	issues = append(issues, c.checkNetworkListeners(call, realPkg, funcName, reviewed)...)
	issues = append(issues, c.checkDirectSyscalls(call, realPkg, funcName, reviewed)...)

	return issues
}

func (c *CodeAnalyzer) checkCommandExecution(call *ast.CallExpr, realPkg, funcName string, reviewed map[string]bool) []Issue {
	if (realPkg == "os/exec" && (funcName == "Command" || funcName == "CommandContext")) ||
		(realPkg == "os" && funcName == "StartProcess") ||
		(realPkg == "syscall" && (funcName == "Exec" || funcName == "ForkExec")) {
		if !reviewed["SECURITY-PROCESS-EXEC"] {
			return []Issue{{
				ID:          "SECURITY-PROCESS-EXEC",
				Category:    Security,
				Severity:    High,
				Title:       fmt.Sprintf("External process execution via %s.%s", realPkg, funcName),
				Description: "Executing external processes can lead to command injection if input is not sanitized.",
				Location:    c.fset.Position(call.Pos()).String(),
				Suggestion:  "Avoid external processes if possible. If necessary, use hardcoded paths and sanitized arguments.",
				Effort:      MediumEffort,
				Priority:    P1,
				Status:      Pending,
			}}
		}
	}
	return nil
}

func (c *CodeAnalyzer) checkNetworkListeners(call *ast.CallExpr, realPkg, funcName string, reviewed map[string]bool) []Issue {
	if realPkg == "net" && (strings.HasPrefix(funcName, "Listen")) {
		if !reviewed["SECURITY-NET-LISTEN"] {
			return []Issue{{
				ID:          "SECURITY-NET-LISTEN",
				Category:    Security,
				Severity:    Medium,
				Title:       fmt.Sprintf("Network listener created via %s.%s", realPkg, funcName),
				Description: "Opening network ports can expose the application to remote attacks.",
				Location:    c.fset.Position(call.Pos()).String(),
				Suggestion:  "Ensure the listener is bound to localhost unless external access is required.",
				Effort:      MediumEffort,
				Priority:    P2,
				Status:      Pending,
			}}
		}
	}
	return nil
}

func (c *CodeAnalyzer) checkDirectSyscalls(call *ast.CallExpr, realPkg, funcName string, reviewed map[string]bool) []Issue {
	if realPkg == "syscall" && strings.HasPrefix(funcName, "Syscall") {
		if !reviewed["SECURITY-DIRECT-SYSCALL"] {
			return []Issue{{
				ID:          "SECURITY-DIRECT-SYSCALL",
				Category:    Security,
				Severity:    Critical,
				Title:       "Direct Syscall usage",
				Description: "Direct system calls bypass Go's safety abstractions.",
				Location:    c.fset.Position(call.Pos()).String(),
				Suggestion:  "Use higher-level abstractions in 'os' or 'net' packages.",
				Effort:      Large,
				Priority:    P0,
				Status:      Pending,
			}}
		}
	}
	return nil
}



// AnalyzePackages performs analysis at the package level.
func (c *CodeAnalyzer) AnalyzePackages(ctx context.Context, path string, ws *Workspace) ([]Issue, error) {
	return c.Analyze(ctx, path, ws)
}

// CheckSecurity performs security-specific analysis.
func (c *CodeAnalyzer) CheckSecurity(ctx context.Context, path string, ws *Workspace) ([]Issue, error) {
	return c.Analyze(ctx, path, ws)
}

// AssessDependencies evaluates the project's dependencies.
func (c *CodeAnalyzer) AssessDependencies(ctx context.Context, path string, ws *Workspace) ([]Issue, error) {
	return c.Analyze(ctx, path, ws)
}

// IdentifyTechnicalDebt finds technical debt markers in the codebase.
func (c *CodeAnalyzer) IdentifyTechnicalDebt(ctx context.Context, path string, ws *Workspace) ([]Issue, error) {
	return c.Analyze(ctx, path, ws)
}
