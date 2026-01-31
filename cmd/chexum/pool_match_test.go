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

func TestPoolMatchingMode_Integration(t *testing.T) {
	colorHandler := color.NewColorHandler()
	errHandler := errors.NewErrorHandler(colorHandler)

	tmpDir, err := os.MkdirTemp("", "pool-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	testHash := "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2"

	t.Run("BasicPoolMatch", func(t *testing.T) {
		cfg := &config.Config{Files: []string{testFile}, Hashes: []string{testHash}, Algorithm: "sha256"}
		out, _ := runPoolTest(t, cfg, colorHandler, errHandler)

		if !strings.Contains(out, "testfile.txt") || !strings.Contains(out, "REFERENCE:") || !strings.Contains(out, testHash) {
			t.Errorf("Output missing expected content: %s", out)
		}
	})

	t.Run("OrphanedHash", func(t *testing.T) {
		orphanHash := "0000000000000000000000000000000000000000000000000000000000000000"
		cfg := &config.Config{Files: []string{testFile}, Hashes: []string{orphanHash}, Algorithm: "sha256"}
		out, _ := runPoolTest(t, cfg, colorHandler, errHandler)

		if !strings.Contains(out, "REFERENCE:") {
			t.Errorf("Expected REFERENCE: in output, got: %s", out)
		}
	})
}

func runPoolTest(t *testing.T, cfg *config.Config, ch *color.Handler, eh *errors.Handler) (string, int) {
	var outBuf, errBuf bytes.Buffer
	streams := &console.Streams{Out: &outBuf, Err: &errBuf}
	exitCode := runStandardHashingMode(cfg, ch, streams, eh)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Err: %s", exitCode, errBuf.String())
	}
	return outBuf.String(), exitCode
}
