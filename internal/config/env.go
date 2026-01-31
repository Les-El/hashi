package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

// EnvParser implements the Parser interface for environment variables.
type EnvParser struct {
	ExtraVars map[string]string
	FlagSet   *pflag.FlagSet
}

// NewEnvParser creates a new EnvParser.
func NewEnvParser(extra map[string]string, fs *pflag.FlagSet) *EnvParser {
	return &EnvParser{
		ExtraVars: extra,
		FlagSet:   fs,
	}
}

// Parse implements the Parser interface.
func (p *EnvParser) Parse(cfg *Config) error {
	env := LoadEnvConfig(p.ExtraVars)
	env.ApplyEnvConfig(cfg, p.FlagSet)
	return nil
}

// LoadEnvConfig reads environment variables into an EnvConfig struct.
// It accepts an optional map of extra variables (e.g. from a .env file)
// which take precedence over actual environment variables.
func LoadEnvConfig(extra map[string]string) *EnvConfig {
	get := func(key string) string {
		if val, ok := extra[key]; ok {
			return val
		}
		return os.Getenv(key)
	}

	env := &EnvConfig{
		NoColor:    get("NO_COLOR") != "",
		Debug:      parseBool(get("DEBUG")),
		TmpDir:     get("TMPDIR"),
		Home:       get("HOME"),
		ConfigHome: get("XDG_CONFIG_HOME"),

		ChexumConfig:         get("CHEXUM_CONFIG"),
		ChexumAlgorithm:      get("CHEXUM_ALGORITHM"),
		ChexumOutputFormat:   get("CHEXUM_OUTPUT_FORMAT"),
		ChexumDryRun:         parseBool(get("CHEXUM_DRY_RUN")),
		ChexumRecursive:      parseBool(get("CHEXUM_RECURSIVE")),
		ChexumHidden:         parseBool(get("CHEXUM_HIDDEN")),
		ChexumVerbose:        parseBool(get("CHEXUM_VERBOSE")),
		ChexumQuiet:          parseBool(get("CHEXUM_QUIET")),
		ChexumBool:           parseBool(get("CHEXUM_BOOL")),
		ChexumPreserveOrder:  parseBool(get("CHEXUM_PRESERVE_ORDER")),
		ChexumMatchRequired:  parseBool(get("CHEXUM_MATCH_REQUIRED")),
		ChexumAnyMatch:       parseBool(get("CHEXUM_ANY_MATCH")),
		ChexumAllMatch:       parseBool(get("CHEXUM_ALL_MATCH")),
		ChexumManifest:       get("CHEXUM_MANIFEST"),
		ChexumOnlyChanged:    parseBool(get("CHEXUM_ONLY_CHANGED")),
		ChexumOutputManifest: get("CHEXUM_OUTPUT_MANIFEST"),

		ChexumOutputFile: get("CHEXUM_OUTPUT_FILE"),
		ChexumAppend:     parseBool(get("CHEXUM_APPEND")),
		ChexumForce:      parseBool(get("CHEXUM_FORCE")),

		ChexumLogFile: get("CHEXUM_LOG_FILE"),
		ChexumLogJSON: get("CHEXUM_LOG_JSON"),

		ChexumHelp:    parseBool(get("CHEXUM_HELP")),
		ChexumVersion: parseBool(get("CHEXUM_VERSION")),

		ChexumJobs:           parseInt(get("CHEXUM_JOBS")),
		ChexumBlacklistFiles: get("CHEXUM_BLACKLIST_FILES"),
		ChexumBlacklistDirs:  get("CHEXUM_BLACKLIST_DIRS"),
		ChexumWhitelistFiles: get("CHEXUM_WHITELIST_FILES"),
		ChexumWhitelistDirs:  get("CHEXUM_WHITELIST_DIRS"),
	}

	return env
}

func parseBool(val string) bool {
	val = strings.ToLower(val)
	return val == "1" || val == "true" || val == "yes" || val == "on"
}

func parseInt(val string) int {
	if val == "" {
		return 0
	}
	var res int
	fmt.Sscanf(val, "%d", &res)
	return res
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
	env.applyBasicEnvConfig(cfg, flagSet)
	env.applyOutputEnvConfig(cfg, flagSet)
	env.applyBlacklistEnvConfig(cfg)
	env.applyWhitelistEnvConfig(cfg)
}

