// Package config handles configuration and argument parsing for hashi.
//
// It supports multiple configuration sources with the following precedence:
// flags > environment variables > project config > user config > system config
//
// The package uses spf13/pflag for POSIX-compliant flag parsing, supporting
// both short (-v) and long (--verbose) flag formats.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/spf13/pflag"
	"github.com/Les-El/hashi/internal/conflict"
	"github.com/Les-El/hashi/internal/security"
)

// Exit codes for scripting support
const (
	ExitSuccess        = 0   // All files processed successfully
	ExitNoMatches      = 1   // No matches found (with --match-required)
	ExitPartialFailure = 2   // Some files failed to process
	ExitInvalidArgs    = 3   // Invalid arguments or flags
	ExitFileNotFound   = 4   // One or more files not found
	ExitPermissionErr  = 5   // Permission denied
	ExitIntegrityFail  = 6   // Archive integrity verification failed
	ExitInterrupted    = 130 // Interrupted by Ctrl-C (128 + SIGINT)
)

// ConfigCommandError is returned when a user tries to use a config subcommand.
type ConfigCommandError struct{}

func (e *ConfigCommandError) Error() string {
	return `Error: hashi does not support config subcommands

Configuration must be done by manually editing config files.

Hashi auto-loads config from these standard locations:
  • .hashi.toml (project-specific)
  • hashi/config.toml (in XDG config directory)
  • .hashi/config.toml (traditional dotfile)

For configuration documentation and examples, see:
  https://github.com/[your-repo]/hashi#configuration`
}

func (e *ConfigCommandError) ExitCode() int {
	return ExitInvalidArgs
}

// EnvConfig holds environment variable configuration.
type EnvConfig struct {
	NoColor     bool   // NO_COLOR environment variable
	Debug       bool   // DEBUG environment variable
	TmpDir      string // TMPDIR environment variable
	Home        string // HOME environment variable
	ConfigHome  string // XDG_CONFIG_HOME environment variable
	HashiConfig string // HASHI_CONFIG environment variable
	
	HashiAlgorithm     string // HASHI_ALGORITHM
	HashiOutputFormat  string // HASHI_OUTPUT_FORMAT
	HashiRecursive     bool   // HASHI_RECURSIVE
	HashiHidden        bool   // HASHI_HIDDEN
	HashiVerbose       bool   // HASHI_VERBOSE
	HashiQuiet         bool   // HASHI_QUIET
	HashiPreserveOrder bool   // HASHI_PRESERVE_ORDER
	
	HashiBlacklistFiles string // HASHI_BLACKLIST_FILES
	HashiBlacklistDirs  string // HASHI_BLACKLIST_DIRS
	HashiWhitelistFiles string // HASHI_WHITELIST_FILES
	HashiWhitelistDirs  string // HASHI_WHITELIST_DIRS
}

// Config holds all configuration options for hashi.
type Config struct {
	Files  []string
	Hashes []string

	Recursive     bool
	Hidden        bool
	Algorithm     string
	Verbose       bool
	Quiet         bool
	Bool          bool
	PreserveOrder bool
	Raw           bool

	MatchRequired bool

	OutputFormat string
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

	ConfigFile string

	BlacklistFiles []string
	BlacklistDirs  []string
	WhitelistFiles []string
	WhitelistDirs  []string

	ShowHelp    bool
	ShowVersion bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Algorithm:    "sha256",
		OutputFormat: "default",
		MinSize:      0,
		MaxSize:      -1, // No limit
	}
}

// WriteError returns a generic error message for security-sensitive write failures.
func WriteError() error {
	return fmt.Errorf("Unknown write/append error")
}

// WriteErrorWithVerbose returns either a generic or detailed error message.
func WriteErrorWithVerbose(verbose bool, verboseDetails string) error {
	if verbose {
		return fmt.Errorf("%s", verboseDetails)
	}
	return WriteError()
}

// FileSystemError returns a generic error for file system operations.
func FileSystemError(verbose bool, verboseDetails string) error {
	if verbose {
		return fmt.Errorf("%s", verboseDetails)
	}
	return WriteError()
}

