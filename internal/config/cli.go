package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/Les-El/chexum/internal/conflict"
	"github.com/spf13/pflag"
)

// CLIParser implements the Parser interface for command-line flags.
type CLIParser struct {
	Args    []string
	FlagSet *pflag.FlagSet
}

// NewCLIParser creates a new CLIParser.
func NewCLIParser(args []string, fs *pflag.FlagSet) *CLIParser {
	return &CLIParser{
		Args:    args,
		FlagSet: fs,
	}
}

// Parse implements the Parser interface.
func (p *CLIParser) Parse(cfg *Config) error {
	// 1. Define and parse flags
	defineFlags(p.FlagSet, cfg)
	if err := parseFlags(p.FlagSet, p.Args); err != nil {
		return err
	}

	// 2. Validate basic flag values
	if err := validateBasicFlags(cfg, p.FlagSet); err != nil {
		return err
	}

	// 3. Handle remaining positional arguments
	if err := handleArguments(cfg, p.FlagSet); err != nil {
		return err
	}

	return nil
}

// ParseArgs parses command-line arguments and returns a Config.
func ParseArgs(args []string) (*Config, []conflict.Warning, error) {
	cfg := DefaultConfig()
	fs := pflag.NewFlagSet("chexum", pflag.ContinueOnError)

	// Load .env variables
	envVars, _ := LoadDotEnv("")

	// Create MultiParser with priority: CLI > File > Env
	// (Note: File overrides Env because it's applied after Env in the loop if both check pflag.Changed)
	mp := &MultiParser{
		Parsers: []Parser{
			NewCLIParser(args, fs),
			NewEnvParser(envVars, fs),
			NewFileParser("", fs), // Path will be taken from cfg.ConfigFile during Parse
		},
	}

	if err := mp.Parse(cfg); err != nil {
		return nil, nil, err
	}

	// Resolve final state and validate
	if cfg.OutputFormat == "json" && cfg.Append {
		return nil, nil, fmt.Errorf("JSON output with --append is not supported. Use JSONL format for appending.")
	}

	return finalizeConfig(cfg, args, fs)
}

func defineFlags(flagSet *pflag.FlagSet, cfg *Config) {
	flagSet.BoolVarP(&cfg.Recursive, "recursive", "r", false, "Process directories recursively")
	flagSet.BoolVarP(&cfg.Hidden, "hidden", "H", false, "Include hidden files")
	flagSet.StringVarP(&cfg.Algorithm, "algorithm", "a", "sha256", "Hash algorithm")
	flagSet.BoolVar(&cfg.DryRun, "dry-run", false, "Preview files without hashing")
	flagSet.BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable verbose output")
	flagSet.BoolVarP(&cfg.Quiet, "quiet", "q", false, "Suppress stdout")
	flagSet.BoolVarP(&cfg.Bool, "bool", "b", false, "Boolean output mode")
	flagSet.BoolVar(&cfg.PreserveOrder, "preserve-order", false, "Keep input order")
	flagSet.BoolVar(&cfg.MatchRequired, "match-required", false, "Exit 0 only if matches found")
	_ = flagSet.MarkDeprecated("match-required", "use --any-match instead")
	flagSet.BoolVar(&cfg.AnyMatch, "any-match", false, "Exit 0 if at least one match is found")
	flagSet.BoolVar(&cfg.AllMatch, "all-match", false, "Exit 0 only if all files match")
	flagSet.BoolVar(&cfg.KeepTmp, "keep-tmp", false, "Keep temporary files after execution")
	flagSet.StringVarP(&cfg.OutputFormat, "format", "f", "default", "Output format")
	flagSet.StringVarP(&cfg.OutputFile, "output", "o", "", "Write output to file")
	flagSet.BoolVar(&cfg.Append, "append", false, "Append to output file")
	flagSet.BoolVar(&cfg.Force, "force", false, "Overwrite without prompting")

	flagSet.BoolVar(&cfg.JSON, "json", false, "Output in JSON format")
	flagSet.BoolVar(&cfg.JSONL, "jsonl", false, "Output in JSONL format")
	flagSet.BoolVar(&cfg.Plain, "plain", false, "Output in plain format")
	flagSet.BoolVar(&cfg.CSV, "csv", false, "Output in CSV format")

	flagSet.StringVar(&cfg.LogFile, "log-file", "", "File for logging")
	flagSet.StringVar(&cfg.LogJSON, "log-json", "", "File for JSON logging")

	flagSet.StringSliceVarP(&cfg.Include, "include", "i", nil, "Glob patterns to include")
	flagSet.StringSliceVarP(&cfg.Exclude, "exclude", "e", nil, "Glob patterns to exclude")

	flagSet.StringVar(&cfg.Manifest, "manifest", "", "Baseline manifest for incremental ops")
	flagSet.BoolVar(&cfg.OnlyChanged, "only-changed", false, "Only process files changed from manifest")
	flagSet.StringVar(&cfg.OutputManifest, "output-manifest", "", "Save results as a manifest")

	// Add placeholders for string-based filters that need parsing
	flagSet.String("min-size", "0", "Minimum file size")
	flagSet.String("max-size", "-1", "Maximum file size")
	flagSet.String("modified-after", "", "Date")
	flagSet.String("modified-before", "", "Date")

	flagSet.StringVarP(&cfg.ConfigFile, "config", "c", "", "Path to config file")
	flagSet.IntVarP(&cfg.Jobs, "jobs", "j", 0, "Number of parallel jobs (0 = auto)")
	flagSet.BoolVar(&cfg.Test, "test", false, "Run system diagnostics")
	flagSet.BoolVarP(&cfg.ShowHelp, "help", "h", false, "Show help")
	flagSet.BoolVarP(&cfg.ShowVersion, "version", "V", false, "Show version")
}

