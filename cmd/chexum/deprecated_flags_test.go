package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestDeprecatedFlag_MatchRequired(t *testing.T) {
	// We use the binary built in TestMain
	cmd := exec.Command(binaryName, "--match-required", "nonexistent-file.txt")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run command. It might fail because of nonexistent-file, but we care about stderr.
	_ = cmd.Run()

	output := stderr.String()
	if !strings.Contains(output, "Flag --match-required has been deprecated, use --any-match instead") {
		t.Errorf("Expected deprecation warning in stderr, got: %q", output)
	}
}
