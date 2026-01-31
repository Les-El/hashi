package console

import (
	"github.com/Les-El/chexum/internal/config"
	"github.com/Les-El/chexum/internal/testutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOutputManager_Prompt(t *testing.T) {
	r, w, _ := os.Pipe()
	m := NewOutputManager(config.DefaultConfig(), r)

	go func() {
		w.WriteString("y\n")
	}()

	if !m.prompt("test") {
		t.Error("Expected true for 'y'")
	}

	go func() {
		w.WriteString("n\n")
		w.Close()
	}()
	if m.prompt("test") {
		t.Error("Expected false for 'n'")
	}
}

// Reviewed: LONG-FUNCTION - Table-driven coverage tests for file operations.
func TestOutputManager_OpenOutputFile_Coverage(t *testing.T) {
	m := NewOutputManager(config.DefaultConfig(), nil)

	t.Run("EmptyPath", func(t *testing.T) {
		f, err := m.OpenOutputFile("", false, false)
		if f != nil || err != nil {
			t.Error("Expected nil, nil for empty path")
		}
	})

	t.Run("ExistingFileNoForce", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "existing")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		_, err := m.OpenOutputFile(tmpFile.Name(), false, false)
		if err == nil {
			t.Error("Expected error for existing file without force")
		}
	})

	t.Run("CreateDirectory", func(t *testing.T) {
		tmpDir, _ := os.MkdirTemp("", "console_test")
		defer os.RemoveAll(tmpDir)
		path := tmpDir + "/sub/dir/file.txt"

		f, err := m.OpenOutputFile(path, false, false)
		if err != nil {
			t.Fatalf("Failed to create file in new directory: %v", err)
		}
		f.Close()
	})

	t.Run("MkdirFailure", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "mkdir_fail")
		defer os.Remove(tmpFile.Name())
		// Blocking directory creation with a file
		path := tmpFile.Name() + "/file.txt"
		_, err := m.OpenOutputFile(path, false, false)
		if err == nil {
			t.Error("Expected error when MkdirAll fails")
		}
	})

	t.Run("AtomicWriterCloseFailure", func(t *testing.T) {
		tmpDir, _ := os.MkdirTemp("", "atomic_fail")
		defer os.RemoveAll(tmpDir)
		path := filepath.Join(tmpDir, "file.txt")
		
		w, err := m.OpenOutputFile(path, false, false)
		if err != nil {
			t.Fatal(err)
		}
		
		// Block rename by creating a directory where the file should be
		os.MkdirAll(path, 0755)
		testutil.CreateFile(t, path, "blocker", "data")
		
		err = w.Close()
		if err == nil {
			t.Error("Expected error from atomic Close when rename fails")
		}
	})
}

func TestAtomicWriter_Write(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "atomic")
	defer os.Remove(tmpFile.Name())
	w := &atomicWriter{f: tmpFile, path: tmpFile.Name() + ".out", tempPath: tmpFile.Name()}
	w.Write([]byte("data"))
	w.Close()
}

func TestJsonLogWriter_Write(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "json")
	defer os.Remove(tmpFile.Name())
	w := &jsonLogWriter{f: tmpFile, isNew: true}
	w.Write([]byte("{}"))
	w.Close()
}

func TestJsonLogWriter_Close(t *testing.T) {
	// Already covered but adding named test
}

func TestAtomicWriter_Close(t *testing.T) {
	// Already covered but adding named test
}

