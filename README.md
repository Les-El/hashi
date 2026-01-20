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

## Installation and Uninstallation

This section details how to install and uninstall `hashi` on different operating systems, including how to manage its presence in your system's PATH and remove any associated configuration or log files.

### Linux

#### Installation Method 1: Using `go install` (Recommended if Go is installed)

This method fetches, compiles, and installs `hashi` directly from the source repository. Ensure you have a Go development environment (Go 1.16+ recommended) set up.

```bash
go install github.com/Les-El/hashi/cmd/hashi@latest
```

#### Uninstallation Method 1 (for `go install`):

To remove `hashi` installed via `go install`:

```bash
rm "$(go env GOPATH)"/bin/hashi
```

#### Installation Method 2: Building from Source

For more control or if `go install` is not suitable, you can build `hashi` directly from its source code. Ensure you have Git and a Go development environment (Go 1.16+ recommended) set up.

```bash
git clone https://github.com/Les-El/hashi.git # Clone the repository into your preferred development directory
cd hashi # Navigate to the cloned project root
go build -o hashi ./cmd/hashi # Build the executable from the project root
sudo mv hashi /usr/local/bin/ # Move the executable from the project root
```
*Note: The `sudo mv` command places the `hashi` executable in a standard system-wide location.*

#### Uninstallation Method 2 (for building from source):

To remove `hashi` if you installed it by moving the compiled binary to `/usr/local/bin`:

```bash
sudo rm /usr/local/bin/hashi
```

#### How to add to PATH (Linux):

If `hashi` is not recognized after installation, you may need to add its location to your system's PATH.

If installed via `go install`, the executable is usually placed in `$(go env GOPATH)/bin`. Add the following line to your shell profile (e.g., `~/.bashrc`, `~/.zshrc`):

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```
Then, reload your profile: `source ~/.bashrc` (or the appropriate file for your shell).

If you moved the `hashi` binary to `/usr/local/bin`, it should already be in your PATH. If not, ensure `/usr/local/bin` is included in your shell's PATH variable.

#### Removing Configuration and Logs (Linux):

`hashi` may create configuration files or logs. To remove them:

```bash
# Remove global configuration directory (XDG standard)
rm -rf ~/.config/hashi

# Remove traditional dotfile configuration
rm -rf ~/.hashi

# Remove any local project configuration
rm .hashi.toml # This removes the .hashi.toml file from your current working directory
```

### Windows

#### Installation Method 1: Using `go install` (Recommended if Go is installed)

This method fetches, compiles, and installs `hashi.exe` directly from the source repository. Ensure you have a Go development environment (Go 1.16+ recommended) set up.

```powershell
go install github.com/Les-El/hashi/cmd/hashi@latest
```
*Note: On Windows, the executable will be named `hashi.exe`.*

#### Uninstallation Method 1 (for `go install`):

To remove `hashi.exe` installed via `go install`:

```powershell
Remove-Item "$(go env GOPATH)\bin\hashi.exe"
```
Or, if using Command Prompt:
```cmd
del "%GOPATH%\bin\hashi.exe"
```

#### Installation Method 2: Building from Source

For more control or if `go install` is not suitable, you can build `hashi` directly from its source code. Ensure you have Git and a Go development environment (Go 1.16+ recommended) set up.

```powershell
git clone https://github.com/Les-El/hashi.git # Clone the repository into your preferred development directory
cd hashi # Navigate to the cloned project root
go build -o hashi.exe ./cmd/hashi # Build the executable from the project root
# Then manually move hashi.exe to a desired location, e.g., C:\Program Files\hashi
# Move-Item hashi.exe "C:\Program Files\hashi\"
```

#### Uninstallation Method 2 (for building from source):

To remove `hashi.exe` if you installed it by manually moving the compiled binary:

```powershell
# Assuming hashi.exe was moved to C:\Program Files\hashi
Remove-Item "C:\Program Files\hashi\hashi.exe"
```
Or, if using Command Prompt:
```cmd
del "C:\Program Files\hashi\hashi.exe"
```

#### How to add to PATH (Windows):

If `hashi.exe` is not recognized after installation, you will need to add its location to your system's PATH environment variable.

1.  Open the **Start Search**, type in "env", and choose **"Edit the system environment variables"**.
2.  Click the **"Environment Variables..."** button.
3.  Under **"User variables"** (for current user only) or **"System variables"** (for all users), find **"Path"**, select it, and click **"Edit..."**.
4.  Click **"New"** and add the full path to the directory where `hashi.exe` is located (e.g., `C:\Users\YourUser\go\bin` if using `go install`, or `C:\Program Files\hashi` if you moved it manually).
5.  Click **OK** on all windows and restart your terminal (Command Prompt, PowerShell, etc.) for changes to take effect.

#### Removing Configuration and Logs (Windows):

`hashi` may create configuration files or logs. These are typically found in your user's application data directory or the current working directory.

```powershell
# Remove user-specific configuration (example path)
Remove-Item -Path "$env:APPDATA\hashi" -Recurse -Force -ErrorAction SilentlyContinue

# Remove local project configuration
Remove-Item -Path ".\.hashi.toml" -ErrorAction SilentlyContinue # This removes the .hashi.toml file from your current working directory
```
*Note: Actual paths may vary. Check `~/.config/hashi` or `~/.hashi` if using a Linux-like environment on Windows (e.g., WSL, Git Bash).*

### MacOS

#### Installation Method 1: Using `go install` (Recommended if Go is installed)

This method fetches, compiles, and installs `hashi` directly from the source repository. Ensure you have a Go development environment (Go 1.16+ recommended) set up.

```bash
go install github.com/Les-El/hashi/cmd/hashi@latest
```

#### Uninstallation Method 1 (for `go install`):

To remove `hashi` installed via `go install`:

```bash
rm "$(go env GOPATH)"/bin/hashi
```

#### Installation Method 2: Building from Source

For more control or if `go install` is not suitable, you can build `hashi` directly from its source code. Ensure you have Git and a Go development environment (Go 1.16+ recommended) set up.

```bash
git clone https://github.com/Les-El/hashi.git # Clone the repository into your preferred development directory
cd hashi # Navigate to the cloned project root
go build -o hashi ./cmd/hashi # Build the executable from the project root
sudo mv hashi /usr/local/bin/ # Move the executable from the project root
```
*Note: The `sudo mv` command places the `hashi` executable in a standard system-wide location.*

#### Uninstallation Method 2 (for building from source):

To remove `hashi` if you installed it by moving the compiled binary to `/usr/local/bin`:

```bash
sudo rm /usr/local/bin/hashi
```

#### How to add to PATH (MacOS):

If `hashi` is not recognized after installation, you may need to add its location to your system's PATH.

If installed via `go install`, the executable is usually placed in `$(go env GOPATH)/bin`. Add the following line to your shell profile (e.g., `~/.bash_profile`, `~/.zshrc`):

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```
Then, reload your profile: `source ~/.bash_profile` (or the appropriate file for your shell).

If you moved the `hashi` binary to `/usr/local/bin`, it should already be in your PATH. If not, ensure `/usr/local/bin` is included in your shell's PATH variable.

#### Removing Configuration and Logs (MacOS):

`hashi` may create configuration files or logs. To remove them:

```bash
# Remove global configuration directory (XDG standard)
rm -rf ~/.config/hashi

# Remove traditional dotfile configuration
rm -rf ~/.hashi

# Remove any local project configuration
rm .hashi.toml
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
