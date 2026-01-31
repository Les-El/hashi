// Package progress provides progress bar functionality for chexum.
//
// Progress indicators are shown for operations taking longer than 100ms.
// The progress bar displays percentage, count, and ETA. It automatically
// hides when output is not a TTY.
package progress

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
)

// Bar wraps a progress bar with TTY-aware display.
type Bar struct {
	bar       *progressbar.ProgressBar
	total     int64
	current   int64
	startTime time.Time
	isTTY     bool
	enabled   bool
	threshold time.Duration
	writer    io.Writer
}

// Options configures the progress bar behavior.
type Options struct {
	// Total is the total number of items to process
	Total int64
	// Description is shown before the progress bar
	Description string
	// ShowBytes shows bytes processed instead of count
	ShowBytes bool
	// Threshold is the minimum duration before showing progress (default 100ms)
	Threshold time.Duration
	// Writer is where to write progress output (default os.Stderr)
	Writer io.Writer
}

// DefaultOptions returns default progress bar options.
func DefaultOptions() *Options {
	return &Options{
		Threshold: 100 * time.Millisecond,
		Writer:    os.Stderr,
	}
}

// NewBar creates a new progress bar with the given options.
func NewBar(opts *Options) *Bar {
	opts = ensureDefaults(opts)
	isTTY := checkTTY(opts.Writer)

	b := &Bar{
		total:     opts.Total,
		startTime: time.Now(),
		isTTY:     isTTY,
		enabled:   false,
		threshold: opts.Threshold,
		writer:    opts.Writer,
	}

	if isTTY {
		b.bar = progressbar.NewOptions64(opts.Total, configureBar(opts)...)
	}
	return b
}

func ensureDefaults(opts *Options) *Options {
	if opts == nil {
		opts = DefaultOptions()
	}
	if opts.Threshold == 0 {
		opts.Threshold = 100 * time.Millisecond
	}
	if opts.Writer == nil {
		opts.Writer = os.Stderr
	}
	return opts
}

func checkTTY(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

func configureBar(opts *Options) []progressbar.Option {
	barOpts := []progressbar.Option{
		progressbar.OptionSetDescription(opts.Description),
		progressbar.OptionSetWriter(opts.Writer),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	}
	if opts.ShowBytes {
		barOpts = append(barOpts, progressbar.OptionShowBytes(true))
	}
	return barOpts
}

// Add increments the progress by n.
func (b *Bar) Add(n int64) {
	b.current += n

	// Check if we should enable the progress bar
	if !b.enabled && time.Since(b.startTime) > b.threshold {
		b.enabled = true
	}

	// Update the progress bar if enabled and on TTY
	if b.enabled && b.bar != nil {
		b.bar.Add64(n)
	}
}

// Increment increments the progress by 1.
func (b *Bar) Increment() {
	b.Add(1)
}

// SetCurrent sets the current progress value.
func (b *Bar) SetCurrent(n int64) {
	delta := n - b.current
	if delta > 0 {
		b.Add(delta)
	}
}

// Finish completes the progress bar.
func (b *Bar) Finish() {
	if b.bar != nil {
		b.bar.Finish()
	}
}

// Clear removes the progress bar from display.
func (b *Bar) Clear() {
	if b.bar != nil {
		b.bar.Clear()
	}
}

// WriteMessage prints a message to the progress bar's writer,
// ensuring the bar is cleared before printing and redrawn after.
func (b *Bar) WriteMessage(msg string) {
	if b.IsEnabled() {
		b.Clear()
		fmt.Fprintln(b.writer, msg)
		// We don't force a redraw here, as the next Add/Increment
		// or periodic update from the library will handle it.
		// If we wanted immediate redraw we'd need bar.Render()
	} else {
		fmt.Fprintln(b.writer, msg)
	}
}

// IsEnabled returns whether the progress bar is currently being displayed.
func (b *Bar) IsEnabled() bool {
	return b.enabled && b.isTTY
}

// IsTTY returns whether output is going to a terminal.
func (b *Bar) IsTTY() bool {
	return b.isTTY
}

// ETA returns the estimated time remaining.
func (b *Bar) ETA() time.Duration {
	if b.current == 0 {
		return 0
	}

	elapsed := time.Since(b.startTime)
	rate := float64(b.current) / elapsed.Seconds()
	remaining := float64(b.total-b.current) / rate

	return time.Duration(remaining * float64(time.Second))
}

// Percentage returns the current progress as a percentage.
func (b *Bar) Percentage() float64 {
	if b.total == 0 {
		return 0
	}
	return float64(b.current) / float64(b.total) * 100
}

// String returns a string representation of the progress.
func (b *Bar) String() string {
	return fmt.Sprintf("%.1f%% (%d/%d)", b.Percentage(), b.current, b.total)
}
