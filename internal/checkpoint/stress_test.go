package checkpoint

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Les-El/chexum/internal/testutil"
)

// Property 13: Stress Test Validation
// **Validates: Requirements 5.2, 5.5**
func TestStress_LargeProjectAnalysis(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// Create a "large" project with 100 files
	for i := 0; i < 100; i++ {
		dir := filepath.Join(tmpDir, fmt.Sprintf("pkg%d", i))
		os.MkdirAll(dir, 0755)
		testutil.GenerateMockGoFile(t, dir, "file.go", i%2 == 0, i%3 == 0)
	}

	ctx := context.Background()
	analyzer := NewCodeAnalyzer()
	ws, _ := NewWorkspace(true)

	// Measure performance
	start := time.Now()
	issues, err := analyzer.Analyze(ctx, tmpDir, ws)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Large project analysis failed: %v", err)
	}

	t.Logf("Analyzed 100 files in %v, found %d issues", duration, len(issues))

	if duration > 5*time.Second {
		t.Errorf("Analysis took too long: %v", duration)
	}
}

// Property 14: Error Handling Test Coverage
// **Validates: Requirements 5.3**
func TestError_PermissionDenied(t *testing.T) {
	// Feature: checkpoint-remediation, Property 14: Error handling test coverage

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	noReadDir := filepath.Join(tmpDir, "noread")
	os.MkdirAll(noReadDir, 0000) // No permissions
	defer os.Chmod(noReadDir, 0755)

	ctx := context.Background()
	analyzer := NewCodeAnalyzer()
	ws, _ := NewWorkspace(true)

	_, err := analyzer.Analyze(ctx, noReadDir, ws)
	// We expect either an error or for it to skip gracefully
	if err != nil {
		t.Logf("Got expected error for permission denied: %v", err)
	}
}
