// Package color provides TTY detection and color handling for chexum.
//
// It automatically detects whether output is going to a terminal and
// respects the NO_COLOR environment variable. Colors are used to
// highlight important information in human-readable output.
package color

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// Color represents a text color for output.
type Color int

const (
	// ColorNone represents no color (default terminal color)
	ColorNone Color = iota
	// ColorGreen for matches and success
	ColorGreen
	// ColorRed for errors and mismatches
	ColorRed
	// ColorYellow for warnings
	ColorYellow
	// ColorBlue for information
	ColorBlue
	// ColorCyan for file paths
	ColorCyan
	// ColorGray for secondary information
	ColorGray
)

// Handler manages color output based on TTY detection and environment.
type Handler struct {
	enabled bool
	isTTY   bool

	// Color functions
	green  func(format string, a ...interface{}) string
	red    func(format string, a ...interface{}) string
	yellow func(format string, a ...interface{}) string
	blue   func(format string, a ...interface{}) string
	cyan   func(format string, a ...interface{}) string
	gray   func(format string, a ...interface{}) string
}

// NewColorHandler creates a new color handler with automatic TTY detection.
func NewColorHandler() *Handler {
	h := &Handler{}
	h.detectTTY()
	h.setupColors()
	return h
}

// detectTTY checks if stdout is a terminal and if NO_COLOR is set.
func (h *Handler) detectTTY() {
	// Check if stdout is a terminal
	h.isTTY = term.IsTerminal(int(os.Stdout.Fd()))

	// Check NO_COLOR environment variable
	_, noColor := os.LookupEnv("NO_COLOR")

	// Enable colors only if TTY and NO_COLOR not set
	h.enabled = h.isTTY && !noColor
}

// setupColors initializes the color functions.
func (h *Handler) setupColors() {
	if h.enabled {
		// Force color output for the color functions
		color.NoColor = false

		h.green = color.New(color.FgGreen).SprintfFunc()
		h.red = color.New(color.FgRed).SprintfFunc()
		h.yellow = color.New(color.FgYellow).SprintfFunc()
		h.blue = color.New(color.FgBlue).SprintfFunc()
		h.cyan = color.New(color.FgCyan).SprintfFunc()
		h.gray = color.New(color.FgHiBlack).SprintfFunc()
	} else {
		// Disable color output
		color.NoColor = true

		// No-op functions when colors are disabled
		noOp := func(format string, a ...interface{}) string {
			if len(a) == 0 {
				return format
			}
			return fmt.Sprintf(format, a...)
		}
		h.green = noOp
		h.red = noOp
		h.yellow = noOp
		h.blue = noOp
		h.cyan = noOp
		h.gray = noOp
	}
}

// IsEnabled returns whether color output is enabled.
func (h *Handler) IsEnabled() bool {
	return h.enabled
}

// IsTTY returns whether stdout is a terminal.
func (h *Handler) IsTTY() bool {
	return h.isTTY
}

// SetEnabled manually enables or disables color output.
func (h *Handler) SetEnabled(enabled bool) {
	h.enabled = enabled
	h.setupColors()
}

// Colorize applies the specified color to the text.
func (h *Handler) Colorize(text string, c Color) string {
	switch c {
	case ColorGreen:
		return h.green("%s", text)
	case ColorRed:
		return h.red("%s", text)
	case ColorYellow:
		return h.yellow("%s", text)
	case ColorBlue:
		return h.blue("%s", text)
	case ColorCyan:
		return h.cyan("%s", text)
	case ColorGray:
		return h.gray("%s", text)
	default:
		return text
	}
}

// Green returns text in green (for matches, success).
func (h *Handler) Green(text string) string {
	return h.Colorize(text, ColorGreen)
}

// Red returns text in red (for errors, mismatches).
func (h *Handler) Red(text string) string {
	return h.Colorize(text, ColorRed)
}

// Yellow returns text in yellow (for warnings).
func (h *Handler) Yellow(text string) string {
	return h.Colorize(text, ColorYellow)
}

// Blue returns text in blue (for information).
func (h *Handler) Blue(text string) string {
	return h.Colorize(text, ColorBlue)
}

// Cyan returns text in cyan (for file paths).
func (h *Handler) Cyan(text string) string {
	return h.Colorize(text, ColorCyan)
}

// Gray returns text in gray (for secondary information).
func (h *Handler) Gray(text string) string {
	return h.Colorize(text, ColorGray)
}

// Success formats a success message with a green checkmark.
func (h *Handler) Success(message string) string {
	return h.Green("✓") + " " + message
}

// Error formats an error message with a red X.
func (h *Handler) Error(message string) string {
	return h.Red("✗") + " " + message
}

// Warning formats a warning message with a yellow exclamation.
func (h *Handler) Warning(message string) string {
	return h.Yellow("!") + " " + message
}

// Info formats an info message with a blue i.
func (h *Handler) Info(message string) string {
	return h.Blue("ℹ") + " " + message
}
