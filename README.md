# hashi

A command-line hash comparison tool that follows industry-standard CLI design guidelines.

## Core Principles

1. **Developer Continuity**: This project is documented so any competent developer can pick it up and continue work. Code without context is technical debt.

2. **User-First Design**: Every feature answers the question: "What functionality does the user need, and what behavior does the user expect?"

3. **No Lock-Out**: Users must never be locked out of functionality due to design choices. Default behaviors always have escape hatches (e.g., `--raw` bypasses ZIP auto-verification).

## Features

- **Human-first design**: Clear, colorized output with progress indicators
- **Multiple output formats**: Default (grouped), verbose, JSON, and plain (for scripting)
- **Flexible input**: Files, directories, stdin, and hash strings
- **Archive verification**: ZIP file integrity checking via CRC32
- **Robust error handling**: User-friendly messages with actionable suggestions
- **Script-friendly**: Meaningful exit codes and quiet mode for automation

## Installation

```bash
go install github.com/example/hashi/cmd/hashi@latest
```

Or build from source:

```bash
git clone https://github.com/example/hashi.git
cd hashi
go build -o hashi ./cmd/hashi
```

## Quick Start

```bash
# Hash all files in current directory and display matches
hashi

# Compare two files
hashi file1.txt file2.txt

# Boolean check: do files match? (outputs true/false)
hashi -b file1.txt file2.txt

# Compare a file and a hash string
hashi file.txt e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855

# Validate a hash string and detect its algorithm
hashi d41d8cd98f00b204e9800998ecf8427e

# Recursively hash a directory
hashi -r /path/to/dir

# Output as JSON
hashi --json *.txt

# Verify ZIP file integrity (CRC32)
hashi archive.zip

# Hash ZIP file as raw bytes (instead of verifying)
hashi --raw archive.zip
```

## Configuration

### Auto-Discovery

`hashi` automatically searches for a configuration file in the following locations (in priority order):

1.  `./.hashi.toml` (Project-specific)
2.  `$XDG_CONFIG_HOME/hashi/config.toml` (XDG standard)
3.  `~/.config/hashi/config.toml` (XDG fallback)
4.  `~/.hashi/config.toml` (Traditional dotfile)

### Precedence Hierarchy

Settings are applied in the following order (highest to lowest):

1.  **Command-line flags** (e.g., `--algorithm=sha512`)
2.  **Environment variables** (e.g., `HASHI_ALGORITHM=md5`)
3.  **Configuration file** (from the list above)
4.  **Built-in defaults** (SHA-256, grouped output)

*Note: Explicitly set flags always override environment variables, even if the flag value matches the default.*

## Environment Variables

| Variable | Description |
|----------|-------------|
| `NO_COLOR` | Disable color output |
| `DEBUG` | Enable debug logging |
| `HASHI_CONFIG` | Path to a specific config file |
| `HASHI_ALGORITHM` | Default algorithm (sha256, md5, sha1, sha512, blake2b) |
| `HASHI_OUTPUT_FORMAT` | Default format (default, verbose, json, plain) |
| `HASHI_RECURSIVE` | Enable recursion by default (true/false) |
| `HASHI_HIDDEN` | Include hidden files by default (true/false) |
| `HASHI_BLACKLIST_FILES` | Comma-separated patterns to block from output |

## Security and Safety

- **Read-Only**: hashi never modifies source files.
- **Path Validation**: Prevents directory traversal attacks (`..` sequences).
- **Write Protection**: Automatically blocks writing output or logs to sensitive files (like `.env` or `.ssh/`) and directories.
- **Obfuscated Errors**: Security-sensitive failures (like permission denied on a blacklisted path) return generic errors to prevent information leakage, unless `--verbose` is enabled.
- **Integrity vs. Authenticity**: ZIP verification confirms **integrity** (bits are correct) but NOT **authenticity** (proof of origin).

## Troubleshooting

If you encounter errors, use the `--verbose` flag for detailed information:

```bash
# Generic error (security protection)
$ hashi --output .env go.mod
Error: output file: Unknown write/append error

# Detailed error with --verbose
$ hashi --verbose --output .env go.mod
Error: output file: security policy violation: cannot write to configuration file: .env
```

See [docs/user/error-handling.md](docs/user/error-handling.md) for comprehensive troubleshooting guidance.

## Output Formats

### Default (grouped by hash)

Files with matching hashes are grouped together:

```
file1.txt    e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
file4.pdf    e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855

file2.doc    a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890

file3.jpg    9876543210abcdef9876543210abcdef9876543210abcdef9876543210abcdef
```

### JSON (`--json`)

```json
{
  "processed": 4,
  "duration_ms": 234,
  "match_groups": [
    {
      "hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
      "count": 2,
      "files": ["file1.txt", "file4.pdf"]
    }
  ],
  "unmatched": [
    {"file": "file2.doc", "hash": "a1b2c3d4..."},
    {"file": "file3.jpg", "hash": "9876543210..."}
  ],
  "errors": []
}
```

### Plain (`--plain`)

Tab-separated, one file per line (for grep/awk/cut):

```
file1.txt	e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
file2.doc	a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | No matches found (with `--match-required`) |
| 2 | Some files failed to process |
| 3 | Invalid arguments |
| 4 | File not found |
| 5 | Permission denied |
| 6 | Archive integrity verification failed |
| 130 | Interrupted (Ctrl-C) |

## Environment Variables

- `NO_COLOR`: Disable color output
- `DEBUG`: Enable debug logging
- `HASHI_CONFIG`: Default config file path

## Security Note

**Integrity vs. Authenticity**

hashi's ZIP verification confirms **integrity** (data not corrupted) using CRC32 checksums. This does NOT verify **authenticity** (data not tampered with). A malicious actor can craft a file with valid CRC32 checksums.

For authenticity verification, use cryptographic signatures (GPG, etc.) in addition to hash verification.

## Dependencies

- [github.com/fatih/color](https://github.com/fatih/color) - Color output
- [github.com/schollz/progressbar](https://github.com/schollz/progressbar) - Progress bars
- [github.com/spf13/pflag](https://github.com/spf13/pflag) - POSIX-compliant flag parsing
- [github.com/joho/godotenv](https://github.com/joho/godotenv) - .env file support
- [golang.org/x/term](https://pkg.go.dev/golang.org/x/term) - Terminal detection

## License

MIT License - see LICENSE file for details.
