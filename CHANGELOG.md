# Changelog


## [1.0.1] - 2026-01-18

### Added
- **Archive Verification**: Deep ZIP file integrity checking using CRC32 (via `--verify` flag)
- **Flag Conflict Resolution**: Implemented "Pipeline of Intent" for predictable flag interactions
- **Boolean Mode**: New `--bool` flag for simple `true`/`false` output
- **Raw Mode**: New `--raw` flag to bypass special file handling

### Fixed
- **CRITICAL**: Fixed configuration precedence violation where explicit default flags were incorrectly overridden by environment variables
  - Explicit flags now always override environment variables, even when flag value equals built-in default
  - Maintains correct precedence hierarchy: flags > env vars > config files > defaults

## [1.0.0] - 2026-01-17
- Internal release — hash_machine depreciated

### Notes

- Adds core hashi utilities and CLI entrypoint
- Basic tests and examples included