// HandleFileWriteError processes file writing errors.
func HandleFileWriteError(err error, verbose bool, path string) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	
	if strings.Contains(errStr, "permission denied") || 
	   strings.Contains(errStr, "access is denied") {
		return FileSystemError(verbose, fmt.Sprintf("permission denied writing to %s", path))
	}
	
	if strings.Contains(errStr, "no space left") || 
	   strings.Contains(errStr, "disk full") {
		return FileSystemError(verbose, fmt.Sprintf("insufficient disk space for %s", path))
	}
	
	if strings.Contains(errStr, "network") || 
	   strings.Contains(errStr, "connection") ||
	   strings.Contains(errStr, "timeout") {
		return FileSystemError(verbose, fmt.Sprintf("network error writing to %s", path))
	}
	
	if strings.Contains(errStr, "file name too long") || 
	   strings.Contains(errStr, "path too long") {
		return FileSystemError(verbose, fmt.Sprintf("path too long: %s", path))
	}
	
	return err
}

// validateOutputPath validates that an output path is safe.
func validateOutputPath(path string, cfg *Config) error {
	opts := security.Options{
		Verbose:        cfg.Verbose,
		BlacklistFiles: cfg.BlacklistFiles,
		BlacklistDirs:  cfg.BlacklistDirs,
		WhitelistFiles: cfg.WhitelistFiles,
		WhitelistDirs:  cfg.WhitelistDirs,
	}
	return security.ValidateOutputPath(path, opts)
}

// LoadEnvConfig reads environment variables.
func LoadEnvConfig() *EnvConfig {
	env := &EnvConfig{
		NoColor:    os.Getenv("NO_COLOR") != "",
		Debug:      parseBoolEnv("DEBUG"),
		TmpDir:     os.Getenv("TMPDIR"),
		Home:       os.Getenv("HOME"),
		ConfigHome: os.Getenv("XDG_CONFIG_HOME"),
		
		HashiConfig:        os.Getenv("HASHI_CONFIG"),
		HashiAlgorithm:     os.Getenv("HASHI_ALGORITHM"),
		HashiOutputFormat:  os.Getenv("HASHI_OUTPUT_FORMAT"),
		HashiRecursive:     parseBoolEnv("HASHI_RECURSIVE"),
		HashiHidden:        parseBoolEnv("HASHI_HIDDEN"),
		HashiVerbose:       parseBoolEnv("HASHI_VERBOSE"),
		HashiQuiet:         parseBoolEnv("HASHI_QUIET"),
		HashiPreserveOrder: parseBoolEnv("HASHI_PRESERVE_ORDER"),
		
		HashiBlacklistFiles: os.Getenv("HASHI_BLACKLIST_FILES"),
		HashiBlacklistDirs:  os.Getenv("HASHI_BLACKLIST_DIRS"),
		HashiWhitelistFiles: os.Getenv("HASHI_WHITELIST_FILES"),
		HashiWhitelistDirs:  os.Getenv("HASHI_WHITELIST_DIRS"),
	}
	
	return env
}

func parseBoolEnv(key string) bool {
	val := strings.ToLower(os.Getenv(key))
	return val == "1" || val == "true" || val == "yes" || val == "on"
}

func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// ApplyEnvConfig applies environment variable configuration to a Config.
func (env *EnvConfig) ApplyEnvConfig(cfg *Config, flagSet *pflag.FlagSet) {
	if !flagSet.Changed("algorithm") && env.HashiAlgorithm != "" {
		cfg.Algorithm = env.HashiAlgorithm
	}
	if !flagSet.Changed("format") && env.HashiOutputFormat != "" {
		cfg.OutputFormat = env.HashiOutputFormat
	}
	if !flagSet.Changed("recursive") && env.HashiRecursive {
		cfg.Recursive = env.HashiRecursive
	}
	if !flagSet.Changed("hidden") && env.HashiHidden {
		cfg.Hidden = env.HashiHidden
	}
	if !flagSet.Changed("verbose") && env.HashiVerbose {
		cfg.Verbose = env.HashiVerbose
	}
	if !flagSet.Changed("quiet") && env.HashiQuiet {
		cfg.Quiet = env.HashiQuiet
	}
	if !flagSet.Changed("preserve-order") && env.HashiPreserveOrder {
		cfg.PreserveOrder = env.HashiPreserveOrder
	}
	
	if env.HashiBlacklistFiles != "" {
		patterns := parseCommaSeparated(env.HashiBlacklistFiles)
		cfg.BlacklistFiles = append(cfg.BlacklistFiles, patterns...)
	}
	if env.HashiBlacklistDirs != "" {
		patterns := parseCommaSeparated(env.HashiBlacklistDirs)
		cfg.BlacklistDirs = append(cfg.BlacklistDirs, patterns...)
	}
	if env.HashiWhitelistFiles != "" {
		patterns := parseCommaSeparated(env.HashiWhitelistFiles)
		cfg.WhitelistFiles = append(cfg.WhitelistFiles, patterns...)
	}
	if env.HashiWhitelistDirs != "" {
		patterns := parseCommaSeparated(env.HashiWhitelistDirs)
		cfg.WhitelistDirs = append(cfg.WhitelistDirs, patterns...)
	}
}

