// Package conflict implements the "Pipeline of Intent" state machine for
// resolving flag configurations.
//
// DESIGN PRINCIPLE: Pipeline of Intent
// ------------------------------------
// Traditional CLI flag handling often relies on a "Conflict Matrix" (if A then not B).
// This scales poorly (O(N^2)) and leads to "Matrix Hell" as flags are added.
//
// The Pipeline of Intent replaces this with a three-phase state machine:
// 1. COLLECT: Scan all arguments to identify the user's goals (Mode, Format, Verbosity).
// 2. CONSTRUCT: Build the desired state using "Last One Wins" for format and "Highest Precedence" for modes.
// 3. RESOLVE: Apply overrides (e.g. --bool always forces --quiet) and emit warnings for suppressed flags.
//
// This reduces complexity to O(N) and provides a predictable, linear behavior
// that is easy for users to understand and developers to maintain.
package conflict

import (
	"fmt"
	"strings"
)

// Mode defines the operational mode of the application.
type Mode string

const (
	ModeStandard Mode = "standard"
	ModeBool     Mode = "bool" // --bool
)

// Format defines the data output format (stdout).
type Format string

const (
	FormatDefault Format = "default"
	FormatJSON    Format = "json"    // --json or --format=json
	FormatJSONL   Format = "jsonl"   // --jsonl or --format=jsonl
	FormatPlain   Format = "plain"   // --plain or --format=plain
	FormatCSV     Format = "csv"     // --csv or --format=csv
	FormatVerbose Format = "verbose" // --format=verbose
)

// Verbosity defines the logging level (stderr).
type Verbosity string

const (
	VerbosityNormal  Verbosity = "normal"
	VerbosityQuiet   Verbosity = "quiet"   // --quiet
	VerbosityVerbose Verbosity = "verbose" // --verbose
)

// RunState represents the finalized, resolved behavior of the application.
type RunState struct {
	Mode      Mode
	Format    Format
	Verbosity Verbosity
}

// Intent represents a user's specific request for a behavior.
type intent struct {
	Type     string // "mode", "format", "verbosity"
	Value    string // e.g. "json", "jsonl", "quiet"
	Position int    // Index in os.Args
	Flag     string // The actual flag used (e.g., "--json")
}

// Warning represents a non-fatal conflict resolution.
type Warning struct {
	Message string
}

// ResolveState processes detected flags to produce a consistent RunState.
func ResolveState(flagSet map[string]bool, lastFormatIntent string) (*RunState, []Warning, error) {
	warnings := make([]Warning, 0)

	// Phase 2: State Construction
	state := &RunState{
		Mode:      ModeStandard,
		Format:    FormatDefault,
		Verbosity: VerbosityNormal,
	}

	// 2a. Determine Mode
	if flagSet["bool"] {
		state.Mode = ModeBool
		state.Verbosity = VerbosityQuiet
	}

	// 2b. Determine Format
	formatWarn := state.determineFormat(lastFormatIntent)
	if formatWarn != "" {
		warnings = append(warnings, Warning{Message: formatWarn})
	}

	// 2c. Determine Verbosity
	verbosityWarns := state.determineVerbosity(flagSet)
	warnings = append(warnings, verbosityWarns...)

	return state, warnings, nil
}

func (s *RunState) determineFormat(intent string) string {
	if intent == "" {
		return ""
	}

	if s.Mode == ModeBool && intent != "" && intent != "default" {
		return fmt.Sprintf("--bool overrides --%s", intent)
	}

	if s.Mode != ModeBool {
		switch intent {
		case "json":
			s.Format = FormatJSON
		case "jsonl":
			s.Format = FormatJSONL
		case "plain":
			s.Format = FormatPlain
		case "verbose":
			s.Format = FormatVerbose
		case "default":
			s.Format = FormatDefault
		default:
			s.Format = Format(intent)
		}
	}
	return ""
}

func (s *RunState) determineVerbosity(flagSet map[string]bool) []Warning {
	var warns []Warning
	if s.Mode == ModeBool {
		return nil
	}

	if flagSet["quiet"] {
		s.Verbosity = VerbosityQuiet
		if flagSet["verbose"] {
			warns = append(warns, Warning{Message: "--quiet overrides --verbose"})
		}
	} else if flagSet["verbose"] {
		s.Verbosity = VerbosityVerbose
		if s.Format == FormatDefault {
			s.Format = FormatVerbose
		}
	}
	return warns
}

// FormatAllWarnings formats all warnings for display.
func FormatAllWarnings(warnings []Warning) string {
	if len(warnings) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, warn := range warnings {
		sb.WriteString(fmt.Sprintf("Warning: %s\n", warn.Message))
	}
	return sb.String()
}
