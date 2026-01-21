// Package main provides the entry point for the hashi CLI tool.
//
// hashi is a command-line hash comparison tool that computes and compares
// cryptographic hashes. It follows industry-standard CLI design guidelines
// for human-first design, composability, and robustness.
//
// Usage:
//
//	// hashi [flags] [files...]
//
// When run with no arguments, hashi processes all non-hidden files in the
// current directory.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Les-El/hashi/internal/archive"
	"github.com/Les-El/hashi/internal/color"
	"github.com/Les-El/hashi/internal/config"
	"github.com/Les-El/hashi/internal/conflict"
	"github.com/Les-El/hashi/internal/console"
	"github.com/Les-El/hashi/internal/errors"
	"github.com/Les-El/hashi/internal/hash"
	"github.com/Les-El/hashi/internal/output"
	"github.com/Les-El/hashi/internal/progress"
	"github.com/Les-El/hashi/internal/signals"
)

func main() {
	// Set up signal handling for graceful Ctrl-C interruption
	sigHandler := signals.NewSignalHandler(func() {
		// Cleanup function - called on first Ctrl-C
		// Note: The main cleanup (streams) is handled by defer in main,
		// but this callback is for immediate signal response if needed.
	})
	sigHandler.Start()

	// Initialize color handler for TTY-aware output
	colorHandler := color.NewColorHandler()

	// Initialize error handler
	errHandler := errors.NewErrorHandler(colorHandler)

	// Parse command-line arguments
	cfg, warnings, err := config.ParseArgs(os.Args[1:])
	if err != nil {
		// Streams are not initialized yet, so we use standard stderr
		fmt.Fprintln(os.Stderr, errHandler.FormatError(err))
		os.Exit(config.ExitInvalidArgs)
	}

	// Initialize Global Split Streams (Stdout=Data, Stderr=Context)
	streams, cleanup, err := console.InitStreams(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing I/O: %v\n", err)
		os.Exit(config.ExitInvalidArgs)
	}
	defer cleanup()

	// Display any warnings from conflict resolution (Context -> Stderr)
	if len(warnings) > 0 && !cfg.Quiet {
		fmt.Fprint(streams.Err, conflict.FormatAllWarnings(warnings))
	}

	// Handle help flag (User requested info -> Stdout)
	if cfg.ShowHelp {
		fmt.Fprintln(streams.Out, config.HelpText())
		os.Exit(config.ExitSuccess)
	}

	// Handle version flag (User requested info -> Stdout)
	if cfg.ShowVersion {
		fmt.Fprintln(streams.Out, config.VersionText())
		os.Exit(config.ExitSuccess)
	}

	// 4. Handle stdin file list expansion
	if cfg.HasStdinMarker() {
		cfg.Files = expandStdinFiles(cfg.Files)
	}

	// 5. Discover/Expand files based on paths and options
	// We skip discovery if we are ONLY validating hash strings (no files provided)
	if len(cfg.Files) > 0 || len(cfg.Hashes) == 0 {
		discOpts := hash.DiscoveryOptions{
			Recursive:      cfg.Recursive,
			Hidden:         cfg.Hidden,
			Include:        cfg.Include,
			Exclude:        cfg.Exclude,
			MinSize:        cfg.MinSize,
			MaxSize:        cfg.MaxSize,
			ModifiedAfter:  cfg.ModifiedAfter,
			ModifiedBefore: cfg.ModifiedBefore,
		}
		discovered, err := hash.DiscoverFiles(cfg.Files, discOpts)
		if err != nil {
			fmt.Fprintln(streams.Err, errHandler.FormatError(err))
			os.Exit(errors.DetermineDiscoveryExitCode(err))
		}
		cfg.Files = discovered
	}

	// Determine operation mode based on arguments
	
	// Edge case handling for file + hash comparison mode (Requirements 25.4, 25.5, 25.6)
	if len(cfg.Hashes) > 0 {
		// Check for multiple files with hash strings
		if len(cfg.Files) > 1 {
			fmt.Fprintln(streams.Err, errHandler.FormatError(
				fmt.Errorf("Cannot compare multiple files with hash strings. Use one file at a time.")))
			os.Exit(config.ExitInvalidArgs)
		}
		
		// Check for stdin marker with hash strings
		if cfg.HasStdinMarker() {
			fmt.Fprintln(streams.Err, errHandler.FormatError(
				fmt.Errorf("Cannot use stdin input with hash comparison")))
			os.Exit(config.ExitInvalidArgs)
		}
	}
	
	if len(cfg.Files) == 0 && len(cfg.Hashes) > 0 {
		// Hash validation mode: no files, only hash strings
		exitCode := runHashValidationMode(cfg, colorHandler, streams)
		os.Exit(exitCode)
	}

	if len(cfg.Files) == 1 && len(cfg.Hashes) == 1 {
		// File + hash comparison mode: one file, one hash string
		exitCode := runFileHashComparisonMode(cfg, colorHandler, streams)
		os.Exit(exitCode)
	}

	// Archive verification mode (Task 33)
	if cfg.Verify {
		exitCode := runArchiveVerificationMode(cfg, colorHandler, streams, errHandler)
		os.Exit(exitCode)
	}

	// Standard file processing mode
	if len(cfg.Files) > 0 {
		exitCode := runStandardHashingMode(cfg, colorHandler, streams, errHandler)
		os.Exit(exitCode)
	}

	os.Exit(config.ExitSuccess)
}