// LoadDotEnv loads environment variables from a .env file.
func LoadDotEnv(path string) error {
	if path == "" {
		path = ".env"
	}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open .env file: %w", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf(".env line %d: invalid format (expected KEY=VALUE): %s", lineNum, line)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %w", err)
	}
	return nil
}

type ConfigFile struct {
	Recursive     *bool   `toml:"recursive,omitempty"`
	Hidden        *bool   `toml:"hidden,omitempty"`
	Algorithm     *string `toml:"algorithm,omitempty"`
	Verbose       *bool   `toml:"verbose,omitempty"`
	Quiet         *bool   `toml:"quiet,omitempty"`
	Bool          *bool   `toml:"bool,omitempty"`
	PreserveOrder *bool   `toml:"preserve_order,omitempty"`
	Raw           *bool   `toml:"raw,omitempty"`
	MatchRequired *bool   `toml:"match_required,omitempty"`
	OutputFormat  *string `toml:"output_format,omitempty"`
	OutputFile    *string `toml:"output_file,omitempty"`
	Append        *bool   `toml:"append,omitempty"`
	Force         *bool   `toml:"force,omitempty"`
	LogFile       *string `toml:"log_file,omitempty"`
	LogJSON       *string `toml:"log_json,omitempty"`
	Include       []string `toml:"include,omitempty"`
	Exclude       []string `toml:"exclude,omitempty"`
	MinSize       *string  `toml:"min_size,omitempty"`
	MaxSize       *string  `toml:"max_size,omitempty"`
	BlacklistFiles []string `toml:"blacklist_files,omitempty"`
	BlacklistDirs  []string `toml:"blacklist_dirs,omitempty"`
	WhitelistFiles []string `toml:"whitelist_files,omitempty"`
	WhitelistDirs  []string `toml:"whitelist_dirs,omitempty"`
	Files          []string `toml:"files,omitempty"`
}

func LoadConfigFile(path string) (*ConfigFile, error) {
	if path == "" {
		return &ConfigFile{}, nil
	}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ConfigFile{}, nil
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()
	if strings.HasSuffix(strings.ToLower(path), ".toml") {
		return loadTOMLConfig(file)
	}
	return loadTextConfig(file)
}

func loadTOMLConfig(file *os.File) (*ConfigFile, error) {
	var cfg ConfigFile
	if _, err := toml.DecodeReader(file, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse TOML config: %w", err)
	}
	return &cfg, nil
}

