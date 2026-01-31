# chexum Flags and Arguments

This document provides a concise reference for all command-line flags and positional arguments supported by `chexum`.

## Usage Syntax

```bash
chexum [flags] [files or hashes...]
```

## Positional Arguments

`chexum` smartly differentiates between file paths and hash strings provided as positional arguments:

- **Files/Directories**: Any argument that exists on the filesystem as a file or directory.
- **Hashes**: Any argument that looks like a cryptographic hash (hexadecimal characters of specific lengths: 32, 40, 64, or 128 characters).
- **Stdin Marker (`-`)**: A special argument that tells `chexum` to read file paths from standard input.

## Command-Line Flags

### Core Operation

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--recursive` | `-r` | `false` | Process directories recursively |
| `--hidden` | `-H` | `false` | Include hidden files (starting with `.`) |
| `--algorithm` | `-a` | `sha256` | Hash algorithm to use (`sha256`, `sha512`, `md5`, `sha1`, `blake2b`) |
| `--jobs` | `-j` | `0` (Auto) | Number of parallel hashing jobs to run |
| `--dry-run` | | `false` | Preview files and estimate time without hashing |
| `--config` | `-c` | | Path to a custom configuration file |

### Filtering

| Flag | Short | Description |
|------|-------|-------------|
| `--include` | `-i` | Glob pattern to include (e.g., `"*.go"`) |
| `--exclude` | `-e` | Glob pattern to exclude (e.g., `"node_modules/*"`) |
| `--min-size` | | Minimum file size (e.g., `100KB`, `1MB`, `1GB`) |
| `--max-size` | | Maximum file size (e.g., `10MB`, `500MB`) |
| `--modified-after` | | Filter files modified after date (`YYYY-MM-DD`) |
| `--modified-before`| | Filter files modified before date (`YYYY-MM-DD`) |

### Output and Formatting

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--format` | `-f` | `default` | Output format (`default`, `json`, `jsonl`, `plain`, `verbose`, `csv`) |
| `--json` | | | Shortcut for `--format json` |
| `--jsonl` | | | Shortcut for `--format jsonl` |
| `--plain` | | | Shortcut for `--format plain` |
| `--csv` | | | Shortcut for `--format csv` |
| `--quiet` | `-q` | `false` | Suppress non-essential output (like progress bars) |
| `--verbose`| `-v` | `false` | Show detailed debug information and errors |
| `--bool` | `-b` | `false` | Only output `true` or `false` (identical mode) |
| `--output` | `-o` | | Write results to a specific file |
| `--append` | | `false` | Combined with `--output`, appends instead of overwriting |
| `--force` | | `false` | Overwrite existing output files without prompting |

### Incremental Hashing

| Flag | Short | Description |
|------|-------|-------------|
| `--manifest` | | Use a previously saved manifest as a baseline |
| `--only-changed` | | Only process files that differ from the manifest |
| `--output-manifest` | | Save hashing results as a structural manifest for later use |

### Miscellaneous

| Flag | Short | Description |
|------|-------|-------------|
| `--preserve-order` | | Maintain strict input/alphabetical order in output |
| `--any-match` | | Exit with 0 if at least one match was found |
| `--all-match` | | Exit with 0 only if all files match |
| `--test` | | Run project diagnostics (internal testing) |
| `--help` | `-h` | Show this help message |
| `--version` | `-V` | Show application version |

## Environment Variables

The following environment variables can also be used to configure `chexum`:

- `CHEXUM_ALGORITHM`: Sets the default algorithm (overridden by `-a`)
- `CHEXUM_FORMAT`: Sets the default output format (overridden by `-f`)
- `CHEXUM_QUIET`: If `true`, enables quiet mode
- `CHEXUM_VERBOSE`: If `true`, enables verbose mode
- `CHEXUM_JOBS`: Sets the default number of workers
- `CHEXUM_SKIP_CLEANUP`: If set, prevents periodic cleanup of temp repositories
