package config

import (
	"fmt"
	"testing"
)

func TestWriteError(t *testing.T) {
	err := WriteError()
	if err.Error() != "Unknown write/append error" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestConfigCommandError(t *testing.T) {
	err := &ConfigCommandError{}
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
	if err.ExitCode() != ExitInvalidArgs {
		t.Errorf("Expected exit code %d, got %d", ExitInvalidArgs, err.ExitCode())
	}
}

func TestError(t *testing.T)    { TestConfigCommandError(t) }
func TestExitCode(t *testing.T) { TestConfigCommandError(t) }

func TestHandleFileWriteError(t *testing.T) {
	tests := []struct {
		err      error
		verbose  bool
		expected string
	}{
		{fmt.Errorf("permission denied"), false, "Unknown write/append error"},
		{fmt.Errorf("permission denied"), true, "permission denied writing to path"},
		{fmt.Errorf("disk full"), true, "insufficient disk space for path"},
		{fmt.Errorf("network timeout"), true, "network error writing to path"},
		{fmt.Errorf("file name too long"), true, "path too long: path"},
		{fmt.Errorf("other"), true, "other"},
		{nil, true, ""},
	}

	for _, tt := range tests {
		got := HandleFileWriteError(tt.err, tt.verbose, "path")
		if tt.err == nil {
			if got != nil {
				t.Errorf("expected nil, got %v", got)
			}
			continue
		}
		if got.Error() != tt.expected {
			t.Errorf("HandleFileWriteError(%v, %v) = %v; want %v", tt.err, tt.verbose, got, tt.expected)
		}
	}
}

func TestWriteErrorWithVerbose(t *testing.T) {
	if WriteErrorWithVerbose(false, "secret").Error() != "Unknown write/append error" {
		t.Error("expected obfuscated error")
	}
	if WriteErrorWithVerbose(true, "secret").Error() != "secret" {
		t.Error("expected verbose error")
	}
}

func TestFileSystemError(t *testing.T) {
	if FileSystemError(false, "secret").Error() != "Unknown write/append error" {
		t.Error("expected obfuscated error")
	}
	if FileSystemError(true, "secret").Error() != "secret" {
		t.Error("expected verbose error")
	}
}
