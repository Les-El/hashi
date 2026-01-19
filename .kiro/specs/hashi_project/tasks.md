# Implementation Plan: CLI Guidelines Review for Hashi

## Overview

This implementation plan transforms the `hashi` CLI tool to follow industry-standard CLI design guidelines. The work is organized into phases, starting with critical improvements that significantly impact usability, followed by important enhancements, and finally nice-to-have features.

The implementation follows a test-driven approach with both unit tests and property-based tests to ensure correctness and robustness.

## IMPORTANT: Fresh Start Approach

**This is a complete rewrite, not a refactor.** The existing `main.go` and `main_test.go` contain legacy code that does not align with the new architecture. 

**For all developers:**
- Do NOT attempt to integrate or refactor the old code
- The old code has been backed up in a separate cloned repository
- Start fresh with the new package structure defined in Task 1
- Reference the design.md for component interfaces and architecture

**Rationale:** The new spec introduces modular architecture, new components (color handling, progress bars, signal handling, archive verification, conflict detection), and significantly different behavior. Attempting to refactor would create more confusion than starting clean.

## Tasks

- [x] 1. Set up enhanced project structure and dependencies (FRESH START)
  - **NOTE: This is a fresh implementation. The existing main.go contains legacy code that will be replaced entirely. Do not attempt to refactor or integrate the old code.**
  - Archive or delete existing main.go and main_test.go (backup exists in cloned repo)
  - Create new package structure:
    - `cmd/hashi/main.go` - Entry point
    - `internal/config/` - Configuration and argument parsing
    - `internal/hash/` - Hash computation logic
    - `internal/output/` - Formatters (default, verbose, JSON, plain)
    - `internal/color/` - TTY detection and color handling
    - `internal/progress/` - Progress bar component
    - `internal/errors/` - Error handling and formatting
    - `internal/signals/` - Signal handling (Ctrl-C)
    - `internal/archive/` - ZIP verification
    - `internal/conflict/` - Flag conflict detection
  - Add dependencies to go.mod:
    - `github.com/fatih/color` or similar for color output
    - `github.com/schollz/progressbar/v3` or similar for progress bars
    - `github.com/spf13/pflag` or similar for POSIX-compliant flag parsing
    - `github.com/joho/godotenv` for .env file support
  - Set up testing framework with property-based testing support (use `testing/quick` from stdlib)
  - Review README.md for project structure overview, correcting and adding as needed+
  - _Requirements: All_

- [x] 2. Implement TTY detection and color output system
  - [x] 2.1 Create ColorHandler with TTY detection
    - Implement TTY detection using `term.IsTerminal()`
    - Check `NO_COLOR` environment variable
    - Provide methods for colorizing text (green, red, yellow, blue, cyan, gray)
    - _Requirements: 2.1, 2.2, 2.3, 5.2_
  
  - [x] 2.2 Write property test for color output
    - **Property 2: Color output respects TTY detection**
    - **Validates: Requirements 2.2, 2.3**
  
  - [x] 2.3 Write unit tests for ColorHandler
    - Test TTY detection logic
    - Test NO_COLOR environment variable
    - Test color code generation
    - _Requirements: 2.1, 2.2, 2.3_

- [x] 3. Implement enhanced argument parser with full flag support
  - [x] 3.1 Create Config struct with all new fields
    - Add Quiet, PreserveOrder, MatchRequired, NoMatchRequired fields
    - Add OutputFormat field with validation
    - Add all filtering fields (Include, Exclude, MinSize, MaxSize, dates)
    - _Requirements: 4.1, 4.2, 4.4, 4.5_
  
  - [x] 3.2 Implement ParseArgs function with short and long flags
    - Support both `-v` and `--verbose` for all flags
    - Accept flags in any order
    - Support `-` for stdin
    - Support `--flag=value` and `--flag value` syntax
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_
  
  - [x] 3.3 Write property test for flag order independence
    - **Property 6: Flags accept any order**
    - **Validates: Requirements 4.4**
  
  - [x] 3.4 Write unit tests for argument parsing
    - Test short and long flag variants
    - Test stdin support with `-`
    - Test flag validation
    - Test error cases
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [x] 3.5 **HUMAN-IN-THE-LOOP: Review and incorporate flag logic design**
  - [x] 3.5.1 Review docs/design/flag_logic.md for flag precedence hierarchy
    - **DECISION MADE:** Adopt matrix-based conflict resolution system
    - **DECISION MADE:** --bool overrides --quiet (no conflict)
    - **DECISION MADE:** Eliminate --no-match-required flag (use shell negation)
    - **DECISION MADE:** Remove unused ConflictType: Requires
    - Updated requirements document with detailed flag precedence rules
    - Updated design document with conflict resolution strategies
    - _Requirements: 17.1, 26.3, 26.4_
  
  - [x] 3.5.2 Review docs/design/configuration.md for configuration system design
    - **DECISION MADE:** Switch from JSON to TOML configuration format
    - **DECISION MADE:** Skip bookmarks and templates (violates hashi principles)
    - **DECISION MADE:** Use simple, secure config display (no progressive disclosure)
    - **DECISION MADE:** Add per-directory configuration support
    - **DECISION MADE:** Skip configuration validation commands (no real purpose)
    - Update requirements document with comprehensive configuration features
    - Update design document with security model and validation
    - _Requirements: 5.1, 5.2, 5.3, 23.1, 23.2, 23.3_
  
  - [x] 3.5.3 Update spec documents based on decisions
    - Updated requirements.md with flag precedence rules and TOML config
    - Updated design.md with matrix-based conflict resolution
    - Updated design.md with simplified configuration system
    - Removed --no-match-required from all specifications
    - Added new requirement for flag precedence and override system
    - Ensured all design decisions are captured in requirements