func parseFlags(fs *pflag.FlagSet, args []string) error {
	if err := fs.Parse(args); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "unknown flag: ") {
			unknown := strings.TrimPrefix(errMsg, "unknown flag: ")
			if suggestion := SuggestFlag(unknown); suggestion != "" {
				return fmt.Errorf("%s (Did you mean %s?)", errMsg, suggestion)
			}
		}
		return err
	}
	return nil
}

func validateBasicFlags(cfg *Config, fs *pflag.FlagSet) error {
	var err error
	if minStr, _ := fs.GetString("min-size"); minStr != "" {
		if cfg.MinSize, err = parseSize(minStr); err != nil {
			return fmt.Errorf("invalid --min-size: %w", err)
		}
	}
	if maxStr, _ := fs.GetString("max-size"); maxStr != "" {
		if cfg.MaxSize, err = parseSize(maxStr); err != nil {
			return fmt.Errorf("invalid --max-size: %w", err)
		}
	}
	if afterStr, _ := fs.GetString("modified-after"); afterStr != "" {
		if cfg.ModifiedAfter, err = parseDate(afterStr); err != nil {
			return fmt.Errorf("invalid --modified-after: %w", err)
		}
	}
	if beforeStr, _ := fs.GetString("modified-before"); beforeStr != "" {
		if cfg.ModifiedBefore, err = parseDate(beforeStr); err != nil {
			return fmt.Errorf("invalid --modified-before: %w", err)
		}
	}

	if cfg.Jobs < 0 {
		return fmt.Errorf("number of jobs cannot be negative")
	}

	return nil
}

func handleArguments(cfg *Config, fs *pflag.FlagSet) error {
	remainingArgs := fs.Args()
	for _, arg := range remainingArgs {
		if arg == "config" {
			return &ConfigCommandError{}
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
			return nil
		}
	}

	files, hashes, unknowns, err := ClassifyArguments(remainingArgs, cfg.Algorithm)
	if err != nil {
		return err
	}
	cfg.Files = files
	cfg.Hashes = hashes
	cfg.Unknowns = unknowns
	return nil
}

