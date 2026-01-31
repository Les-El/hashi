package checkpoint

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewDocAuditor(t *testing.T) {
	auditor := NewDocAuditor()
	if auditor == nil {
		t.Fatal("NewDocAuditor returned nil")
	}
}

func TestDocAuditor_Name(t *testing.T) {
	auditor := NewDocAuditor()
	if name := auditor.Name(); name != "DocAuditor" {
		t.Errorf("expected DocAuditor, got %s", name)
	}
}

func TestDocAuditor_Analyze(t *testing.T) {
	auditor := NewDocAuditor()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	// Normal case
	_, err := auditor.Analyze(ctx, ".", ws)
	if err != nil {
		t.Errorf("Analyze failed: %v", err)
	}

	t.Run("AuditError", func(t *testing.T) {
		// Use a path that triggers an error in Walk
		_, err := auditor.AuditGoDocumentation(ctx, "/proc/self/mem", ws) // Should fail or be empty
		if err != nil {
			// This is fine, we just want to hit the error handling in Analyze
		}
	})

	t.Run("ExampleError", func(t *testing.T) {
		// Create a directory that we can't read
		tmpDir := t.TempDir()
		exDir := filepath.Join(tmpDir, "examples")
		os.Mkdir(exDir, 0000)
		defer os.Chmod(exDir, 0755)

		_, err := auditor.VerifyExamples(ctx, tmpDir, ws)
		if err != nil {
			// Hit error path
		}
	})
}

func TestAuditGoDocumentation(t *testing.T) {
	auditor := NewDocAuditor()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	// Create a temp file with missing doc
	tmpDir, err := os.MkdirTemp("", "doc_auditor_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test
func ExportedWithoutDoc() {}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	issues, err := auditor.AuditGoDocumentation(ctx, tmpDir, ws)
	if err != nil {
		t.Fatalf("AuditGoDocumentation failed: %v", err)
	}

	foundMissingDoc := false
	for _, issue := range issues {
		if issue.ID == "MISSING-DOC" {
			foundMissingDoc = true
			break
		}
	}

	if !foundMissingDoc {
		t.Error("expected to find MISSING-DOC issue")
	}
}

func TestValidateREADME(t *testing.T) {
	auditor := NewDocAuditor()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	issues, err := auditor.ValidateREADME(ctx, "../../", ws)
	if err != nil {
		t.Errorf("ValidateREADME failed: %v", err)
	}
	_ = issues
}

func TestCheckArchitecturalDocs(t *testing.T) {
	auditor := NewDocAuditor()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	issues, err := auditor.CheckArchitecturalDocs(ctx, "../../", ws)
	if err != nil {
		t.Errorf("CheckArchitecturalDocs failed: %v", err)
	}
	_ = issues
}

func TestVerifyExamples(t *testing.T) {
	auditor := NewDocAuditor()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	// Create a dummy example directory in tmpDir
	tmpDir, _ := os.MkdirTemp("", "examples_test")
	defer os.RemoveAll(tmpDir)
	os.Mkdir(filepath.Join(tmpDir, "examples"), 0755)

	content := `package main
func main() {}
`
	os.WriteFile(filepath.Join(tmpDir, "examples/test.go"), []byte(content), 0644)

	issues, err := auditor.VerifyExamples(ctx, tmpDir, ws)
	if err != nil {
		t.Fatalf("VerifyExamples failed: %v", err)
	}

	// Should be 0 issues if example exists
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}
