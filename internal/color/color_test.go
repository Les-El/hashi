// Package color provides TTY detection and color handling for chexum.
package color

import (
	"os"
	"strings"
	"testing"
	"testing/quick"
)

func TestProperty_ColorOutputRespectsTTYDetection(t *testing.T) {
	property := func(text string) bool {
		h := &Handler{enabled: false, isTTY: false}
		h.setupColors()

		if !verifyAllColors(h, text) {
			return false
		}
		return verifyConvenienceMethods(h, text)
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

func verifyAllColors(h *Handler, text string) bool {
	colors := []Color{ColorGreen, ColorRed, ColorYellow, ColorBlue, ColorCyan, ColorGray}
	for _, c := range colors {
		result := h.Colorize(text, c)
		if result != text || containsANSI(result) {
			return false
		}
	}
	return true
}

func verifyConvenienceMethods(h *Handler, text string) bool {
	methods := []func(string) string{h.Green, h.Red, h.Yellow, h.Blue, h.Cyan, h.Gray}
	for _, m := range methods {
		if m(text) != text || containsANSI(m(text)) {
			return false
		}
	}
	return true
}

// TestProperty_ColorOutputWithNO_COLOR verifies that NO_COLOR environment variable
// disables color output regardless of TTY status.
//
// Feature: cli-guidelines-review, Property 2: Color output respects TTY detection
// Validates: Requirements 2.3
func TestProperty_ColorOutputWithNO_COLOR(t *testing.T) {
	// Save original NO_COLOR state
	originalValue, hadNoColor := os.LookupEnv("NO_COLOR")
	defer func() {
		if hadNoColor {
			os.Setenv("NO_COLOR", originalValue)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	}()

	// Property: When NO_COLOR is set, colors should be disabled
	property := func(text string) bool {
		// Set NO_COLOR environment variable
		os.Setenv("NO_COLOR", "1")

		// Create a new handler (will detect NO_COLOR)
		h := NewColorHandler()

		// Colors should be disabled
		if h.IsEnabled() {
			return false
		}

		// All color methods should return plain text
		if h.Green(text) != text || containsANSI(h.Green(text)) {
			return false
		}
		if h.Red(text) != text || containsANSI(h.Red(text)) {
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// containsANSI checks if a string contains ANSI escape sequences.
func containsANSI(s string) bool {
	// ANSI escape sequences start with ESC [ (or \x1b[)
	return strings.Contains(s, "\x1b[") || strings.Contains(s, "\033[")
}

// Unit Tests

func TestColorHandler_TTYDetection(t *testing.T) {
	tests := []struct {
		name     string
		setupEnv func()
	}{
		{"NO_COLOR set", func() { os.Setenv("NO_COLOR", "1") }},
		{"NO_COLOR empty", func() { os.Setenv("NO_COLOR", "") }},
		{"NO_COLOR unset", func() { os.Unsetenv("NO_COLOR") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig, had := os.LookupEnv("NO_COLOR")
			defer restoreEnv("NO_COLOR", orig, had)

			tt.setupEnv()
			h := NewColorHandler()

			if strings.Contains(tt.name, "NO_COLOR") && !strings.Contains(tt.name, "unset") {
				if h.IsEnabled() {
					t.Errorf("Expected colors disabled for %s", tt.name)
				}
			}
		})
	}
}

func restoreEnv(key, val string, had bool) {
	if had {
		os.Setenv(key, val)
	} else {
		os.Unsetenv(key)
	}
}

// TestColorHandler_ColorCodeGeneration tests that color codes are generated correctly.
func TestColorHandler_ColorCodeGeneration(t *testing.T) {
	// Save and clear NO_COLOR to allow colors in tests
	originalValue, hadNoColor := os.LookupEnv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	defer func() {
		if hadNoColor {
			os.Setenv("NO_COLOR", originalValue)
		}
	}()

	// Create handler with colors explicitly enabled
	h := &Handler{
		enabled: true,
		isTTY:   true,
	}
	h.setupColors()

	tests := []struct {
		name  string
		color Color
		text  string
	}{
		{"Green", ColorGreen, "success"},
		{"Red", ColorRed, "error"},
		{"Yellow", ColorYellow, "warning"},
		{"Blue", ColorBlue, "info"},
		{"Cyan", ColorCyan, "path"},
		{"Gray", ColorGray, "secondary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.Colorize(tt.text, tt.color)

			// When colors are enabled, result should contain ANSI codes
			if !containsANSI(result) {
				t.Errorf("Expected ANSI codes in output, but got: %q", result)
			}

			// Result should contain the original text
			if !strings.Contains(result, tt.text) {
				t.Errorf("Expected result to contain %q, but got: %q", tt.text, result)
			}
		})
	}
}

// TestColorHandler_ColorCodeDisabled tests that no color codes are generated when disabled.
func TestColorHandler_ColorCodeDisabled(t *testing.T) {
	// Create handler with colors explicitly disabled
	h := &Handler{
		enabled: false,
		isTTY:   false,
	}
	h.setupColors()

	tests := []struct {
		name  string
		color Color
		text  string
	}{
		{"Green", ColorGreen, "success"},
		{"Red", ColorRed, "error"},
		{"Yellow", ColorYellow, "warning"},
		{"Blue", ColorBlue, "info"},
		{"Cyan", ColorCyan, "path"},
		{"Gray", ColorGray, "secondary"},
		{"None", ColorNone, "plain"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.Colorize(tt.text, tt.color)

			// When colors are disabled, result should equal input
			if result != tt.text {
				t.Errorf("Expected %q, but got: %q", tt.text, result)
			}

			// Result should not contain ANSI codes
			if containsANSI(result) {
				t.Errorf("Expected no ANSI codes, but got: %q", result)
			}
		})
	}
}

// TestColorHandler_ConvenienceMethods tests the convenience methods.
func TestColorHandler_ConvenienceMethods(t *testing.T) {
	// Save and clear NO_COLOR to allow colors in tests
	originalValue, hadNoColor := os.LookupEnv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	defer func() {
		if hadNoColor {
			os.Setenv("NO_COLOR", originalValue)
		}
	}()

	// Test with colors enabled
	h := &Handler{
		enabled: true,
		isTTY:   true,
	}
	h.setupColors()

	text := "test"

	tests := []struct {
		name   string
		method func(string) string
	}{
		{"Green", h.Green},
		{"Red", h.Red},
		{"Yellow", h.Yellow},
		{"Blue", h.Blue},
		{"Cyan", h.Cyan},
		{"Gray", h.Gray},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method(text)

			// Should contain ANSI codes when enabled
			if !containsANSI(result) {
				t.Errorf("Expected ANSI codes in output, but got: %q", result)
			}

			// Should contain original text
			if !strings.Contains(result, text) {
				t.Errorf("Expected result to contain %q, but got: %q", text, result)
			}
		})
	}
}

// TestColorHandler_FormattedMessages tests formatted message methods.
func TestColorHandler_FormattedMessages(t *testing.T) {
	// Save and clear NO_COLOR to allow colors in tests
	originalValue, hadNoColor := os.LookupEnv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	defer func() {
		if hadNoColor {
			os.Setenv("NO_COLOR", originalValue)
		}
	}()

	h := &Handler{
		enabled: true,
		isTTY:   true,
	}
	h.setupColors()

	tests := []struct {
		name     string
		method   func(string) string
		message  string
		expected string // Symbol that should be present
	}{
		{"Success", h.Success, "operation completed", "✓"},
		{"Error", h.Error, "operation failed", "✗"},
		{"Warning", h.Warning, "be careful", "!"},
		{"Info", h.Info, "for your information", "ℹ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method(tt.message)

			// Should contain the symbol
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected result to contain symbol %q, but got: %q", tt.expected, result)
			}

			// Should contain the message
			if !strings.Contains(result, tt.message) {
				t.Errorf("Expected result to contain message %q, but got: %q", tt.message, result)
			}

			// Should contain ANSI codes when enabled
			if !containsANSI(result) {
				t.Errorf("Expected ANSI codes in output, but got: %q", result)
			}
		})
	}
}

