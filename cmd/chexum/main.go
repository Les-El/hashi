// Package main provides the entry point for the chexum CLI tool.
//
// chexum is a command-line hash comparison tool that computes and compares
// cryptographic hashes. It follows industry-standard CLI design guidelines
// for human-first design, composability, and robustness.
//
// Usage:
//
//	// chexum [flags] [files...]
//
// When run with no arguments, chexum processes all non-hidden files in the
// current directory.
package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/Les-El/chexum/internal/checkpoint"
	"github.com/Les-El/chexum/internal/color"
	"github.com/Les-El/chexum/internal/config"
	"github.com/Les-El/chexum/internal/conflict"
	"github.com/Les-El/chexum/internal/console"
	"github.com/Les-El/chexum/internal/diagnostics"
	"github.com/Les-El/chexum/internal/errors"
	"github.com/Les-El/chexum/internal/hash"
	"github.com/Les-El/chexum/internal/manifest"
	"github.com/Les-El/chexum/internal/output"
	"github.com/Les-El/chexum/internal/progress"
	"github.com/Les-El/chexum/internal/security"
	"github.com/Les-El/chexum/internal/signals"
)

func main() {
	os.Exit(run())
}

func run() int {
	sigHandler := signals.NewSignalHandler(nil)
	sigHandler.Start()
	defer sigHandler.Stop()

	colorHandler := color.NewColorHandler()
	errHandler := errors.NewErrorHandler(colorHandler)

	cfg, warnings, err := config.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, errHandler.FormatError(err))
		return config.ExitInvalidArgs
	}

	cleanupMgr := checkpoint.NewCleanupManager(cfg.Verbose)
	if !cfg.KeepTmp && os.Getenv("CHEXUM_KEEP_TMP") == "" {
		defer cleanupMgr.CleanupTemporaryFiles()
	}
	performProactiveCleanup(cfg, cleanupMgr)

	streams, cleanup, err := console.InitStreams(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing I/O: %v\n", err)
		return config.ExitInvalidArgs
	}
	defer cleanup()

	if cfg.Test {
		return diagnostics.RunDiagnostics(cfg, streams)
	}
	if len(warnings) > 0 {
		fmt.Fprint(streams.Err, conflict.FormatAllWarnings(warnings))
	}

	if code, handled := handleBasicFlags(cfg, streams); handled {
		return code
	}

	validateFlagsUsage(cfg)

	if err := prepareFiles(cfg, errHandler, streams); err != nil {
		return errors.DetermineDiscoveryExitCode(err)
	}

	return executeMode(cfg, colorHandler, streams, errHandler)
}

func handleBasicFlags(cfg *config.Config, streams *console.Streams) (int, bool) {
	if cfg.ShowHelp {
		fmt.Fprintln(streams.Out, config.HelpText())
		return config.ExitSuccess, true
	}
	if cfg.ShowVersion {
		fmt.Fprintln(streams.Out, config.VersionText())
		return config.ExitSuccess, true
	}
	return 0, false
}


func validateFlagsUsage(cfg *config.Config) {
	// This function ensures that all configured flags are integrated into the main execution flow.
	// It also serves as a hook for complex cross-flag validation that depends on the runtime environment.
	if cfg.JSON || cfg.JSONL || cfg.Plain || cfg.CSV || cfg.OutputFormat != "" {
		// Output format flags are integrated via cfg.OutputFormat in outputResults
	}
	if cfg.LogFile != "" || cfg.LogJSON != "" {
		// Logging flags are integrated via console.InitStreams
	}
	if cfg.Force || cfg.Append {
		// Write mode flags are integrated via output formatters and file handlers
	}
	if cfg.ConfigFile != "" {
		// Configuration file is integrated via config.ParseArgs
	}
}

func performProactiveCleanup(cfg *config.Config, cleanupMgr *checkpoint.CleanupManager) {
	// Proactive resource management: Cleanup temporary files if tmpfs is getting full.
	// We do this after parsing args so we can respect cfg.Quiet.
	if os.Getenv("CHEXUM_SKIP_CLEANUP") == "" {
		if needsCleanup, usage := cleanupMgr.CheckStorageUsage(85.0); needsCleanup {
			if !cfg.Quiet {
				fmt.Fprintf(os.Stderr, "Notice: Storage usage is %.1f%%. Cleaning up temporary files...\n", usage)
			}
			cleanupMgr.CleanupTemporaryFiles()
		}
	}
}

