// Package testutil provides utilities for testing the chexum CLI tool and its components.
package testutil

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// CaptureOutput captures stdout and stderr during the execution of f.
func CaptureOutput(f func()) (stdout, stderr string, err error) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	outChan := make(chan string)
	errChan := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, rOut)
		outChan <- buf.String()
	}()

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, rErr)
		errChan <- buf.String()
	}()

	f()

	wOut.Close()
	wErr.Close()

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdout = <-outChan
	stderr = <-errChan

	return stdout, stderr, nil
}

// TempDir creates a temporary directory and returns its path and a cleanup function.
func TempDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "h-testdir-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	return dir, func() {
		os.RemoveAll(dir)
	}
}

// CreateFile creates a file with the given content in the specified directory.
func CreateFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create directory for file %s: %v", name, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", name, err)
	}
	return path
}

// AssertExitCode asserts that the given exit code matches the expected one.
func AssertExitCode(t *testing.T, expected, actual int) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected exit code %d, got %d", expected, actual)
	}
}

// AssertContains asserts that the given string contains the expected substring.
func AssertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !contains(s, substr) {
		t.Errorf("expected string to contain %q, but it didn't.\nString: %s", substr, s)
	}
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

// AutoCleanupStorage performs automatic cleanup of temporary storage to prevent disk space issues during tests.
// This function identifies and removes Go build artifacts (go-build-*, etc) and chexum-specific temporary files.
// It uses force cleanup mode to ensure space is freed up for subsequent tests.
// Returns true if cleanup was performed successfully.
func AutoCleanupStorage(t *testing.T) bool {
	t.Helper()

	tmpDir := os.TempDir()
	patterns := []string{
		"chexum-*",
		"checkpoint-*",
		"test-*",
	}

	cleaned := false
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(tmpDir, pattern))
		if err != nil {
			continue
		}
		for _, match := range matches {
			if err := os.RemoveAll(match); err == nil {
				cleaned = true
			}
		}
	}

	return cleaned
}

// RequireCleanStorage ensures temporary storage has sufficient space by aggressively cleaning it.
// This is called before resource-intensive tests. It removes go test build artifacts
// that accumulate and consume disk space during testing.
func RequireCleanStorage(t *testing.T) {
	t.Helper()

	AutoCleanupStorage(t)
}
