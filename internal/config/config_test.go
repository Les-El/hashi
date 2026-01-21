// Package config tests for argument parsing and configuration.
package config

import (
	"strings"
	"testing"
	"testing/quick"
)

// Helper for contains check
func containsSlice(slice []string, item string) bool {
	for _, s := range slice {
		if strings.HasPrefix(s, item) {
			return true
		}
	}
	return false
}

// TestParseArgs_Bool tests the --bool flag.
func TestParseArgs_Bool(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"short bool", []string{"-b"}, true},
		{"long bool", []string{"--bool"}, true},
		{"no bool", []string{"file.txt"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}
			if cfg.Bool != tt.want {
				t.Errorf("Bool = %v, want %v", cfg.Bool, tt.want)
			}
			
			// Bool now overrides and implies Quiet behavior
			// Bool should automatically set Quiet=true
			if tt.want && !cfg.Quiet {
				t.Errorf("Bool=true should automatically set Quiet=true, got Quiet=%v", cfg.Quiet)
			}
		})
	}
}

// TestParseArgs_BoolWithMatchFlags tests bool combined with match requirement flags.
func TestParseArgs_BoolWithMatchFlags(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		wantMatchRequired bool
	}{
		{
			name:              "bool alone (no match flags)",
			args:              []string{"-b"},
			wantMatchRequired: false,
		},
		{
			name:              "bool with match-required",
			args:              []string{"-b", "--match-required"},
			wantMatchRequired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}
			if cfg.MatchRequired != tt.wantMatchRequired {
				t.Errorf("MatchRequired = %v, want %v", cfg.MatchRequired, tt.wantMatchRequired)
			}
		})
	}
}

// TestParseArgs_Help tests that help flags are recognized.
func TestParseArgs_Help(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"short help", []string{"-h"}, true},
		{"long help", []string{"--help"}, true},
		{"no help", []string{"file.txt"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}
			if cfg.ShowHelp != tt.want {
				t.Errorf("ShowHelp = %v, want %v", cfg.ShowHelp, tt.want)
			}
		})
	}
}

// TestParseArgs_Verbose tests verbose flag parsing.
func TestParseArgs_Verbose(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"short verbose", []string{"-v"}, true},
		{"long verbose", []string{"--verbose"}, true},
		{"no verbose", []string{"file.txt"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}
			if cfg.Verbose != tt.want {
				t.Errorf("Verbose = %v, want %v", cfg.Verbose, tt.want)
			}
		})
	}
}

// TestParseArgs_OutputFormat tests output format flag parsing.
func TestParseArgs_OutputFormat(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"default format", []string{}, "default"},
		{"json shorthand", []string{"--json"}, "json"},
		{"plain shorthand", []string{"--plain"}, "plain"},
		{"format flag json", []string{"--format=json"}, "json"},
		{"format flag plain", []string{"--format=plain"}, "plain"},
		{"format flag verbose", []string{"--format=verbose"}, "verbose"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}
			if cfg.OutputFormat != tt.want {
				t.Errorf("OutputFormat = %v, want %v", cfg.OutputFormat, tt.want)
			}
		})
	}
}

// TestParseArgs_Files tests that positional arguments are collected as files.
func TestParseArgs_Files(t *testing.T) {
	args := []string{"file1.txt", "file2.txt", "file3.txt"}
	cfg, _, err := ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs() error = %v", err)
	}

	if len(cfg.Files) != 3 {
		t.Errorf("Files count = %d, want 3", len(cfg.Files))
	}

	for i, want := range args {
		if cfg.Files[i] != want {
			t.Errorf("Files[%d] = %v, want %v", i, cfg.Files[i], want)
		}
	}
}

// TestParseArgs_Algorithm tests algorithm flag parsing.
func TestParseArgs_Algorithm(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"default algorithm", []string{}, "sha256"},
		{"short md5", []string{"-a", "md5"}, "md5"},
		{"long sha1", []string{"--algorithm=sha1"}, "sha1"},
		{"long sha512", []string{"--algorithm", "sha512"}, "sha512"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}
			if cfg.Algorithm != tt.want {
				t.Errorf("Algorithm = %v, want %v", cfg.Algorithm, tt.want)
			}
		})
	}
}

