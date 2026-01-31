package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCaptureOutput(t *testing.T) {
	stdout, stderr, err := CaptureOutput(func() {
		fmt.Print("hello stdout")
		fmt.Fprint(os.Stderr, "hello stderr")
	})

	if err != nil {
		t.Fatalf("CaptureOutput failed: %v", err)
	}

	if stdout != "hello stdout" {
		t.Errorf("expected stdout %q, got %q", "hello stdout", stdout)
	}

	if stderr != "hello stderr" {
		t.Errorf("expected stderr %q, got %q", "hello stderr", stderr)
	}
}

func TestTempDir(t *testing.T) {
	dir, cleanup := TempDir(t)
	defer cleanup()

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("expected temp dir %s to exist", dir)
	}

	if !filepath.HasPrefix(filepath.Base(dir), "h-testdir-") {
		t.Errorf("expected temp dir to have prefix h-testdir-, got %s", filepath.Base(dir))
	}

	cleanup()
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("expected temp dir %s to be removed after cleanup", dir)
	}
}

func TestCreateFile(t *testing.T) {
	dir, cleanup := TempDir(t)
	defer cleanup()

	path := CreateFile(t, dir, "test.txt", "hello world")

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if string(content) != "hello world" {
		t.Errorf("expected content %q, got %q", "hello world", string(content))
	}
}

func TestAssertExitCode(t *testing.T) {
	// This just checks if it doesn't fail when codes match
	// Since it uses t.Errorf, we can't easily test the failure case without a mock testing.T
	AssertExitCode(t, 0, 0)
}

func TestAssertContains(t *testing.T) {
	AssertContains(t, "hello world", "world")
}

func TestAutoCleanupStorage(t *testing.T) {
	// Create a file that matches the pattern
	path := filepath.Join(os.TempDir(), "chexum-cleanup-test")
	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file in /tmp: %v", err)
	}

	cleaned := AutoCleanupStorage(t)
	if !cleaned {
		t.Errorf("expected AutoCleanupStorage to return true (cleaned something)")
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected file %s to be removed", path)
	}
}

func TestRequireCleanStorage(t *testing.T) {
	// Smoke test
	RequireCleanStorage(t)
}