func loadTextConfig(file *os.File) (*ConfigFile, error) {
	cfg := &ConfigFile{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cfg.Files = append(cfg.Files, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading text config: %w", err)
	}
	return cfg, nil
}

func (cf *ConfigFile) ApplyConfigFile(cfg *Config) error {
	if cf.Recursive != nil && !cfg.Recursive {
		cfg.Recursive = *cf.Recursive
	}
	if cf.Hidden != nil && !cfg.Hidden {
		cfg.Hidden = *cf.Hidden
	}
	if cf.Verbose != nil && !cfg.Verbose {
		cfg.Verbose = *cf.Verbose
	}
	if cf.Quiet != nil && !cfg.Quiet {
		cfg.Quiet = *cf.Quiet
	}
	if cf.Bool != nil && !cfg.Bool {
		cfg.Bool = *cf.Bool
	}
	if cf.PreserveOrder != nil && !cfg.PreserveOrder {
		cfg.PreserveOrder = *cf.PreserveOrder
	}
	if cf.Raw != nil && !cfg.Raw {
		cfg.Raw = *cf.Raw
	}
	if cf.MatchRequired != nil && !cfg.MatchRequired {
		cfg.MatchRequired = *cf.MatchRequired
	}
	if cf.Append != nil && !cfg.Append {
		cfg.Append = *cf.Append
	}
	if cf.Force != nil && !cfg.Force {
		cfg.Force = *cf.Force
	}
	
	if cf.Algorithm != nil && cfg.Algorithm == "sha256" {
		cfg.Algorithm = *cf.Algorithm
	}
	if cf.OutputFormat != nil && cfg.OutputFormat == "default" {
		cfg.OutputFormat = *cf.OutputFormat
	}
	if cf.OutputFile != nil && cfg.OutputFile == "" {
		cfg.OutputFile = *cf.OutputFile
	}
	if cf.LogFile != nil && cfg.LogFile == "" {
		cfg.LogFile = *cf.LogFile
	}
	if cf.LogJSON != nil && cfg.LogJSON == "" {
		cfg.LogJSON = *cf.LogJSON
	}
	
	if cf.MinSize != nil && cfg.MinSize == 0 {
		size, err := parseSize(*cf.MinSize)
		if err != nil {
			return fmt.Errorf("invalid min_size in config: %w", err)
		}
		cfg.MinSize = size
	}
	if cf.MaxSize != nil && cfg.MaxSize == -1 {
		size, err := parseSize(*cf.MaxSize)
		if err != nil {
			return fmt.Errorf("invalid max_size in config: %w", err)
		}
		cfg.MaxSize = size
	}
	
	if len(cf.Include) > 0 && len(cfg.Include) == 0 {
		cfg.Include = cf.Include
	}
	if len(cf.Exclude) > 0 && len(cfg.Exclude) == 0 {
		cfg.Exclude = cf.Exclude
	}
	
	if len(cf.BlacklistFiles) > 0 {
		cfg.BlacklistFiles = append(cfg.BlacklistFiles, cf.BlacklistFiles...)
	}
	if len(cf.BlacklistDirs) > 0 {
		cfg.BlacklistDirs = append(cfg.BlacklistDirs, cf.BlacklistDirs...)
	}
	if len(cf.WhitelistFiles) > 0 {
		cfg.WhitelistFiles = append(cfg.WhitelistFiles, cf.WhitelistFiles...)
	}
	if len(cf.WhitelistDirs) > 0 {
		cfg.WhitelistDirs = append(cfg.WhitelistDirs, cf.WhitelistDirs...)
	}
	
	if len(cf.Files) > 0 && len(cf.Files) == 0 {
		cfg.Files = cf.Files
	}
	
	return nil
}

var ValidOutputFormats = []string{"default", "verbose", "json", "plain"}
var ValidAlgorithms = []string{"sha256", "md5", "sha1", "sha512", "blake2b"}

func ValidateOutputFormat(format string) error {
	for _, valid := range ValidOutputFormats {
		if format == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid output format %q: must be one of %s", format, strings.Join(ValidOutputFormats, ", "))
}

func ValidateAlgorithm(algorithm string) error {
	for _, valid := range ValidAlgorithms {
		if algorithm == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid algorithm %q: must be one of %s", algorithm, strings.Join(ValidAlgorithms, ", "))
}

// ValidateConfig validates the configuration and returns an error if invalid.
func ValidateConfig(cfg *Config) ([]conflict.Warning, error) {
	warnings := make([]conflict.Warning, 0)
	
	opts := security.Options{
		Verbose:        cfg.Verbose,
		BlacklistFiles: cfg.BlacklistFiles,
		BlacklistDirs:  cfg.BlacklistDirs,
		WhitelistFiles: cfg.WhitelistFiles,
		WhitelistDirs:  cfg.WhitelistDirs,
	}

	// 1. Security validation of inputs
	if err := security.ValidateInputs(cfg.Files, cfg.Hashes, opts); err != nil {
		return warnings, err
	}

	// 2. Format and algorithm validation
	if err := ValidateOutputFormat(cfg.OutputFormat); err != nil {
		return warnings, err
	}

	if err := ValidateAlgorithm(cfg.Algorithm); err != nil {
		return warnings, err
	}

	if cfg.MinSize < 0 {
		return warnings, fmt.Errorf("min-size must be non-negative, got %d", cfg.MinSize)
	}
	if cfg.MaxSize != -1 && cfg.MaxSize < 0 {
		return warnings, fmt.Errorf("max-size must be non-negative or -1 (no limit), got %d", cfg.MaxSize)
	}
	if cfg.MaxSize != -1 && cfg.MinSize > cfg.MaxSize {
		return warnings, fmt.Errorf("min-size (%d) cannot be greater than max-size (%d)", cfg.MinSize, cfg.MaxSize)
	}

	if err := validateOutputPath(cfg.OutputFile, cfg); err != nil {
		return warnings, fmt.Errorf("output file: %w", err)
	}
	
	if err := validateOutputPath(cfg.LogFile, cfg); err != nil {
		return warnings, fmt.Errorf("log file: %w", err)
	}
	
	if err := validateOutputPath(cfg.LogJSON, cfg); err != nil {
		return warnings, fmt.Errorf("JSON log file: %w", err)
	}

	return warnings, nil
}

func parseSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" || s == "-1" {
		return -1, nil
	}

	suffixes := []struct {
		suffix string
		mult   int64
	}{
		{"TB", 1024 * 1024 * 1024 * 1024},
		{"GB", 1024 * 1024 * 1024},
		{"MB", 1024 * 1024},
		{"KB", 1024},
		{"T", 1024 * 1024 * 1024 * 1024},
		{"G", 1024 * 1024 * 1024},
		{"M", 1024 * 1024},
		{"K", 1024},
		{"B", 1},
	}

	for _, s2 := range suffixes {
		if strings.HasSuffix(s, s2.suffix) {
			numStr := strings.TrimSuffix(s, s2.suffix)
			num, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid size %q: %w", s, err)
			}
			return int64(num * float64(s2.mult)), nil
		}
	}

	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q: must be a number or include unit (KB, MB, GB)", s)
	}
	return num, nil
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z07:00",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date %q: use format YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS", s)
}

