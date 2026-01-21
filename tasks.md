# Implementation Plan: ZIP Verification Correction

## Overview

This implementation plan corrects the ZIP file handling behavior in hashi to align with project principles. The key change is making ZIP verification **opt-in** via the `--verify` flag, ensuring consistent file processing behavior regardless of file type.

## Tasks

- [x] 1. Add --verify flag to configuration system
  - Add `Verify bool` field to Config struct in `internal/config/config.go`
  - Add flag definition in `setupFlags()` function
  - Add environment variable support (`HASHI_VERIFY`)
  - Add TOML configuration support
  - _Requirements: 15.2, 15.8_

- [x] 1.1 Write unit tests for --verify flag parsing
  - Test flag parsing with various combinations
  - Test environment variable precedence
  - Test TOML configuration loading
  - _Requirements: 15.2_

- [x] 2. Update conflict resolution system
  - Remove obsolete `--raw` vs `--verify` conflict rule from `internal/conflict/conflict.go`
  - Update conflict matrix to reflect new flag behavior
  - Ensure `--verify` works with all output formats
  - _Requirements: 17.2_

- [x] 2.1 Write property test for conflict resolution
  - **Property 25: Mutually exclusive flags are rejected**
  - **Validates: Requirements 17.1**

- [x] 3. Implement archive verification integration in main.go
  - Replace TODO comment in `cmd/hashi/main.go` (line 140-142)
  - Add logic to detect when `--verify` flag is used
  - Route ZIP files to archive verification when `--verify` is specified
  - Route all files (including ZIP) to standard hashing by default
  - _Requirements: 15.1, 15.2, 15.3_

- [x] 3.1 Write property test for consistent file processing
  - **Property 23: ZIP files processed consistently by default**
  - **Validates: Requirements 15.1**

- [x] 3.2 Write property test for verify flag behavior
  - **Property 24: Verify flag enables ZIP verification**
  - **Validates: Requirements 15.2**

- [x] 4. Update archive verification for mixed file types
  - Modify `internal/archive/archive.go` to handle mixed ZIP/non-ZIP files
  - When `--verify` is used on non-ZIP files, process with standard hashing
  - Ensure proper error handling for unsupported file types
  - _Requirements: 15.8_

- [x] 4.1 Write property test for CRC32 algorithm enforcement
  - **Property 22: ZIP verification uses CRC32 only**
  - **Validates: Requirements 15.5, 20.3**

- [x] 4.2 Write property test for multiple ZIP verification
  - **Property 26: Multiple ZIP verification returns single boolean**
  - **Validates: Requirements 15.3**

- [x] 5. Update help text and documentation
  - Update flag descriptions in `internal/config/config.go`
  - Add `--verify` flag to help text examples
  - Update flag precedence documentation
  - _Requirements: 15.7_

- [x] 6. Checkpoint - Ensure all tests pass
  - Run full test suite including property-based tests
  - Verify integration between archive module and main application
  - Test with real ZIP files to ensure CRC32 verification works
  - Ask the user if questions arise.

- [x] 7. Update comprehensive test suite expectations
  - Update test cases 4.2.9 and 4.2.10 in `comprehensive_test_suite_2026-01-20.md`
  - Change expected behavior from auto-verification to opt-in verification
  - Update test steps to use `--verify` flag for ZIP verification
  - _Requirements: 15.1, 15.2_

- [x] 8. Preserve --raw flag as code stub
  - Keep `Raw bool` field in Config struct with comment "Reserved for future use"
  - Keep flag definition but mark as no-op in current implementation
  - Add comment explaining it's a code stub for potential future functionality
  - _Requirements: Future-proofing_

- [x] 8.1 Write unit test for --raw flag stub
  - Test that `--raw` flag is parsed but has no effect on current behavior
  - Document expected behavior for future implementation
  - _Requirements: Future-proofing_

- [x] 9. Final integration testing
  - Test `hashi file.zip` produces standard SHA-256 hash
  - Test `hashi --verify file.zip` performs CRC32 verification
  - Test mixed file types with `--verify` flag
  - Test all output formats work with verification mode
  - Verify exit codes match specification (0 for pass, 6 for fail)
  - _Requirements: 15.1, 15.2, 15.3, 15.6_

- [x] 10. Final checkpoint - Complete verification
  - Ensure all tests pass, ask the user if questions arise.
  - Verify documentation matches implementation
  - Confirm README examples work as documented

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- The `--raw` flag is preserved as a code stub for future extensibility
- Integration focuses on making ZIP verification opt-in while maintaining consistent file processing