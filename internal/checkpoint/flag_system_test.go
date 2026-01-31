package checkpoint

import (
	"context"
	"go/ast"
	"go/parser"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewFlagSystem(t *testing.T) {
	fs := NewFlagSystem()
	if fs == nil {
		t.Fatal("NewFlagSystem returned nil")
	}
}

func TestFlagSystem_Name(t *testing.T) {
	fs := NewFlagSystem()
	if name := fs.Name(); name != "FlagSystem" {
		t.Errorf("expected FlagSystem, got %s", name)
	}
}

func TestFlagSystem_Analyze(t *testing.T) {
	fs := NewFlagSystem()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)
	issues, err := fs.Analyze(ctx, "../../", ws)
	if err != nil {
		t.Errorf("Analyze failed: %v", err)
	}
	if len(issues) == 0 {
		t.Log("No issues found, which is expected if everything is clean.")
	}
}

func TestCatalogFlags(t *testing.T) {
	fs := NewFlagSystem()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)
	_, err := fs.CatalogFlags(ctx, "../../", ws)
	if err != nil {
		t.Logf("CatalogFlags failed (possibly config.go missing): %v", err)
	}
}

func TestClassifyImplementation(t *testing.T) {
	fs := NewFlagSystem()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)
	mockFlags := []FlagStatus{{LongForm: "verbose", DefinedInCode: true}}
	if _, err := os.Stat("../../internal/config/config.go"); err == nil {
		_, err = fs.ClassifyImplementation(ctx, "../../", ws, mockFlags)
		if err != nil {
			t.Errorf("ClassifyImplementation failed: %v", err)
		}
	}
}

func TestPerformCrossReferenceAnalysis(t *testing.T) {
	fs := NewFlagSystem()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)
	mockFlags := []FlagStatus{{LongForm: "verbose", DefinedInCode: true}}
	_, err := fs.PerformCrossReferenceAnalysis(ctx, "../../", ws, mockFlags)
	if err != nil {
		t.Errorf("PerformCrossReferenceAnalysis failed: %v", err)
	}
}

func TestDetectConflicts(t *testing.T) {
	fs := NewFlagSystem()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)
	mockFlags := []FlagStatus{{LongForm: "verbose", DefinedInCode: true}}
	_, err := fs.DetectConflicts(ctx, ws, mockFlags)
	if err != nil {
		t.Errorf("DetectConflicts failed: %v", err)
	}
}

func TestValidateFunctionality(t *testing.T) {
	fs := NewFlagSystem()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)
	mockFlags := []FlagStatus{{LongForm: "verbose", DefinedInCode: true}}
	_, err := fs.ValidateFunctionality(ctx, ws, mockFlags)
	if err != nil {
		t.Errorf("ValidateFunctionality failed: %v", err)
	}
}

func TestGenerateStatusReport(t *testing.T) {
	fs := NewFlagSystem()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)
	mockFlags := []FlagStatus{
		{LongForm: "verbose", DefinedInCode: true, Status: FullyImplemented},
		{LongForm: "hidden", DefinedInCode: true, Status: PartiallyImplemented},
		{LongForm: "recursive", DefinedInCode: true, Status: NeedsRepair},
	}
	report, err := fs.GenerateStatusReport(ctx, ws, mockFlags)
	if err != nil {
		t.Errorf("GenerateStatusReport failed: %v", err)
	}
	if report == "" {
		t.Error("GenerateStatusReport returned empty string")
	}
}

func TestFlagSystem_ParseFlagCall(t *testing.T) {
	fs := NewFlagSystem()

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"BoolVarP", `fs.BoolVarP(&v, "test-flag", "t", false, "usage")`, "test-flag"},
		{"StringVar", `fs.StringVar(&v, "test-flag", "def", "usage")`, "test-flag"},
	}

	for _, tc := range tests {
		expr, _ := parser.ParseExpr(tc.code)
		call := expr.(*ast.CallExpr)
		status, ok := fs.parseFlagCall(call)
		if !ok || status.LongForm != tc.expected {
			t.Errorf("%s: expected %s, got %s", tc.name, tc.expected, status.LongForm)
		}
	}
}

func TestFlagSystem_ExtractPotentialFlags(t *testing.T) {
	fs := NewFlagSystem()
	content := "some text with --flag-one and --flag-2"
	flags := fs.extractPotentialFlags(content)
	if len(flags) != 2 {
		t.Errorf("expected 2 flags, got %d", len(flags))
	}
}

func TestClassifyImplementation_SpecialCases(t *testing.T) {
	fs := NewFlagSystem()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	tmpDir := t.TempDir()
	setupSpecialCasesMockEnv(t, tmpDir)

	flags := []FlagStatus{
		{LongForm: "json"}, {LongForm: "jsonl"}, {LongForm: "help"},
		{LongForm: "version"}, {LongForm: "config"}, {LongForm: "log-json"},
		{LongForm: "format"}, {LongForm: "csv"},
	}

	classified, err := fs.ClassifyImplementation(ctx, tmpDir, ws, flags)
	if err != nil {
		t.Fatalf("ClassifyImplementation failed: %v", err)
	}

	for _, f := range classified {
		t.Run(f.LongForm, func(t *testing.T) {
			if f.Status != FullyImplemented {
				t.Errorf("Flag --%s expected FullyImplemented, got %s", f.LongForm, f.Status)
			}
		})
	}
}

func setupSpecialCasesMockEnv(t *testing.T, tmpDir string) {
	configDir := filepath.Join(tmpDir, "internal", "config")
	os.MkdirAll(configDir, 0755)

	configContent := `package config
type Config struct {
    JSON bool; JSONL bool; ShowHelp bool; ShowVersion bool
    ConfigFile string; LogJSON string; OutputFormat string; CSV bool
}
func (cfg *Config) Dummy() {
    _ = cfg.JSON; _ = cfg.JSONL; _ = cfg.ShowHelp; _ = cfg.ShowVersion
    _ = cfg.ConfigFile; _ = cfg.LogJSON; _ = cfg.OutputFormat; _ = cfg.CSV
}
`
	os.WriteFile(filepath.Join(configDir, "config.go"), []byte(configContent), 0644)

	mainDir := filepath.Join(tmpDir, "cmd", "chexum")
	os.MkdirAll(mainDir, 0755)
	mainContent := `package main
func main() {
    var cfg config.Config
    _ = cfg.JSON; _ = cfg.JSONL; _ = cfg.ShowHelp; _ = cfg.ShowVersion
    _ = cfg.ConfigFile; _ = cfg.LogJSON; _ = cfg.OutputFormat; _ = cfg.CSV
}
`
	os.WriteFile(filepath.Join(mainDir, "main.go"), []byte(mainContent), 0644)
}

func TestFlagSystem_ReadFilesCombined(t *testing.T) {
	fs := NewFlagSystem()
	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "f1.txt")
	os.WriteFile(f1, []byte("content1"), 0644)
	content := fs.readFilesCombined([]string{f1})
	if !strings.Contains(content, "content1") {
		t.Error("expected content1 in combined output")
	}
}

func TestFlagSystem_ReportConflictIssues(t *testing.T) {
	fs := NewFlagSystem()
	flag := FlagStatus{
		LongForm: "test",
		ConflictDetails: []FlagConflict{
			{Type: OrphanedFlag, Severity: ConflictHigh, Description: "desc"},
		},
	}
	issues := fs.reportConflictIssues(".", "internal/config", flag)
	if len(issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(issues))
	}
}
