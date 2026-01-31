package checkpoint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverPackageFiles(t *testing.T) {
	tmpDir := t.TempDir()
	pkgDir := filepath.Join(tmpDir, "internal", "testpkg")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(pkgDir, "file.go"), []byte("package testpkg"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := discoverPackageFiles(tmpDir, "internal/testpkg")
	if err != nil {
		t.Fatalf("discoverPackageFiles failed: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}

	// Error path
	_, err = discoverPackageFiles(tmpDir, "non-existent")
	if err == nil {
		t.Error("expected error for non-existent package")
	}
}

func TestDiscoverCorePackages_NoInternal(t *testing.T) {
	tmpDir := t.TempDir()
	pkgs, err := discoverCorePackages(tmpDir)
	if err != nil {
		t.Fatalf("discoverCorePackages failed: %v", err)
	}
	if len(pkgs) != 0 {
		t.Errorf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestDiscoverPackageByName_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := discoverPackageByName(tmpDir, "non-existent")
	if err == nil {
		t.Error("expected error for non-existent package")
	}
}