- [x] 3.6 Implement config command rejection
  - [x] 3.6.1 Add config command detection to ParseArgs
    - Detect when ANY positional argument is "config" (not just first position)
    - Return helpful error message with config file names (no full paths for security)
    - Include link to configuration documentation
    - Exit with code 3 (invalid arguments)
    - _Requirements: 27.1, 27.2, 27.3, 27.4, 27.5_
  
  - [x] 3.6.2 Write property test for config command rejection
    - **Property 35: Config command is rejected with helpful error**
    - **Validates: Requirements 27.1, 27.5**
  
  - [x] 3.6.3 Write unit tests for config command handling
    - Test `hashi config` is rejected
    - Test `hashi config set` is rejected
    - Test `hashi config get` is rejected
    - Test `hashi file.txt config` is rejected (config anywhere in args)
    - Test `hashi --verbose config --json` is rejected (config anywhere in args)
    - Test error message contains config file names (not full paths)
    - Test error message contains documentation link
    - Test exit code is 3
    - _Requirements: 27.1, 27.2, 27.3, 27.4, 27.5_

- [x] 4. Implement environment variable and configuration system
  - [x] 4.1 Create EnvConfig struct and LoadEnvConfig function
    - Read NO_COLOR, DEBUG, TMPDIR, HOME, XDG_CONFIG_HOME
    - Support HASHI_* prefixed variables
    - _Requirements: 5.2_
  
  - [x] 4.2 Implement .env file support
    - Parse .env files in working directory
    - Merge with system environment variables
    - _Requirements: 5.4_
  
  - [x] 4.3 Implement config file loading (TOML/text)
    - Parse TOML config files
    - Parse text config files (one arg per line)
    - Merge with flags and env vars
    - _Requirements: 5.1, 5.3_
  
  - [x] 4.4 Write property test for configuration precedence
    - **Property 7: Configuration precedence is correct**
    - **Validates: Requirements 5.3**
  
  - [x] 4.5 Write unit tests for configuration system
    - Test env var loading
    - Test .env file parsing
    - Test config file parsing
    - Test precedence rules
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [x] 5. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 6. Implement output formatters with new default format
  - [x] 6.1 Create OutputFormatter interface and implementations
    - Implement DefaultFormatter (grouped by matches with blank lines)
    - Implement PreserveOrderFormatter (input order maintained)
    - Implement VerboseFormatter (detailed with summaries)
    - Implement JSONFormatter (machine-readable)
    - Implement PlainFormatter (tab-separated for scripting)
    - _Requirements: 2.5, 7.1, 7.2, 7.4, 7.5_
  
  - [x] 6.2 Write property test for default output grouping
    - **Property 19: Default output groups by matches**
    - **Validates: Requirements 2.5**
  
  - [x] 6.3 Write property test for preserve-order flag
    - **Property 20: Preserve-order flag maintains input order**
    - **Validates: Requirements 2.5**
  
  - [x] 6.4 Write property test for JSON validity
    - **Property 10: JSON output is valid**
    - **Validates: Requirements 7.1**
  
  - [x] 6.5 Write property test for plain output format
    - **Property 11: Plain output is line-based**
    - **Validates: Requirements 7.2**
  
  - [x] 6.6 Write unit tests for all formatters
    - Test each formatter with sample data
    - Test edge cases (empty results, single file, many matches)
    - Verify output format correctness
    - _Requirements: 2.5, 7.1, 7.2, 7.4, 7.5_

- [x] 7. Implement progress indicators for long operations
  - [x] 7.1 Create ProgressBar component
    - Show progress bar for operations >100ms
    - Display percentage, count, and ETA
    - Support both file-level and batch-level progress
    - Hide progress when output is not a TTY
    - _Requirements: 2.4, 6.2, 6.3_
  
  - [x] 7.2 Write property test for progress indicators
    - **Property 3: Progress indicators appear for long operations**
    - **Validates: Requirements 2.4, 6.2**
  
  - [x] 7.3 Write unit tests for progress bar
    - Test progress calculation
    - Test ETA calculation
    - Test TTY detection for progress display
    - _Requirements: 2.4, 6.2, 6.3_

- [x] 8. Implement enhanced error handling with user-friendly messages
  - [x] 8.1 Create ErrorHandler component
    - Format errors with clear explanations
    - Provide actionable suggestions for common errors
    - Group similar errors to reduce noise
    - Sanitize error messages to avoid leaking sensitive paths
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6_
  
  - [x] 8.2 Write property test for error message quality
    - **Property 4: Error messages are human-readable**
    - **Validates: Requirements 3.1, 3.6**
  
  - [x] 8.3 Write property test for error grouping
    - **Property 5: Similar errors are grouped**
    - **Validates: Requirements 3.5**
  
  - [x] 8.4 Write unit tests for error handling
    - Test error message formatting
    - Test suggestion generation
    - Test error grouping logic
    - Test path sanitization
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6_

- [x] 9. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 10. Implement signal handling for graceful interruption
  - [x] 10.1 Create SignalHandler component
    - Catch SIGINT (Ctrl-C) signal
    - Exit immediately on first Ctrl-C with status message
    - Skip cleanup and exit on second Ctrl-C
    - Display what cleanup was skipped
    - _Requirements: 6.4, 6.5, 12.1, 12.2, 12.3_
  
  - [x] 10.2 Write property test for interrupted operations
    - **Property 14: Interrupted operations leave recoverable state**
    - **Validates: Requirements 12.4, 12.5**
  
  - [x] 10.3 Write unit tests for signal handling
    - Test SIGINT handling
    - Test cleanup timeout
    - Test double Ctrl-C behavior
    - _Requirements: 6.4, 6.5, 12.1, 12.2, 12.3_

