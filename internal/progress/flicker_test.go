package progress

import (
	"bytes"
	"testing"
	"strings"
	"time"
	"github.com/schollz/progressbar/v3"
)

func TestWriteMessage_FlickerPrevention(t *testing.T) {
	t.Run("clears bar if enabled and TTY", func(t *testing.T) {
		buf := &bytes.Buffer{}
		opts := &Options{
			Total:     100,
			Writer:    buf,
			Threshold: 0,
		}
		bar := NewBar(opts)
		bar.startTime = time.Now().Add(-1 * time.Second)
		bar.isTTY = true // Force TTY for test
		bar.Add(1)       // Enable it
		
		if !bar.IsEnabled() {
			t.Fatal("bar should be enabled and TTY")
		}
		
		// Manually set inner bar if it's nil
		if bar.bar == nil {
			bar.bar = progressbar.Default(100)
		}
		
		msg := "error occurred"
		bar.WriteMessage(msg)
		
		// We can't easily check if it was cleared and redrawn because the underlying
		// library handles terminal escape codes, but we can verify the message
		// was written to the writer.
		if !strings.Contains(buf.String(), msg) {
			t.Errorf("expected %q in output, got %q", msg, buf.String())
		}
	})
}

func TestFinishAndClear_Coverage(t *testing.T) {
	// These usually do nothing if bar is nil (non-TTY)
	// We already have some coverage, but let's ensure they don't panic
	// and try to get that 100% up by forcing bar to be non-nil
	
	buf := &bytes.Buffer{}
	opts := &Options{
		Total:     100,
		Writer:    buf,
		Threshold: 0,
	}
	bar := NewBar(opts)
	
	// Manually initialize the inner bar
	bar.bar = progressbar.NewOptions(100, progressbar.OptionSetWriter(buf))
	bar.isTTY = true
	bar.enabled = true
	
	t.Run("Finish does not panic", func(t *testing.T) {
		bar.Finish()
	})
	
	t.Run("Clear does not panic", func(t *testing.T) {
		bar.Clear()
	})
}

func TestBar_Add_WithInnerBar_Coverage(t *testing.T) {
	buf := &bytes.Buffer{}
	opts := &Options{
		Total:     100,
		Writer:    buf,
		Threshold: 0,
	}
	bar := NewBar(opts)
	bar.startTime = time.Now().Add(-1 * time.Second)
	
	// Manually initialize inner bar
	bar.bar = progressbar.NewOptions(100, progressbar.OptionSetWriter(buf))
	bar.isTTY = true
	
	t.Run("Add with inner bar", func(t *testing.T) {
		bar.Add(10)
		if bar.current != 10 {
			t.Errorf("Expected current 10, got %d", bar.current)
		}
		if !bar.enabled {
			t.Error("Expected bar to be enabled")
		}
	})
}

func TestFinishAndClear_NilBar_Coverage(t *testing.T) {
	bar := &Bar{bar: nil}
	
	t.Run("Finish with nil bar does not panic", func(t *testing.T) {
		bar.Finish()
	})
	
	t.Run("Clear with nil bar does not panic", func(t *testing.T) {
		bar.Clear()
	})
}

func TestETA_EdgeCases(t *testing.T) {
	t.Run("zero progress", func(t *testing.T) {
		b := &Bar{current: 0}
		if b.ETA() != 0 {
			t.Error("ETA should be 0 when current is 0")
		}
	})
	
	t.Run("slow progress", func(t *testing.T) {
		b := &Bar{
			total:     100,
			current:   1,
			startTime: time.Now().Add(-100 * time.Hour),
		}
		if b.ETA() <= 0 {
			t.Error("ETA should be positive for slow progress")
		}
	})
}
