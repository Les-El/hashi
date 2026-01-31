package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

// TestApplyEnvConfig tests the ApplyEnvConfig function.
func TestApplyEnvConfig(t *testing.T) {
	t.Run("ApplyDefaults", func(t *testing.T) {
		os.Setenv("CHEXUM_ALGORITHM", "md5")
		os.Setenv("CHEXUM_RECURSIVE", "true")
		os.Setenv("CHEXUM_BLACKLIST_FILES", "file1.log,file2.tmp")
		defer os.Clearenv()

		cfg := DefaultConfig()
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("algorithm", "", "")
		fs.Bool("recursive", false, "")
		fs.StringSlice("blacklist-files", []string{}, "")

		envCfg := LoadEnvConfig(nil)
		envCfg.ApplyEnvConfig(cfg, fs)

		if cfg.Algorithm != "md5" {
			t.Errorf("Expected algorithm to be md5, got %s", cfg.Algorithm)
		}
		if !cfg.Recursive {
			t.Errorf("Expected recursive to be true, got %v", cfg.Recursive)
		}
		if len(cfg.BlacklistFiles) != 2 {
			t.Errorf("Expected 2 blacklist files, got %v", cfg.BlacklistFiles)
		}
	})

	t.Run("FlagsOverrideEnv", func(t *testing.T) {
		os.Setenv("CHEXUM_ALGORITHM", "md5")
		defer os.Clearenv()

		cfg := DefaultConfig()
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("algorithm", "sha1", "")
		fs.Set("algorithm", "sha1")
		cfg.Algorithm = "sha1"

		envCfg := LoadEnvConfig(nil)
		envCfg.ApplyEnvConfig(cfg, fs)

		if cfg.Algorithm != "sha1" {
			t.Errorf("Expected algorithm to be sha1 (from flag), got %s", cfg.Algorithm)
		}
	})
}

func TestLoadEnvConfig(t *testing.T) {
	env := LoadEnvConfig(nil)
	if env == nil {
		t.Error("LoadEnvConfig() returned nil")
	}
}

func TestLoadDotEnv(t *testing.T) {
	tempDir := t.TempDir()
	dotEnvPath := filepath.Join(tempDir, ".env")

	t.Run("ValidDotEnvFile", func(t *testing.T) {
		content := []byte("KEY1=value1\nKEY2=\"value 2\"\n#comment\nKEY3='value3'")
		err := os.WriteFile(dotEnvPath, content, 0644)
		if err != nil {
			t.Fatalf("Failed to create .env file: %v", err)
		}

		os.Clearenv()
		vars, err := LoadDotEnv(dotEnvPath)
		if err != nil {
			t.Errorf("LoadDotEnv() error = %v", err)
		}

		if vars["KEY1"] != "value1" {
			t.Errorf("KEY1 = %s, want value1", vars["KEY1"])
		}
	})

	t.Run("NonExistentDotEnvFile", func(t *testing.T) {
		os.Clearenv()
		_, err := LoadDotEnv("non_existent.env")
		if err != nil {
			t.Errorf("LoadDotEnv() error for non-existent file = %v", err)
		}
	})

	t.Run("InvalidFormatDotEnvFile", func(t *testing.T) {
		content := []byte("KEY1=value1\nINVALID_LINE\nKEY2=value2")
		os.WriteFile(dotEnvPath, content, 0644)
		defer os.Remove(dotEnvPath)

		os.Clearenv()
		_, err := LoadDotEnv(dotEnvPath)
		if err == nil || !strings.Contains(err.Error(), "invalid format") {
			t.Errorf("LoadDotEnv() expected error for invalid format, got %v", err)
		}
	})
}

func TestNewEnvParser(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	p := NewEnvParser(nil, fs)
	if p == nil {
		t.Fatal("NewEnvParser returned nil")
	}
}
