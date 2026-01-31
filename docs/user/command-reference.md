# Command Reference

This document provides a complete reference for all command-line flags available in `chexum`.

## Core Flags

### `--recursive`, `-r`
Recursively traverse directories.
- **Default**: false

### `--hidden`, `-H`
Include hidden files and directories in the analysis.
- **Default**: false

### `--algorithm`, `-a`
Specify the hashing algorithm to use (`sha256`, `sha1`, `md5`, `sha512`, `blake2b`).
- **Default**: `sha256`

### `--dry-run`
Preview the files that would be processed without actually computing any hashes. Useful for verifying include/exclude patterns and estimating workload.
- **Behavior**: Discovers files, applies filters, and displays a summary including total file count, aggregate size, and estimated hashing time.
- **Default**: false

### `--jobs`, `-j`
Number of parallel processing jobs to use.
- **0 (Auto)**: Calculate based on the **Neighborhood Policy** (default).
- **Positive Integer**: Use exactly that many workers.
- **Neighborhood Policy**: chexum aims to use most available cores while leaving headroom for system responsiveness (e.g., N-1 cores on quad-core systems, N-2 on larger systems, capped at 32).
- **Environment Variable**: `CHEXUM_JOBS`

### `--config`, `-c`
Path to a configuration file.

### `--test`
Run system diagnostics and performance checks to troubleshoot environment issues or verify hardware capabilities.
- **Checks**: OS/Architecture info, CPU count, Go version, algorithm sanity check, and detailed inspection of provided file or hash arguments.
- **Default**: false

### `--help`, `-h`
Show help text and exit.

### `--version`, `-V`
Display the version of `chexum` and exit.

## Filtering Flags

### `--include`, `-i`
Glob patterns to include. Can be specified multiple times.

### `--exclude`, `-e`
Glob patterns to exclude. Can be specified multiple times.

### `--min-size`
Minimum file size (e.g., 100KB, 1MB).

### `--max-size`
Maximum file size (e.g., 1GB).

### `--modified-after`
Only process files modified after this date (YYYY-MM-DD).

### `--modified-before`
Only process files modified before this date (YYYY-MM-DD).

## Output Control

### `--quiet`, `-q`
Suppresses all non-essential output. Only critical errors will be displayed.
- **Default**: false

### `--verbose`, `-v`
Enable verbose logging.

### `--bool`, `-b`
Output only a boolean result (`true` or `false`) indicating success or failure.
- **Default**: false

### `--format`, `-f`
Specify the output format (`default`, `verbose`, `json`, `jsonl`, `plain`, `csv`).
- **Default**: `default`

### `--json`
Shortcut for `--format json`.

### `--jsonl`
Shortcut for `--format jsonl`.

### `--plain`
Shortcut for `--format plain`.

### `--csv`
Shortcut for `--format csv`.

### `--output`, `-o`
Write output results to the specified file.

### `--append`
Append results to the specified output file instead of overwriting it.

### `--force`
Overwrite files without prompting.

### `--log-file`
File for logging context and errors.

### `--log-json`
File for JSON formatted logging.

## Advanced & Incremental Behavior

### `--preserve-order`
Ensure that file discovery and processing maintain alphabetical order.
- **Default**: false

### `--any-match`
Exit 0 if at least one match is found. Replaces the deprecated `--match-required`.
- **Default**: false

### `--all-match`
Exit 0 only if ALL provided files match a hash in the provided pool or are identical.
- **Default**: false

### `--keep-tmp`
Keep temporary files and workspaces after execution (useful for debugging).

### `--manifest`
Baseline manifest for incremental operations.

### `--only-changed`
Only process files that have changed relative to the manifest.

### `--output-manifest`
Save the results as a new manifest file.