// runStandardHashingMode processes multiple files, computing hashes and formatting output.
func runStandardHashingMode(cfg *config.Config, colorHandler *color.Handler, streams *console.Streams, errHandler *errors.Handler) int {
	// 1. Initialize hash computer
	computer, err := hash.NewComputer(cfg.Algorithm)
	if err != nil {
		fmt.Fprintln(streams.Err, errHandler.FormatError(err))
		return config.ExitInvalidArgs
	}

	// 2. Initialize progress bar if needed
	var bar *progress.Bar
	if !cfg.Quiet && !cfg.Bool {
		bar = progress.NewBar(&progress.Options{
			Total:       int64(len(cfg.Files)),
			Description: "Hashing files...",
			Writer:      streams.Err, // Progress goes to Context (stderr)
		})
		defer bar.Finish()
	}

	results := &hash.Result{
		Entries: make([]hash.Entry, 0, len(cfg.Files)),
	}

	start := time.Now()

	// 3. Process files
	for _, path := range cfg.Files {
		entry, err := computer.ComputeFile(path)
		if err != nil {
			results.Errors = append(results.Errors, err)
			results.Entries = append(results.Entries, hash.Entry{
				Original: path,
				Error:    err,
			})
			if !cfg.Quiet {
				// Don't let error messages overwrite progress bar if possible
				if bar != nil && bar.IsEnabled() {
					fmt.Fprint(streams.Err, "\r\033[K") // Clear line
				}
				fmt.Fprintln(streams.Err, errHandler.FormatError(err))
			}
		} else {
			results.Entries = append(results.Entries, *entry)
			results.FilesProcessed++
			results.BytesProcessed += entry.Size
		}

		if bar != nil {
			bar.Increment()
		}
	}

	results.Duration = time.Since(start)

	// 4. Group results for default output
	results.Matches, results.Unmatched = groupResults(results.Entries)

	// 5. Format and output results (Data -> Stdout)
	if cfg.Bool {
		success := false
		if cfg.MatchRequired {
			success = len(results.Matches) > 0
		} else {
			// All files must match (one group, zero unmatched)
			// Special case for single file: always true if no errors
			if len(results.Entries) == 1 && len(results.Errors) == 0 {
				success = true
			} else {
				success = len(results.Matches) == 1 && len(results.Unmatched) == 0
			}
		}
		if success {
			fmt.Fprintln(streams.Out, "true")
		} else {
			fmt.Fprintln(streams.Out, "false")
		}
	} else if !cfg.Quiet {
		formatter := output.NewFormatter(cfg.OutputFormat, cfg.PreserveOrder)
		fmt.Fprintln(streams.Out, formatter.Format(results))
	}

	// 6. Determine exit code
	return errors.DetermineExitCode(cfg, results)
}

