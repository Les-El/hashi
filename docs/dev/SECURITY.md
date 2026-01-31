# Security Policy

## Our Credos
1. **Bad Actor Assumption**: We assume a bad actor controls all the input, sees all the output, and has researched the repository.
2. **chexum Can't Change chexum**: chexum never, under any circumstances, modifies the files it is hashing, nor can it overwrite its own configuration or binary.
3. **Read-Only Informational Tool**: Chexum is a read-only informational tool that can only write or append to `.txt`, `.json`, `.jsonl`, `.csv`, and `.log` files.
4. **Natural Prey**: A CLI tool designed to help automate algorithmic hashing is the natural prey of bad actors. We code defensively.

## Principles

## Threat Model
Chexum identifies and mitigates the following threats:
- **Arbitrary File Overwrite**: Prevented by strict output path validation and an extension whitelist.
- **Information Leakage**: Sensitive paths are sanitized in error messages, and security policy violations are obfuscated by default.
- **Directory Traversal**: All paths are cleaned and validated to prevent escaping the intended scope.
- **Symlink Exploitation**: Chexum refuses to write output to symlinks.
- **Terminal Escape Injection**: Chexum sanitizes non-printable characters in output paths to protect the user's terminal emulator.
- **Resource Exhaustion**: Chexum uses streaming I/O to maintain a constant memory footprint regardless of file size and restricts concurrent operations via worker pools.

## Reporting a Vulnerability
If you discover a security vulnerability, please do not open a public issue. Instead, please report it via the project's security contact.