// TestParseArgs_FlagOrderIndependence is a property-based test that verifies
// flags can be provided in any order and produce the same result.
// Property 6: Flags accept any order
// Validates: Requirements 4.4
func TestParseArgs_FlagOrderIndependence(t *testing.T) {
	// Test that specific flag combinations work in different orders
	// Note: Avoid mutually exclusive combinations like --json with --verbose
	testCases := [][]string{
		{"-v", "file.txt"},
		{"--verbose", "file.txt"},
		{"file.txt", "-v"},
		{"file.txt", "--verbose"},
	}

	var expected *Config
	for i, args := range testCases {
		cfg, _, err := ParseArgs(args)
		if err != nil {
			t.Fatalf("ParseArgs(%v) error = %v", args, err)
		}

		if i == 0 {
			expected = cfg
		} else {
			// Compare relevant fields
			if cfg.Verbose != expected.Verbose {
				t.Errorf("Order %d: Verbose = %v, want %v", i, cfg.Verbose, expected.Verbose)
			}
			if cfg.OutputFormat != expected.OutputFormat {
				t.Errorf("Order %d: OutputFormat = %v, want %v", i, cfg.OutputFormat, expected.OutputFormat)
			}
			if len(cfg.Files) != len(expected.Files) {
				t.Errorf("Order %d: Files count = %d, want %d", i, len(cfg.Files), len(expected.Files))
			}
		}
	}
}

// TestParseArgs_FlagOrderIndependence_Property is a property-based test using testing/quick.
// It generates random permutations of flags and verifies they produce equivalent configs.
// Property 6: Flags accept any order
// Validates: Requirements 4.4
func TestParseArgs_FlagOrderIndependence_Property(t *testing.T) {
	// Define test flag sets that should produce the same result regardless of order
	// Use --flag=value syntax to keep flag-value pairs together during permutation
	// Note: We avoid invalid combinations like --quiet with --verbose or --json with --verbose
	flagSets := [][]string{
		{"-v", "-r"},
		{"--plain", "--hidden"},
		{"--json", "--recursive"},
		{"--algorithm=md5", "--preserve-order"},
		{"-r", "--hidden", "--preserve-order"},
	}

	for _, flags := range flagSets {
		// Generate all permutations and verify they produce equivalent configs
		permutations := generatePermutations(flags)
		
		var baseConfig *Config
		for i, perm := range permutations {
			cfg, _, err := ParseArgs(perm)
			if err != nil {
				t.Fatalf("ParseArgs(%v) error = %v", perm, err)
			}

			if i == 0 {
				baseConfig = cfg
			} else {
				if !configsEquivalent(baseConfig, cfg) {
					t.Errorf("Flag order affected result:\n  Order 0: %v\n  Order %d: %v", flags, i, perm)
				}
			}
		}
	}
}

// generatePermutations generates all permutations of a string slice.
// Limited to small slices to avoid combinatorial explosion.
func generatePermutations(arr []string) [][]string {
	if len(arr) <= 1 {
		return [][]string{arr}
	}
	
	// Limit to first 24 permutations (4! = 24) to keep tests fast
	var result [][]string
	permute(arr, 0, &result, 24)
	return result
}

func permute(arr []string, start int, result *[][]string, limit int) {
	if len(*result) >= limit {
		return
	}
	if start == len(arr) {
		perm := make([]string, len(arr))
		copy(perm, arr)
		*result = append(*result, perm)
		return
	}
	for i := start; i < len(arr); i++ {
		arr[start], arr[i] = arr[i], arr[start]
		permute(arr, start+1, result, limit)
		arr[start], arr[i] = arr[i], arr[start]
	}
}

// configsEquivalent checks if two configs have equivalent flag values.
// It ignores the Files field since file order may differ.
func configsEquivalent(a, b *Config) bool {
	return a.Recursive == b.Recursive &&
		a.Hidden == b.Hidden &&
		a.Algorithm == b.Algorithm &&
		a.Verbose == b.Verbose &&
		a.Quiet == b.Quiet &&
		a.PreserveOrder == b.PreserveOrder &&
		a.MatchRequired == b.MatchRequired &&
		a.OutputFormat == b.OutputFormat &&
		a.OutputFile == b.OutputFile &&
		a.Append == b.Append &&
		a.Force == b.Force
}