// runArchiveVerificationMode verifies the integrity of archive files.
func runArchiveVerificationMode(cfg *config.Config, colorHandler *color.Handler, streams *console.Streams, errHandler *errors.Handler) int {
	verifier := archive.NewVerifier()
	verifier.SetVerbose(cfg.Verbose)

	allPassed := true
	anyProcessed := false

	// results for standard hashing if needed (mixed mode)
	hResults := &hash.Result{
		Entries: make([]hash.Entry, 0),
	}
	computer, _ := hash.NewComputer(cfg.Algorithm)

	for _, path := range cfg.Files {
		if verifier.IsArchiveFile(path) {
			anyProcessed = true
			result, err := verifier.VerifyZIP(path)
			if err != nil {
				fmt.Fprintln(streams.Err, errHandler.FormatError(err))
				allPassed = false
				continue
			}

			if !cfg.Quiet {
				if cfg.Bool {
					// In bool mode, we don't print individual results here
				} else {
					fmt.Fprint(streams.Out, verifier.FormatResult(result, true)) // Always verbose for verify mode?
				}
			}

			if !result.Passed {
				allPassed = false
			}
		} else {
			// Process non-ZIP with standard hashing (Requirement 15.8)
			anyProcessed = true
			entry, err := computer.ComputeFile(path)
			if err != nil {
				fmt.Fprintln(streams.Err, errHandler.FormatError(err))
				allPassed = false
				continue
			}
			hResults.Entries = append(hResults.Entries, *entry)
			
			if !cfg.Quiet && !cfg.Bool {
				fmt.Fprintf(streams.Out, "%-12s %s\n", path, entry.Hash)
			}
		}
	}

	if cfg.Bool {
		if allPassed {
			fmt.Fprintln(streams.Out, "true")
		} else {
			fmt.Fprintln(streams.Out, "false")
		}
	}

	if !anyProcessed {
		return config.ExitSuccess
	}

	if !allPassed {
		return config.ExitIntegrityFail
	}
	return config.ExitSuccess
}

// groupResults categorizes entries into matches and unique hashes.
func groupResults(entries []hash.Entry) ([]hash.MatchGroup, []hash.Entry) {
	groups := make(map[string][]hash.Entry)
	for _, e := range entries {
		if e.Error == nil {
			groups[e.Hash] = append(groups[e.Hash], e)
		}
	}

	var matches []hash.MatchGroup
	var unmatched []hash.Entry

	for h, groupEntries := range groups {
		if len(groupEntries) > 1 {
			matches = append(matches, hash.MatchGroup{
				Hash:    h,
				Entries: groupEntries,
				Count:   len(groupEntries),
			})
		} else {
			unmatched = append(unmatched, groupEntries[0])
		}
	}

	return matches, unmatched
}

// runHashValidationMode validates hash strings and displays results.
// This mode is triggered when no files are provided, only hash strings.
// Requirements: 24.1, 24.2, 24.3
func runHashValidationMode(cfg *config.Config, colorHandler *color.Handler, streams *console.Streams) int {
	allValid := true
	
	for _, hashStr := range cfg.Hashes {
		// Detect possible algorithms for this hash string
		algorithms := hash.DetectHashAlgorithm(hashStr)
		
		if len(algorithms) == 0 {
			// Invalid hash format
			allValid = false
			if !cfg.Quiet {
				// Context/Error -> Stderr
				fmt.Fprintf(streams.Err, "%s %s - Invalid hash format\n", 
					colorHandler.Red("✗"), hashStr)
				fmt.Fprintf(streams.Err, "  Hash strings must contain only hexadecimal characters (0-9, a-f, A-F)\n")
				fmt.Fprintf(streams.Err, "  and have a valid length (8, 32, 40, 64, or 128 characters)\n")
			}
		} else {
			// Valid hash format
			if !cfg.Quiet {
				// Context/Info -> Stderr
				fmt.Fprintf(streams.Err, "%s %s - Valid hash\n", 
					colorHandler.Green("✓"), hashStr)
				
				if len(algorithms) == 1 {
					fmt.Fprintf(streams.Err, "  Algorithm: %s\n", algorithms[0])
				} else {
					fmt.Fprintf(streams.Err, "  Possible algorithms: %s\n", formatAlgorithmList(algorithms))
				}
			}
		}
	}
	
	// Exit with appropriate code
	if allValid {
		return config.ExitSuccess
	} else {
		return config.ExitInvalidArgs
	}
}

