// Package errors provides tests for error handling and formatting.
package errors

import (
	"errors"
	"os"
	"strings"
	"testing"
	"testing/quick"

	"github.com/Les-El/hashi/internal/color"
)

// Feature: cli-guidelines-review, Property 4: Error messages are human-readable
// Validates: Requirements 3.1, 3.6
//
// This property test verifies that error messages are human-readable by checking
// that they don't contain technical jargon like stack traces, raw error codes,
// or overly technical language in non-verbose mode.
func TestProperty_ErrorMessagesAreHumanReadable(t *testing.T) {
	// Create a color handler with colors disabled for consistent testing
	colorHandler := color.NewColorHandler()
	colorHandler.SetEnabled(false)
	
	handler := NewErrorHandler(colorHandler)
	handler.SetVerbose(false) // Non-verbose mode
	
	// Property: For any error, the formatted message should not contain
	// technical jargon like stack traces, hex addresses, or raw syscall names
	property := func(errMsg string) bool {
		// Skip empty strings
		if errMsg == "" {
			return true
		}
		
		// Create a test error
		testErr := errors.New(errMsg)
		formatted := handler.FormatError(testErr)
		
		// Check that formatted message doesn't contain technical jargon
		technicalPatterns := []string{
			"0x",           // Hex addresses
			"syscall",      // System call references
			"goroutine",    // Stack trace indicators
			"panic:",       // Panic messages
			"runtime.",     // Runtime package references
		}
		
		for _, pattern := range technicalPatterns {
			if strings.Contains(formatted, pattern) {
				return false
			}
		}
		
		// Check that the message is not just the raw error
		// (it should be formatted with context)
		if formatted == errMsg {
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

// Feature: cli-guidelines-review, Property 4: Error messages include suggestions
// Validates: Requirements 3.2
//
// This property test verifies that common error types include actionable suggestions.
func TestProperty_ErrorMessagesIncludeSuggestions(t *testing.T) {
	colorHandler := color.NewColorHandler()
	colorHandler.SetEnabled(false)
	
	handler := NewErrorHandler(colorHandler)
	
	// Property: For any file-related error, the formatted message should include
	// a suggestion for how to fix it
	property := func(filename string) bool {
		// Skip empty filenames
		if filename == "" {
			return true
		}
		
		// Create common error types
		errors := []error{
			NewFileNotFoundError(filename),
			NewPermissionError(filename),
		}
		
		for _, err := range errors {
			formatted := handler.FormatError(err)
			
			// Check that the formatted message contains some guidance
			// (suggestions typically contain words like "try", "check", "use")
			guidanceWords := []string{"try", "check", "use", "tip", "help"}
			hasGuidance := false
			
			lowerFormatted := strings.ToLower(formatted)
			for _, word := range guidanceWords {
				if strings.Contains(lowerFormatted, word) {
					hasGuidance = true
					break
				}
			}
			
			if !hasGuidance {
				return false
			}
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

// Feature: cli-guidelines-review, Property 4: Paths are sanitized
// Validates: Requirements 3.4
//
// This property test verifies that sensitive paths are sanitized in error messages.
func TestProperty_PathsAreSanitized(t *testing.T) {
	colorHandler := color.NewColorHandler()
	colorHandler.SetEnabled(false)
	
	handler := NewErrorHandler(colorHandler)
	
	// Get home directory for testing
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}
	
	// Property: For any path in the home directory, it should be displayed with ~
	property := func(filename string) bool {
		// Skip empty filenames
		if filename == "" {
			return true
		}
		
		// Create a path in the home directory
		fullPath := home + "/" + filename
		
		// Create an error with this path
		err := NewFileNotFoundError(fullPath)
		formatted := handler.FormatError(err)
		
		// The formatted message should contain ~ instead of the full home path
		// (unless the path is very short and doesn't need sanitization)
		if strings.Contains(formatted, home) && len(home) > 10 {
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

// Feature: cli-guidelines-review, Property 5: Similar errors are grouped
// Validates: Requirements 3.5
//
// This property test verifies that errors of the same type are grouped together
// rather than being repeated individually.
func TestProperty_SimilarErrorsAreGrouped(t *testing.T) {
	// Property: For any list of errors containing multiple errors of the same type,
	// GroupErrors should group them by type
	property := func(numFileNotFound, numPermission, numOther uint8) bool {
		// Limit the numbers to reasonable ranges
		if numFileNotFound > 20 {
			numFileNotFound = numFileNotFound % 20
		}
		if numPermission > 20 {
			numPermission = numPermission % 20
		}
		if numOther > 20 {
			numOther = numOther % 20
		}
		
		// Create a list of errors
		var errs []error
		
		// Add file not found errors
		for i := uint8(0); i < numFileNotFound; i++ {
			errs = append(errs, NewFileNotFoundError("file"+string(rune(i))+".txt"))
		}
		
		// Add permission errors
		for i := uint8(0); i < numPermission; i++ {
			errs = append(errs, NewPermissionError("file"+string(rune(i))+".txt"))
		}
		
		// Add other errors
		for i := uint8(0); i < numOther; i++ {
			errs = append(errs, errors.New("unknown error "+string(rune(i))))
		}
		
		// Group the errors
		groups := GroupErrors(errs)
		
		// Verify that errors are grouped correctly
		// Count errors in each group
		var fileNotFoundCount, permissionCount, otherCount int
		
		for errType, groupErrs := range groups {
			switch errType {
			case ErrorTypeFileNotFound:
				fileNotFoundCount = len(groupErrs)
			case ErrorTypePermission:
				permissionCount = len(groupErrs)
			case ErrorTypeUnknown:
				otherCount = len(groupErrs)
			}
		}
		
		// Verify counts match
		if fileNotFoundCount != int(numFileNotFound) {
			return false
		}
		if permissionCount != int(numPermission) {
			return false
		}
		if otherCount != int(numOther) {
			return false
		}
		
		// Verify that the total number of errors is preserved
		totalInGroups := 0
		for _, groupErrs := range groups {
			totalInGroups += len(groupErrs)
		}
		
		if totalInGroups != len(errs) {
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

// Feature: cli-guidelines-review, Property 5: Grouping preserves all errors
// Validates: Requirements 3.5
//
// This property test verifies that grouping errors doesn't lose any errors.
func TestProperty_GroupingPreservesAllErrors(t *testing.T) {
	// Property: For any list of errors, grouping should preserve all errors
	property := func(numErrors uint8) bool {
		// Limit to reasonable range
		if numErrors > 50 {
			numErrors = numErrors % 50
		}
		
		// Create a list of mixed errors
		var errs []error
		for i := uint8(0); i < numErrors; i++ {
			switch i % 3 {
			case 0:
				errs = append(errs, NewFileNotFoundError("file.txt"))
			case 1:
				errs = append(errs, NewPermissionError("file.txt"))
			case 2:
				errs = append(errs, errors.New("other error"))
			}
		}
		
		// Group the errors
		groups := GroupErrors(errs)
		
		// Count total errors in groups
		totalInGroups := 0
		for _, groupErrs := range groups {
			totalInGroups += len(groupErrs)
		}
		
		// Verify no errors were lost
		return totalInGroups == len(errs)
	}
	
	config := &quick.Config{
		MaxCount: 100,
	}
	
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// Unit Tests for Error Handling
// Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6

// TestErrorHandler_FormatError tests basic error formatting.
func TestErrorHandler_FormatError(t *testing.T) {
	colorHandler := color.NewColorHandler()
	colorHandler.SetEnabled(false)
	
	handler := NewErrorHandler(colorHandler)
	
	tests := []struct {
		name     string
		err      error
		wantText string
	}{
		{
			name:     "nil error",
			err:      nil,
			wantText: "",
		},
		{
			name:     "file not found error",
			err:      NewFileNotFoundError("document.pdf"),
			wantText: "Cannot find file",
		},
		{
			name:     "permission error",
			err:      NewPermissionError("secret.txt"),
			wantText: "Cannot read file",
		},
		{
			name:     "invalid hash error",
			err:      NewInvalidHashError("abc123", "sha256", 64),
			wantText: "Invalid sha256 hash",
		},
		{
			name:     "standard error",
			err:      errors.New("something went wrong"),
			wantText: "something went wrong",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.FormatError(tt.err)
			
			if tt.wantText == "" && result != "" {
				t.Errorf("Expected empty string, got %q", result)
			}
			
			if tt.wantText != "" && !strings.Contains(result, tt.wantText) {
				t.Errorf("Expected result to contain %q, got %q", tt.wantText, result)
			}
		})
	}
}

// TestErrorHandler_SuggestFix tests suggestion generation.
func TestErrorHandler_SuggestFix(t *testing.T) {
	colorHandler := color.NewColorHandler()
	colorHandler.SetEnabled(false)
	
	handler := NewErrorHandler(colorHandler)
	
	tests := []struct {
		name           string
		err            error
		wantSuggestion bool
	}{
		{
			name:           "file not found has suggestion",
			err:            NewFileNotFoundError("missing.txt"),
			wantSuggestion: true,
		},
		{
			name:           "permission error has suggestion",
			err:            NewPermissionError("protected.txt"),
			wantSuggestion: true,
		},
		{
			name:           "invalid hash has suggestion",
			err:            NewInvalidHashError("short", "sha256", 64),
			wantSuggestion: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := handler.SuggestFix(tt.err)
			
			if tt.wantSuggestion && suggestion == "" {
				t.Errorf("Expected suggestion, got empty string")
			}
			
			if !tt.wantSuggestion && suggestion != "" {
				t.Errorf("Expected no suggestion, got %q", suggestion)
			}
		})
	}
}

// TestErrorHandler_VerboseMode tests verbose error output.
func TestErrorHandler_VerboseMode(t *testing.T) {
	colorHandler := color.NewColorHandler()
	colorHandler.SetEnabled(false)
	
	handler := NewErrorHandler(colorHandler)
	
	// Create an error with an underlying cause
	originalErr := errors.New("underlying cause")
	wrappedErr := &Error{
		Type:       ErrorTypeFileNotFound,
		Message:    "Cannot find file: test.txt",
		Suggestion: "Check the path",
		Original:   originalErr,
	}
	
	// Test non-verbose mode
	handler.SetVerbose(false)
	nonVerbose := handler.FormatError(wrappedErr)
	
	if strings.Contains(nonVerbose, "underlying cause") {
		t.Errorf("Non-verbose output should not contain original error details")
	}
	
	// Test verbose mode
	handler.SetVerbose(true)
	verbose := handler.FormatError(wrappedErr)
	
	if !strings.Contains(verbose, "underlying cause") {
		t.Errorf("Verbose output should contain original error details")
	}
}

// TestGroupErrors tests error grouping functionality.
func TestGroupErrors(t *testing.T) {
	tests := []struct {
		name       string
		errors     []error
		wantGroups int
	}{
		{
			name:       "empty list",
			errors:     []error{},
			wantGroups: 0,
		},
		{
			name: "single error type",
			errors: []error{
				NewFileNotFoundError("file1.txt"),
				NewFileNotFoundError("file2.txt"),
				NewFileNotFoundError("file3.txt"),
			},
			wantGroups: 1,
		},
		{
			name: "multiple error types",
			errors: []error{
				NewFileNotFoundError("file1.txt"),
				NewPermissionError("file2.txt"),
				NewFileNotFoundError("file3.txt"),
			},
			wantGroups: 2,
		},
		{
			name: "mixed standard and custom errors",
			errors: []error{
				NewFileNotFoundError("file1.txt"),
				errors.New("unknown error"),
				NewPermissionError("file2.txt"),
			},
			wantGroups: 3,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := GroupErrors(tt.errors)
			
			if len(groups) != tt.wantGroups {
				t.Errorf("Expected %d groups, got %d", tt.wantGroups, len(groups))
			}
			
			// Verify all errors are accounted for
			totalInGroups := 0
			for _, groupErrs := range groups {
				totalInGroups += len(groupErrs)
			}
			
			if totalInGroups != len(tt.errors) {
				t.Errorf("Expected %d total errors, got %d", len(tt.errors), totalInGroups)
			}
		})
	}
}

// TestSanitizePath tests path sanitization.
func TestSanitizePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}
	
	tests := []struct {
		name     string
		path     string
		wantHome bool // Should contain ~ instead of home path
	}{
		{
			name:     "path in home directory",
			path:     home + "/documents/file.txt",
			wantHome: true,
		},
		{
			name:     "relative path",
			path:     "file.txt",
			wantHome: false,
		},
		{
			name:     "current directory path",
			path:     "./file.txt",
			wantHome: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizePath(tt.path)
			
			if tt.wantHome {
				if !strings.Contains(result, "~") {
					t.Errorf("Expected path to contain ~, got %q", result)
				}
				if strings.Contains(result, home) {
					t.Errorf("Expected home path to be replaced with ~, got %q", result)
				}
			}
		})
	}
}

// TestNewFileNotFoundError tests file not found error creation.
func TestNewFileNotFoundError(t *testing.T) {
	err := NewFileNotFoundError("missing.txt")
	
	if err.Type != ErrorTypeFileNotFound {
		t.Errorf("Expected ErrorTypeFileNotFound, got %v", err.Type)
	}
	
	if !strings.Contains(err.Message, "missing.txt") {
		t.Errorf("Expected message to contain filename, got %q", err.Message)
	}
	
	if err.Suggestion == "" {
		t.Errorf("Expected suggestion to be non-empty")
	}
}

// TestNewPermissionError tests permission error creation.
func TestNewPermissionError(t *testing.T) {
	err := NewPermissionError("protected.txt")
	
	if err.Type != ErrorTypePermission {
		t.Errorf("Expected ErrorTypePermission, got %v", err.Type)
	}
	
	if !strings.Contains(err.Message, "protected.txt") {
		t.Errorf("Expected message to contain filename, got %q", err.Message)
	}
	
	if err.Suggestion == "" {
		t.Errorf("Expected suggestion to be non-empty")
	}
}

// TestNewInvalidHashError tests invalid hash error creation.
func TestNewInvalidHashError(t *testing.T) {
	err := NewInvalidHashError("abc123", "sha256", 64)
	
	if err.Type != ErrorTypeInvalidHash {
		t.Errorf("Expected ErrorTypeInvalidHash, got %v", err.Type)
	}
	
	if !strings.Contains(err.Message, "sha256") {
		t.Errorf("Expected message to contain algorithm, got %q", err.Message)
	}
	
	if !strings.Contains(err.Suggestion, "64") {
		t.Errorf("Expected suggestion to mention expected length, got %q", err.Suggestion)
	}
}

// TestNewConfigError tests config error creation.
func TestNewConfigError(t *testing.T) {
	err := NewConfigError("Invalid configuration value")
	
	if err.Type != ErrorTypeConfig {
		t.Errorf("Expected ErrorTypeConfig, got %v", err.Type)
	}
	
	if err.Message != "Invalid configuration value" {
		t.Errorf("Expected message to match input, got %q", err.Message)
	}
	
	if err.Suggestion == "" {
		t.Errorf("Expected suggestion to be non-empty")
	}
}

// TestError_Unwrap tests error unwrapping.
func TestError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := &Error{
		Type:     ErrorTypeUnknown,
		Message:  "wrapped message",
		Original: originalErr,
	}
	
	unwrapped := wrappedErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Expected unwrapped error to be original, got %v", unwrapped)
	}
}

// TestClassifyError tests error classification.
func TestClassifyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantType ErrorType
	}{
		{
			name:     "file not found",
			err:      os.ErrNotExist,
			wantType: ErrorTypeFileNotFound,
		},
		{
			name:     "permission denied",
			err:      os.ErrPermission,
			wantType: ErrorTypePermission,
		},
		{
			name:     "unknown error",
			err:      errors.New("unknown"),
			wantType: ErrorTypeUnknown,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyError(tt.err)
			if result != tt.wantType {
				t.Errorf("Expected %v, got %v", tt.wantType, result)
			}
		})
	}
}
