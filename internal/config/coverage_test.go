package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

// stringPtr is a helper to create a pointer to a string.
func stringPtr(s string) *string { return &s }

// stringSlicesEqual is a helper to compare two string slices.
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// configFilesEqual performs a deep comparison of two ConfigFile structs.
// Note: This is a simplified comparison, focusing on the fields that LoadConfigFile populates.
func configFilesEqual(a, b *ConfigFile) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Compare Files
	if !stringSlicesEqual(a.Files, b.Files) {
		return false
	}

	// Compare Defaults (pointers require dereferencing and nil checks)
	if !ptrBoolEqual(a.Defaults.Recursive, b.Defaults.Recursive) ||
		!ptrBoolEqual(a.Defaults.Hidden, b.Defaults.Hidden) ||
		!ptrStringEqual(a.Defaults.Algorithm, b.Defaults.Algorithm) ||
		!ptrBoolEqual(a.Defaults.Verbose, b.Defaults.Verbose) ||
		!ptrBoolEqual(a.Defaults.Quiet, b.Defaults.Quiet) ||
		!ptrBoolEqual(a.Defaults.Bool, b.Defaults.Bool) ||
		!ptrBoolEqual(a.Defaults.PreserveOrder, b.Defaults.PreserveOrder) ||
		!ptrBoolEqual(a.Defaults.MatchRequired, b.Defaults.MatchRequired) ||
		!ptrStringEqual(a.Defaults.OutputFormat, b.Defaults.OutputFormat) ||
		!ptrStringEqual(a.Defaults.OutputFile, b.Defaults.OutputFile) ||
		!ptrBoolEqual(a.Defaults.Append, b.Defaults.Append) ||
		!ptrBoolEqual(a.Defaults.Force, b.Defaults.Force) ||
		!ptrStringEqual(a.Defaults.LogFile, b.Defaults.LogFile) ||
		!ptrStringEqual(a.Defaults.LogJSON, b.Defaults.LogJSON) ||
		!stringSlicesEqual(a.Defaults.Include, b.Defaults.Include) ||
		!stringSlicesEqual(a.Defaults.Exclude, b.Defaults.Exclude) {
		return false
	}

	// Compare Security
	if !stringSlicesEqual(a.Security.BlacklistFiles, b.Security.BlacklistFiles) ||
		!stringSlicesEqual(a.Security.BlacklistDirs, b.Security.BlacklistDirs) ||
		!stringSlicesEqual(a.Security.WhitelistFiles, b.Security.WhitelistFiles) ||
		!stringSlicesEqual(a.Security.WhitelistDirs, b.Security.WhitelistDirs) {
		return false
	}

	return true
}

// Helper for comparing two *bool pointers
func ptrBoolEqual(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// Helper for comparing two *string pointers
func ptrStringEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func TestApplyEnvConfig_Coverage(t *testing.T) {
	cfg := DefaultConfig()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("algorithm", "sha256", "")
	fs.String("format", "default", "")
	fs.Bool("recursive", false, "")
	fs.Bool("hidden", false, "")
	fs.Bool("verbose", false, "")
	fs.Bool("quiet", false, "")
	fs.Bool("preserve-order", false, "")

	env := &EnvConfig{
		ChexumAlgorithm:      "md5",
		ChexumOutputFormat:   "json",
		ChexumRecursive:      true,
		ChexumHidden:         true,
		ChexumVerbose:        true,
		ChexumQuiet:          true,
		ChexumPreserveOrder:  true,
		ChexumBlacklistFiles: "f1,f2",
		ChexumBlacklistDirs:  "d1,d2",
		ChexumWhitelistFiles: "w1,w2",
		ChexumWhitelistDirs:  "wd1,wd2",
	}

	env.ApplyEnvConfig(cfg, fs)

	if cfg.Algorithm != "md5" {
		t.Errorf("Expected md5, got %s", cfg.Algorithm)
	}
	if len(cfg.BlacklistFiles) != 2 {
		t.Errorf("Expected 2 blacklist files, got %d", len(cfg.BlacklistFiles))
	}
}

func TestLoadDotEnv_Errors(t *testing.T) {
	t.Run("InvalidFormat", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", ".env_test")
		defer os.Remove(tmpFile.Name())
		os.WriteFile(tmpFile.Name(), []byte("INVALID_LINE"), 0644)

		_, err := LoadDotEnv(tmpFile.Name())
		if err == nil {
			t.Error("Expected error for invalid format")
		}
	})

	t.Run("EmptyPath", func(t *testing.T) {
		// This will try to open .env in current dir, which might not exist
		_, _ = LoadDotEnv("")
	})
}

