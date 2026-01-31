package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/Les-El/chexum/internal/color"
	"github.com/Les-El/chexum/internal/config"
	"github.com/Les-El/chexum/internal/console"
	"github.com/Les-El/chexum/internal/errors"
	"github.com/Les-El/chexum/internal/hash"
)

func TestRun_Basic(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name     string
		args     []string
		expected int
	}{
		{"Help", []string{"chexum", "--help"}, config.ExitSuccess},
		{"Version", []string{"chexum", "--version"}, config.ExitSuccess},
		{"InvalidFlag", []string{"chexum", "--no-such-flag"}, config.ExitInvalidArgs},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			if code := run(); code != tt.expected {
				t.Errorf("run() = %d; want %d", code, tt.expected)
			}
		})
	}
}

func TestPrepareFiles_Coverage(t *testing.T) {
	colorHandler := color.NewColorHandler()
	errHandler := errors.NewErrorHandler(colorHandler)
	var outBuf, errBuf bytes.Buffer
	streams := &console.Streams{Out: &outBuf, Err: &errBuf}

	t.Run("EmptyFilesAndNoHashes", func(t *testing.T) {
		cfg := &config.Config{}
		err := prepareFiles(cfg, errHandler, streams)
		if err != nil {
			t.Errorf("prepareFiles() error = %v", err)
		}
	})

	t.Run("WithFiles", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "chexum_test_prepare_*.txt")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		cfg := &config.Config{Files: []string{tmpFile.Name()}}
		err := prepareFiles(cfg, errHandler, streams)
		if err != nil {
			t.Errorf("prepareFiles() error = %v", err)
		}
		if len(cfg.Files) == 0 {
			t.Error("Expected files to be preserved/discovered")
		}
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		cfg := &config.Config{Files: []string{"non_existent_file_12345.txt"}}
		err := prepareFiles(cfg, errHandler, streams)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})
}

func TestExecuteMode_Coverage(t *testing.T) {
	colorHandler := color.NewColorHandler()
	errHandler := errors.NewErrorHandler(colorHandler)
	var outBuf, errBuf bytes.Buffer
	streams := &console.Streams{Out: &outBuf, Err: &errBuf}

	t.Run("MultipleFilesWithHashes", func(t *testing.T) {
		cfg := &config.Config{
			Files:  []string{"file1.txt", "file2.txt"},
			Hashes: []string{"hash1"},
		}
		code := executeMode(cfg, colorHandler, streams, errHandler)
		if code != config.ExitInvalidArgs {
			t.Errorf("executeMode() = %d; want %d", code, config.ExitInvalidArgs)
		}
	})

	t.Run("StdinWithHashes", func(t *testing.T) {
		cfg := &config.Config{
			Files:  []string{"-"},
			Hashes: []string{"hash1"},
		}
		code := executeMode(cfg, colorHandler, streams, errHandler)
		if code != config.ExitInvalidArgs {
			t.Errorf("executeMode() = %d; want %d", code, config.ExitInvalidArgs)
		}
	})

	t.Run("NoFilesNoHashes", func(t *testing.T) {
		cfg := &config.Config{}
		code := executeMode(cfg, colorHandler, streams, errHandler)
		if code != config.ExitSuccess {
			t.Errorf("executeMode() = %d; want %d", code, config.ExitSuccess)
		}
	})
}

func TestExpandStdinFiles_Coverage(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		w.WriteString("file1.txt\n")
		w.WriteString("  file2.txt  \n")
		w.WriteString("\n")
		w.Close()
	}()

	files := []string{"-", "existing.txt"}
	result := expandStdinFiles(files)

	expected := []string{"existing.txt", "file1.txt", "file2.txt"}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d files, got %d", len(expected), len(result))
	}
	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("Expected result[%d] = %s, got %s", i, expected[i], result[i])
		}
	}
}

func TestRunStandardHashingMode_InvalidAlgorithm(t *testing.T) {
	colorHandler := color.NewColorHandler()
	errHandler := errors.NewErrorHandler(colorHandler)
	var outBuf, errBuf bytes.Buffer
	streams := &console.Streams{Out: &outBuf, Err: &errBuf}

	cfg := &config.Config{
		Algorithm: "invalid-alg",
		Files:     []string{"file.txt"},
	}
	code := runStandardHashingMode(cfg, colorHandler, streams, errHandler)
	if code != config.ExitInvalidArgs {
		t.Errorf("Expected ExitInvalidArgs, got %d", code)
	}
}

