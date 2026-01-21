package hash

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hashi-discovery-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test structure
	// tmpDir/
	//   file1.txt
	//   .hidden_file
	//   sub/
	//     file2.txt
	//     .hidden_sub/
	//       file3.txt
	
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".hidden_file"), []byte("hidden"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "sub"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "sub", "file2.txt"), []byte("2"), 0644)
	os.Mkdir(filepath.Join(tmpDir, ".hidden_sub"), 0755)
	os.WriteFile(filepath.Join(tmpDir, ".hidden_sub", "file3.txt"), []byte("3"), 0644)

	t.Run("default discovery (non-recursive, no hidden)", func(t *testing.T) {
		opts := DiscoveryOptions{Recursive: false, Hidden: false, MaxSize: -1}
		files, err := DiscoverFiles([]string{tmpDir}, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		// Should find only file1.txt
		if len(files) != 1 {
			t.Errorf("expected 1 file, got %d: %v", len(files), files)
		} else if filepath.Base(files[0]) != "file1.txt" {
			t.Errorf("expected file1.txt, got %s", files[0])
		}
	})

	t.Run("recursive discovery", func(t *testing.T) {
		opts := DiscoveryOptions{Recursive: true, Hidden: false, MaxSize: -1}
		files, err := DiscoverFiles([]string{tmpDir}, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		// Should find file1.txt and file2.txt
		if len(files) != 2 {
			t.Errorf("expected 2 files, got %d: %v", len(files), files)
		}
	})

	t.Run("include hidden files", func(t *testing.T) {
		opts := DiscoveryOptions{Recursive: true, Hidden: true, MaxSize: -1}
		files, err := DiscoverFiles([]string{tmpDir}, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		// Should find file1.txt, .hidden_file, file2.txt, file3.txt
		if len(files) != 4 {
			t.Errorf("expected 4 files, got %d: %v", len(files), files)
		}
	})
}
