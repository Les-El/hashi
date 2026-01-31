package config

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
)

func TestMultiParser(t *testing.T) {
	cfg := DefaultConfig()
	p1 := &mockParser{val: "p1"}
	p2 := &mockParser{val: "p2"}

	mp := &MultiParser{
		Parsers: []Parser{p1, p2},
	}

	if err := mp.Parse(cfg); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if cfg.Algorithm != "p2" {
		t.Errorf("expected algorithm to be p2 (last parser wins if no pflag.Changed check), got %s", cfg.Algorithm)
	}
}

func TestParse(t *testing.T) { TestMultiParser(t) }

type mockParser struct {
	val string
}

func (m *mockParser) Parse(cfg *Config) error {
	cfg.Algorithm = m.val
	return nil
}

func TestCLIParser(t *testing.T) {
	cfg := DefaultConfig()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	args := []string{"--algorithm", "md5", "cli.go"}

	parser := NewCLIParser(args, fs)
	if err := parser.Parse(cfg); err != nil {
		t.Fatalf("CLIParser.Parse failed: %v", err)
	}

	if cfg.Algorithm != "md5" {
		t.Errorf("expected algorithm md5, got %s", cfg.Algorithm)
	}

	if len(cfg.Files) != 1 || cfg.Files[0] != "cli.go" {
		t.Errorf("expected files [cli.go], got %v", cfg.Files)
	}
}

func TestEnvParser(t *testing.T) {
	cfg := DefaultConfig()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	extra := map[string]string{
		"CHEXUM_ALGORITHM": "sha1",
	}

	parser := NewEnvParser(extra, fs)
	if err := parser.Parse(cfg); err != nil {
		t.Fatalf("EnvParser.Parse failed: %v", err)
	}

	if cfg.Algorithm != "sha1" {
		t.Errorf("expected algorithm sha1, got %s", cfg.Algorithm)
	}
}

func TestFileParser(t *testing.T) {
	cfg := DefaultConfig()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	defineFlags(fs, cfg)

	// Create a dummy config file
	tmpFile, err := os.CreateTemp("", "testconfig*.toml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `
[defaults]
algorithm = "md5"
recursive = true
`
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	parser := NewFileParser(tmpFile.Name(), fs)
	if err := parser.Parse(cfg); err != nil {
		t.Fatalf("FileParser.Parse failed: %v", err)
	}

	if cfg.Algorithm != "md5" {
		t.Errorf("expected algorithm md5 from config file, got %s", cfg.Algorithm)
	}
	if !cfg.Recursive {
		t.Error("expected recursive true from config file")
	}
}