// formatAlgorithmList formats a list of algorithms for display.
func formatAlgorithmList(algorithms []string) string {
	if len(algorithms) == 0 {
		return ""
	}
	if len(algorithms) == 1 {
		return algorithms[0]
	}
	
	result := ""
	for i, alg := range algorithms {
		if i > 0 {
			if i == len(algorithms)-1 {
				result += " or "
			} else {
				result += ", "
			}
		}
		result += alg
	}
	return result
}

// runFileHashComparisonMode compares a file's hash against a provided hash string.
// This mode is triggered when exactly one file and one hash string are provided.
// Requirements: 25.1, 25.2, 25.3
func runFileHashComparisonMode(cfg *config.Config, colorHandler *color.Handler, streams *console.Streams) int {
	filePath := cfg.Files[0]
	expectedHash := cfg.Hashes[0] // Already normalized to lowercase by ClassifyArguments
	
	// Create hash computer
	computer, err := hash.NewComputer(cfg.Algorithm)
	if err != nil {
		if !cfg.Quiet {
			fmt.Fprintf(streams.Err, "%s Failed to initialize hash computer: %v\n", 
				colorHandler.Red("✗"), err)
		}
		return config.ExitInvalidArgs
	}
	
	// Compute file hash
	entry, err := computer.ComputeFile(filePath)
	if err != nil {
		if !cfg.Quiet {
			fmt.Fprintf(streams.Err, "%s Failed to compute hash for %s: %v\n", 
				colorHandler.Red("✗"), filePath, err)
		}
		// Return appropriate exit code based on error type
		if os.IsNotExist(err) {
			return config.ExitFileNotFound
		}
		if os.IsPermission(err) {
			return config.ExitPermissionErr
		}
		return config.ExitPartialFailure
	}
	
	// Compare hashes (case-insensitive)
	computedHash := strings.ToLower(entry.Hash)
	expectedHashLower := strings.ToLower(expectedHash)
	
	if computedHash == expectedHashLower {
		// Hashes match
		if cfg.Bool {
			// Boolean output mode: just output "true" -> Data (Stdout)
			fmt.Fprintln(streams.Out, "true")
		} else if !cfg.Quiet {
			// Regular output: display PASS -> Data (Stdout)
			fmt.Fprintf(streams.Out, "%s %s\n", colorHandler.Green("PASS"), filePath)
		}
		return config.ExitSuccess
	} else {
		// Hashes don't match
		if cfg.Bool {
			// Boolean output mode: just output "false" -> Data (Stdout)
			fmt.Fprintln(streams.Out, "false")
		} else if !cfg.Quiet {
			// Regular output: display FAIL -> Data (Stdout)
			fmt.Fprintf(streams.Out, "%s %s\n", colorHandler.Red("FAIL"), filePath)
			fmt.Fprintf(streams.Out, "  Expected: %s\n", expectedHash)
			fmt.Fprintf(streams.Out, "  Computed: %s\n", entry.Hash)
		}
				return config.ExitNoMatches // Exit code 1 for mismatch
			}
		}
		
		// expandStdinFiles reads file paths from stdin and adds them to the file list.
		func expandStdinFiles(files []string) []string {
			var result []string
			
			// Remove the "-" marker
			for _, f := range files {
				if f != "-" {
					result = append(result, f)
				}
			}
			
			// Read from stdin
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				path := strings.TrimSpace(scanner.Text())
				if path != "" {
					result = append(result, path)
				}
			}
			
			return result
		}
		