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
)

// FileParser implements the Parser interface for config files.
type FileParser struct {
	Path    string
	FlagSet *pflag.FlagSet
}

// NewFileParser creates a new FileParser.
func NewFileParser(path string, fs *pflag.FlagSet) *FileParser {
	return &FileParser{
		Path:    path,
		FlagSet: fs,
	}
}

// Parse implements the Parser interface.
func (p *FileParser) Parse(cfg *Config) error {
	path := p.Path
	if path == "" {
		path = cfg.ConfigFile
	}
	if path == "" {
		path = FindConfigFile()
	}
	if path == "" {
		return nil
	}

	configFile, err := LoadConfigFile(path)
	if err != nil {
		return fmt.Errorf("failed to load config file %s: %w", path, err)
	}
	if err := configFile.ApplyConfigFile(cfg, p.FlagSet); err != nil {
		return fmt.Errorf("failed to apply config file %s: %w", path, err)
	}
	return nil
}

type ConfigFile struct {
	Defaults struct {
		Recursive     *bool    `toml:"recursive,omitempty"`
		Hidden        *bool    `toml:"hidden,omitempty"`
		Algorithm     *string  `toml:"algorithm,omitempty"`
		Verbose       *bool    `toml:"verbose,omitempty"`
		Quiet         *bool    `toml:"quiet,omitempty"`
		Bool          *bool    `toml:"bool,omitempty"`
		PreserveOrder *bool    `toml:"preserve_order,omitempty"`
		MatchRequired *bool    `toml:"match_required,omitempty"`
		AnyMatch      *bool    `toml:"any_match,omitempty"`
		AllMatch      *bool    `toml:"all_match,omitempty"`
		OutputFormat  *string  `toml:"output_format,omitempty"`
		OutputFile    *string  `toml:"output_file,omitempty"`
		Append        *bool    `toml:"append,omitempty"`
		Force         *bool    `toml:"force,omitempty"`
		LogFile       *string  `toml:"log_file,omitempty"`
		LogJSON       *string  `toml:"log_json,omitempty"`
		Include       []string `toml:"include,omitempty"`
		Exclude       []string `toml:"exclude,omitempty"`
		MinSize       *string  `toml:"min_size,omitempty"`
		MaxSize       *string  `toml:"max_size,omitempty"`
	} `toml:"defaults"`
	Security struct {
		BlacklistFiles []string `toml:"blacklist_files,omitempty"`
		BlacklistDirs  []string `toml:"blacklist_dirs,omitempty"`
		WhitelistFiles []string `toml:"whitelist_files,omitempty"`
		WhitelistDirs  []string `toml:"whitelist_dirs,omitempty"`
	} `toml:"security"`
	Files []string `toml:"files,omitempty"`
}

// LoadConfigFile reads and parses the configuration file at the given path.
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

// ApplyConfigFile merges the file configuration into the main Config, respecting flag precedence.
func (cf *ConfigFile) ApplyConfigFile(cfg *Config, flagSet *pflag.FlagSet) error {
	cf.applyBoolDefaults(cfg, flagSet)
	cf.applyStringDefaults(cfg, flagSet)

	if err := cf.applySizeDefaults(cfg, flagSet); err != nil {
		return err
	}

	cf.applyListDefaults(cfg, flagSet)
	cf.applySecurityDefaults(cfg)

	if len(cf.Files) > 0 && len(cfg.Files) == 0 {
		cfg.Files = cf.Files
	}

	return nil
}

func (cf *ConfigFile) applyBoolDefaults(cfg *Config, flagSet *pflag.FlagSet) {
	d := cf.Defaults
	boolFlags := []struct {
		val  *bool
		name string
		ptr  *bool
	}{
		{d.Recursive, "recursive", &cfg.Recursive},
		{d.Hidden, "hidden", &cfg.Hidden},
		{d.Verbose, "verbose", &cfg.Verbose},
		{d.Quiet, "quiet", &cfg.Quiet},
		{d.Bool, "bool", &cfg.Bool},
		{d.PreserveOrder, "preserve-order", &cfg.PreserveOrder},
		{d.MatchRequired, "match-required", &cfg.MatchRequired},
		{d.AnyMatch, "any-match", &cfg.AnyMatch},
		{d.AllMatch, "all-match", &cfg.AllMatch},
		{d.Append, "append", &cfg.Append},
		{d.Force, "force", &cfg.Force},
	}

	for _, f := range boolFlags {
		if f.val != nil && !flagSet.Changed(f.name) {
			*f.ptr = *f.val
		}
	}
}

func (cf *ConfigFile) applyStringDefaults(cfg *Config, flagSet *pflag.FlagSet) {
	d := cf.Defaults
	stringFlags := []struct {
		val  *string
		name string
		ptr  *string
	}{
		{d.Algorithm, "algorithm", &cfg.Algorithm},
		{d.OutputFormat, "format", &cfg.OutputFormat},
		{d.OutputFile, "output", &cfg.OutputFile},
		{d.LogFile, "log-file", &cfg.LogFile},
		{d.LogJSON, "log-json", &cfg.LogJSON},
	}

	for _, f := range stringFlags {
		if f.val != nil && !flagSet.Changed(f.name) {
			*f.ptr = *f.val
		}
	}
}

func (cf *ConfigFile) applySizeDefaults(cfg *Config, flagSet *pflag.FlagSet) error {
	d := cf.Defaults
	if d.MinSize != nil && !flagSet.Changed("min-size") {
		size, err := parseSize(*d.MinSize)
		if err != nil {
			return fmt.Errorf("invalid min_size in config: %w", err)
		}
		cfg.MinSize = size
	}
	if d.MaxSize != nil && !flagSet.Changed("max-size") {
		size, err := parseSize(*d.MaxSize)
		if err != nil {
			return fmt.Errorf("invalid max_size in config: %w", err)
		}
		cfg.MaxSize = size
	}
	return nil
}

func (cf *ConfigFile) applyListDefaults(cfg *Config, flagSet *pflag.FlagSet) {
	d := cf.Defaults
	if len(d.Include) > 0 && !flagSet.Changed("include") {
		cfg.Include = d.Include
	}
	if len(d.Exclude) > 0 && !flagSet.Changed("exclude") {
		cfg.Exclude = d.Exclude
	}
}

func (cf *ConfigFile) applySecurityDefaults(cfg *Config) {
	s := cf.Security
	cfg.BlacklistFiles = append(cfg.BlacklistFiles, s.BlacklistFiles...)
	cfg.BlacklistDirs = append(cfg.BlacklistDirs, s.BlacklistDirs...)
	cfg.WhitelistFiles = append(cfg.WhitelistFiles, s.WhitelistFiles...)
	cfg.WhitelistDirs = append(cfg.WhitelistDirs, s.WhitelistDirs...)
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

// FindConfigFile searches standard locations for a configuration file.
func FindConfigFile() string {
	locations := []string{
		"./.chexum.toml",
	}
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		locations = append(locations, filepath.Join(xdgConfigHome, "chexum", "config.toml"))
	}
	if home := os.Getenv("HOME"); home != "" {
		locations = append(locations, filepath.Join(home, ".config", "chexum", "config.toml"))
	}
	if home := os.Getenv("HOME"); home != "" {
		locations = append(locations, filepath.Join(home, ".chexum", "config.toml"))
	}
	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			return location
		}
	}
	return ""
}