// ParseArgs parses command-line arguments and returns a Config.
func ParseArgs(args []string) (*Config, []conflict.Warning, error) {
	cfg := DefaultConfig()
	fs := pflag.NewFlagSet("hashi", pflag.ContinueOnError)

	fs.BoolVarP(&cfg.Recursive, "recursive", "r", false, "Process directories recursively")
	fs.BoolVar(&cfg.Hidden, "hidden", false, "Include hidden files")
	fs.StringVarP(&cfg.Algorithm, "algorithm", "a", "sha256", "Hash algorithm")
	fs.BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable verbose output")
	fs.BoolVarP(&cfg.Quiet, "quiet", "q", false, "Suppress stdout")
	fs.BoolVarP(&cfg.Bool, "bool", "b", false, "Boolean output mode")
	fs.BoolVar(&cfg.PreserveOrder, "preserve-order", false, "Keep input order")
	fs.BoolVar(&cfg.Raw, "raw", false, "Treat files as raw bytes")
	fs.BoolVar(&cfg.MatchRequired, "match-required", false, "Exit 0 only if matches found")
	fs.StringVarP(&cfg.OutputFormat, "format", "f", "default", "Output format")
	fs.StringVarP(&cfg.OutputFile, "output", "o", "", "Write output to file")
	fs.BoolVar(&cfg.Append, "append", false, "Append to output file")
	fs.BoolVar(&cfg.Force, "force", false, "Overwrite without prompting")

	var jsonOutput, plainOutput bool
	fs.BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	fs.BoolVar(&plainOutput, "plain", false, "Output in plain format")

	fs.StringVar(&cfg.LogFile, "log-file", "", "File for logging")
	fs.StringVar(&cfg.LogJSON, "log-json", "", "File for JSON logging")

	fs.StringSliceVarP(&cfg.Include, "include", "i", nil, "Glob patterns to include")
	fs.StringSliceVarP(&cfg.Exclude, "exclude", "e", nil, "Glob patterns to exclude")
	
	var minSizeStr, maxSizeStr string
	fs.StringVar(&minSizeStr, "min-size", "0", "Minimum file size")
	fs.StringVar(&maxSizeStr, "max-size", "-1", "Maximum file size")

	var modifiedAfterStr, modifiedBeforeStr string
	fs.StringVar(&modifiedAfterStr, "modified-after", "", "Date")
	fs.StringVar(&modifiedBeforeStr, "modified-before", "", "Date")

	fs.StringVarP(&cfg.ConfigFile, "config", "c", "", "Path to config file")
	fs.BoolVarP(&cfg.ShowHelp, "help", "h", false, "Show help")
	fs.BoolVarP(&cfg.ShowVersion, "version", "V", false, "Show version")

	if err := fs.Parse(args); err != nil {
		// Try to suggest a correction for unknown flag errors
		errMsg := err.Error()
		if strings.Contains(errMsg, "unknown flag: ") {
			unknown := strings.TrimPrefix(errMsg, "unknown flag: ")
			suggestion := SuggestFlag(unknown)
			if suggestion != "" {
				return nil, nil, fmt.Errorf("%s (Did you mean %s?)", errMsg, suggestion)
			}
		}
		return nil, nil, err
	}

	var err error
	cfg.MinSize, err = parseSize(minSizeStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid --min-size: %w", err)
	}
	cfg.MaxSize, err = parseSize(maxSizeStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid --max-size: %w", err)
	}
	cfg.ModifiedAfter, err = parseDate(modifiedAfterStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid --modified-after: %w", err)
	}
	cfg.ModifiedBefore, err = parseDate(modifiedBeforeStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid --modified-before: %w", err)
	}

	remainingArgs := fs.Args()
	for _, arg := range remainingArgs {
		if arg == "config" {
			return nil, nil, &ConfigCommandError{}
		}
	}

	if len(remainingArgs) > 0 && allArgsAreNonExistentFiles(remainingArgs) {
		hasHashLikeArgs := false
		for _, arg := range remainingArgs {
			if looksLikeHashString(arg) {
				hasHashLikeArgs = true
				break
			}
		}
		if hasHashLikeArgs {
			cfg.Files = []string{}
			cfg.Hashes = remainingArgs
		} else {
			files, hashes, err := ClassifyArguments(remainingArgs, cfg.Algorithm)
			if err != nil {
				return nil, nil, err
			}
			cfg.Files = files
			cfg.Hashes = hashes
		}
	} else {
		files, hashes, err := ClassifyArguments(remainingArgs, cfg.Algorithm)
		if err != nil {
			return nil, nil, err
		}
		cfg.Files = files
		cfg.Hashes = hashes
	}

	envCfg := LoadEnvConfig()
	envCfg.ApplyEnvConfig(cfg, fs)

	configPath := cfg.ConfigFile
	if configPath == "" {
		configPath = FindConfigFile()
	}
	if configPath != "" {
		configFile, err := LoadConfigFile(configPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load config file %s: %w", configPath, err)
		}
		if err := configFile.ApplyConfigFile(cfg); err != nil {
			return nil, nil, fmt.Errorf("failed to apply config file %s: %w", configPath, err)
		}
	}

	// CONFLICT RESOLUTION LOGIC
	
	// 2. Resolve final state using the Pipeline of Intent
	flagSet := map[string]bool{
		"json":    jsonOutput,
		"plain":   plainOutput,
		"quiet":   cfg.Quiet,
		"verbose": cfg.Verbose,
		"bool":    cfg.Bool,
		"raw":     cfg.Raw,
	}

	state, resolveWarnings, err := conflict.ResolveState(args, flagSet, cfg.OutputFormat)
	if err != nil {
		return nil, nil, err
	}

	// 3. Apply resolved state back to Config
	cfg.OutputFormat = string(state.Format)
	cfg.Quiet = (state.Verbosity == conflict.VerbosityQuiet)
	cfg.Verbose = (state.Verbosity == conflict.VerbosityVerbose)
	if state.Mode == conflict.ModeBool {
		cfg.Bool = true
		cfg.Quiet = true 
	}
	if state.Mode == conflict.ModeRaw {
		cfg.Raw = true
	}

	// 4. Final Validation (non-conflict checks)
	validationWarnings, err := ValidateConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	allWarnings := append(resolveWarnings, validationWarnings...)

	return cfg, allWarnings, nil
}

