# Community TODOs: Cross-Platform Verification

The current version of `hashi` (v0.0.19) has been fully verified and signed off for **Linux**. To ensure a high-quality cross-platform experience, we need community assistance to verify installation and functionality on **Windows** and **macOS**.

## Windows Verification

- [ ] **1. Installation Testing**
    - [ ] Verify `go install github.com/Les-El/hashi/cmd/hashi@latest` works in PowerShell and CMD.
    - [ ] Verify "Building from Source" instructions in `README.md` are accurate for Windows.
    - [ ] Verify manual `PATH` addition steps for Windows are clear and correct.

- [ ] **2. Functionality Testing**
    - [ ] Verify basic hashing of files and directories.
    - [ ] Verify boolean mode (`-b`) and quiet mode (`-q`).
    - [ ] Verify environment variable precedence (`HASHI_ALGORITHM`, etc.) using `$env:VAR` syntax.

- [ ] **3. Uninstallation Testing**
    - [ ] Verify binary removal using `Remove-Item` (PowerShell) and `del` (CMD).
    - [ ] Verify removal of configuration/log data from `$env:APPDATA\hashi`.

## macOS Verification

- [ ] **1. Installation Testing**
    - [ ] Verify `go install github.com/Les-El/hashi/cmd/hashi@latest` works in Zsh/Bash.
    - [ ] Verify "Building from Source" instructions for macOS.

- [ ] **2. Functionality Testing**
    - [ ] Verify basic hashing and recursive directory scanning.
    - [ ] Verify TTY color detection works as expected in Terminal.app and iTerm2.

- [ ] **3. Uninstallation Testing**
    - [ ] Verify binary removal from `/usr/local/bin/` or `GOPATH/bin`.
    - [ ] Verify removal of `~/.config/hashi` and `~/.hashi`.

## Reporting Results

If you encounter issues on these platforms, please:
1. Run with `--verbose` to capture detailed error messages.
2. Open a GitHub Issue with the prefix `[OS-Verification]` (e.g., `[Windows-Verification] Bug in PATH instructions`).
3. Include your OS version, Go version, and the exact command that failed.
