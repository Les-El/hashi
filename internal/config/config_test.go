package config

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}
	if cfg.Algorithm != "sha256" {
		t.Errorf("expected sha256, got %s", cfg.Algorithm)
	}
}

func TestHasStdinMarker(t *testing.T) {
	cfg := &Config{Files: []string{"a.txt", "-", "b.txt"}}
	if !cfg.HasStdinMarker() {
		t.Error("expected true")
	}
	cfg.Files = []string{"a.txt"}
	if cfg.HasStdinMarker() {
		t.Error("expected false")
	}
}

func TestFilesWithoutStdin(t *testing.T) {
	cfg := &Config{Files: []string{"a.txt", "-", "b.txt"}}
	res := cfg.FilesWithoutStdin()
	if len(res) != 2 || res[0] != "a.txt" || res[1] != "b.txt" {
		t.Errorf("unexpected result: %v", res)
	}
}