func TestIsSuccess_Coverage(t *testing.T) {
	tests := []struct {
		name     string
		results  *hash.Result
		cfg      *config.Config
		expected bool
	}{
		{
			"MatchRequired_WithMatches",
			&hash.Result{Matches: []hash.MatchGroup{{}}},
			&config.Config{MatchRequired: true},
			true,
		},
		{
			"MatchRequired_NoMatches",
			&hash.Result{Matches: []hash.MatchGroup{}},
			&config.Config{MatchRequired: true},
			false,
		},
		{
			"SingleFile_Success",
			&hash.Result{Entries: []hash.Entry{{}}, Errors: nil},
			&config.Config{},
			true,
		},
		{
			"SingleFile_Error",
			&hash.Result{Entries: []hash.Entry{{}}, Errors: []error{os.ErrNotExist}},
			&config.Config{},
			false,
		},
		{
			"MultipleFiles_AllMatch",
			&hash.Result{Matches: []hash.MatchGroup{{}}, Unmatched: []hash.Entry{}},
			&config.Config{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSuccess(tt.results, tt.cfg); got != tt.expected {
				t.Errorf("isSuccess() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFormatAlgorithmList_Coverage(t *testing.T) {
	tests := []struct {
		name string
		algs []string
		want string
	}{
		{"Empty", []string{}, ""},
		{"One", []string{"sha256"}, "sha256"},
		{"Two", []string{"md5", "sha1"}, "md5 or sha1"},
		{"Three", []string{"md5", "sha1", "sha256"}, "md5, sha1 or sha256"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatAlgorithmList(tt.algs); got != tt.want {
				t.Errorf("formatAlgorithmList() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReportValidHash_Coverage(t *testing.T) {
	colorHandler := color.NewColorHandler()
	colorHandler.SetEnabled(false)
	var outBuf, errBuf bytes.Buffer
	streams := &console.Streams{Out: &outBuf, Err: &errBuf}

	t.Run("OneAlgorithm", func(t *testing.T) {
		errBuf.Reset()
		reportValidHash("abc", []string{"sha256"}, &config.Config{}, colorHandler, streams)
		if !bytes.Contains(errBuf.Bytes(), []byte("Algorithm: sha256")) {
			t.Errorf("Expected algorithm in output, got %s", errBuf.String())
		}
	})

	t.Run("MultipleAlgorithms", func(t *testing.T) {
		errBuf.Reset()
		reportValidHash("abc", []string{"md5", "sha1"}, &config.Config{}, colorHandler, streams)
		if !bytes.Contains(errBuf.Bytes(), []byte("Possible algorithms: md5 or sha1")) {
			t.Errorf("Expected possible algorithms in output, got %s", errBuf.String())
		}
	})
}

func TestReportInvalidHash_Coverage(t *testing.T) {
	colorHandler := color.NewColorHandler()
	colorHandler.SetEnabled(false)
	var outBuf, errBuf bytes.Buffer
	streams := &console.Streams{Out: &outBuf, Err: &errBuf}

	reportInvalidHash("abc", &config.Config{}, colorHandler, streams)
	if !bytes.Contains(errBuf.Bytes(), []byte("Invalid hash format")) {
		t.Errorf("Expected invalid hash message, got %s", errBuf.String())
	}
}

func TestFormatSize_Coverage(t *testing.T) {
	tests := []struct {
		size int64
		want string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, tt := range tests {
		got := formatSize(tt.size)
		if got != tt.want {
			t.Errorf("formatSize(%d) = %q, want %q", tt.size, got, tt.want)
		}
	}
}

func TestRunDryRunMode_Coverage(t *testing.T) {
	colorHandler := color.NewColorHandler()
	colorHandler.SetEnabled(false)
	var outBuf, errBuf bytes.Buffer
	streams := &console.Streams{Out: &outBuf, Err: &errBuf}

	t.Run("FileStatError", func(t *testing.T) {
		cfg := &config.Config{Files: []string{"non_existent_file"}}
		code := runDryRunMode(cfg, colorHandler, streams)
		if code != config.ExitSuccess {
			t.Errorf("Expected success even with stat error, got %d", code)
		}
		if !bytes.Contains(errBuf.Bytes(), []byte("non_existent_file")) {
			t.Errorf("Expected filename in error output, got %s", errBuf.String())
		}
	})
}

func TestProcessEntry_Error_Coverage(t *testing.T) {
	colorHandler := color.NewColorHandler()
	errHandler := errors.NewErrorHandler(colorHandler)
	var outBuf, errBuf bytes.Buffer
	streams := &console.Streams{Out: &outBuf, Err: &errBuf}
	results := &hash.Result{}

	entry := hash.Entry{
		Original: "non_existent",
		Error:    os.ErrNotExist,
	}

	processEntry(entry, results, nil, &config.Config{}, streams, errHandler)
	if len(results.Errors) == 0 {
		t.Error("Expected error in results")
	}
}

func TestHandleComparisonError_Coverage(t *testing.T) {
	colorHandler := color.NewColorHandler()
	colorHandler.SetEnabled(false)
	var outBuf, errBuf bytes.Buffer
	streams := &console.Streams{Out: &outBuf, Err: &errBuf}

	handleComparisonError(os.ErrNotExist, "Failed test", &config.Config{}, colorHandler, streams)
	if !bytes.Contains(errBuf.Bytes(), []byte("Failed test")) {
		t.Errorf("Expected message in stderr, got %s", errBuf.String())
	}
}