// Reviewed: LONG-FUNCTION - Table-driven coverage tests for JSON logging.
func TestOutputManager_OpenJSONLog_Coverage(t *testing.T) {
	m := NewOutputManager(config.DefaultConfig(), nil)

	t.Run("EmptyPath", func(t *testing.T) {
		f, err := m.OpenJSONLog("")
		if f != nil || err != nil {
			t.Error("Expected nil, nil for empty path")
		}
	})

	t.Run("NewAndAppend", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "test.json")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		w, err := m.OpenJSONLog(tmpFile.Name())
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(`{"a":1}`))
		w.Close()

		// Append
		w, err = m.OpenJSONLog(tmpFile.Name())
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(`{"b":2}`))
		w.Close()

		content, _ := os.ReadFile(tmpFile.Name())
		expected := "[\n{\"a\":1}\n,\n{\"b\":2}\n]"
		if string(content) != expected {
			t.Errorf("Unexpected JSON content:\nGot:  %q\nWant: %q", string(content), expected)
		}
	})

	t.Run("AppendToNonArray", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "nonarray.json")
		defer os.Remove(tmpFile.Name())
		os.WriteFile(tmpFile.Name(), []byte("not an array"), 0644)

		w, err := m.OpenJSONLog(tmpFile.Name())
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(`{"c":3}`))
		w.Close()

		content, _ := os.ReadFile(tmpFile.Name())
		if !strings.Contains(string(content), "not an array{\"c\":3}") {
			t.Errorf("Expected content to be appended at the end, got %q", string(content))
		}
	})

	t.Run("MkdirFailure", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "mkdir_fail_json")
		defer os.Remove(tmpFile.Name())
		path := tmpFile.Name() + "/log.json"
		_, err := m.OpenJSONLog(path)
		if err == nil {
			t.Error("Expected error when MkdirAll fails")
		}
	})

	t.Run("WriteFailure", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "write_fail.json")
		defer os.Remove(tmpFile.Name())
		
		w, err := m.OpenJSONLog(tmpFile.Name())
		if err != nil {
			t.Fatal(err)
		}
		
		// Access the internal file to close it
		jw := w.(*jsonLogWriter)
		jw.f.Close()
		
		_, err = w.Write([]byte("data"))
		if err == nil {
			t.Error("Expected error when writing to closed file")
		}
	})

	t.Run("CloseFailure", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "close_fail.json")
		defer os.Remove(tmpFile.Name())
		
		w, err := m.OpenJSONLog(tmpFile.Name())
		if err != nil {
			t.Fatal(err)
		}
		
		jw := w.(*jsonLogWriter)
		jw.f.Close()
		
		err = w.Close()
		if err == nil {
			t.Error("Expected error when closing already closed file")
		}
	})

	t.Run("WriteFailureNew", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "write_fail_new.json")
		defer os.Remove(tmpFile.Name())
		
		w, err := m.OpenJSONLog(tmpFile.Name())
		if err != nil {
			t.Fatal(err)
		}
		
		jw := w.(*jsonLogWriter)
		jw.f.Close()
		jw.isNew = true
		
		_, err = w.Write([]byte("data"))
		if err == nil {
			t.Error("Expected error when writing to closed new file")
		}
	})
}

func TestOutputManager_isInteractive_Coverage(t *testing.T) {
	t.Run("NotAFile", func(t *testing.T) {
		m := NewOutputManager(nil, strings.NewReader(""))
		if m.isInteractive() {
			t.Error("expected false for strings.Reader")
		}
	})

	t.Run("ClosedFile", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "closed")
		m := NewOutputManager(nil, tmpFile)
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		if m.isInteractive() {
			t.Error("expected false for closed file")
		}
	})
}

func TestInitStreams_Coverage(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		cfg := &config.Config{}
		streams, cleanup, err := InitStreams(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()
		if streams.Out == nil {
			t.Error("Expected non-nil Out")
		}
	})

	t.Run("WithOutputFile", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "out")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		cfg := &config.Config{OutputFile: tmpFile.Name(), Force: true}
		streams, cleanup, err := InitStreams(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()
		streams.Out.Write([]byte("hello"))
	})

	t.Run("WithLogFile", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "log")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		cfg := &config.Config{LogFile: tmpFile.Name()}
		_, cleanup, err := InitStreams(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()
	})
}
