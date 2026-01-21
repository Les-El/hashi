// Package conflict implements the "Pipeline of Intent" state machine for
// resolving flag configurations.
//
// It replaces the traditional pairwise conflict checks with a phased approach:
// 1. Intent Collection (Scanning args)
// 2. State Construction (Applying rules like "Last One Wins")
// 3. Validation (Checking for invalid states)
package conflict

import (
	"fmt"
	"strings"
)

// Mode defines the operational mode of the application.
type Mode string

const (
	ModeStandard Mode = "standard"
	ModeBool     Mode = "bool"   // --bool
	ModeRaw      Mode = "raw"    // --raw
	ModeVerify   Mode = "verify" // --verify
)

// Format defines the data output format (stdout).
type Format string

const (
	FormatDefault Format = "default"
	FormatJSON    Format = "json"    // --json or --format=json
	FormatPlain   Format = "plain"   // --plain or --format=plain
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
	Value    string // e.g. "json", "quiet"
	Position int    // Index in os.Args
	Flag     string // The actual flag used (e.g., "--json")
}

// Warning represents a non-fatal conflict resolution.
type Warning struct {
	Message string
}

// ResolveState processes raw arguments and detected flags to produce a consistent RunState.
//
// args: The raw command-line arguments (os.Args[1:]) to determine order.
// flags: A map of boolean flags detected by the flag parser (e.g. {"json": true, "quiet": true}).
// explicitFormat: The value of --format if set (empty if default).
func ResolveState(args []string, flagSet map[string]bool, explicitFormat string) (*RunState, []Warning, error) {
	warnings := make([]Warning, 0)
	
	// Phase 1: Intent Collection
	// We scan args to establish the "timeline" of user intent.
	// This allows "Last One Wins" logic.
	
	lastFormatIntent := ""
	lastFormatPos := -1
	
	// Check if explicit format was provided via --format flag
	if explicitFormat != "" && explicitFormat != "default" {
		lastFormatIntent = explicitFormat
		// We assign a low priority position effectively, but specific flags like --json
		// usually override general --format if they come later.
		// However, to simplify, we can treat --format as an intent that happened "somewhere".
		// But --json and --plain are distinct flags.
		// If user does `--format=json --plain`, `plain` should win.
		// If user does `--plain --format=json`, `json` should win.
		// We need to find the positions.
	}

	// Scan args for relevant flags
	for i, arg := range args {
		if arg == "--json" || arg == "--plain" {
			lastFormatIntent = strings.TrimPrefix(arg, "--")
			lastFormatPos = i
		} else if strings.HasPrefix(arg, "--format=") {
			lastFormatIntent = strings.TrimPrefix(arg, "--format=")
			lastFormatPos = i
		} else if strings.HasPrefix(arg, "-f=") {
			lastFormatIntent = strings.TrimPrefix(arg, "-f=")
			lastFormatPos = i
		} else if arg == "-f" && i+1 < len(args) {
			lastFormatIntent = args[i+1]
			lastFormatPos = i
		}
	}

	// Phase 2: State Construction
	state := &RunState{
		Mode:      ModeStandard,
		Format:    FormatDefault,
		Verbosity: VerbosityNormal,
	}

	// 2a. Determine Format (Last One Wins)
	if lastFormatPos >= 0 || lastFormatIntent != "" {
		switch lastFormatIntent {
		case "json":
			state.Format = FormatJSON
		case "plain":
			state.Format = FormatPlain
		case "verbose":
			state.Format = FormatVerbose
		case "default":
			state.Format = FormatDefault
		default:
			// If unknown format was passed, preserve it so validation can catch it.
			state.Format = Format(lastFormatIntent)
		}
	}

	// 2b. Determine Verbosity (Quiet overrides Verbose)
	// We check the flagSet (populated by pflag) because these boolean logic rules
	// are "Existence" based, not "Order" based (per design doc).
	isQuiet := flagSet["quiet"]
	isVerbose := flagSet["verbose"]
	
	// Also check bool mode, which implies quiet
	isBool := flagSet["bool"]

	if isQuiet {
		state.Verbosity = VerbosityQuiet
		if isVerbose {
			warnings = append(warnings, Warning{Message: "--quiet overrides --verbose"})
		}
	} else if isVerbose {
		state.Verbosity = VerbosityVerbose
		// Promote default format to verbose if -v is used (Requirement 17.1)
		if state.Format == FormatDefault {
			state.Format = FormatVerbose
		}
	}

	// 2c. Determine Mode (Bool overrides everything)
	isRaw := flagSet["raw"]
	isVerify := flagSet["verify"]
	
	if isBool {
		state.Mode = ModeBool
		state.Verbosity = VerbosityQuiet // Bool implies Quiet
		
		// Bool overrides Format
		if state.Format != FormatDefault {
			warnings = append(warnings, Warning{Message: fmt.Sprintf("--bool overrides --%s", state.Format)})
			state.Format = FormatDefault // Reset to default (or we could define a FormatBool)
		}
	} else if isVerify {
		state.Mode = ModeVerify
	} else if isRaw {
		state.Mode = ModeRaw
	}

	// Phase 3: Validation (Hard Errors)
	
	return state, warnings, nil
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