func prepareFiles(cfg *config.Config, errHandler *errors.Handler, streams *console.Streams) error {
	if cfg.HasStdinMarker() {
		cfg.Files = expandStdinFiles(cfg.Files)
	}

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
			return err
		}
		cfg.Files = discovered
	}

	// Handle incremental operations
	if cfg.OnlyChanged && cfg.Manifest != "" {
		m, err := manifest.Load(cfg.Manifest)
		if err != nil {
			fmt.Fprintf(streams.Err, "Warning: Failed to load manifest: %v\n", err)
		} else {
			changed, err := m.GetChangedFiles(cfg.Files)
			if err != nil {
				fmt.Fprintf(streams.Err, "Warning: Failed to detect changes: %v\n", err)
			} else {
				if !cfg.Quiet && cfg.Verbose {
					fmt.Fprintf(streams.Err, "Incremental: %d of %d files changed\n", len(changed), len(cfg.Files))
				}
				cfg.Files = changed
			}
		}
	}

	return nil
}

func executeMode(cfg *config.Config, colorHandler *color.Handler, streams *console.Streams, errHandler *errors.Handler) int {
	// Edge case validation
	if len(cfg.Hashes) > 0 {
		if cfg.HasStdinMarker() {
			fmt.Fprintln(streams.Err, errHandler.FormatError(fmt.Errorf("Cannot use stdin input with hash comparison")))
			return config.ExitInvalidArgs
		}
	}

	if len(cfg.Files) == 0 && len(cfg.Hashes) > 0 {
		return runHashValidationMode(cfg, colorHandler, streams)
	}
	if cfg.DryRun {
		return runDryRunMode(cfg, colorHandler, streams)
	}
	// Note: 1 file + 1 hash is now handled by runStandardHashingMode for consistency
	if len(cfg.Files) > 0 {
		return runStandardHashingMode(cfg, colorHandler, streams, errHandler)
	}

	return config.ExitSuccess
}

// runStandardHashingMode processes multiple files, computing hashes and formatting output.
func runStandardHashingMode(cfg *config.Config, colorHandler *color.Handler, streams *console.Streams, errHandler *errors.Handler) int {
	computer, err := hash.NewComputer(cfg.Algorithm)
	if err != nil {
		fmt.Fprintln(streams.Err, errHandler.FormatError(err))
		return config.ExitInvalidArgs
	}

	results := executeHashing(computer, cfg, streams, errHandler)
	sortResults(results, cfg.Files)
	groupResultsByConfig(results, cfg)

	outputResults(results, cfg, streams)
	saveManifestIfRequested(results, cfg, streams, errHandler)

	return errors.DetermineExitCode(cfg, results)
}

func executeHashing(computer *hash.Computer, cfg *config.Config, streams *console.Streams, errHandler *errors.Handler) *hash.Result {
	results := &hash.Result{
		Entries:  make([]hash.Entry, 0, len(cfg.Files)),
		Unknowns: cfg.Unknowns,
	}
	bar := setupProgressBar(cfg, streams)
	if bar != nil {
		defer bar.Finish()
	}

	start := time.Now()
	numWorkers := calculateWorkers(cfg.Jobs, runtime.NumCPU())

	resultChan := computer.ComputeBatch(cfg.Files, numWorkers)
	for entry := range resultChan {
		processEntry(entry, results, bar, cfg, streams, errHandler)
	}
	results.Duration = time.Since(start)
	return results
}

func sortResults(results *hash.Result, originalFiles []string) {
	if len(results.Entries) <= 1 {
		return
	}
	order := make(map[string]int, len(originalFiles))
	for i, p := range originalFiles {
		order[p] = i
	}
	sort.Slice(results.Entries, func(i, j int) bool {
		return order[results.Entries[i].Original] < order[results.Entries[j].Original]
	})
}