// TestDefaultConfig tests that DefaultConfig returns sensible defaults.
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Algorithm != "sha256" {
		t.Errorf("Algorithm = %v, want sha256", cfg.Algorithm)
	}
	if cfg.OutputFormat != "default" {
		t.Errorf("OutputFormat = %v, want default", cfg.OutputFormat)
	}
	if cfg.MaxSize != -1 {
		t.Errorf("MaxSize = %v, want -1", cfg.MaxSize)
	}
}

// TestHelpText tests that help text is non-empty and contains key sections.
func TestHelpText(t *testing.T) {
	help := HelpText()

	if len(help) == 0 {
		t.Error("HelpText() returned empty string")
	}

	// Check for key sections
	sections := []string{"EXAMPLES", "USAGE", "FLAGS", "EXIT CODES"}
	for _, section := range sections {
		if !contains(help, section) {
			t.Errorf("HelpText() missing section: %s", section)
		}
	}
}

// TestVersionText tests that version text is non-empty.
func TestVersionText(t *testing.T) {
	version := VersionText()

	if len(version) == 0 {
		t.Error("VersionText() returned empty string")
	}

	if !contains(version, "hashi") {
		t.Error("VersionText() should contain 'hashi'")
	}
}

// Property-based test: parsing should not panic on random input
func TestParseArgs_NoPanic(t *testing.T) {
	f := func(args []string) bool {
		// Filter out nil strings
		filtered := make([]string, 0, len(args))
		for _, arg := range args {
			if arg != "" {
				filtered = append(filtered, arg)
			}
		}

		// This should not panic
		_, _, _ = ParseArgs(filtered)
		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestParseArgs_AbbreviationRejection verifies that arbitrary abbreviations are rejected.
// Property 15: Abbreviated flags are rejected
func TestParseArgs_AbbreviationRejection(t *testing.T) {
	tests := []struct {
		arg     string
		wantErr bool
	}{
		{"--verb", true}, // Abbreviation of --verbose
		{"--help", false}, // Exact match
		{"-v", false},     // Exact short flag
		{"--vers", true},  // Abbreviation of --version
	}

	for _, tt := range tests {
		_, _, err := ParseArgs([]string{tt.arg})
		if tt.wantErr && err == nil {
			t.Errorf("ParseArgs(%q) expected error for abbreviation, got nil", tt.arg)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("ParseArgs(%q) unexpected error: %v", tt.arg, err)
		}
	}
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestParseArgs_ShortAndLongFlags tests that both short and long flag variants work.
// Validates: Requirements 4.1, 4.2
func TestParseArgs_ShortAndLongFlags(t *testing.T) {
	tests := []struct {
		name      string
		shortArgs []string
		longArgs  []string
		check     func(*Config) bool
	}{
		{
			name:      "recursive",
			shortArgs: []string{"-r"},
			longArgs:  []string{"--recursive"},
			check:     func(c *Config) bool { return c.Recursive },
		},
		{
			name:      "verbose",
			shortArgs: []string{"-v"},
			longArgs:  []string{"--verbose"},
			check:     func(c *Config) bool { return c.Verbose },
		},
		{
			name:      "quiet",
			shortArgs: []string{"-q"},
			longArgs:  []string{"--quiet"},
			check:     func(c *Config) bool { return c.Quiet },
		},
		{
			name:      "help",
			shortArgs: []string{"-h"},
			longArgs:  []string{"--help"},
			check:     func(c *Config) bool { return c.ShowHelp },
		},
		{
			name:      "version",
			shortArgs: []string{"-V"},
			longArgs:  []string{"--version"},
			check:     func(c *Config) bool { return c.ShowVersion },
		},
		{
			name:      "algorithm",
			shortArgs: []string{"-a", "md5"},
			longArgs:  []string{"--algorithm=md5"},
			check:     func(c *Config) bool { return c.Algorithm == "md5" },
		},
		{
			name:      "format",
			shortArgs: []string{"-f", "json"},
			longArgs:  []string{"--format=json"},
			check:     func(c *Config) bool { return c.OutputFormat == "json" },
		},
		{
			name:      "output",
			shortArgs: []string{"-o", "out.txt"},
			longArgs:  []string{"--output=out.txt"},
			check:     func(c *Config) bool { return c.OutputFile == "out.txt" },
		},
		{
			name:      "config",
			shortArgs: []string{"-c", "config.toml"},
			longArgs:  []string{"--config=config.toml"},
			check:     func(c *Config) bool { return c.ConfigFile == "config.toml" },
		},
		{
			name:      "include",
			shortArgs: []string{"-i", "*.txt"},
			longArgs:  []string{"--include=*.txt"},
			check:     func(c *Config) bool { return len(c.Include) == 1 && c.Include[0] == "*.txt" },
		},
		{
			name:      "exclude",
			shortArgs: []string{"-e", "*.log"},
			longArgs:  []string{"--exclude=*.log"},
			check:     func(c *Config) bool { return len(c.Exclude) == 1 && c.Exclude[0] == "*.log" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_short", func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.shortArgs)
			if err != nil {
				t.Fatalf("ParseArgs(%v) error = %v", tt.shortArgs, err)
			}
			if !tt.check(cfg) {
				t.Errorf("Short flag %v did not set expected value", tt.shortArgs)
			}
		})

		t.Run(tt.name+"_long", func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.longArgs)
			if err != nil {
				t.Fatalf("ParseArgs(%v) error = %v", tt.longArgs, err)
			}
			if !tt.check(cfg) {
				t.Errorf("Long flag %v did not set expected value", tt.longArgs)
			}
		})
	}
}

// TestParseArgs_StdinSupport tests that "-" is recognized as stdin marker.
// Validates: Requirements 4.3
func TestParseArgs_StdinSupport(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		hasStdin bool
	}{
		{"stdin only", []string{"-"}, true},
		{"stdin with files", []string{"file1.txt", "-", "file2.txt"}, true},
		{"stdin with flags", []string{"-v", "-"}, true},
		{"no stdin", []string{"file1.txt", "file2.txt"}, false},
		{"empty args", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs(%v) error = %v", tt.args, err)
			}
			if cfg.HasStdinMarker() != tt.hasStdin {
				t.Errorf("HasStdinMarker() = %v, want %v", cfg.HasStdinMarker(), tt.hasStdin)
			}
		})
	}
}