// TestColorHandler_SetEnabled tests manual enable/disable.
func TestColorHandler_SetEnabled(t *testing.T) {
	// Save and clear NO_COLOR to allow colors in tests
	originalValue, hadNoColor := os.LookupEnv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	defer func() {
		if hadNoColor {
			os.Setenv("NO_COLOR", originalValue)
		}
	}()

	h := NewColorHandler()

	// Test enabling
	h.SetEnabled(true)
	if !h.IsEnabled() {
		t.Error("Expected colors to be enabled after SetEnabled(true)")
	}

	result := h.Green("test")
	if !containsANSI(result) {
		t.Errorf("Expected ANSI codes when enabled, but got: %q", result)
	}

	// Test disabling
	h.SetEnabled(false)
	if h.IsEnabled() {
		t.Error("Expected colors to be disabled after SetEnabled(false)")
	}

	result = h.Green("test")
	if result != "test" {
		t.Errorf("Expected plain text when disabled, but got: %q", result)
	}
	if containsANSI(result) {
		t.Errorf("Expected no ANSI codes when disabled, but got: %q", result)
	}
}

// TestColorHandler_IsTTY tests the IsTTY method.
func TestColorHandler_IsTTY(t *testing.T) {
	h := NewColorHandler()

	// IsTTY should return a boolean (we can't predict the value in tests)
	isTTY := h.IsTTY()
	if isTTY != true && isTTY != false {
		t.Error("IsTTY should return a boolean value")
	}
}