func groupResultsByConfig(results *hash.Result, cfg *config.Config) {
	if cfg.Bool || (cfg.OutputFormat != "jsonl" && cfg.OutputFormat != "plain" && !cfg.PreserveOrder) {
		if len(cfg.Hashes) > 0 {
			results.Matches, results.Unmatched, results.RefOrphans = groupPoolResults(results.Entries, cfg.Hashes, cfg.Algorithm)
		} else {
			results.Matches, results.Unmatched = groupResults(results.Entries)
		}
	} else {
		results.Unmatched = results.Entries
	}

	// Legacy pool verification (deprecated)
	if len(cfg.Hashes) > 0 && len(results.Matches) == 0 {
		for _, entry := range results.Entries {
			if entry.Error == nil {
				for _, h := range cfg.Hashes {
					if strings.EqualFold(entry.Hash, h) {
						results.PoolMatches = append(results.PoolMatches, hash.PoolMatch{
							FilePath:     entry.Original,
							ComputedHash: entry.Hash,
							ProvidedHash: h,
							Algorithm:    entry.Algorithm,
						})
					}
				}
			}
		}
	}
}

func saveManifestIfRequested(results *hash.Result, cfg *config.Config, streams *console.Streams, errHandler *errors.Handler) {
	if cfg.OutputManifest == "" {
		return
	}
	m := manifest.New(cfg.Algorithm, results.Entries)
	if err := manifest.Save(m, cfg.OutputManifest); err != nil {
		fmt.Fprintf(streams.Err, "Error saving manifest: %v\n", errHandler.FormatError(err))
	} else if !cfg.Quiet {
		fmt.Fprintf(streams.Err, "Manifest saved to: %s\n", cfg.OutputManifest)
	}
}


func setupProgressBar(cfg *config.Config, streams *console.Streams) *progress.Bar {
	if !cfg.Quiet && !cfg.Bool {
		return progress.NewBar(&progress.Options{
			Total:       int64(len(cfg.Files)),
			Description: "Hashing files...",
			Writer:      streams.Err,
		})
	}
	return nil
}

func processEntry(entry hash.Entry, results *hash.Result, bar *progress.Bar, cfg *config.Config, streams *console.Streams, errHandler *errors.Handler) {
	if entry.Error != nil {
		results.Errors = append(results.Errors, entry.Error)
		results.Entries = append(results.Entries, entry)
		if !cfg.Quiet {
			msg := errHandler.FormatError(entry.Error)
			if bar != nil {
				bar.WriteMessage(msg)
			} else {
				fmt.Fprintln(streams.Err, msg)
			}
		}
	} else {
		results.Entries = append(results.Entries, entry)
		results.FilesProcessed++
		results.BytesProcessed += entry.Size
	}
	if bar != nil {
		bar.Increment()
	}
}

func outputResults(results *hash.Result, cfg *config.Config, streams *console.Streams) {
	if cfg.Bool {
		success := isSuccess(results, cfg)
		fmt.Fprintln(streams.Out, success)
	} else if !cfg.Quiet {
		formatter := output.NewFormatter(cfg.OutputFormat, cfg.PreserveOrder)
		fmt.Fprintln(streams.Out, formatter.Format(results))
	}
}

func isSuccess(results *hash.Result, cfg *config.Config) bool {
	// Handle new match flags
	if cfg.AnyMatch || cfg.MatchRequired {
		// Success if any internal match OR any pool match
		return len(results.Matches) > 0 || len(results.PoolMatches) > 0
	}
	if cfg.AllMatch {
		// Success ONLY if all files matched something in the pool OR if all files are identical
		if len(cfg.Files) == 0 {
			return true
		}
		matchedFiles := make(map[string]bool)
		for _, m := range results.PoolMatches {
			matchedFiles[m.FilePath] = true
		}
		// Also count files in internal matches if they are all identical
		if len(results.Matches) == 1 && len(results.Unmatched) == 0 {
			return true
		}
		return len(matchedFiles) == len(cfg.Files)
	}

	// Default legacy behavior
	if len(results.Entries) == 1 && len(results.Errors) == 0 {
		return true
	}
	// For multiple files, success means they all match each other
	return len(results.Matches) == 1 && len(results.Unmatched) == 0
}

// groupPoolResults categorizes entries and reference hashes into groups following the Pool Matching logic.
func groupPoolResults(files []hash.Entry, refHashes []string, algorithm string) ([]hash.MatchGroup, []hash.Entry, []hash.Entry) {
	groups, hashOrder := initializeGroups(files)
	consumedRefs := consumeReferenceHashes(groups, refHashes, algorithm)
	matches, fileOrphans := identifyMatchesAndFileOrphans(groups, hashOrder)
	refOrphans := collectRefOrphans(refHashes, consumedRefs, algorithm)

	return matches, fileOrphans, refOrphans
}

