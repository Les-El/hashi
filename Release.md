# Release v0.0.19

This release renumbers the project to v0.0.19 to reflect its pre-1.0 status and a reduction in scope. ZIP integrity verification has been removed to maintain hashi's core principles of simplicity and security.

## ✨ Features
- **Advanced Filtering**: Full implementation of file inclusion/exclusion (--include, --exclude), size-based filtering (--min-size, --max-size), and date-based filtering (--modified-after, --modified-before).
- **Input Flexibility**: The stdin marker - is now fully supported for reading lists of file paths from standard input.
- **Cross-Compiled Binaries**: Pre-built binaries are now available for Linux, Windows, and macOS across both amd64 and arm64 architectures.

## 🐛 Bug Fixes
- **Stdin expansion**: Fixed a bug where - was treated as a literal filename instead of an input stream marker.
- **Configuration Precedence**: Environment variables now correctly override project-specific .hashi.toml files.
- **Error Obfuscation**: Improved error messaging for invalid hash strings, removing unnecessary security obfuscation for non-sensitive input validation.
- **Exit Codes**: Corrected exit codes for discovery failures (now returns 4 for file-not-found).
- **TOML Support**: Fixed an issue where settings under the [defaults] section in TOML config files were ignored.

## 📦 Available Binaries (dist/)

| Platform | Architecture | Filename |
| :--- | :--- | :--- |
| **Linux** | amd64 | `hashi-linux-amd64` |
| **Linux** | arm64 | `hashi-linux-arm64` |
| **Windows** | amd64 | `hashi-windows-amd64.exe` |
| **Windows** | arm64 | `hashi-windows-arm64.exe` |
| **macOS** | amd64 | `hashi-darwin-amd64` |
| **macOS** | arm64 | `hashi-darwin-arm64` |

## 🚀 Installation

### Via Go (Recommended)
```bash
GOPROXY=direct go install github.com/Les-El/hashi/cmd/hashi@latest
```

### Via Binary Download
1. Download the appropriate binary for your system from the dist/ directory.
2. (Linux/macOS) Make the binary executable: `chmod +x hashi-<platform>-<arch>`
3. Move the binary to a directory in your system's PATH (e.g., `/usr/local/bin/hashi`).

---
*Verified and Signed Off for Linux production release on 2026-01-21.*