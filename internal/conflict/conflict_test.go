// Package conflict tests for the "Pipeline of Intent" logic.
package conflict_test

import (
	"testing"
	"testing/quick"

	"hashi/internal/conflict"
)

// Feature: cli-guidelines-review, Property 25: Mutually exclusive flags are rejected
// **Validates: Requirements 17.1**
func TestProperty_MutuallyExclusiveFlagsAreRejected(t *testing.T) {
	// Property: Certain flag combinations must always return an error
	property := func(raw, verify bool) bool {
		if raw && verify {
			flagSet := map[string]bool{
				"raw": true, "verify": true,
			}
			_, _, err := conflict.ResolveState(nil, flagSet, "", true, false)
			return err != nil
		}
		return true
	}

	if err := quick.Check(property, nil); err != nil {
		t.Errorf("Property failed: %v", err)
	}
}

// TestConflictResolution_PipelineOfIntent tests the conflict resolution logic.
// **Validates: Pipeline of Intent (Conflict Resolution)**
func TestConflictResolution_PipelineOfIntent(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flagSet        map[string]bool
		explicitFormat string
		hasHashes      bool
		expectedState  conflict.RunState
		hasWarning     bool
	}{
		{
			name: "Default State",
			args: []string{},
			flagSet: map[string]bool{
				"json": false, "plain": false, "quiet": false, "verbose": false, "bool": false, "raw": false,
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
			name: "Last One Wins: JSON vs Plain (JSON wins)",
			args: []string{"--plain", "--json"},
			flagSet: map[string]bool{
				"plain": true, "json": true,
			},
			expectedState: conflict.RunState{
				Mode:      conflict.ModeStandard,
				Format:    conflict.FormatJSON,
				Verbosity: conflict.VerbosityNormal,
			},
		},
		{
			name: "Last One Wins: JSON vs Plain (Plain wins)",
			args: []string{"--json", "--plain"},
			flagSet: map[string]bool{
				"plain": true, "json": true,
			},
			expectedState: conflict.RunState{
				Mode:      conflict.ModeStandard,
				Format:    conflict.FormatPlain,
				Verbosity: conflict.VerbosityNormal,
			},
		},
		{
			name: "Format Flag: Verbose",
			args: []string{"--format=verbose"},
			explicitFormat: "verbose",
			flagSet: map[string]bool{},
			expectedState: conflict.RunState{
				Mode:      conflict.ModeStandard,
				Format:    conflict.FormatVerbose,
				Verbosity: conflict.VerbosityNormal,
			},
		},
		{
			name: "Raw Mode",
			args: []string{"--raw"},
			flagSet: map[string]bool{
				"raw": true,
			},
			expectedState: conflict.RunState{
				Mode:      conflict.ModeRaw,
				Format:    conflict.FormatDefault,
				Verbosity: conflict.VerbosityNormal,
			},
		},
		{
			name: "Verify Mode",
			args: []string{"--verify"},
			flagSet: map[string]bool{
				"verify": true,
			},
			expectedState: conflict.RunState{
				Mode:      conflict.ModeVerify,
				Format:    conflict.FormatDefault,
				Verbosity: conflict.VerbosityNormal,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state, warnings, err := conflict.ResolveState(tt.args, tt.flagSet, tt.explicitFormat, true, tt.hasHashes)
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

// TestConflictResolution_HardErrors tests mutually exclusive flag errors.
func TestConflictResolution_HardErrors(t *testing.T) {
	tests := []struct {
		name      string
		flagSet   map[string]bool
		hasHashes bool
	}{
		{
			name: "Raw vs Verify",
			flagSet: map[string]bool{
				"raw": true, "verify": true,
			},
		},
		{
			name: "Verify vs Hash Strings",
			flagSet: map[string]bool{
				"verify": true,
			},
			hasHashes: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := conflict.ResolveState(nil, tt.flagSet, "", true, tt.hasHashes)
			if err == nil {
				t.Error("Expected error for mutually exclusive flags, got nil")
			}
		})
	}
}