// TestParseArgs_FilesWithoutStdin tests that FilesWithoutStdin removes stdin marker.
func TestParseArgs_FilesWithoutStdin(t *testing.T) {
	cfg, _, err := ParseArgs([]string{"file1.txt", "-", "file2.txt"})
	if err != nil {
		t.Fatalf("ParseArgs() error = %v", err)
	}

	files := cfg.FilesWithoutStdin()
	if len(files) != 2 {
		t.Errorf("FilesWithoutStdin() returned %d files, want 2", len(files))
	}
	if files[0] != "file1.txt" || files[1] != "file2.txt" {
		t.Errorf("FilesWithoutStdin() = %v, want [file1.txt file2.txt]", files)
	}
}

// TestParseArgs_FlagValidation tests that invalid flag values are rejected.
// Validates: Requirements 4.5
func TestParseArgs_FlagValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		wantWarning bool
		errMsg      string
		warnMsg     string
	}{
		{
			name:    "invalid output format",
			args:    []string{"--format=invalid"},
			wantErr: true,
			errMsg:  "invalid output format",
		},
		{
			name:    "invalid algorithm",
			args:    []string{"--algorithm=invalid"},
			wantErr: true,
			errMsg:  "invalid algorithm",
		},
		{
			name:        "quiet overrides verbose (warning)",
			args:        []string{"--quiet", "--verbose"},
			wantErr:     false,
			wantWarning: true,
			warnMsg:     "--quiet overrides --verbose",
		},
		{
			name:        "json overrides verbose (no warning, split streams)",
			args:        []string{"--json", "--verbose"},
			wantErr:     false,
			wantWarning: false,
		},
		{
			name:        "json and plain - last wins (no warning, handled by parser)",
			args:        []string{"--json", "--plain"},
			wantErr:     false,
			wantWarning: false,
		},
		{
			name:    "invalid min-size",
			args:    []string{"--min-size=abc"},
			wantErr: true,
			errMsg:  "invalid",
		},
		{
			name:    "invalid date format",
			args:    []string{"--modified-after=not-a-date"},
			wantErr: true,
			errMsg:  "invalid",
		},
		{
			name:    "valid flags",
			args:    []string{"-v", "-r"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, warnings, err := ParseArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseArgs(%v) expected error, got nil", tt.args)
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ParseArgs(%v) error = %v, want error containing %q", tt.args, err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ParseArgs(%v) unexpected error = %v", tt.args, err)
				}
			}
			
			if tt.wantWarning {
				if len(warnings) == 0 {
					t.Errorf("ParseArgs(%v) expected warning, got none", tt.args)
				} else if tt.warnMsg != "" && !contains(warnings[0].Message, tt.warnMsg) {
					t.Errorf("ParseArgs(%v) warning = %v, want warning containing %q", tt.args, warnings[0].Message, tt.warnMsg)
				}
			}
		})
	}
}

