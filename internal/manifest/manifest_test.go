package manifest

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Les-El/chexum/internal/hash"
)

func TestNew(t *testing.T) {
	entries := []hash.Entry{
		{Original: "f1.txt", Hash: "h1", Size: 10, ModTime: time.Now()},
	}
	m := New("sha256", entries)
	if len(m.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(m.Files))
	}
}

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")
	
	m := New("sha256", nil)
	Save(m, manifestPath)

	m2, err := Load(manifestPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if m2.Algorithm != "sha256" {
		t.Errorf("Expected sha256, got %s", m2.Algorithm)
	}

	t.Run("NonExistent", func(t *testing.T) {
		_, err := Load("missing.json")
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Malformed", func(t *testing.T) {
		malformedPath := filepath.Join(tmpDir, "malformed.json")
		os.WriteFile(malformedPath, []byte("{"), 0644)
		_, err := Load(malformedPath)
		if err == nil {
			t.Error("Expected error")
		}
	})
}

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	m := New("sha256", nil)
	
	t.Run("Success", func(t *testing.T) {
		path := filepath.Join(tmpDir, "manifest.json")
		if err := Save(m, path); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	})

	t.Run("MkdirFailure", func(t *testing.T) {
		blockedPath := filepath.Join(tmpDir, "blocked")
		os.WriteFile(blockedPath, []byte("junk"), 0644)
		path := filepath.Join(blockedPath, "manifest.json")
		if err := Save(m, path); err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Readonly", func(t *testing.T) {
		roDir := filepath.Join(tmpDir, "readonly")
		os.Mkdir(roDir, 0555)
		path := filepath.Join(roDir, "manifest.json")
		err := Save(m, path)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("TempFileFailure", func(t *testing.T) {
		path := filepath.Join(tmpDir, "fail.json")
		// Block the temp path with a directory
		os.MkdirAll(path+".tmp", 0755)
		err := Save(m, path)
		if err == nil {
			t.Error("Expected error when temp file creation fails")
		}
	})

	t.Run("RenameFailure", func(t *testing.T) {
		path := filepath.Join(tmpDir, "rename_fail")
		// Block rename by creating a directory at the target path
		os.MkdirAll(path, 0755)
		err := Save(m, path)
		if err == nil {
			t.Error("Expected error when rename fails")
		}
	})
}

func TestGetChangedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "f1.txt")
	os.WriteFile(f1, []byte("c1"), 0644)
	stat1, _ := os.Stat(f1)

	m := New("sha256", []hash.Entry{
		{Original: f1, Hash: "h1", Size: stat1.Size(), ModTime: stat1.ModTime()},
	})

	t.Run("Unchanged", func(t *testing.T) {
		changed, _ := m.GetChangedFiles([]string{f1})
		if len(changed) != 0 {
			t.Errorf("Expected 0 changed, got %v", changed)
		}
	})

	t.Run("Modified", func(t *testing.T) {
		os.WriteFile(f1, []byte("c1-mod"), 0644)
		os.Chtimes(f1, time.Now().Add(time.Hour), time.Now().Add(time.Hour))
		changed, _ := m.GetChangedFiles([]string{f1})
		if len(changed) != 1 {
			t.Errorf("Expected 1 changed, got %v", changed)
		}
	})

	t.Run("ModifiedSize", func(t *testing.T) {
		os.WriteFile(f1, []byte("longer content"), 0644)
		changed, _ := m.GetChangedFiles([]string{f1})
		if len(changed) != 1 {
			t.Errorf("Expected 1 changed (size), got %v", changed)
		}
	})

	t.Run("Missing", func(t *testing.T) {
		missing := filepath.Join(tmpDir, "missing.txt")
		changed, _ := m.GetChangedFiles([]string{missing})
		if len(changed) != 1 {
			t.Errorf("Expected 1 changed (missing), got %v", changed)
		}
	})

	t.Run("Directory", func(t *testing.T) {
		subDir := filepath.Join(tmpDir, "subdir")
		os.Mkdir(subDir, 0755)
		changed, _ := m.GetChangedFiles([]string{subDir})
		if len(changed) != 0 {
			t.Errorf("Expected 0 changed (dir), got %v", changed)
		}
	})
}

func TestManifest(t *testing.T) {
	// Wrapper test to keep original entry point if needed
	t.Run("New", TestNew)
	t.Run("Load", TestLoad)
	t.Run("Save", TestSave)
	t.Run("GetChangedFiles", TestGetChangedFiles)
}