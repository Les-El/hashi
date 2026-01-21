# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.19] - 2026-01-20

### Fixed
- Updated `hashi --help` examples to correctly reflect the use of the `--verify` flag for ZIP archive integrity checking.

## [1.0.18] - 2026-01-20

### Fixed
- Improved error messages by removing technical prefixes like `lstat` from file-not-found errors.
- Corrected exit codes: non-existent file arguments now return exit code 4 (`ExitFileNotFound`) instead of 2.

## [1.0.17] - 2026-01-20

### Fixed
- Fixed bug where configuration file settings under the `[defaults]` section were not being correctly applied.
- Aligned configuration file parsing with the documented TOML format.

## [1.0.16] - 2026-01-20

### Fixed
- Corrected configuration precedence: Environment variables now correctly override project configuration files (`.hashi.toml`).
- Improved configuration system to correctly respect explicit command-line flags even when a configuration file is present.

## [1.0.15] - 2026-01-20

### Added
- Implemented full file filtering capabilities:
    - Inclusion/Exclusion patterns via `--include` and `--exclude`.
    - File size filtering via `--min-size` and `--max-size`.
    - Date-based filtering via `--modified-after` and `--modified-before`.

### Fixed
- Fixed bug where boolean mode (`-b`) produced no output when processing standard files.
- Refined verbose output mode to automatically promote the default format to 'verbose'.

## [1.0.14] - 2026-01-20

### Added
- Added `--verify` flag for opt-in ZIP archive integrity checking. ZIP files are now processed with standard hashing by default.

## [1.0.13] - 2026-01-20

### Fixed
- Fixed bug where the stdin marker `-` was not correctly expanding file paths from standard input.
- Improved error messaging for invalid hash strings (previously showed generic security obfuscation message).

## [1.0.12] - 2026-01-20

### Fixed
- Fixed a bug that caused `hashi` to fail silently when installed via `go install` and run with no arguments in a directory.

### Changed
- Improved and clarified installation, uninstallation, and configuration instructions in `README.md`.
- Consolidated duplicate "Environment Variables" and "Security Note" sections in `README.md`.

## [1.0.11] - 2026-01-18

### Fixed
- **CRITICAL**: Fixed configuration precedence violation where explicit default flags were incorrectly overridden by environment variables
- Explicit flags now always override environment variables, even when flag value equals built-in default
- Corrected logic for file/hash classification to prevent misinterpretation of arguments

### Changed
- Refined conflict resolution logic for output format flags (`--json`, `--plain`, `--verbose`, etc.)
- Improved error message for unknown flags with "Did you mean...?" suggestions
- Reorganized `main.go` into distinct modes for clarity and robustness
- Enhanced `README.md` with more detailed sections on configuration precedence and troubleshooting

### Added
- `--match-required` flag to exit with status 0 only if matches are found
- Support for reading file paths from stdin using `-`
- `CHANGELOG.md` to track project history

## [1.0.0] - 2026-01-15

### Added
- Initial release of `hashi`
- Core hashing functionality for multiple algorithms (SHA-256, SHA-512, MD5, etc.)
- File and directory processing (recursive and non-recursive)
- Multiple output formats (default, verbose, JSON, plain)
- Configuration file support (`.hashi.toml`)
- Environment variable support
- Basic error handling and exit codes
- Colorized output with TTY detection
- Progress bar for long operations
- Security features (path validation, sensitive file protection)
- Comprehensive `README.md` with usage and installation instructions
