# hashi

hi! I'm hashi

I'm an intuitive CLI tool for hashing. I hope I can help!

## AI Declaration

I was 100% written by AI. The hashi project began development on the Kiro IDE, using Claude for planning and coding. Delevlopment work was a mix of Kiro, and Gemini CLI via VS Code. 

## Features

- **Human-first design**: I aim to be intuitative and easy to use. I offer colorized output, progress. indicators on long batches, and helpful error messages
- **Multiple output formats**: You can ask me show results on the screen in a number of different ways. At the same time I can save reports in different formats to multiple file types (plain, verbose, JSON, csv, etc.)
- **Flexible input**: I try to offer a forgiving syntax without brittle flag order requirements. I also easily accepts arguemts as files, directories, or a text manafest of files and directories. 
- **Script-friendly**: I'll give your scripts meaningful exit codes, a granular quiet mode, and lots of ways to easily return a bool. 

## Installation 

### Universal Installation (Linux / Windows / MacOS)

The easiest way to get me up and running is to use the command-line installation. Make sure you have [Go](https://go.dev/dl/) installed on your system, then open your terminal and enter the following
```bash
go install github.com/Les-El/hashi/cmd/hashi@latest
```

### Installing by Building from Source

To build `hashi` locally from the source code, first ensure that you have both [Git](https://git-scm.com/install/) and [Go](https://go.dev/dl/) installed on your system.


#### On Linux and MacOS
Open your terminal application and navigate to your preferred development directory, then enter:
```bash
git clone https://github.com/Les-El/hashi.git # Clone the repository
cd hashi # Navigate to the cloned project root
go build -o hashi ./cmd/hashi # Build the executable
sudo mv hashi /usr/local/bin/ # Move the executable
```
*Note: The `sudo mv` command places the `hashi` executable in a standard system-wide location.*

#### On Windows using PowerShell
Open the PowerShell application, navigate to your preferred development directory, and enter:
```powershell
git clone https://github.com/Les-El/hashi.git # Clone the repository
cd hashi # Navigate to the cloned project root
go build -o hashi.exe ./cmd/hashi # Build the executable
Move-Item hashi.exe "C:\Program Files\hashi\" # Move hashi.exe
```

## Adding to Path
If you want to be able to call the tool from any directory by just typing `hashi`, you may need to add it to your PATH

### Adding to PATH in Linux and MacOS

#### If you used the `go install` method

The executable is usually placed in `$(go env GOPATH)/bin`. Add the following line to your shell profile (e.g., `~/.bashrc`, `~/.zshrc`):

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```
Then, reload your profile: `source ~/.bashrc` (or the appropriate file for your shell).

#### If you built `hashi` from source

If you moved the `hashi` binary to `/usr/local/bin`, it should already be in your PATH. If not, ensure `/usr/local/bin` is included in your shell's PATH variable.

### Adding to PATH in Windows

If `hashi.exe` is not recognized after installation, you will need to add its location to your system's PATH environment variable.

1.  Open the **Start Search**, type in "env", and choose **"Edit the system environment variables"**.
2.  Click the **"Environment Variables..."** button.
3.  Under **"User variables"** (for current user only) or **"System variables"** (for all users), find **"Path"**, select it, and click **"Edit..."**.
4.  Click **"New"** and add the full path to the directory where `hashi.exe` is located (e.g., `C:\Users\YourUser\go\bin` if using `go install`, or `C:\Program Files\hashi` if you moved it manually).
5.  Click **OK** on all windows and restart your terminal (Command Prompt, PowerShell, etc.) for changes to take effect.


## Uninstallation
_Please note that after uninstalling `hashi`, you may also want to remove Configuration and Log files; instructions below_

### Uninstallation after using the `go install` method above
Open the terminal and enter the following
#### On Linux and MacOS
```
rm "$(go env GOPATH)"/bin/hashi
```
#### On Windows using PowerShell or Windows Terminal
```
Remove-Item "$(go env GOPATH)\bin\hashi.exe"
```
#### On Windows using Command Prompt (cmd)
```
del "%GOPATH%\bin\hashi.exe"
```
---

### Uninstalling After Building from Source
To remove `hashi` if you installed it by moving the compiled binary to `/usr/local/bin`, open the terminal application and enter

#### On Linux and MacOS
```bash
sudo rm /usr/local/bin/hashi
```
#### On Windows PowerShell or Windows Terminal
```powershell
# Assuming hashi.exe was moved to C:\Program Files\hashi
Remove-Item "C:\Program Files\hashi\hashi.exe"
```
#### On Windows Command Prompt
```cmd
del "C:\Program Files\hashi\hashi.exe"
```

## Removing Configuration and Log

`hashi` may create configuration files or logs. To remove them after uninstalling the tool, open your terminal application and enter the following

### On Linux and MacOS
```bash
# Remove global configuration directory (XDG standard)
rm -rf ~/.config/hashi

# Remove traditional dotfile configuration
rm -rf ~/.hashi

# Remove any local project configuration
rm .hashi.toml # This removes the .hashi.toml file from your current working directory
```

### On Windows

```powershell
# Remove user-specific configuration (example path)
Remove-Item -Path "$env:APPDATA\hashi" -Recurse -Force -ErrorAction SilentlyContinue

# Remove local project configuration
Remove-Item -Path ".\.hashi.toml" -ErrorAction SilentlyContinue # This removes the .hashi.toml file from your current working directory
```
*Note: Actual paths may vary. Check `~/.config/hashi` or `~/.hashi` if using a Linux-like environment on Windows (e.g., WSL, Git Bash).*


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
| 130 | Interrupted (Ctrl-C) |

## Dependencies

- [github.com/fatih/color](https://github.com/fatih/color) - Color output
- [github.com/schollz/progressbar](https://github.com/schollz/progressbar) - Progress bars
- [github.com/spf13/pflag](https://github.com/spf13/pflag) - POSIX-compliant flag parsing
- [github.com/joho/godotenv](https://github.com/joho/godotenv) - .env file support
- [golang.org/x/term](https://pkg.go.dev/golang.org/x/term) - Terminal detection

## AI Steering
Certain phrases were used to attempt to keep AI on track. Some of them include:

* "No Lock-Out: Users must never be locked out of functionality due to design choices. Default behaviors always have escape hatches." 

* "hashi is a read-only informational tool that can only create or append to txt, json, log, and csv files."

* (In regards to configuration files and settings) "hashi can't change hashi"

* "I want to defend against a bad actor that can send any input and view any output."

* "A CLI tool designed to help automate algorithmic hashing is the natural prey of bad actors."

* (When discussing documentation) "Any competent developer should be able to pick up the project and continue work immediately."

* "Before adding a new feature we must ask ourselves, does it bring value to the user? Is it worth the added complexity? Is this something the user wants, and how do they expect it to be delivered?"



## License

MIT License - see LICENSE file for details.
