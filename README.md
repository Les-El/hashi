# chexum User Documentation

Welcome to the chexum user documentation. This guide will help you get the most out of chexum, from basic usage to advanced scripting and CI/CD integration.

## AI Disclosure
As of v0.5.1 chexum is 100% AI written. Tools used include Kiro, Gemini CLI, Roo Code, and Antigrativy. 
chexum is not ready for production use. Please review all code and documentation carefully before use.
Security and reliability are very important to the chexum project. Please report any issues you find on the [GitHub Issues page](https://github.com/Les-El/chexum/issues).

## Overview

`chexum` is a human-first, intuitive Command Line Interface (CLI) tool for hashing, built with Go. It aims to provide a robust, script-friendly, and easy-to-use alternative to traditional hashing utilities.

## Installation

### Prerequisites
- [Go](https://go.dev/doc/install) 1.24.0 or later.

### Building from Source
To build `chexum` for your current environment:
```bash
git clone https://github.com/Les-El/chexum.git
cd chexum
go build -o chexum ./cmd/chexum
```

#### Cross-Platform Builds
You can build for other operating systems and architectures using environment variables:

**Linux (64-bit)**
```bash
GOOS=linux GOARCH=amd64 go build -o chexum ./cmd/chexum
```

**Windows (64-bit)**
```bash
GOOS=windows GOARCH=amd64 go build -o chexum.exe ./cmd/chexum
```

**macOS (Intel)**
```bash
GOOS=darwin GOARCH=amd64 go build -o chexum ./cmd/chexum
```

**macOS (Apple Silicon)**
```bash
GOOS=darwin GOARCH=arm64 go build -o chexum ./cmd/chexum
```

### Adding to PATH
To run `chexum` from anywhere, add the directory containing the binary to your system's PATH.

#### Linux / macOS
Move the binary to a directory already in your PATH (e.g., `/usr/local/bin`):
```bash
sudo mv chexum /usr/local/bin/
```
Or add the directory to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):
```bash
export PATH=$PATH:/path/to/chexum-folder
```

#### Windows
1. Open the **Start Menu**, search for "Edit the system environment variables", and select it.
2. Click the **Environment Variables** button.
3. Under **System variables**, find **Path**, select it, and click **Edit**.
4. Click **New** and add the full path to the folder where `chexum.exe` is located.
5. Click **OK** on all windows to save.

### Uninstallation
To completely remove `chexum` from your system:

1. **Delete the binary:** Remove the `chexum` (or `chexum.exe`) file from your system.
2. **Remove from PATH:** If you manually added `chexum` to your PATH, revert those changes.
3. **Delete Configuration:** Remove the configuration directories where `chexum` stores its settings:
   - **Linux / macOS:** 
     - `~/.chexum/`
     - `~/.config/chexum/`
   - **Windows:** 
     - `%USERPROFILE%\.chexum\`
     - `%USERPROFILE%\.config\chexum\`

For more detailed setup options, please refer to the **[Getting Started](docs/user/getting-started.md)** guide.

## Usage
Basic usage examples:
```bash
./chexum [files...]
./chexum -r [directory]
./chexum --json [files...]
```
For more details, see the **[Examples](docs/user/examples.md)** and **[Command Reference](docs/user/command-reference.md)**.


## Documentation Index

### Core Guides
- **[getting-started.md](docs/user/getting-started.md)** - Installation and your first hash
- **[examples.md](docs/user/examples.md)** - Common usage patterns and recipes
- **[command-reference.md](docs/user/command-reference.md)** - Complete flag and option reference
- **[flags-and-arguments.md](docs/user/flags-and-arguments.md)** - Quick lookup for flags and arguments

### Feature Deep Dives
- **[filtering.md](docs/user/filtering.md)** - Detailed guide to include/exclude patterns, size, and date filters
- **[incremental.md](docs/user/incremental.md)** - How to use manifests for high-performance hashing in CI/CD
- **[dry-run.md](docs/user/dry-run.md)** - Previewing operations and estimating time
- **[performance.md](docs/user/performance.md)** - Performance optimization and benchmarks
- **[output-formats.md](docs/user/output-formats.md)** - Understanding JSON and JSONL output
- **[csv_output.md](docs/user/csv_output.md)** - Detailed guide for CSV output and integration

### Automation and Configuration
- **[scripting.md](docs/user/scripting.md)** - Integrating chexum into bash, PowerShell, and more
- **[configuration.md](docs/user/configuration.md)** - Managing config files and environment variables (Coming soon)
- **[project_master_guide.md](docs/user/project_master_guide.md)** - Comprehensive architectural and project overview

### Troubleshooting
- **[error-handling.md](docs/user/error-handling.md)** - Troubleshooting and understanding error messages
- **[test-space-management.md](docs/user/test-space-management.md)** - Managing disk space during large test runs

## Key Features

### Human-First Design
chexum is designed to be intuitive and easy to use. Considerations include colorized output when a TTY is detected, progress bars for long-running operations, and helpful error messages. Arguments and modifiers are processed through a "pipeline of intent," making syntax more flexible and easier on the fingers.

### Security Minded
The chexum project assumes that any bad actor will have control over all inputs and be able to view all outputs, in addition to having researched the repository. Defenses include restricted write operations, sanitized inputs, file tranversal protections, non-standard character handling, 

If you choose to contribute to chexum, please consider that security is a top priority.

### Machine-Friendly
chexum was built with people in mind. Part of that is making your automation and scripting easier. The tool has output support for JSON, JSONL, and csv files. stdout and stderr are split into separate streams, exit codes are consistent, the boolean flag implies quiet, and there are advanced filtering options to limit output to only what you need.
* Freindly Neighbor Policy: chexum aims to limit its resource and respect the environment it runs in.  

### Performance at Scale
With advanced filtering and incremental hashing via manifests, chexum can handle codebases with hundreds of thousands of files efficiently.

## Getting Help

If you're ever stuck, remember:
- `chexum --help` for a quick flag reference
- `chexum --verbose` for detailed error information
- Check the [docs/user/error-handling.md](docs/user/error-handling.md) guide for common solutions

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
Third-party library licenses are documented in [THIRD_PARTY_LICENSES.md](THIRD_PARTY_LICENSES.md).
