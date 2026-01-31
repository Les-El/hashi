package progress

import (
	"bytes"
	"testing"
	"strings"
	"time"
)

func TestWriteMessage(t *testing.T) {
	t.Run("writes to buffer when disabled", func(t *testing.T) {
		buf := &bytes.Buffer{}
		opts := &Options{
			Total:   100,
			Writer:  buf,
		}
		bar := NewBar(opts)
		
		msg := "test message"
		bar.WriteMessage(msg)
		
		if !strings.Contains(buf.String(), msg) {
			t.Errorf("expected %q in output, got %q", msg, buf.String())
		}
	})

	t.Run("writes to buffer when enabled but not TTY", func(t *testing.T) {
		buf := &bytes.Buffer{}
		opts := &Options{
			Total:     100,
			Writer:    buf,
			Threshold: 0,
		}
		bar := NewBar(opts)
		bar.startTime = time.Now().Add(-1 * time.Second)
		bar.Add(1) // Enable it
		
		if !bar.enabled {
			t.Fatal("bar should be enabled")
		}
		
		msg := "test message"
		bar.WriteMessage(msg)
		
		if !strings.Contains(buf.String(), msg) {
			t.Errorf("expected %q in output, got %q", msg, buf.String())
		}
	})
}
