package config

import (
	"time"
)

// Exit codes for scripting support
const (
	ExitSuccess        = 0   // All files processed successfully
	ExitNoMatches      = 1   // No matches found (with --any-match or --all-match)
	ExitPartialFailure = 2   // Some files failed to process
	ExitInvalidArgs    = 3   // Invalid arguments or flags
	ExitFileNotFound   = 4   // One or more files not found
	ExitPermissionErr  = 5   // Permission denied
	ExitInterrupted    = 130 // Interrupted by Ctrl-C (128 + SIGINT)
)

// ConfigCommandError is returned when a user tries to use a config subcommand.
type ConfigCommandError struct{}

// Error returns the formatted error message.
func (e *ConfigCommandError) Error() string {
	return `Error: chexum does not support config subcommands

Configuration must be done by manually editing config files.

Chexum auto-loads config from these standard locations:
  • .chexum.toml (project-specific)
  • chexum/config.toml (in XDG config directory)
  • .chexum/config.toml (traditional dotfile)

For configuration documentation and examples, see:
  https://github.com/[your-repo]/chexum#configuration`
}

// ExitCode returns the appropriate exit code for the error.
func (e *ConfigCommandError) ExitCode() int {
	return ExitInvalidArgs
}

// EnvConfig holds environment variable configuration.
type EnvConfig struct {
	NoColor      bool   // NO_COLOR environment variable
	Debug        bool   // DEBUG environment variable
	TmpDir       string // TMPDIR environment variable
	Home         string // HOME environment variable
	ConfigHome   string // XDG_CONFIG_HOME environment variable
	ChexumConfig string // CHEXUM_CONFIG environment variable

	ChexumAlgorithm      string // CHEXUM_ALGORITHM
	ChexumOutputFormat   string // CHEXUM_OUTPUT_FORMAT
	ChexumDryRun         bool   // CHEXUM_DRY_RUN
	ChexumRecursive      bool   // CHEXUM_RECURSIVE
	ChexumHidden         bool   // CHEXUM_HIDDEN
	ChexumVerbose        bool   // CHEXUM_VERBOSE
	ChexumQuiet          bool   // CHEXUM_QUIET
	ChexumBool           bool   // CHEXUM_BOOL
	ChexumPreserveOrder  bool   // CHEXUM_PRESERVE_ORDER
	ChexumMatchRequired  bool   // CHEXUM_MATCH_REQUIRED (deprecated, use CHEXUM_ANY_MATCH)
	ChexumAnyMatch       bool   // CHEXUM_ANY_MATCH
	ChexumAllMatch       bool   // CHEXUM_ALL_MATCH
	ChexumManifest       string // CHEXUM_MANIFEST
	ChexumOnlyChanged    bool   // CHEXUM_ONLY_CHANGED
	ChexumOutputManifest string // CHEXUM_OUTPUT_MANIFEST

	ChexumOutputFile string // CHEXUM_OUTPUT_FILE
	ChexumAppend     bool   // CHEXUM_APPEND
	ChexumForce      bool   // CHEXUM_FORCE

	ChexumLogFile string // CHEXUM_LOG_FILE
	ChexumLogJSON string // CHEXUM_LOG_JSON

	ChexumHelp    bool // CHEXUM_HELP
	ChexumVersion bool // CHEXUM_VERSION

	ChexumJobs int // CHEXUM_JOBS

	ChexumBlacklistFiles string // CHEXUM_BLACKLIST_FILES
	ChexumBlacklistDirs  string // CHEXUM_BLACKLIST_DIRS
	ChexumWhitelistFiles string // CHEXUM_WHITELIST_FILES
	ChexumWhitelistDirs  string // CHEXUM_WHITELIST_DIRS
}

// Config holds all configuration options for chexum.
type Config struct {
	Input       InputConfig
	Output      OutputConfig
	Processing  ProcessingConfig
	Incremental IncrementalConfig
	Security    SecurityConfig
	Discovery   DiscoveryConfig

	ConfigFile  string
	ShowHelp    bool
	ShowVersion bool

	// Deprecated: Moving to structured fields
	Files  []string
	Hashes []string

	Recursive     bool
	Hidden        bool
	Algorithm     string
	DryRun        bool
	Verbose       bool
	Quiet         bool
	Bool          bool
	PreserveOrder bool
	Jobs          int
	Test          bool

	MatchRequired bool
	AnyMatch      bool
	AllMatch      bool
	KeepTmp       bool

	OutputFormat string
	JSON         bool
	JSONL        bool
	Plain        bool
	CSV          bool
	OutputFile   string
	Append       bool
	Force        bool

	LogFile string
	LogJSON string

	Include        []string
	Exclude        []string
	MinSize        int64
	MaxSize        int64
	ModifiedAfter  time.Time
	ModifiedBefore time.Time

	Manifest       string
	OnlyChanged    bool
	OutputManifest string

	BlacklistFiles []string
	BlacklistDirs  []string
	WhitelistFiles []string
	WhitelistDirs  []string

	Unknowns []string
}

// InputConfig holds file discovery and filtering options.
type InputConfig struct {
	Files          []string
	Hashes         []string
	Include        []string
	Exclude        []string
	MinSize        int64
	MaxSize        int64
	ModifiedAfter  time.Time
	ModifiedBefore time.Time
}

// OutputConfig holds output formatting and destination options.
type OutputConfig struct {
	Format     string
	JSON       bool
	JSONL      bool
	Plain      bool
	OutputFile string
	Append     bool
	Force      bool
	LogFile    string
	LogJSON    string
}

// ProcessingConfig holds core processing behavior options.
type ProcessingConfig struct {
	Recursive     bool
	Hidden        bool
	Algorithm     string
	DryRun        bool
	Verbose       bool
	Quiet         bool
	Bool          bool
	PreserveOrder bool
	MatchRequired bool
	Jobs          int
}

// IncrementalConfig holds options for incremental hashing.
type IncrementalConfig struct {
	Manifest       string
	OnlyChanged    bool
	OutputManifest string
}

// SecurityConfig holds security policy overrides.
type SecurityConfig struct {
	BlacklistFiles []string
	BlacklistDirs  []string
	WhitelistFiles []string
	WhitelistDirs  []string
}

// DiscoveryConfig holds paths used by analysis engines.
type DiscoveryConfig struct {
	InternalPath string
	MainEntry    string
	DocsPath     string
}

// ValidatedConfig is a marker type for a configuration that has been validated.
type ValidatedConfig struct {
	*Config
}