// TestParseArgs_ErrorCases tests various error scenarios.
func TestParseArgs_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"unknown flag", []string{"--unknown-flag"}, true},
		{"missing flag value", []string{"--algorithm"}, true},
		{"invalid size with unit", []string{"--min-size=10XB"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ParseArgs(tt.args)
			if tt.wantErr && err == nil {
				t.Errorf("ParseArgs(%v) expected error, got nil", tt.args)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ParseArgs(%v) unexpected error = %v", tt.args, err)
			}
		})
	}
}

// TestParseArgs_SizeUnits tests human-readable size parsing.
func TestParseArgs_SizeUnits(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantSize int64
	}{
		{"bytes", []string{"--min-size=100"}, 100},
		{"kilobytes", []string{"--min-size=1KB"}, 1024},
		{"megabytes", []string{"--min-size=1MB"}, 1024 * 1024},
		{"gigabytes", []string{"--min-size=1GB"}, 1024 * 1024 * 1024},
		{"short K", []string{"--min-size=10K"}, 10 * 1024},
		{"short M", []string{"--min-size=10M"}, 10 * 1024 * 1024},
		{"decimal", []string{"--min-size=1.5MB"}, int64(1.5 * 1024 * 1024)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs(%v) error = %v", tt.args, err)
			}
			if cfg.MinSize != tt.wantSize {
				t.Errorf("MinSize = %d, want %d", cfg.MinSize, tt.wantSize)
			}
		})
	}
}

// TestParseArgs_DateParsing tests date flag parsing.
func TestParseArgs_DateParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantYear int
	}{
		{"simple date", []string{"--modified-after=2024-01-15"}, 2024},
		{"with time", []string{"--modified-after=2023-06-01T12:00:00"}, 2023},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs(%v) error = %v", tt.args, err)
			}
			if cfg.ModifiedAfter.Year() != tt.wantYear {
				t.Errorf("ModifiedAfter.Year() = %d, want %d", cfg.ModifiedAfter.Year(), tt.wantYear)
			}
		})
	}
}

// TestParseArgs_FlagValueSyntax tests both --flag=value and --flag value syntax.
// Validates: Requirements 4.4
func TestParseArgs_FlagValueSyntax(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"equals syntax", []string{"--algorithm=md5"}, "md5"},
		{"space syntax", []string{"--algorithm", "md5"}, "md5"},
		{"short with space", []string{"-a", "sha1"}, "sha1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs(%v) error = %v", tt.args, err)
			}
			if cfg.Algorithm != tt.want {
				t.Errorf("Algorithm = %v, want %v", cfg.Algorithm, tt.want)
			}
		})
	}
}