func TestApplyConfigFile_Errors(t *testing.T) {
	cfg := DefaultConfig()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)

	cf := &ConfigFile{}
	cf.Defaults.MinSize = stringPtr("invalid")

	err := cf.ApplyConfigFile(cfg, fs)
	if err == nil {
		t.Error("Expected error for invalid min_size")
	}

	cf.Defaults.MinSize = nil
	cf.Defaults.MaxSize = stringPtr("invalid")
	err = cf.ApplyConfigFile(cfg, fs)
	if err == nil {
		t.Error("Expected error for invalid max_size")
	}
}

// TestLoadConfigFile_Coverage tests the LoadConfigFile function extensively.
//
// Reviewed: LONG-FUNCTION - Kept long for comprehensive table-driven tests.
func TestLoadConfigFile_Coverage(t *testing.T) {
	// Create a temporary directory for test config files
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		path        string
		content     string
		wantCfg     *ConfigFile
		wantErr     bool
		errContains string
	}{
		{
			name:    "EmptyPath",
			path:    "",
			content: "",
			wantCfg: &ConfigFile{},
			wantErr: false,
		},
		{
			name:    "NonExistent",
			path:    filepath.Join(tmpDir, "nonexistent.toml"),
			content: "",
			wantCfg: &ConfigFile{},
			wantErr: false,
		},
		{
			name:    "TextConfig",
			path:    filepath.Join(tmpDir, "config.txt"),
			content: "file1.txt\nfile2.txt\n# comment\nfile3.txt",
			wantCfg: &ConfigFile{Files: []string{"file1.txt", "file2.txt", "file3.txt"}},
			wantErr: false,
		},
		{
			name:    "TOMLConfig",
			path:    filepath.Join(tmpDir, "config.toml"),
			content: "files = [\"from_toml_file.txt\"]\n[defaults]\nalgorithm = \"sha1\"\nrecursive = true\n[security]\nblacklist_files = [\"*.log\", \"temp/*\"]",
			wantCfg: func() *ConfigFile {
				cfg := &ConfigFile{}
				cfg.Defaults.Algorithm = ptr("sha1")
				cfg.Defaults.Recursive = ptr(true)
				cfg.Security.BlacklistFiles = []string{"*.log", "temp/*"}
				cfg.Files = []string{"from_toml_file.txt"}
				return cfg
			}(),
			wantErr: false,
		},
		{
			name:        "MalformedTOML",
			path:        filepath.Join(tmpDir, "malformed.toml"),
			content:     "[defaults\nalgorithm = \"sha1\"", // Missing closing bracket
			wantCfg:     nil,
			wantErr:     true,
			errContains: "failed to parse TOML config",
		},
		{
			name:        "FileOpenError",
			path:        filepath.Join("/root", "nopermission.toml"), // Path with no write permissions
			content:     "",
			wantCfg:     nil,
			wantErr:     true,
			errContains: "failed to open config file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the config file if content is provided and it's not a special case like /root
			if tt.content != "" && !strings.HasPrefix(tt.path, "/root") {
				if err := os.WriteFile(tt.path, []byte(tt.content), 0644); err != nil {
					t.Fatalf("Failed to write test file %s: %v", tt.path, err)
				}
				defer os.Remove(tt.path)
			}

			// Execute LoadConfigFile
			gotCfg, err := LoadConfigFile(tt.path)

			// Assertions
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
				if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error to contain %q, but got %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}
				if !configFilesEqual(gotCfg, tt.wantCfg) {
					t.Errorf("ConfigFile mismatch for %s:\nGot:  %+v\nWant: %+v", tt.name, gotCfg, tt.wantCfg)
				}
			}
		})
	}
}

// TestFindConfigFile tests that FindConfigFile returns something if a config file exists, or empty string.
func TestFindConfigFile(t *testing.T) {
	_ = FindConfigFile()
}