- [x] 11. Implement hash algorithm detection
  - [x] 11.1 Add DetectHashAlgorithm function to internal/hash
    - Validate string contains only hex characters (0-9, a-f, A-F)
    - Map lengths to algorithms: 8→CRC32, 32→MD5, 40→SHA-1, 64→SHA-256, 128→SHA-512/BLAKE2b
    - Return all matching algorithms (may be multiple for ambiguous lengths)
    - Return empty slice for invalid format
    - _Requirements: 21.1, 21.2_
  
  - [x] 11.2 Write property test for hex validation
    - **Property 27: Hash algorithm detection validates hex characters**
    - **Validates: Requirements 21.1**
  
  - [x] 11.3 Write property test for algorithm identification
    - **Property 28: Hash algorithm detection identifies correct algorithms by length**
    - **Validates: Requirements 21.2**
  
  - [x] 11.4 Write unit tests for hash detection
    - Test valid hex strings of each length
    - Test invalid hex characters
    - Test ambiguous lengths (128 chars)
    - Test edge cases (empty string, wrong length)
    - _Requirements: 21.1, 21.2, 21.5_

- [x] 12. Implement argument classification (files vs hash strings)
  - [x] 12.1 Add ClassifyArguments function to internal/config
    - Check filesystem existence FIRST (files take precedence)
    - Use DetectHashAlgorithm for non-file arguments
    - Return helpful errors when hash doesn't match current algorithm
    - Normalize hash strings to lowercase
    - _Requirements: 22.1, 22.2, 22.3, 22.4, 22.5, 22.6_
  
  - [x] 12.2 Integrate ClassifyArguments into ParseArgs
    - Call after flag parsing
    - Populate cfg.Files and cfg.Hashes separately
    - Return error if hash algorithm mismatch detected
    - _Requirements: 22.5_
  
  - [x] 12.3 Write property test for file precedence
    - **Property 29: Argument classification prioritizes files over hash strings**
    - **Validates: Requirements 22.1, 22.2**
  
  - [x] 12.4 Write property test for case normalization
    - **Property 30: Hash strings are normalized to lowercase**
    - **Validates: Requirements 22.6**
  
  - [x] 12.5 Write unit tests for argument classification
    - Test file vs hash disambiguation
    - Test algorithm mismatch errors
    - Test case normalization
    - Test non-existent files treated as files (will error later)
    - _Requirements: 22.1, 22.2, 22.3, 22.4, 22.5, 22.6_

- [x] 13. Implement config file auto-discovery
  **CRITICAL BUG DISCOVERED**: Current implementation has a fundamental flaw in precedence logic that violates the documented "flags > env vars > config files" hierarchy. See Task 47 for details and fix.
  - [x] 13.1 Add FindConfigFile function to internal/config
    - Search locations in order: ./.hashi.json, $XDG_CONFIG_HOME/hashi/config.json, ~/.config/hashi/config.json, ~/.hashi/config.json
    - Return first file found, or empty string if none
    - Handle empty XDG_CONFIG_HOME and HOME gracefully
    - _Requirements: 23.1, 23.5_
  
  - [x] 13.2 Integrate auto-discovery into ParseArgs
    - Call FindConfigFile if --config not specified
    - Load and apply config file if found
    - Maintain precedence: flags > env vars > config file > defaults
    - _Requirements: 23.2, 23.3, 23.4_
  
  - [x] 13.3 Update help text for configuration section
    - Document config file locations and search order
    - Document config file format (TOML)
    - Document all environment variables
    - _Requirements: 23.6_
  
  - [x] 13.4 Write property test for auto-discovery
    - **Property 31: Config auto-discovery finds first available config**
    - **Validates: Requirements 23.1, 23.2**
  
  - [x] 13.5 Write unit tests for config auto-discovery
    - Test priority order of config locations
    - Test with missing HOME/XDG_CONFIG_HOME
    - Test --config override
    - Test config precedence
    - **NOTE**: Tests use non-default values (e.g., sha512) to avoid the design limitation where pflag cannot distinguish between "user explicitly set flag to default value" vs "flag not set". This is acceptable since the precedence logic works correctly when flags are explicitly set.
    - _Requirements: 23.1, 23.2, 23.3, 23.4, 23.5_

- [x] 14. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 15. Implement hash string validation mode
  - [x] 15.1 Add hash validation mode to main.go
    - Detect mode: no files, only hash strings
    - Validate hash format using DetectHashAlgorithm
    - Display valid/invalid status with algorithm info
    - _Requirements: 24.1, 24.2, 24.3_
  
  - [x] 15.2 Implement output for validation mode
    - Show checkmark for valid, X for invalid
    - List possible algorithms for valid hashes
    - Show helpful error for invalid format
    - _Requirements: 24.2, 24.3, 24.4_
  
  - [x] 15.3 Write property test for validation mode
    - **Property 32: Hash validation mode reports correct algorithms**
    - **Validates: Requirements 24.2, 24.4**
  
  - [x] 15.4 Write unit tests for validation mode
    - Test valid hash strings of each algorithm
    - Test invalid hash strings
    - Test ambiguous lengths
    - Test exit codes (0 valid, 3 invalid)
    - _Requirements: 24.1, 24.2, 24.3, 24.4, 24.5_