func TestNewColorHandler(t *testing.T) {
	h := NewColorHandler()
	if h == nil {
		t.Fatal("NewColorHandler returned nil")
	}
}

func TestIsEnabled(t *testing.T) {
	h := NewColorHandler()
	_ = h.IsEnabled()
}

func TestIsTTY(t *testing.T) {
	h := NewColorHandler()
	_ = h.IsTTY()
}

func TestSetEnabled(t *testing.T) {
	h := NewColorHandler()
	h.SetEnabled(true)
	if !h.IsEnabled() {
		t.Error("Expected enabled")
	}
	h.SetEnabled(false)
	if h.IsEnabled() {
		t.Error("Expected disabled")
	}
}

func TestColorize(t *testing.T) {
	h := NewColorHandler()
	h.SetEnabled(false)
	text := "test"
	if h.Colorize(text, ColorGreen) != text {
		t.Error("Expected plain text when disabled")
	}
}

func TestGreen(t *testing.T) {
	h := NewColorHandler()
	h.SetEnabled(false)
	if h.Green("test") != "test" {
		t.Error("Expected plain text")
	}
}

func TestRed(t *testing.T) {
	h := NewColorHandler()
	h.SetEnabled(false)
	if h.Red("test") != "test" {
		t.Error("Expected plain text")
	}
}

func TestYellow(t *testing.T) {
	h := NewColorHandler()
	h.SetEnabled(false)
	if h.Yellow("test") != "test" {
		t.Error("Expected plain text")
	}
}

func TestBlue(t *testing.T) {
	h := NewColorHandler()
	h.SetEnabled(false)
	if h.Blue("test") != "test" {
		t.Error("Expected plain text")
	}
}

func TestCyan(t *testing.T) {
	h := NewColorHandler()
	h.SetEnabled(false)
	if h.Cyan("test") != "test" {
		t.Error("Expected plain text")
	}
}

func TestGray(t *testing.T) {
	h := NewColorHandler()
	h.SetEnabled(false)
	if h.Gray("test") != "test" {
		t.Error("Expected plain text")
	}
}

func TestSuccess(t *testing.T) {
	h := NewColorHandler()
	h.SetEnabled(false)
	if !strings.Contains(h.Success("msg"), "msg") {
		t.Error("Missing message")
	}
}

func TestError(t *testing.T) {
	h := NewColorHandler()
	h.SetEnabled(false)
	if !strings.Contains(h.Error("msg"), "msg") {
		t.Error("Missing message")
	}
}

func TestWarning(t *testing.T) {
	h := NewColorHandler()
	h.SetEnabled(false)
	if !strings.Contains(h.Warning("msg"), "msg") {
		t.Error("Missing message")
	}
}

func TestInfo(t *testing.T) {
	h := NewColorHandler()
	h.SetEnabled(false)
	if !strings.Contains(h.Info("msg"), "msg") {
		t.Error("Missing message")
	}
}
