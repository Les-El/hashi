package config

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
)

func TestApplyConfigFile(t *testing.T) {
	t.Run("ApplyBoolDefaults", func(t *testing.T) {
		cfg := DefaultConfig()
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.Bool("recursive", false, "")
		cf := &ConfigFile{}
		cf.Defaults.Recursive = ptr(true)

		_ = cf.ApplyConfigFile(cfg, fs)
		if !cfg.Recursive {
			t.Errorf("Expected Recursive to be true, got %v", cfg.Recursive)
		}
	})

	t.Run("ApplyStringDefaults", func(t *testing.T) {
		cfg := DefaultConfig()
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("algorithm", "", "")
		cf := &ConfigFile{}
		cf.Defaults.Algorithm = ptr("md5")

		_ = cf.ApplyConfigFile(cfg, fs)
		if cfg.Algorithm != "md5" {
			t.Errorf("Expected Algorithm to be md5, got %s", cfg.Algorithm)
		}
	})
}

func TestLoadConfigFile(t *testing.T) {
	path := "test_config_simple.toml"
	content := "[defaults]\nalgorithm = \"md5\""
	os.WriteFile(path, []byte(content), 0644)
	defer os.Remove(path)

	cfg, err := LoadConfigFile(path)
	if err != nil {
		t.Errorf("LoadConfigFile() error = %v", err)
	}

	if *cfg.Defaults.Algorithm != "md5" {
		t.Errorf("Algorithm = %s, want md5", *cfg.Defaults.Algorithm)
	}
}

func ptr[T any](v T) *T {
	return &v
}

func TestNewFileParser(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	p := NewFileParser("config.toml", fs)
	if p == nil {
		t.Fatal("NewFileParser returned nil")
	}
}

func TestFileParser_Parse(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	defineFlags(fs, &Config{}) // define flags for ApplyConfigFile to find them

	path := "test_parser.toml"
	os.WriteFile(path, []byte("[defaults]\nrecursive = true"), 0644)
	defer os.Remove(path)

	p := NewFileParser(path, fs)
	cfg := DefaultConfig()
	if err := p.Parse(cfg); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !cfg.Recursive {
		t.Error("expected recursive=true")
	}
}