- [x] 16. Implement file + hash comparison mode
  - [x] 16.1 Add comparison mode to main.go
    - Detect mode: one file, one hash string
    - Compute file hash and compare to provided hash
    - Display PASS/FAIL result
    - _Requirements: 25.1, 25.2, 25.3_
  
  - [x] 16.2 Add edge case handling
    - Error if multiple files with hash strings
    - Error if stdin marker with hash strings
    - Handle file not found and permission errors
    - _Requirements: 25.4, 25.5, 25.6_
  
  - [x] 16.3 Write property test for comparison mode
    - **Property 33: File+hash comparison returns correct exit codes**
    - **Validates: Requirements 25.2, 25.3**
  
  - [x] 16.4 Write unit tests for comparison mode
    - Test matching hash
    - Test mismatching hash
    - Test algorithm mismatch error
    - Test file not found
    - Test multiple files error
    - _Requirements: 25.1, 25.2, 25.3, 25.4, 25.5, 25.6_

- [x] 17. Implement boolean output flag (--bool)
  - [x] 17.1 Add BoolOutput field to Config struct
    - Add --bool / -b flag to ParseArgs
    - Implement override behavior (--bool overrides other output flags)
    - _Requirements: 26.3, 26.4_
  
  - [x] 17.2 Implement bool output in comparison mode
    - Output only "true" or "false" when --bool is set
    - Maintain correct exit codes
    - _Requirements: 26.1, 26.2_
  
  - [x] 17.3 Update help text for --bool flag
    - Add to OUTPUT FORMATS section
    - Document override behavior
    - _Requirements: 26.5_
  
  - [x] 17.4 Write property test for bool output
    - **Property 34: Bool output produces only true/false**
    - **Validates: Requirements 26.1**
  
  - [x] 17.5 Write unit tests for bool output
    - Test true output on match
    - Test false output on mismatch
    - Test override behavior (--bool overrides --quiet, --verbose, --json, --plain)
    - _Requirements: 26.1, 26.2, 26.3, 26.4_

- [x] 18. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 19. Implement file output manager with safety features
  - [x] 19.1 Create FileOutputManager component
    - Check if file exists before writing
    - Prompt for confirmation unless --force
    - Support append mode with --append
    - Implement atomic writes (temp file + rename)
    - Handle JSON log append to maintain validity
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_
  
  - [x] 19.2 Write property test for append mode
    - **Property 12: Append mode preserves existing content**
    - **Validates: Requirements 8.3**
  
  - [x] 19.3 Write property test for JSON log append
    - **Property 13: JSON log append maintains validity**
    - **Validates: Requirements 8.5**
  
  - [x] 19.4 Write unit tests for file output manager
    - Test overwrite protection
    - Test force flag
    - Test append mode
    - Test atomic writes
    - Test JSON log append
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [x] 20. Implement exit code handler with scripting support
  - [x] 20.1 Create ExitCodeHandler component
    - Define exit code constants
    - Implement DetermineExitCode logic
    - Support --match-required flag
    - [x] Support --no-match-required flag (Note: Eliminated in Task 3.5.1, following that decision)
    - Handle error cases with specific codes
    - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.5, 13.6, 13.7, 13.8, 13.9_
  
  - [x] 20.2 Write property test for exit codes
    - **Property 16: Exit codes reflect processing status**
    - **Validates: Requirements 13.1, 13.2, 13.3, 13.4, 13.9**
  
  - [x] 20.3 Write property test for match-required flag
    - **Property 17: Match-required flag controls exit code**
    - **Validates: Requirements 13.5, 13.6**
  
  - [x] 20.4 Write property test for no-match-required flag (Note: Eliminated in Task 3.5.1)
  
  - [x] 20.5 Write unit tests for exit code handler
    - Test exit code determination for various scenarios
    - Test match-required logic
    - Test no-match-required logic (Note: Eliminated in Task 3.5.1)
    - Test error code mapping
    - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.5, 13.6, 13.7, 13.8, 13.9_

- [x] 21. Implement quiet mode for boolean scripting
  - [x] 21.1 Add quiet mode support to output system
    - Suppress all stdout when --quiet is set
    - Still output errors to stderr
    - Work with all exit code flags
    - _Requirements: 2.8_
  
  - [x] 21.2 Write property test for quiet mode
    - **Property 21: Quiet mode suppresses stdout**
    - **Validates: Requirements 2.8**
  
  - [x] 21.3 Write unit tests for quiet mode
    - Test stdout suppression
    - Test stderr still works
    - Test interaction with exit codes
    - _Requirements: 2.8_

- [x] 22. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 23. Implement security features
  - [x] 23.1 Add path validation
    - Validate all file paths
    - Prevent directory traversal attacks
    - Resolve to absolute paths safely
    - _Requirements: 6.1_
  
  - [x] 23.2 Add input validation
    - Validate all user inputs before processing
    - Check hash format validity
    - Validate size limits (Handled in config validation)
    - Validate date formats (Handled in config validation)
    - _Requirements: 6.1_
  
  - [x] 23.3 Add security integration testing
    - Test configurable blacklist/whitelist system in broader project context
    - Verify security patterns work with all output modes (--output, --log-file, --log-json)
    - Test environment variable and config file integration
    - Validate security error messages in different verbosity modes
    - _Requirements: 6.1_
  
  - [x] 23.4 Write property test for input validation
    - **Property 8: Input validation occurs before processing**
    - **Validates: Requirements 6.1**
  
  - [x] 23.5 Write unit tests for security features
    - Test path validation
    - Test input validation
    - Test error message sanitization (Handled in errors package)
    - Test resource limits (Handled in config)
    - _Requirements: 6.1_