// TestValidateOutputFormat tests output format validation.
func TestValidateOutputFormat(t *testing.T) {
	tests := []struct {
		format  string
		wantErr bool
	}{
		{"default", false},
		{"verbose", false},
		{"json", false},
		{"plain", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			err := ValidateOutputFormat(tt.format)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateOutputFormat(%q) expected error", tt.format)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateOutputFormat(%q) unexpected error = %v", tt.format, err)
			}
		})
	}
}

// TestValidateAlgorithm tests algorithm validation.
func TestValidateAlgorithm(t *testing.T) {
	tests := []struct {
		algorithm string
		wantErr   bool
	}{
		{"sha256", false},
		{"md5", false},
		{"sha1", false},
		{"sha512", false},
		{"invalid", true},
		{"SHA256", true}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.algorithm, func(t *testing.T) {
			err := ValidateAlgorithm(tt.algorithm)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateAlgorithm(%q) expected error", tt.algorithm)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateAlgorithm(%q) unexpected error = %v", tt.algorithm, err)
			}
		})
	}
}

// TestParseArgs_BoolOverridesBehavior tests that --bool overrides other output flags.
func TestParseArgs_BoolOverridesBehavior(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantBool    bool
		wantQuiet   bool
		wantVerbose bool
		wantFormat  string
	}{
		{
			name:        "bool alone",
			args:        []string{"--bool"},
			wantBool:    true,
			wantQuiet:   true, // Bool implies Quiet
			wantVerbose: false,
			wantFormat:  "default",
		},
		{
			name:        "bool with quiet",
			args:        []string{"--bool", "--quiet"},
			wantBool:    true,
			wantQuiet:   true, // Bool overrides and implies Quiet
			wantVerbose: false,
			wantFormat:  "default",
		},
		{
			name:        "bool with verbose",
			args:        []string{"--bool", "--verbose"},
			wantBool:    true,
			wantQuiet:   true, // Bool overrides Verbose and implies Quiet
			wantVerbose: false,
			wantFormat:  "default",
		},
		{
			name:        "bool with json",
			args:        []string{"--bool", "--json"},
			wantBool:    true,
			wantQuiet:   true, // Bool overrides JSON and implies Quiet
			wantVerbose: false,
			wantFormat:  "default", // Bool overrides format
		},
		{
			name:        "bool with plain",
			args:        []string{"--bool", "--plain"},
			wantBool:    true,
			wantQuiet:   true, // Bool overrides Plain and implies Quiet
			wantVerbose: false,
			wantFormat:  "default", // Bool overrides format
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, warnings, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}
			
			// Check that Bool is set correctly
			if cfg.Bool != tt.wantBool {
				t.Errorf("Bool = %v, want %v", cfg.Bool, tt.wantBool)
			}
			
			// Check that Bool implies Quiet
			if cfg.Quiet != tt.wantQuiet {
				t.Errorf("Quiet = %v, want %v", cfg.Quiet, tt.wantQuiet)
			}
			
			// Check that Bool overrides Verbose
			if cfg.Verbose != tt.wantVerbose {
				t.Errorf("Verbose = %v, want %v", cfg.Verbose, tt.wantVerbose)
			}
			
			// Check that Bool overrides OutputFormat
			if cfg.OutputFormat != tt.wantFormat {
				t.Errorf("OutputFormat = %v, want %v", cfg.OutputFormat, tt.wantFormat)
			}
			
			// Check for override warnings
			// We expect warnings ONLY if Format is overridden (e.g. --bool overrides --json)
			// We do NOT expect warnings for Verbosity (e.g. --bool implies quiet, overriding --verbose is implicit/natural)
			expectWarning := false
			if tt.wantFormat == "default" && (containsSlice(tt.args, "--json") || containsSlice(tt.args, "--plain") || containsSlice(tt.args, "--format")) {
				expectWarning = true
			}

			if expectWarning && len(warnings) == 0 {
				t.Errorf("Expected override warning for %v, got none", tt.args)
			}
			if !expectWarning && len(warnings) > 0 {
				t.Errorf("Unexpected warning for %v: %v", tt.args, warnings)
			}
		})
	}
}