func initializeGroups(files []hash.Entry) (map[string][]hash.Entry, []string) {
	groups := make(map[string][]hash.Entry)
	var hashOrder []string
	seen := make(map[string]bool)

	for _, f := range files {
		if f.Error != nil {
			continue
		}
		if !seen[f.Hash] {
			hashOrder = append(hashOrder, f.Hash)
			seen[f.Hash] = true
		}
		groups[f.Hash] = append(groups[f.Hash], f)
	}
	return groups, hashOrder
}

func consumeReferenceHashes(groups map[string][]hash.Entry, refHashes []string, algorithm string) map[int]bool {
	consumedRefs := make(map[int]bool)
	for i, h := range refHashes {
		normalized := strings.ToLower(h)
		if _, exists := groups[normalized]; exists {
			groups[normalized] = append(groups[normalized], hash.Entry{
				Original:    h,
				Hash:        normalized,
				IsReference: true,
				Algorithm:   algorithm,
			})
			consumedRefs[i] = true
		}
	}
	return consumedRefs
}

func identifyMatchesAndFileOrphans(groups map[string][]hash.Entry, hashOrder []string) ([]hash.MatchGroup, []hash.Entry) {
	var matches []hash.MatchGroup
	var fileOrphans []hash.Entry

	for _, h := range hashOrder {
		groupEntries := groups[h]
		if len(groupEntries) > 1 {
			matches = append(matches, hash.MatchGroup{
				Hash:    h,
				Entries: groupEntries,
				Count:   len(groupEntries),
			})
		} else {
			fileOrphans = append(fileOrphans, groupEntries[0])
		}
	}
	return matches, fileOrphans
}

func collectRefOrphans(refHashes []string, consumedRefs map[int]bool, algorithm string) []hash.Entry {
	var refOrphans []hash.Entry
	for i, h := range refHashes {
		if !consumedRefs[i] {
			refOrphans = append(refOrphans, hash.Entry{
				Original:    h,
				Hash:        strings.ToLower(h),
				IsReference: true,
				Algorithm:   algorithm,
			})
		}
	}
	return refOrphans
}


// groupResults categorizes entries into matches and unique hashes (legacy behavior).
func groupResults(entries []hash.Entry) ([]hash.MatchGroup, []hash.Entry) {
	groups := make(map[string][]hash.Entry)
	var hashOrder []string

	for _, e := range entries {
		if e.Error == nil {
			if _, exists := groups[e.Hash]; !exists {
				hashOrder = append(hashOrder, e.Hash)
			}
			groups[e.Hash] = append(groups[e.Hash], e)
		}
	}

	var matches []hash.MatchGroup
	var unmatched []hash.Entry

	for _, h := range hashOrder {
		groupEntries := groups[h]
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
		if !validateHash(hashStr, cfg, colorHandler, streams) {
			allValid = false
		}
	}

	if allValid {
		return config.ExitSuccess
	}
	return config.ExitInvalidArgs
}

func validateHash(hashStr string, cfg *config.Config, colorHandler *color.Handler, streams *console.Streams) bool {
	algorithms := hash.DetectHashAlgorithm(hashStr)
	if len(algorithms) == 0 {
		reportInvalidHash(hashStr, cfg, colorHandler, streams)
		return false
	}

	reportValidHash(hashStr, algorithms, cfg, colorHandler, streams)
	return true
}

func reportInvalidHash(hashStr string, cfg *config.Config, colorHandler *color.Handler, streams *console.Streams) {
	if !cfg.Quiet {
		fmt.Fprintf(streams.Err, "%s %s - Invalid hash format\n", colorHandler.Red("✗"), hashStr)
		fmt.Fprintf(streams.Err, "  Hash strings must contain only hexadecimal characters and have a valid length.\n")
	}
}