- [x] 24. Implement default behavior (no arguments = current directory)
  - [x] 24.1 Add default file discovery
    - When no arguments provided, scan current directory
    - Respect --recursive flag for subdirectories
    - Respect --hidden flag for hidden files
    - Apply filtering flags (include, exclude, size, dates) (Basic discovery logic implemented)
    - _Requirements: 1.1_
  
  - [x] 24.2 Write property test for default behavior
    - **Property 1: Default behavior processes current directory**
    - **Validates: Requirements 1.1**
  
  - [x] 24.3 Write unit tests for default behavior
    - Test current directory scanning
    - Test recursive scanning
    - Test hidden file handling
    - Test filtering
    - _Requirements: 1.1_

- [x] 25. Implement enhanced help system
  - [x] 25.1 Rewrite help text with formatting
    - Use bold headings for sections
    - Lead with examples
    - Show common flags first
    - Include web documentation link
    - Format for readability
    - Include hash verification and config sections
    - _Requirements: 1.2, 1.3, 1.4, 1.5, 9.1, 9.2, 9.3, 23.6_
  
  - [x] 25.2 Write unit tests for help system
    - Test help flag variants (-h, --help, help)
    - Test help text contains examples
    - Test help text formatting
    - Test help text includes web link
    - _Requirements: 1.2, 1.3, 1.4, 1.5, 9.1, 9.2, 9.3_

- [x] 26. Implement command suggestion for common mistakes
  - [x] 26.1 Add typo detection and suggestions
    - Detect common typos in flags
    - Suggest corrections using edit distance
    - Provide helpful error messages
    - _Requirements: 1.6, 3.3_
  
  - [x] 26.2 Write unit tests for command suggestions
    - Test typo detection
    - Test suggestion generation
    - Test error message quality
    - _Requirements: 1.6, 3.3_

- [x] 27. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 28. Implement idempotence and recovery
  - [x] 28.1 Ensure operations are idempotent
    - Verify running same command twice gives same results
    - Ensure no side effects from repeated runs
    - Support resuming after interruption (Concepts considered, Manifest system Task 41 will enhance)
    - _Requirements: 6.6, 12.4, 12.5_
  
  - [x] 28.2 Write property test for idempotence
    - **Property 9: Operations are idempotent**
    - **Validates: Requirements 6.6**
  
  - [x] 28.3 Write unit tests for idempotence
    - Test repeated execution
    - Test recovery after interruption
    - Test state consistency
    - _Requirements: 6.6, 12.4, 12.5_

- [x] 29. Implement flag abbreviation rejection
  - [x] 29.1 Add validation to reject undefined abbreviations
    - Only accept explicitly defined short flags
    - Reject arbitrary abbreviations
    - Provide clear error messages (Standard pflag behavior verified)
    - _Requirements: 11.5_
  
  - [x] 29.2 Write property test for abbreviation rejection
    - **Property 15: Abbreviated flags are rejected**
    - **Validates: Requirements 11.5**
  
  - [x] 29.3 Write unit tests for abbreviation handling
    - Test defined short flags work
    - Test undefined abbreviations are rejected
    - Test error messages
    - _Requirements: 11.5_

- [x] 30. Integration and wiring
  - [x] 30.1 Wire all components together in main.go
    - Initialize all components
    - Set up signal handling
    - Configure output formatters
    - Set up error handling
    - Implement main execution flow
    - Implement mode selection (validation, comparison, standard)
    - _Requirements: All_
  
  - [x] 30.2 Update main function to use new components
    - Parse arguments with new parser
    - Auto-discover and load configuration
    - Classify arguments into files and hashes
    - Select operation mode based on arguments
    - Process files with progress indicators
    - Format output based on flags
    - Handle errors gracefully
    - Return appropriate exit codes
    - _Requirements: All_
  
  - [x] 30.3 End-to-end security testing with compiled binary
    - Compile hashi binary and test security features in real environment
    - Test environment variable configuration with actual binary
    - Test config file loading with actual binary
    - Verify security error messages in production environment
    - Test glob pattern matching with real file system
    - Validate whitelist override behavior with actual files
    - _Requirements: 6.1_
  
  - [x] 30.4 Write integration tests
    - Test end-to-end workflows
    - Test all output formats
    - Test error scenarios
    - Test signal interruption
    - Test file output
    - Test quiet mode with exit codes
    - Test hash validation mode
    - Test file+hash comparison mode
    - Test config auto-discovery
    - _Requirements: All_

- [x] 31. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 32. Documentation and polish
  - [ ] 32.1 Update README with new features
    - Document all new flags including --bool
    - Document hash verification mode
    - Document file+hash comparison mode
    - Document config auto-discovery
    - Provide usage examples
    - Add security section with integrity vs authenticity explanation
    - Add installation/uninstallation instructions
    - Add troubleshooting section with --verbose flag usage
    - _Requirements: 9.5, 10.3, 20.2, 20.4_
  
  - [ ] 32.2 Create comprehensive user documentation
    - Create docs/user/getting-started.md - Installation and first steps
    - Create docs/user/command-reference.md - Complete flag reference
    - Create docs/user/error-handling.md - Error messages and troubleshooting guide
    - Create docs/user/configuration.md - Config files and environment variables
    - Create docs/user/output-formats.md - Understanding different output formats
    - Document error handling system: validation errors (specific), generic errors (security + system), verbose flag usage
    - Explain why some errors are generic (security) without revealing implementation details
    - Provide clear troubleshooting steps for common issues
    - _Requirements: 3.1, 3.2, 3.3, 3.6, 9.1, 9.2, 9.3_
  
  - [ ] 32.3 Create man page
    - Write comprehensive man page
    - Include all flags and examples
    - Make accessible via `hashi help`
    - _Requirements: 9.4_
  
  - [ ] 32.4 Create web documentation
    - Set up documentation website or GitHub pages
    - Include tutorials and advanced examples
    - Link from help text
    - _Requirements: 1.5, 9.3, 9.5_
  
  - [ ] 32.5 Add deprecation warning framework
    - Create system for deprecation warnings
    - Document migration paths
    - _Requirements: 11.1, 11.4_

