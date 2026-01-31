package console

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/quick"

	"github.com/Les-El/chexum/internal/config"
)

func TestOpenOutputFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()

	t.Run("creates new file", func(t *testing.T) {
		testCreateNewFile(t, tmpDir, cfg)
	})

	t.Run("fails if file exists without force", func(t *testing.T) {
		testFailsIfExistsWithoutForce(t, tmpDir, cfg)
	})

	t.Run("overwrites if force is true", func(t *testing.T) {
		testOverwritesIfForceIsTrue(t, tmpDir, cfg)
	})

	t.Run("appends if append is true", func(t *testing.T) {
		testAppendsIfAppendIsTrue(t, tmpDir, cfg)
	})
}

func testCreateNewFile(t *testing.T, tmpDir string, cfg *config.Config) {
	manager := NewOutputManager(cfg, nil)
	path := filepath.Join(tmpDir, "new.txt")
	if f, err := manager.OpenOutputFile(path, false, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		f.Close()
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file was not created")
	}
}

func testFailsIfExistsWithoutForce(t *testing.T, tmpDir string, cfg *config.Config) {
	manager := NewOutputManager(cfg, strings.NewReader(""))
	path := filepath.Join(tmpDir, "exists.txt")
	os.WriteFile(path, []byte("content"), 0644)
	if _, err := manager.OpenOutputFile(path, false, false); err == nil {
		t.Error("expected error")
	}
}

func testOverwritesIfForceIsTrue(t *testing.T, tmpDir string, cfg *config.Config) {
	manager := NewOutputManager(cfg, nil)
	path := filepath.Join(tmpDir, "overwrite.txt")
	os.WriteFile(path, []byte("old"), 0644)
	if f, err := manager.OpenOutputFile(path, false, true); err == nil {
		f.Write([]byte("new"))
		f.Close()
	}
	content, _ := os.ReadFile(path)
	if string(content) != "new" {
		t.Errorf("got %q", string(content))
	}
}

func testAppendsIfAppendIsTrue(t *testing.T, tmpDir string, cfg *config.Config) {
	manager := NewOutputManager(cfg, nil)
	path := filepath.Join(tmpDir, "append.txt")
	os.WriteFile(path, []byte("first "), 0644)
	if f, err := manager.OpenOutputFile(path, true, false); err == nil {
		f.Write([]byte("second"))
		f.Close()
	}
	content, _ := os.ReadFile(path)
	if string(content) != "first second" {
		t.Errorf("got %q", string(content))
	}
}

func TestNewOutputManager(t *testing.T) {
	cfg := config.DefaultConfig()
	manager := NewOutputManager(cfg, nil)
	if manager == nil {
		t.Fatal("NewOutputManager returned nil")
	}
}

func TestOpenJSONLog(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	manager := NewOutputManager(cfg, nil)
	path := filepath.Join(tmpDir, "log.json")
	f, err := manager.OpenJSONLog(path)
	if err != nil {
		t.Fatalf("OpenJSONLog failed: %v", err)
	}
	f.Close()
}

// TestProperty_AppendModePreservesContent verifies that append mode preserves existing content.
// Property 12: Append mode preserves existing content
func TestProperty_AppendModePreservesContent(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "chexum-prop-*")
	defer os.RemoveAll(tmpDir)

	cfg := config.DefaultConfig()
	manager := NewOutputManager(cfg, nil)

	f := func(initial, addition string) bool {
		path := filepath.Join(tmpDir, "prop_append.txt")
		os.Remove(path)

		// Write initial
		os.WriteFile(path, []byte(initial), 0644)

		// Append
		f, err := manager.OpenOutputFile(path, true, false)
		if err != nil {
			return false
		}
		f.Write([]byte(addition))
		f.Close()

		content, _ := os.ReadFile(path)
		return string(content) == initial+addition
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestProperty_AtomicWritesPreserveOriginal verifies that the original file is preserved if write fails.
// Property 48: Atomic writes preserve original on failure
func TestProperty_AtomicWritesPreserveOriginal(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	manager := NewOutputManager(cfg, nil)

	f := func(original, attempt string) bool {
		path := filepath.Join(tmpDir, "atomic_test.txt")
		os.WriteFile(path, []byte(original), 0644)

		// Open for atomic write (append=false, force=true)
		w, err := manager.OpenOutputFile(path, false, true)
		if err != nil {
			return false
		}

		// Write some data but DON'T Close() yet
		w.Write([]byte(attempt))

		// Check original file - it should still be "original" because rename hasn't happened
		current, _ := os.ReadFile(path)
		if string(current) != original {
			w.Close()
			return false
		}

		// Now close to finalize
		w.Close()

		// Now it should be "attempt"
		final, _ := os.ReadFile(path)
		return string(final) == attempt
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestProperty_JSONLogMaintainsValidity verifies that appending to JSON log produces valid JSON.
// Property 50: JSON log append maintains validity
func TestProperty_JSONLogMaintainsValidity(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	manager := NewOutputManager(cfg, nil)
	path := filepath.Join(tmpDir, "validity.json")

	// Helper to check if file is valid JSON array
	isValidJSON := func(p string) bool {
		data, err := os.ReadFile(p)
		if err != nil {
			return false
		}
		var arr []interface{}
		return json.Unmarshal(data, &arr) == nil
	}

	// Start with empty file
	os.Remove(path)

	// Add entries one by one
	for i := 0; i < 10; i++ {
		w, err := manager.OpenJSONLog(path)
		if err != nil {
			t.Fatalf("OpenJSONLog failed at step %d: %v", i, err)
		}
		entry := fmt.Sprintf(`{"id": %d}`, i)
		w.Write([]byte(entry))
		w.Close()

		if !isValidJSON(path) {
			t.Errorf("JSON invalid at step %d", i)
			data, _ := os.ReadFile(path)
			t.Logf("Content: %s", string(data))
			break
		}
	}
}
