package checkpoint

import (
	"os"
	"testing"

	"github.com/spf13/afero"
)

// Reviewed: LONG-FUNCTION - Integration test with multiple file system assertions and subtests.
func TestNewWorkspace(t *testing.T) {
	t.Run("Mem", func(t *testing.T) {
		ws, err := NewWorkspace(true)
		if err != nil {
			t.Fatalf("failed to create mem workspace: %v", err)
		}

		content := []byte("hello mem")
		if err := ws.WriteFile("test.txt", content); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		read, err := ws.ReadFile("test.txt")
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		if string(read) != string(content) {
			t.Errorf("expected %q, got %q", string(content), string(read))
		}

		if err := ws.Cleanup(); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}

		if ws.Fs != nil {
			t.Error("expected Fs to be nil after cleanup")
		}
	})

	t.Run("Disk", func(t *testing.T) {
		ws, err := NewWorkspace(false)
		if err != nil {
			t.Fatalf("failed to create disk workspace: %v", err)
		}
		defer ws.Cleanup()

		if ws.Root == "" {
			t.Fatal("expected Root path to be set")
		}

		if _, ok := ws.Fs.(*afero.OsFs); !ok {
			t.Error("expected OsFs")
		}

		content := []byte("hello disk")
		if err := ws.WriteFile("test.txt", content); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		// Verify file exists on actual disk
		path := ws.Path("test.txt")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("file should exist at %s", path)
		}

		read, err := ws.ReadFile("test.txt")
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		if string(read) != string(content) {
			t.Errorf("expected %q, got %q", string(content), string(read))
		}

		root := ws.Root
		if err := ws.Cleanup(); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}

		if _, err := os.Stat(root); !os.IsNotExist(err) {
			t.Errorf("expected root dir %s to be removed", root)
		}
	})
}

func TestWorkspace_Cleanup(t *testing.T) {
	ws, _ := NewWorkspace(false)
	root := ws.Root
	if err := ws.Cleanup(); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Errorf("expected root dir %s to be removed", root)
	}
}

func TestWorkspace_WriteFile(t *testing.T) {
	ws, _ := NewWorkspace(true)
	content := []byte("test")
	if err := ws.WriteFile("test.txt", content); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	read, _ := ws.ReadFile("test.txt")
	if string(read) != string(content) {
		t.Errorf("expected %q, got %q", string(content), string(read))
	}

	t.Run("Nested", func(t *testing.T) {
		if err := ws.WriteFile("a/b/c/test.txt", content); err != nil {
			t.Fatalf("nested write failed: %v", err)
		}
		read, _ := ws.ReadFile("a/b/c/test.txt")
		if string(read) != string(content) {
			t.Errorf("expected %q, got %q", string(content), string(read))
		}
	})
}

func TestWorkspace_ReadFile(t *testing.T) {
	ws, _ := NewWorkspace(true)
	ws.WriteFile("test.txt", []byte("test"))
	read, err := ws.ReadFile("test.txt")
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(read) != "test" {
		t.Errorf("expected test, got %s", string(read))
	}

	t.Run("Missing", func(t *testing.T) {
		_, err := ws.ReadFile("missing.txt")
		if err == nil {
			t.Error("expected error for missing file")
		}
	})
}

func TestWorkspace_ErrorPaths(t *testing.T) {
	ws, err := NewWorkspace(false)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Cleanup()

	// Write to invalid path
	err = ws.WriteFile("../out_of_ws.txt", []byte("content"))
	if err == nil {
		t.Error("expected error writing out of workspace")
	}

	// Read from non-existent file
	_, err = ws.ReadFile("non-existent.txt")
	if err == nil {
		t.Error("expected error reading non-existent file")
	}
}

func TestWorkspace_Path(t *testing.T) {
	wsMem, _ := NewWorkspace(true)
	pathMem := wsMem.Path("subdir", "file.txt")
	if pathMem != "/subdir/file.txt" {
		t.Errorf("expected /subdir/file.txt, got %s", pathMem)
	}

	wsDisk, _ := NewWorkspace(false)
	defer wsDisk.Cleanup()
	pathDisk := wsDisk.Path("subdir", "file.txt")
	expected := wsDisk.Root + string(os.PathSeparator) + "subdir" + string(os.PathSeparator) + "file.txt"
	if pathDisk != expected {
		t.Errorf("expected %s, got %s", expected, pathDisk)
	}
}