// HelpText returns the formatted help text.
func HelpText() string {
	return `hashi - A command-line hash comparison tool

EXAMPLES
  hashi                          Hash all files in current directory
  hashi file1.txt file2.txt      Compare hashes of two files
  hashi -b file1.txt file2.txt   Boolean check: do files match? (outputs true/false)
  hashi -r /path/to/dir          Recursively hash directory
  hashi --json *.txt             Output results as JSON
  hashi file.zip                 Verify ZIP file integrity (CRC32)
  hashi --raw file.zip           Hash ZIP file as raw bytes
  hashi -                        Read file list from stdin

USAGE
  hashi [flags] [files...]

FLAGS
  -h, --help                Show this help
  -V, --version             Show version
  -v, --verbose             Enable verbose output
  -q, --quiet               Suppress stdout, only return exit code
  -b, --bool                Boolean output mode (true/false)
  -r, --recursive           Process directories recursively
      --hidden              Include hidden files
  -a, --algorithm string    Hash algorithm: sha256, md5, sha1, sha512, blake2b (default: sha256)
      --preserve-order      Keep input order instead of grouping by hash
      --raw                 Treat files as raw bytes (bypass special handling)

BOOLEAN MODE (-b / --bool)
  Boolean mode outputs just "true" or "false" for scripting use cases.
  It overrides other output formats and implies quiet behavior.

  Default behavior (no match flags):
    hashi -b file1 file2 file3     # true if ALL files match

  With --match-required:
    hashi -b --match-required *.txt    # true if ANY matches found

  Check for uniqueness (using negation):
    ! hashi -b --match-required *.txt  # true if NO matches (all unique)

  Scripting examples:
    hashi -b file1 file2 && echo "match" || echo "different"
    if hashi -b old.txt new.txt; then echo "unchanged"; fi
    RESULT=$(hashi -b file1 file2)  # Capture true/false

OUTPUT FORMATS
  -f, --format string       Output format: default, verbose, json, plain
      --json                Shorthand for --format=json
      --plain               Shorthand for --format=plain
  -b, --bool                Boolean output mode (true/false) - overrides other output formats
  -o, --output string       Write output to file
      --append              Append to output file
      --force               Overwrite without prompting

EXIT CODE CONTROL
      --match-required      Exit 0 only if matches found

FILTERING
  -i, --include strings     Glob patterns to include
  -e, --exclude strings     Glob patterns to exclude
      --min-size string     Minimum file size (e.g., 10KB, 1MB, 1GB)
      --max-size string     Maximum file size (-1 for no limit)
      --modified-after      Only files modified after date (YYYY-MM-DD)
      --modified-before     Only files modified before date (YYYY-MM-DD)

CONFIGURATION
  -c, --config string       Path to config file

  Config File Auto-Discovery (searched in order):
    ./.hashi.toml                        Project-specific (highest priority)
    $XDG_CONFIG_HOME/hashi/config.toml   XDG standard location
    ~/.config/hashi/config.toml          XDG fallback location
    ~/.hashi/config.toml                 Traditional dotfile location

  Config File Format (TOML):
    [defaults]
    algorithm = "sha256"
    output_format = "plain"
    recursive = false
    quiet = false

    [colors]
    enabled = true
    matches = "green"
    mismatches = "red"
    warnings = "yellow"

    [security]
    blacklist_files = ["temp*", "draft*", "backup*"]
    blacklist_dirs = ["cache", "tmp*", "build*"]
    whitelist_files = ["important_config_report.txt"]
    whitelist_dirs = ["results_config"]

  Configuration Precedence (highest to lowest):
    1. Command-line flags (including explicit defaults like --algorithm=sha256)
    2. Environment variables (HASHI_*)
    3. Project config (./.hashi.toml)
    4. User config (~/.config/hashi.toml)
    5. Built-in defaults

  Examples of precedence behavior:
    export HASHI_ALGORITHM=md5
    hashi file.txt                    # Uses md5 (env var overrides default)
    hashi --algorithm=sha256 file.txt # Uses sha256 (explicit flag overrides env var)
    hashi --algorithm=md5 file.txt    # Uses md5 (explicit flag, same as env var)

  Note: Explicit flags always override environment variables, even when the flag
  value equals the built-in default. This ensures predictable behavior.

STDIN SUPPORT
  Use "-" as a file argument to read file paths from stdin:
    find . -name "*.txt" | hashi -
    echo "file1.txt" | hashi -

EXIT CODES
  0   Success (all files processed, or matches found with --match-required)
  1   No matches found (with --match-required)
  2   Some files failed to process
  3   Invalid arguments
  4   File not found
  5   Permission denied
  6   Archive integrity verification failed
  130 Interrupted (Ctrl-C)

ENVIRONMENT VARIABLES
  Standard Variables:
    NO_COLOR                Disable color output
    DEBUG                   Enable debug logging
    TMPDIR                  Temporary directory location
    HOME                    User home directory
    XDG_CONFIG_HOME         XDG config directory

  HASHI_* Variables (override config file settings):
    HASHI_CONFIG            Default config file path
    HASHI_ALGORITHM         Hash algorithm (sha256, md5, sha1, sha512, blake2b)
    HASHI_OUTPUT_FORMAT     Output format (default, verbose, json, plain)
    HASHI_RECURSIVE         Process directories recursively (true/false)
    HASHI_HIDDEN            Include hidden files (true/false)
    HASHI_VERBOSE           Enable verbose output (true/false)
    HASHI_QUIET             Suppress stdout output (true/false)
    HASHI_PRESERVE_ORDER    Maintain input order vs grouping (true/false)

  Security Variables (additive - add to built-in patterns):
    HASHI_BLACKLIST_FILES   File name patterns to block (comma-separated)
    HASHI_BLACKLIST_DIRS    Directory name patterns to block (comma-separated)
    HASHI_WHITELIST_FILES   File name patterns that override blacklists (comma-separated)
    HASHI_WHITELIST_DIRS    Directory name patterns that override blacklists (comma-separated)

  Examples:
    export HASHI_ALGORITHM=sha512
    export HASHI_OUTPUT_FORMAT=json
    export HASHI_RECURSIVE=true
    export HASHI_BLACKLIST_FILES="temp*,draft*,backup*"
    export HASHI_WHITELIST_FILES="important_config_report.txt"

For more information, visit: https://github.com/example/hashi
`
}