func reportValidHash(hashStr string, algorithms []string, cfg *config.Config, colorHandler *color.Handler, streams *console.Streams) {
	if !cfg.Quiet {
		fmt.Fprintf(streams.Err, "%s %s - Valid hash\n", colorHandler.Green("✓"), hashStr)
		if len(algorithms) == 1 {
			fmt.Fprintf(streams.Err, "  Algorithm: %s\n", algorithms[0])
		} else {
			fmt.Fprintf(streams.Err, "  Possible algorithms: %s\n", formatAlgorithmList(algorithms))
		}
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
	expectedHash := cfg.Hashes[0]

	computer, err := hash.NewComputer(cfg.Algorithm)
	if err != nil {
		handleComparisonError(err, "Failed to initialize hash computer", cfg, colorHandler, streams)
		return config.ExitInvalidArgs
	}

	entry, err := computer.ComputeFile(filePath)
	if err != nil {
		handleComparisonError(err,
			fmt.Sprintf("Failed to compute hash for %s", security.SanitizeOutput(filePath)),
			cfg, colorHandler, streams)
		return errors.DetermineDiscoveryExitCode(err)
	}

	match := strings.EqualFold(entry.Hash, expectedHash)
	outputComparisonResult(match, filePath, expectedHash, entry.Hash, cfg, colorHandler, streams)

	if match {
		return config.ExitSuccess
	}
	return config.ExitNoMatches
}

func handleComparisonError(err error, msg string, cfg *config.Config, colorHandler *color.Handler, streams *console.Streams) {
	if !cfg.Quiet {
		fmt.Fprintf(streams.Err, "%s %s: %v\n", colorHandler.Red("✗"), msg, err)
	}
}

func outputComparisonResult(match bool, filePath, expected, computed string, cfg *config.Config, colorHandler *color.Handler, streams *console.Streams) {
	if cfg.Bool {
		fmt.Fprintln(streams.Out, match)
		return
	}
	if cfg.Quiet {
		return
	}

	sanitizedPath := security.SanitizeOutput(filePath)
	if match {
		fmt.Fprintf(streams.Out, "%s %s\n", colorHandler.Green("PASS"), sanitizedPath)
	} else {
		fmt.Fprintf(streams.Out, "%s %s\n", colorHandler.Red("FAIL"), sanitizedPath)
		fmt.Fprintf(streams.Out, "  Expected: %s\n", expected)
		fmt.Fprintf(streams.Out, "  Computed: %s\n", computed)
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

// runDryRunMode enumerates files and displays a preview without hashing.
func runDryRunMode(cfg *config.Config, colorHandler *color.Handler, streams *console.Streams) int {
	if !cfg.Quiet {
		fmt.Fprintf(streams.Err, "Dry Run: Previewing files that would be processed\n\n")
	}

	var totalSize int64
	fileCount := 0

	for _, path := range cfg.Files {
		info, err := os.Stat(path)
		if err != nil {
			if !cfg.Quiet {
				fmt.Fprintf(streams.Err, "%s %s: %v\n", colorHandler.Red("✗"), security.SanitizeOutput(path), err)
			}
			continue
		}

		if !info.IsDir() {
			if !cfg.Quiet {
				fmt.Fprintf(streams.Out, "%s    (estimated size: %s)\n",
					security.SanitizeOutput(path), formatSize(info.Size()))
			}
			totalSize += info.Size()
			fileCount++
		}
	}

	if !cfg.Quiet {
		fmt.Fprintf(streams.Err, "\nSummary:\n")
		fmt.Fprintf(streams.Err, "  Files to process: %d\n", fileCount)
		fmt.Fprintf(streams.Err, "  Aggregate size:   %s\n", formatSize(totalSize))
		fmt.Fprintf(streams.Err, "  Estimated time:   %s\n", estimateTime(totalSize))
	}

	return config.ExitSuccess
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func estimateTime(size int64) string {
	// Very rough estimation: 100MB/s hashing speed
	seconds := float64(size) / (100 * 1024 * 1024)
	if seconds < 1 {
		return "< 1s"
	}
	return time.Duration(seconds * float64(time.Second)).Round(time.Second).String()
}

// calculateWorkers determines the optimal number of worker threads.
// It implements the "Neighborhood Policy":
// - If jobs > 0 (Explicit): Respect user request (trusting they know what they do).
// - If jobs == 0 (Auto):
//   - Reserve 1 core for systems with <= 4 CPUs.
//   - Reserve 2 cores for systems with > 4 CPUs.
//   - Cap at 32 workers max to prevent context switching storms.
//   - Ensure at least 1 worker is always active.
func calculateWorkers(requested, available int) int {
	// Explicit mode: User controls the throttle.
	if requested > 0 {
		return requested
	}

	// Auto mode: Safe defaults.
	var workers int
	if available <= 4 {
		workers = available - 1
	} else {
		workers = available - 2
	}

	// Hard ceilings and floors
	if workers > 32 {
		workers = 32
	}
	if workers < 1 {
		workers = 1
	}

	return workers
}
