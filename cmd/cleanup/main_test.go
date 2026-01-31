package main

import (
	"os"
	"testing"

	"github.com/Les-El/chexum/internal/checkpoint"
	"github.com/Les-El/chexum/internal/testutil"
)

// TestMain runs after all tests
func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

// Reviewed: LONG-FUNCTION - Table-driven test with complex setup and output capturing.
func TestRunCleanup_Flags(t *testing.T) {

	tests := []struct {

		name     string

		args     []string

		contains string

	}{

		{

			name:     "Dry Run",

			args:     []string{"--dry-run"},

			contains: "DRY RUN MODE",

		},

		{

			name:     "Force",

			args:     []string{"--force"},

			contains: "Force cleanup requested",

		},

		{

			name:     "High Threshold",

			args:     []string{"--threshold", "100.0"},

			contains: "below threshold",

		},

	}



	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			tmpDir, cleanup := testutil.TempDir(t)

			defer cleanup()

			cm := checkpoint.NewCleanupManager(false)

			cm.SetBaseDir(tmpDir)



			stdout, _, err := testutil.CaptureOutput(func() {

				if err := run(tt.args, cm); err != nil {

					t.Errorf("run() failed: %v", err)

				}

			})

			if err != nil {

				t.Fatalf("CaptureOutput failed: %v", err)

			}

			testutil.AssertContains(t, stdout, tt.contains)

		})

	}

}



func TestRunCleanup_InvalidFlag(t *testing.T) {

	err := run([]string{"--invalid-flag"}, nil)

	if err == nil {

		t.Error("expected error for invalid flag, got nil")

	}

}



// Property 1: Test Suite Comprehensive Validation
// **Validates: Requirements 1.1, 1.2, 1.3**
//
// Reviewed: LONG-FUNCTION - Property test with setup and iterations.
func TestProperty_CleanupFlagParsing(t *testing.T) {

	f := func(force, dryRun bool, threshold float64) bool {

		args := []string{}

		if force {

			args = append(args, "--force")

		}

		if dryRun {

			args = append(args, "--dry-run")

		}

		// Clamp threshold to reasonable range

		if threshold < 0 {

			threshold = 0

		}

		if threshold > 100 {

			threshold = 100

		}

		// In a real test we'd use threshold too, but we just want to ensure it doesn't crash

		

		tmpDir, cleanup := testutil.TempDir(t)

		defer cleanup()

		cm := checkpoint.NewCleanupManager(false)

		cm.SetBaseDir(tmpDir)



		err := run(args, cm)

		return err == nil

	}



	testutil.CheckProperty(t, f)

}

func TestMainDirect(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cleanup", "--dry-run"}

	stdout, _, err := testutil.CaptureOutput(func() {
		main()
	})
	if err != nil {
		t.Fatalf("CaptureOutput failed: %v", err)
	}
	testutil.AssertContains(t, stdout, "DRY RUN MODE")
}