- [x] 33. Implement archive integrity verification (ZIP CRC32)
  - [x] 33.1 Create ArchiveVerifier component
    - Implement ZIP file detection
    - Implement CRC32 verification for ZIP entries
    - Always use CRC32 regardless of metadata (security hardening)
    - Report which specific entries failed
    - _Requirements: 15.1, 15.4, 15.6, 20.3_
  
  - [x] 33.2 Implement default boolean output for ZIP verification
    - Return exit code only (no stdout) by default
    - Support --verbose for detailed output
    - Support --json for machine-readable details
    - _Requirements: 15.1, 16.1, 16.4_
  
  - [x] 33.3 Implement --raw flag for bypassing special handling
    - Treat ZIP files as raw bytes when --raw specified
    - Apply to other special file types as they're added
    - _Requirements: 15.5, 17.2_
  
  - [x] 33.4 Write property test for ZIP CRC32 verification
    - **Property 22: ZIP verification uses CRC32 only**
    - **Validates: Requirements 15.4, 20.3**
  
  - [x] 33.5 Write property test for boolean output default
    - **Property 23: ZIP verification returns boolean by default**
    - **Validates: Requirements 15.1, 16.1**
  
  - [x] 33.6 Write property test for raw flag
    - **Property 24: Raw flag bypasses special file handling**
    - **Validates: Requirements 15.5, 17.2**
  
  - [x] 33.7 Write property test for multiple ZIP verification
    - **Property 26: Multiple ZIP verification returns single boolean**
    - **Validates: Requirements 15.2**
  
  - [x] 33.8 Write unit tests for archive verification
    - Test valid ZIP verification
    - Test corrupted ZIP detection
    - Test --raw flag behavior
    - Test verbose and JSON output modes
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5, 15.6_

- [x] 34. Implement flag conflict detection
  - [x] 34.1 Create ConflictResolver component
    - Define conflict rules for mutually exclusive flags
    - Implement conflict detection logic
    - Generate clear error messages for conflicts
    - Include --bool conflicts with --quiet and --format
    - _Requirements: 17.1, 17.4, 26.3, 26.4_
  
  - [x] 34.2 Document all known conflicts
    - Output format conflicts (--json vs --plain vs --bool)
    - Verbosity conflicts (--quiet vs --verbose)
    - Match requirement conflicts
    - File handling conflicts (--raw vs default)
    - _Requirements: 17.4, 17.5_
  
  - [x] 34.3 Write property test for conflict detection
    - **Property 25: Mutually exclusive flags are rejected**
    - **Validates: Requirements 17.1**
  
  - [x] 34.4 Write unit tests for conflict resolver
    - Test each known conflict pair
    - Test valid flag combinations
    - Test error message quality
    - _Requirements: 17.1, 17.4_

- [x] 35. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 36. Educational code quality (ongoing)
  - [ ] 36.1 Add comprehensive function documentation
    - Document all exported functions with purpose and examples
    - Explain parameters and return values
    - _Requirements: 18.1_
  
  - [ ] 36.2 Add algorithm explanations
    - Comment complex algorithms step-by-step
    - Explain Go idioms when first used
    - _Requirements: 18.2, 18.3_
  
  - [ ] 36.3 Ensure consistent code style
    - Follow Go formatting conventions
    - Use clear, descriptive naming
    - _Requirements: 18.5_

- [ ] 37. Conflict testing infrastructure (moonshot)
  - [ ] 37.1 Document all anticipated conflicts before implementation
    - List all flag pairs and their interactions
    - Record design decisions in plain English
    - _Requirements: 19.1_
  
  - [ ] 37.2 Create conflict review checklist
    - Template for reviewing new flags
    - Process for updating conflict documentation
    - _Requirements: 19.2_
  
  - [ ] 37.3 Build flag combination test suite
    - Test common flag combinations
    - Test edge cases from conflict documentation
    - _Requirements: 19.3_
  
  - [ ]* 37.4 Build fuzzing tool (moonshot)
    - Generate random flag combinations
    - Generate random file inputs
    - Record and analyze unexpected behaviors
    - _Requirements: 19.4_

- [ ] 38. Final checkpoint - All features complete
  - Ensure all tests pass, ask the user if questions arise.
  - Review all documentation for completeness
  - Verify security considerations are documented

