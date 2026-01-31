package checkpoint

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewCodeAnalyzer(t *testing.T) {
	analyzer := NewCodeAnalyzer()
	if analyzer == nil {
		t.Fatal("NewCodeAnalyzer returned nil")
	}
	if analyzer.fset == nil {
		t.Error("analyzer.fset is nil")
	}
}

func TestCodeAnalyzer_Name(t *testing.T) {
	analyzer := NewCodeAnalyzer()
	if name := analyzer.Name(); name != "CodeAnalyzer" {
		t.Errorf("expected CodeAnalyzer, got %s", name)
	}
}

func TestCodeAnalyzer_Analyze(t *testing.T) {
	analyzer := NewCodeAnalyzer()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	tmpDir := t.TempDir()
	setupCodeAnalyzerTestFile(t, tmpDir)

	issues, err := analyzer.Analyze(ctx, tmpDir, ws)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	expectedIDs := []string{
		"TECH-DEBT",
		"SECURITY-UNSAFE",
		"SECURITY-PROCESS-EXEC",
		"SECURITY-NET-LISTEN",
		"SECURITY-DIRECT-SYSCALL",
	}

	for _, id := range expectedIDs {
		t.Run(id, func(t *testing.T) {
			found := false
			for _, issue := range issues {
				if issue.ID == id {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected to find %s issue", id)
			}
		})
	}
}

func setupCodeAnalyzerTestFile(t *testing.T, tmpDir string) {
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test
import "unsafe"
import "os/exec"
import "net"
import "syscall"

// TODO: implement this
func foo() {
    exec.Command("ls")
    net.Listen("tcp", ":8080")
    syscall.Syscall(0, 0, 0, 0)
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
}

func TestAnalyzePackages(t *testing.T) {
	analyzer := NewCodeAnalyzer()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	issues, err := analyzer.AnalyzePackages(ctx, "../../", ws)
	if err != nil {
		t.Errorf("AnalyzePackages failed: %v", err)
	}
	_ = issues
}

func TestCheckSecurity(t *testing.T) {
	analyzer := NewCodeAnalyzer()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	issues, err := analyzer.CheckSecurity(ctx, "../../", ws)
	if err != nil {
		t.Errorf("CheckSecurity failed: %v", err)
	}
	_ = issues
}

func TestCodeAnalyzer_AssessDependencies(t *testing.T) {
	analyzer := NewCodeAnalyzer()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	issues, err := analyzer.AssessDependencies(ctx, "../../", ws)
	if err != nil {
		t.Errorf("AssessDependencies failed: %v", err)
	}
	_ = issues
}

func TestIdentifyTechnicalDebt(t *testing.T) {
	analyzer := NewCodeAnalyzer()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	// This calls Analyze(ctx, rootPath), so it should find something in the current dir if there are technical debt markers.
	// We just want to make sure it doesn't crash and returns something.
	_, err := analyzer.IdentifyTechnicalDebt(ctx, "../../", ws)
	if err != nil {
		t.Errorf("IdentifyTechnicalDebt failed: %v", err)
	}
}

func TestCodeAnalyzer_Mappers(t *testing.T) {
	if mapGosecSeverity("HIGH") != High {
		t.Error("expected High")
	}
	if mapGosecSeverity("MEDIUM") != Medium {
		t.Error("expected Medium")
	}
	if mapGosecSeverity("LOW") != Low {
		t.Error("expected Low")
	}
	if mapGosecSeverity("INFO") != Info {
		t.Error("expected Info")
	}
	if mapGosecSeverity("UNKNOWN") != Info {
		t.Error("expected Info for unknown")
	}

	if mapGosecPriority("HIGH") != P1 {
		t.Error("expected P1")
	}
	if mapGosecPriority("MEDIUM") != P2 {
		t.Error("expected P2")
	}
	if mapGosecPriority("LOW") != P3 {
		t.Error("expected P3")
	}
	if mapGosecPriority("UNKNOWN") != P3 {
		t.Error("expected P3 for unknown")
	}
}