// Reviewed: NESTED-LOOP - Shorthand bundle parsing is O(N*M) where M is small.
func finalizeConfig(cfg *Config, args []string, flagSet *pflag.FlagSet) (*Config, []conflict.Warning, error) {
	lastFormat := detectLastFormat(cfg, args, flagSet)

	flagSetMap := map[string]bool{
		"json":    cfg.JSON,
		"jsonl":   cfg.JSONL,
		"plain":   cfg.Plain,
		"quiet":   cfg.Quiet,
		"verbose": cfg.Verbose,
		"bool":    cfg.Bool,
	}

	state, resolveWarnings, err := conflict.ResolveState(flagSetMap, lastFormat)
	if err != nil {
		return nil, nil, err
	}

	applyResolvedState(cfg, state)

	validationWarnings, err := ValidateConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	return cfg, append(resolveWarnings, validationWarnings...), nil
}

func detectLastFormat(cfg *Config, args []string, flagSet *pflag.FlagSet) string {
	formatFlagsSet := make(map[string]bool)
	flagSet.Visit(func(f *pflag.Flag) {
		switch f.Name {
		case "json", "jsonl", "plain", "csv", "format":
			formatFlagsSet[f.Name] = true
		}
	})

	var lastFormat string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			name := strings.Split(strings.TrimPrefix(arg, "--"), "=")[0]
			if formatFlagsSet[name] {
				if name == "format" {
					lastFormat = cfg.OutputFormat
				} else {
					lastFormat = name
				}
				if name == "format" && !strings.Contains(arg, "=") && i+1 < len(args) {
					i++
				}
			}
		} else if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") {
			for j := 1; j < len(arg); j++ {
				s := string(arg[j])
				if f := flagSet.ShorthandLookup(s); f != nil && formatFlagsSet[f.Name] {
					if f.Name == "format" {
						lastFormat = cfg.OutputFormat
						if j == len(arg)-1 && i+1 < len(args) {
							i++
						}
						break
					} else {
						lastFormat = f.Name
					}
				}
			}
		}
	}

	if lastFormat == "" {
		lastFormat = cfg.OutputFormat
	}
	return lastFormat
}

func applyResolvedState(cfg *Config, state *conflict.RunState) {
	cfg.OutputFormat = string(state.Format)
	cfg.Quiet = (state.Verbosity == conflict.VerbosityQuiet)
	cfg.Verbose = (state.Verbosity == conflict.VerbosityVerbose)
	if state.Mode == conflict.ModeBool {
		cfg.Bool = true
		cfg.Quiet = true
	}
}

// ClassifyArguments separates arguments into file paths, hash strings, and unknowns.
func ClassifyArguments(args []string, algorithm string) (files []string, hashes []string, unknowns []string, err error) {
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
			// If it's not a file and not a valid hash length, it's an unknown string.
			unknowns = append(unknowns, arg)
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
			// If it looks like a hash but doesn't match the current algorithm,
			// we could treat it as unknown or error.
			// Given the "Pool Matching" intent, let's treat it as unknown for now
			// or keep the error if we want strictness.
			// The user said "INVAlID: for strings that are neither files nor hashes".
			// So length-mismatched hashes should probably be INVALID.
			unknowns = append(unknowns, arg)
		}
	}
	return files, hashes, unknowns, nil
}

func detectHashAlgorithm(hashStr string) []string {
	if !isValidHexString(hashStr) {
		return []string{}
	}
	switch len(hashStr) {
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

// HasStdinMarker checks if the special "-" argument is present in the file list.
func (c *Config) HasStdinMarker() bool {
	for _, file := range c.Files {
		if file == "-" {
			return true
		}
	}
	return false
}

// FilesWithoutStdin returns the list of files excluding the stdin marker "-".
func (c *Config) FilesWithoutStdin() []string {
	result := make([]string, 0, len(c.Files))
	for _, file := range c.Files {
		if file != "-" {
			result = append(result, file)
		}
	}
	return result
}
