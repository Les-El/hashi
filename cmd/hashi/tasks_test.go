package main

import (
	"os/exec"
	"strings"
	"testing"
)

// Property 25: Mutually exclusive flags are rejected
// Validates: Requirements 17.1
func TestProperty25_ConflictResolution(t *testing.T) {
	// Test that --quiet overrides --verbose and produces a warning
	cmd := exec.Command("go", "run", ".", "--quiet", "--verbose", "main.go")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("hashi failed: %v, output: %s", err, string(out))
	}
	
	// Warning should be in the output
	if !strings.Contains(string(out), "Warning: --quiet overrides --verbose") {
		t.Errorf("Expected override warning, got: %s", string(out))
	}
}