func VersionText() string {
	return "hashi version 1.0.11"
}

func (c *Config) HasStdinMarker() bool {
	for _, file := range c.Files {
		if file == "-" {
			return true
		}
	}
	return false
}

func (c *Config) FilesWithoutStdin() []string {
	result := make([]string, 0, len(c.Files))
	for _, file := range c.Files {
		if file != "-" {
			result = append(result, file)
		}
	}
	return result
}

func ClassifyArguments(args []string, algorithm string) (files []string, hashes []string, err error) {
	for _, arg := range args {
		if arg == "" {
			continue
		}
		if arg == "-" {
			files = append(files, arg)
			continue
		}
		if _, err := os.Stat(arg); err == nil {
			files = append(files, arg)
			continue
		}
		detectedAlgorithms := detectHashAlgorithm(arg)
		if len(detectedAlgorithms) == 0 {
			files = append(files, arg)
			continue
		}
		currentAlgorithmFound := false
		for _, detected := range detectedAlgorithms {
			if detected == algorithm {
				currentAlgorithmFound = true
				break
			}
		}
		if currentAlgorithmFound {
			hashes = append(hashes, strings.ToLower(arg))
		} else {
			if len(detectedAlgorithms) == 1 {
				return nil, nil, fmt.Errorf("hash length doesn't match %s (expected %d characters, got %d).\nThis looks like %s. Try: hashi --algo %s [files...] %s",
					algorithm, getExpectedLength(algorithm), len(arg), 
					strings.ToUpper(detectedAlgorithms[0]), detectedAlgorithms[0], arg)
			} else {
				algorithmList := make([]string, len(detectedAlgorithms))
				for i, alg := range detectedAlgorithms {
					algorithmList[i] = strings.ToUpper(alg)
				}
				return nil, nil, fmt.Errorf("hash length doesn't match %s (expected %d characters, got %d).\nCould be: %s\nSpecify algorithm with: hashi --algo [algorithm] [files...] %s",
					algorithm, getExpectedLength(algorithm), len(arg),
					strings.Join(algorithmList, ", "), arg)
			}
		}
	}
	return files, hashes, nil
}