- [ ] 39. Implement filter engine for advanced filtering
  - [ ] 39.1 Create FilterEngine component
    - Implement glob pattern matching for include/exclude
    - Implement size filtering (min/max)
    - Implement date filtering (modified before/after)
    - Support multiple patterns (comma-separated or multiple flags)
    - Exclude takes precedence over include
    - _Requirements: 28.1, 28.2, 28.3, 28.4, 28.5, 28.6, 28.7, 28.8, 28.9_
  
  - [ ] 39.2 Write property test for include patterns
    - **Property 36: Include patterns filter correctly**
    - **Validates: Requirements 28.1**
  
  - [ ] 39.3 Write property test for exclude patterns
    - **Property 37: Exclude patterns filter correctly**
    - **Validates: Requirements 28.2**
  
  - [ ] 39.4 Write property test for exclude precedence
    - **Property 38: Exclude takes precedence over include**
    - **Validates: Requirements 28.8**
  
  - [ ] 39.5 Write property test for size filters
    - **Property 39: Size filters work correctly**
    - **Validates: Requirements 28.3, 28.4**
  
  - [ ] 39.6 Write property test for date filters
    - **Property 40: Date filters work correctly**
    - **Validates: Requirements 28.5, 28.6**
  
  - [ ] 39.7 Write property test for filter combination
    - **Property 41: Multiple filters combine with AND logic**
    - **Validates: Requirements 28.7**
  
  - [ ] 39.8 Write unit tests for filter engine
    - Test glob pattern matching
    - Test size filtering edge cases
    - Test date filtering edge cases
    - Test multiple pattern support
    - Test filter combination logic
    - _Requirements: 28.1, 28.2, 28.3, 28.4, 28.5, 28.6, 28.7, 28.8, 28.9_

- [ ] 40. Implement dry run and preview mode
  - [ ] 40.1 Create DryRunSystem component
    - Enumerate files without computing hashes
    - Calculate total file count and size
    - Estimate processing time based on file sizes
    - Apply all filters during enumeration
    - Display preview with filters applied
    - _Requirements: 29.1, 29.2, 29.3, 29.4, 29.5, 29.6, 29.7_
  
  - [ ] 40.2 Write property test for dry run enumeration
    - **Property 42: Dry run enumerates without hashing**
    - **Validates: Requirements 29.1**
  
  - [ ] 40.3 Write property test for dry run accuracy
    - **Property 43: Dry run shows accurate counts**
    - **Validates: Requirements 29.2, 29.3**
  
  - [ ] 40.4 Write property test for dry run filters
    - **Property 44: Dry run applies filters**
    - **Validates: Requirements 29.7**
  
  - [ ] 40.5 Write unit tests for dry run system
    - Test file enumeration
    - Test size calculation
    - Test time estimation
    - Test filter application
    - Test output formatting
    - _Requirements: 29.1, 29.2, 29.3, 29.4, 29.5, 29.6, 29.7_

- [ ] 41. Implement manifest system for incremental operations
  - [ ] 41.1 Create ManifestSystem component
    - Define JSON manifest format (version, algorithm, created, files array)
    - Implement manifest loading and validation
    - Implement manifest saving with atomic writes
    - Implement change detection (size, mtime comparison)
    - Implement GetChangedFiles function
    - _Requirements: 30.1, 30.2, 30.3, 30.4, 30.5, 30.6, 30.7_
  
  - [ ] 41.2 Write property test for change detection
    - **Property 45: Manifest detects changed files**
    - **Validates: Requirements 30.3**
  
  - [ ] 41.3 Write property test for incremental processing
    - **Property 46: Only-changed processes changed files only**
    - **Validates: Requirements 30.2**
  
  - [ ] 41.4 Write property test for manifest format
    - **Property 47: Manifest format is valid JSON**
    - **Validates: Requirements 30.7**
  
  - [ ] 41.5 Write unit tests for manifest system
    - Test manifest loading and parsing
    - Test manifest saving
    - Test change detection logic
    - Test invalid manifest handling
    - Test missing manifest handling
    - _Requirements: 30.1, 30.2, 30.3, 30.4, 30.5, 30.6, 30.7_

- [ ] 42. Enhance file output manager with atomic writes
  - [ ] 42.1 Update FileOutputManager with atomic write support
    - Implement temp file + rename pattern
    - Add write failure rollback
    - Add path validation before processing
    - Enhance JSON log append to maintain array validity
    - _Requirements: 31.1, 31.2, 31.3, 31.4, 31.5, 31.6, 31.7, 31.8_
  
  - [ ] 42.2 Write property test for atomic writes
    - **Property 48: Atomic writes preserve original on failure**
    - **Validates: Requirements 31.7**
  
  - [ ] 42.3 Write property test for append mode
    - **Property 49: Append mode preserves existing content**
    - **Validates: Requirements 31.3**
  
  - [ ] 42.4 Write property test for JSON log append
    - **Property 50: JSON log append maintains validity**
    - **Validates: Requirements 31.6**
  
  - [ ] 42.5 Write unit tests for enhanced file output
    - Test atomic write success and failure
    - Test rollback on write failure
    - Test path validation
    - Test JSON array maintenance
    - _Requirements: 31.1, 31.2, 31.3, 31.4, 31.5, 31.6, 31.7, 31.8_

- [ ] 43. Checkpoint - Ensure all new feature tests pass
  - Ensure all tests pass for tasks 39-42, ask the user if questions arise.

- [ ] 44. Integrate new features into main execution flow
  - [ ] 44.1 Update main.go to support new modes
    - Add dry run mode selection
    - Add incremental mode selection
    - Integrate filter engine into file enumeration
    - Integrate manifest system for incremental operations
    - _Requirements: 28.*, 29.*, 30.*_
  
  - [ ] 44.2 Update configuration parsing
    - Add new flags: --dry-run, --manifest, --only-changed, --output-manifest
    - Add filter flags: --include, --exclude, --min-size, --max-size, --modified-after, --modified-before
    - Update help text with new flags
    - _Requirements: 28.*, 29.*, 30.*_
  
  - [ ] 44.3 Write integration tests for new features
    - Test dry run mode end-to-end
    - Test incremental operations workflow
    - Test advanced filtering with various combinations
    - Test file output safety features
    - _Requirements: 28.*, 29.*, 30.*, 31.*_

