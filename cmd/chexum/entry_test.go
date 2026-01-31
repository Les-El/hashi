package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/Les-El/chexum/internal/color"
	"github.com/Les-El/chexum/internal/config"
	"github.com/Les-El/chexum/internal/console"
	"github.com/Les-El/chexum/internal/errors"
	"github.com/Les-El/chexum/internal/hash"
	"github.com/Les-El/chexum/internal/progress"
)

func TestProcessEntry_WithBar(t *testing.T) {
	colorHandler := color.NewColorHandler()
	errHandler := errors.NewErrorHandler(colorHandler)
	var outBuf, errBuf bytes.Buffer
	streams := &console.Streams{Out: &outBuf, Err: &errBuf}
	results := &hash.Result{}

	opts := &progress.Options{
		Total:     10,
		Writer:    &errBuf,
		Threshold: 0,
	}
	bar := progress.NewBar(opts)

	t.Run("successful entry", func(t *testing.T) {
		entry := hash.Entry{
			Original: "file.txt",
			Hash:     "abc",
			Size:     100,
		}
		processEntry(entry, results, bar, &config.Config{}, streams, errHandler)
		
		if results.FilesProcessed != 1 {
			t.Errorf("Expected 1 file processed, got %d", results.FilesProcessed)
		}
		if results.BytesProcessed != 100 {
			t.Errorf("Expected 100 bytes processed, got %d", results.BytesProcessed)
		}
	})

	t.Run("error entry with bar", func(t *testing.T) {
		errBuf.Reset()
		entry := hash.Entry{
			Original: "missing.txt",
			Error:    fmt.Errorf("not found"),
		}
		processEntry(entry, results, bar, &config.Config{}, streams, errHandler)
		
		if len(results.Errors) != 1 {
			t.Error("Expected error in results")
		}
		// Since it's not a real TTY, IsEnabled() will be false unless we force it,
		// but bar.WriteMessage should still write to the buffer.
		if !bytes.Contains(errBuf.Bytes(), []byte("not found")) {
			t.Errorf("Expected error message in errBuf, got %q", errBuf.String())
		}
	})
}

func TestProcessEntry_Quiet(t *testing.T) {
	colorHandler := color.NewColorHandler()
	errHandler := errors.NewErrorHandler(colorHandler)
	var outBuf, errBuf bytes.Buffer
	streams := &console.Streams{Out: &outBuf, Err: &errBuf}
	results := &hash.Result{}

	entry := hash.Entry{
		Original: "missing.txt",
		Error:    fmt.Errorf("secret error"),
	}
	
	processEntry(entry, results, nil, &config.Config{Quiet: true}, streams, errHandler)
	
	if errBuf.Len() > 0 {
		t.Errorf("Expected no output in quiet mode, got %q", errBuf.String())
	}
}
