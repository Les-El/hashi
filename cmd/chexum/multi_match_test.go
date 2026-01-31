package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Les-El/chexum/internal/color"
	"github.com/Les-El/chexum/internal/config"
	"github.com/Les-El/chexum/internal/console"
	"github.com/Les-El/chexum/internal/errors"
)

// Reviewed: LONG-FUNCTION - Data-driven test with multiple scenarios.
func TestMultiMatchMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "chexum-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	files := setupMultiMatchFiles(t, tmpDir)

	// Hashes for "content1" and "content2" (SHA256)
	hash1 := "d0b425e00e15a0d36b9b361f02bab63563aed6cb4665083905386c55d5b679fa" // content1
	hash2 := "dab741b6289e7dccc1ed42330cae1accc2b755ce8079c2cd5d4b5366c9f769a6" // content2

	tests := []struct {
		name           string
		args           []string
		expectedStatus int
		expectedOutput []string
	}{
		{
			name:           "10 files + 1 hash (one match)",
			args:           []string{files[0], files[1], hash1},
			expectedStatus: config.ExitSuccess,
			expectedOutput: []string{files[0], "REFERENCE:    " + hash1},
		},
		{
			name:           "1 file + 2 hashes (one match)",
			args:           []string{files[0], hash1, hash2},
			expectedStatus: config.ExitSuccess,
			expectedOutput: []string{files[0], "REFERENCE:    " + hash1, "REFERENCE:    " + hash2},
		},
		{
			name:           "Any Match Flag (success)",
			args:           []string{"--any-match", files[0], files[1], hash1},
			expectedStatus: config.ExitSuccess,
		},
		{
			name:           "Any Match Flag (failure)",
			args:           []string{"--any-match", files[0], files[1], "0000000000000000000000000000000000000000000000000000000000000000"},
			expectedStatus: config.ExitNoMatches,
		},
		{
			name:           "All Match Flag (failure)",
			args:           []string{"--all-match", files[0], files[1], hash1},
			expectedStatus: config.ExitNoMatches,
		},
		{
			name:           "All Match Flag (success)",
			args:           []string{"--all-match", files[0], files[2], hash1},
			expectedStatus: config.ExitSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runMultiMatchSubtest(t, tt.args, tt.expectedStatus, tt.expectedOutput)
		})
	}
}

func setupMultiMatchFiles(t *testing.T, tmpDir string) []string {
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	file3 := filepath.Join(tmpDir, "file3.txt")

	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}
	if err := os.WriteFile(file3, []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to write file3: %v", err)
	}
	return []string{file1, file2, file3}
}

func runMultiMatchSubtest(t *testing.T, args []string, expectedStatus int, expectedOutput []string) {
	cfg, _, err := config.ParseArgs(args)
	if err != nil {
		t.Fatalf("failed to parse args: %v", err)
	}

	var outBuf, errBuf bytes.Buffer
	streams := &console.Streams{
		Out: &outBuf,
		Err: &errBuf,
	}
	colorHandler := color.NewColorHandler()
	errHandler := errors.NewErrorHandler(colorHandler)

	if err := prepareFiles(cfg, errHandler, streams); err != nil {
		t.Fatalf("failed to prepare files: %v", err)
	}

	status := executeMode(cfg, colorHandler, streams, errHandler)

	if status != expectedStatus {
		t.Errorf("expected status %d, got %d", expectedStatus, status)
	}

	output := outBuf.String()
	for _, expected := range expectedOutput {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, but got %q", expected, output)
		}
	}
}