func detectHashAlgorithm(hashStr string) []string {
	if !isValidHexString(hashStr) {
		return []string{}
	}
	switch len(hashStr) {
	case 8:
		return []string{"crc32"}
	case 32:
		return []string{"md5"}
	case 40:
		return []string{"sha1"}
	case 64:
		return []string{"sha256"}
	case 128:
		return []string{"sha512", "blake2b"}
	default:
		return []string{}
	}
}

func isValidHexString(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func getExpectedLength(algorithm string) int {
	switch algorithm {
	case "md5":
		return 32
	case "sha1":
		return 40
	case "sha256":
		return 64
	case "sha512", "blake2b":
		return 128
	default:
		return 64
	}
}

func FindConfigFile() string {
	locations := []string{
		"./.hashi.toml",
	}
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		locations = append(locations, filepath.Join(xdgConfigHome, "hashi", "config.toml"))
	}
	if home := os.Getenv("HOME"); home != "" {
		locations = append(locations, filepath.Join(home, ".config", "hashi", "config.toml"))
	}
	if home := os.Getenv("HOME"); home != "" {
		locations = append(locations, filepath.Join(home, ".hashi", "config.toml"))
	}
	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			return location
		}
	}
	return ""
}

func allArgsAreNonExistentFiles(args []string) bool {
	for _, arg := range args {
		if arg == "" {
			continue
		}
		if arg == "-" {
			return false
		}
		if _, err := os.Stat(arg); err == nil {
			return false
		}
	}
	return true
}

func looksLikeHashString(s string) bool {
	if len(s) == 0 {
		return false
	}
	if len(s) < 4 || len(s) > 256 {
		return false
	}
	if strings.Contains(s, ".") {
		return false
	}
	if strings.Contains(s, "/") || strings.Contains(s, "\\") {
		return false
	}
	if strings.Contains(s, " ") {
		return false
	}
	return true
}