func (env *EnvConfig) applyBasicEnvConfig(cfg *Config, flagSet *pflag.FlagSet) {
	if !flagSet.Changed("algorithm") && env.ChexumAlgorithm != "" {
		cfg.Algorithm = env.ChexumAlgorithm
	}
	if !flagSet.Changed("dry-run") && env.ChexumDryRun {
		cfg.DryRun = env.ChexumDryRun
	}
	if !flagSet.Changed("recursive") && env.ChexumRecursive {
		cfg.Recursive = env.ChexumRecursive
	}
	if !flagSet.Changed("hidden") && env.ChexumHidden {
		cfg.Hidden = env.ChexumHidden
	}
	if !flagSet.Changed("verbose") && env.ChexumVerbose {
		cfg.Verbose = env.ChexumVerbose
	}
	if !flagSet.Changed("quiet") && env.ChexumQuiet {
		cfg.Quiet = env.ChexumQuiet
	}
	if !flagSet.Changed("bool") && env.ChexumBool {
		cfg.Bool = env.ChexumBool
	}
	if !flagSet.Changed("preserve-order") && env.ChexumPreserveOrder {
		cfg.PreserveOrder = env.ChexumPreserveOrder
	}
	if !flagSet.Changed("match-required") && env.ChexumMatchRequired {
		cfg.MatchRequired = env.ChexumMatchRequired
	}
	if !flagSet.Changed("any-match") && env.ChexumAnyMatch {
		cfg.AnyMatch = env.ChexumAnyMatch
	}
	if !flagSet.Changed("all-match") && env.ChexumAllMatch {
		cfg.AllMatch = env.ChexumAllMatch
	}
	if !flagSet.Changed("manifest") && env.ChexumManifest != "" {
		cfg.Manifest = env.ChexumManifest
	}
	if !flagSet.Changed("only-changed") && env.ChexumOnlyChanged {
		cfg.OnlyChanged = env.ChexumOnlyChanged
	}
	if !flagSet.Changed("help") && env.ChexumHelp {
		cfg.ShowHelp = env.ChexumHelp
	}
	if !flagSet.Changed("version") && env.ChexumVersion {
		cfg.ShowVersion = env.ChexumVersion
	}
	if !flagSet.Changed("jobs") && env.ChexumJobs != 0 {
		cfg.Jobs = env.ChexumJobs
	}
}

func (env *EnvConfig) applyOutputEnvConfig(cfg *Config, flagSet *pflag.FlagSet) {
	if !flagSet.Changed("format") && env.ChexumOutputFormat != "" {
		cfg.OutputFormat = env.ChexumOutputFormat
	}
	if !flagSet.Changed("output") && env.ChexumOutputFile != "" {
		cfg.OutputFile = env.ChexumOutputFile
	}
	if !flagSet.Changed("append") && env.ChexumAppend {
		cfg.Append = env.ChexumAppend
	}
	if !flagSet.Changed("force") && env.ChexumForce {
		cfg.Force = env.ChexumForce
	}
	if !flagSet.Changed("log-file") && env.ChexumLogFile != "" {
		cfg.LogFile = env.ChexumLogFile
	}
	if !flagSet.Changed("log-json") && env.ChexumLogJSON != "" {
		cfg.LogJSON = env.ChexumLogJSON
	}
	if !flagSet.Changed("output-manifest") && env.ChexumOutputManifest != "" {
		cfg.OutputManifest = env.ChexumOutputManifest
	}
}

func (env *EnvConfig) applyBlacklistEnvConfig(cfg *Config) {
	if env.ChexumBlacklistFiles != "" {
		patterns := parseCommaSeparated(env.ChexumBlacklistFiles)
		cfg.BlacklistFiles = append(cfg.BlacklistFiles, patterns...)
	}
	if env.ChexumBlacklistDirs != "" {
		patterns := parseCommaSeparated(env.ChexumBlacklistDirs)
		cfg.BlacklistDirs = append(cfg.BlacklistDirs, patterns...)
	}
}

func (env *EnvConfig) applyWhitelistEnvConfig(cfg *Config) {
	if env.ChexumWhitelistFiles != "" {
		patterns := parseCommaSeparated(env.ChexumWhitelistFiles)
		cfg.WhitelistFiles = append(cfg.WhitelistFiles, patterns...)
	}
	if env.ChexumWhitelistDirs != "" {
		patterns := parseCommaSeparated(env.ChexumWhitelistDirs)
		cfg.WhitelistDirs = append(cfg.WhitelistDirs, patterns...)
	}
}

// LoadDotEnv loads environment variables from a .env file into a map.
func LoadDotEnv(path string) (map[string]string, error) {
	if path == "" {
		path = ".env"
	}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("failed to open .env file: %w", err)
	}
	defer file.Close()

	envVars := make(map[string]string)
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
			return nil, fmt.Errorf(".env line %d: invalid format (expected KEY=VALUE): %s", lineNum, line)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		envVars[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading .env file: %w", err)
	}
	return envVars, nil
}
