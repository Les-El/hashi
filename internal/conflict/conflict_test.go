// Package conflict tests for the "Pipeline of Intent" logic.
package conflict_test

import (
	"testing"

	"github.com/Les-El/chexum/internal/conflict"
)

// TestConflictResolution_PipelineOfIntent tests the conflict resolution logic.
// **Validates: Pipeline of Intent (Conflict Resolution)**
// Reviewed: Kept long function because it is a comprehensive table-driven test for conflict resolution states.
var conflictResolutionTests = []struct {
	name           string
	args           []string
	flagSet        map[string]bool
	explicitFormat string
	expectedState  conflict.RunState
	hasWarning     bool
}{
	{
		name: "Default State",
		args: []string{},
		flagSet: map[string]bool{
			"json": false, "plain": false, "quiet": false, "verbose": false, "bool": false,
		},
		expectedState: conflict.RunState{
			Mode:      conflict.ModeStandard,
			Format:    conflict.FormatDefault,
			Verbosity: conflict.VerbosityNormal,
		},
	},
	{
		name: "Quiet overrides Verbose",
		args: []string{"-q", "-v"},
		flagSet: map[string]bool{
			"quiet": true, "verbose": true,
		},
		expectedState: conflict.RunState{
			Mode:      conflict.ModeStandard,
			Format:    conflict.FormatDefault,
			Verbosity: conflict.VerbosityQuiet,
		},
		hasWarning: true,
	},
	{
		name: "Bool overrides Format and implies Quiet",
		args: []string{"--bool", "--json"},
		flagSet: map[string]bool{
			"bool": true, "json": true,
		},
		expectedState: conflict.RunState{
			Mode:      conflict.ModeBool,
			Format:    conflict.FormatDefault, // Reset to default when in bool mode
			Verbosity: conflict.VerbosityQuiet,
		},
		hasWarning: true,
	},
	{
		name: "JSONL support",
		args: []string{"--jsonl"},
		flagSet: map[string]bool{
			"jsonl": true,
		},
		expectedState: conflict.RunState{
			Mode:      conflict.ModeStandard,
			Format:    conflict.FormatJSONL,
			Verbosity: conflict.VerbosityNormal,
		},
	},
	{
		name: "Bool overrides JSONL",
		args: []string{"--bool", "--jsonl"},
		flagSet: map[string]bool{
			"bool": true, "jsonl": true,
		},
		expectedState: conflict.RunState{
			Mode:      conflict.ModeBool,
			Format:    conflict.FormatDefault,
			Verbosity: conflict.VerbosityQuiet,
		},
		hasWarning: true,
	},
	{
		name:           "Format Flag: Verbose",
		args:           []string{"--format=verbose"},
		explicitFormat: "verbose",
		flagSet:        map[string]bool{},
		expectedState: conflict.RunState{
			Mode:      conflict.ModeStandard,
			Format:    conflict.FormatVerbose,
			Verbosity: conflict.VerbosityNormal,
		},
	},
}

func TestConflictResolution_PipelineOfIntent(t *testing.T) {
	for _, tt := range conflictResolutionTests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate how config/cli.go determines lastFormat
			lastFormat := tt.explicitFormat
			if tt.flagSet["json"] {
				lastFormat = "json"
			}
			if tt.flagSet["jsonl"] {
				lastFormat = "jsonl"
			}
			if tt.flagSet["plain"] {
				lastFormat = "plain"
			}

			state, warnings, err := conflict.ResolveState(tt.flagSet, lastFormat)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if state.Mode != tt.expectedState.Mode {
				t.Errorf("Mode: got %v, want %v", state.Mode, tt.expectedState.Mode)
			}
			if state.Format != tt.expectedState.Format {
				t.Errorf("Format: got %v, want %v", state.Format, tt.expectedState.Format)
			}
			if state.Verbosity != tt.expectedState.Verbosity {
				t.Errorf("Verbosity: got %v, want %v", state.Verbosity, tt.expectedState.Verbosity)
			}

			if tt.hasWarning && len(warnings) == 0 {
				t.Error("Expected warnings, got none")
			}
			if !tt.hasWarning && len(warnings) > 0 {
				t.Errorf("Expected no warnings, got %v", warnings)
			}
		})
	}
}

func TestResolveState(t *testing.T) {
	TestConflictResolution_PipelineOfIntent(t)
}

func TestFormatAllWarnings(t *testing.T) {
	warnings := []conflict.Warning{
		{Message: "test warning"},
	}
	result := conflict.FormatAllWarnings(warnings)
	if result == "" {
		t.Error("Expected formatted warnings, got empty string")
	}
}