- [ ] 45. Update documentation for new features
  - [ ] 45.1 Update README with new features
    - Document advanced filtering options
    - Document dry run mode
    - Document incremental operations
    - Provide CI/CD workflow examples
    - _Requirements: 28.*, 29.*, 30.*_
  
  - [ ] 45.2 Update help text
    - Add FILTERING section
    - Add INCREMENTAL OPERATIONS section
    - Add DRY RUN section
    - Include examples for each new feature
    - _Requirements: 28.*, 29.*, 30.*_
  
  - [ ] 45.3 Create advanced usage guide
    - Document filter pattern syntax
    - Document manifest format
    - Document CI/CD integration patterns
    - Document performance optimization tips
    - _Requirements: 28.*, 29.*, 30.*_
  
  - [ ] 45.4 Maintain error handling documentation
    - Update docs/user/error-handling.md when error messages change
    - Ensure all new error types are documented
    - Keep troubleshooting steps current with implementation
    - Update verbose flag examples when error details change
    - Maintain security balance (helpful without revealing implementation)
    - _Requirements: 3.1, 3.2, 3.3, 3.6_

- [ ] 46. Final checkpoint - All refactored features complete
  - Ensure all tests pass, ask the user if questions arise.
  - Review all documentation for completeness
  - Verify new features integrate cleanly with existing functionality
  - Test all feature combinations for conflicts

- [x] 47. **CRITICAL BUG FIX**: Fix configuration precedence violation
  **SEE DOCUMENTATION**: CRITICAL_BUG_REPORT.md
  **SEVERITY**: Critical - Breaks documented precedence hierarchy
  **IMPACT**: Users cannot override environment variables with explicit flags when flag value equals default
  
  **BUG DESCRIPTION**:
  The current ApplyEnvConfig implementation incorrectly overrides explicit flags with environment variables when the flag value equals the default. This violates the documented precedence: flags > env vars > config files.
  
  **AFFECTED FLAGS**:
  - String flags: `--algorithm=sha256`, `--format=default` 
  - Boolean flags: `--recursive=false`, `--verbose=false`, `--quiet=false`, `--hidden=false`, `--preserve-order=false`
  
  **REPRODUCTION**:
  ```bash
  export HASHI_ALGORITHM=md5
  hashi --algorithm=sha256 file.txt  # BUG: Uses md5 instead of sha256
  ```
  
  **ROOT CAUSE**:
  ApplyEnvConfig uses hardcoded default value comparisons:
  ```go
  if cfg.Algorithm == "sha256" && env.HashiAlgorithm != "" {
      cfg.Algorithm = env.HashiAlgorithm  // WRONG: Overrides explicit flag
  }
  ```
  
  - [x] 47.1 Implement pflag.Changed() based precedence detection
    - Use pflag.Changed() to detect if flag was explicitly set by user
    - Modify ApplyEnvConfig to only apply env vars when flag was NOT explicitly set
    - Maintain backward compatibility for config file precedence
    - _Requirements: 5.3 (correct precedence)_
  
  - [x] 47.2 Fix all affected configuration options
    - Fix Algorithm precedence (string flag)
    - Fix OutputFormat precedence (string flag) 
    - Fix Recursive precedence (boolean flag)
    - Fix Hidden precedence (boolean flag)
    - Fix Verbose precedence (boolean flag)
    - Fix Quiet precedence (boolean flag)
    - Fix PreserveOrder precedence (boolean flag)
    - _Requirements: 5.3 (correct precedence)_
  
  - [x] 47.3 Update ParseArgs to track flag changes
    - Modify ParseArgs to use pflag.Changed() after parsing
    - Create a mechanism to pass "changed" information to ApplyEnvConfig
    - Ensure config file precedence still works correctly
    - _Requirements: 5.3 (correct precedence)_
  
  - [x] 47.4 Write comprehensive precedence tests
    - Test explicit default flags override env vars (the bug scenario)
    - Test explicit non-default flags override env vars (should still work)
    - Test env vars override implicit defaults (should still work)
    - Test config files have lowest precedence (should still work)
    - Test all affected flags (Algorithm, OutputFormat, boolean flags)
    - _Requirements: 5.3 (correct precedence)_
  
  - [x] 47.5 Update documentation
    - Update help text to clarify precedence behavior
    - Add examples showing explicit default flag usage
    - Document the fix in CHANGELOG.md
    - _Requirements: 5.3 (correct precedence)_



- [ ]* 48. Post-release review and refinement
  - [ ]* 48.1 Research standard post-release process
    - Research industry best practices for post-release reviews
    - Document common post-release activities (user feedback, metrics, bug reports)
    - Identify applicable practices for hashi project
    - _Requirements: None (post-release planning)_
  
  - [ ]* 48.2 Collect and document ideas, adjustments, and leftovers
    - Compile all feature ideas discussed throughout development
    - Document refinements and improvements identified during implementation
    - Catalog known issues and bug reports from users
    - Organize by priority and effort estimate
    - Create roadmap for future releases
    - _Requirements: None (post-release planning)_



## Notes

- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- Integration tests verify end-to-end workflows
- Security is built in from the start, not added later
- The implementation follows the priority order: Critical → Important → Nice-to-have
- All tests are required for comprehensive quality assurance
- Tasks marked with `*` are optional moonshot goals
- Educational code quality is an ongoing effort throughout development
- Conflict testing infrastructure supports long-term maintainability
- Tasks 11-17 implement hash detection, argument classification, config auto-discovery, and new operation modes
- Tasks 39-46 implement advanced filtering, dry run mode, incremental operations, and enhanced file output safety (added from Use_cases_list.md refactoring)
