# Chexum Developer Guidance

## Coding Standards

### 1. Zero-Gap Robustness
- Every function that can fail MUST return an error.
- Errors must be wrapped with context using `fmt.Errorf("...: %w", err)`.
- Use specialized error types (e.g., `ConfigCommandError`) for errors that require specific exit codes.

### 2. Streaming over Buffering
- Avoid `os.ReadFile` or `ioutil.ReadAll` for file content.
- Use `io.Reader` and `io.Writer` interfaces.
- Use `io.Copy` for efficient data transfer.

### 3. Global Split Streams
- Data output (results) goes to `stdout`.
- Contextual information (logs, errors, progress) goes to `stderr`.
- Use `internal/console` to manage these streams.

## Error Handling

- All user-facing errors should be formatted using the `internal/errors` package.
- Exit codes are defined in `internal/config/types.go` and should be strictly followed.

## Signal Management

- Use `internal/signals` to handle `SIGINT` (Ctrl+C).
- Ensure all long-running operations check for context cancellation or signal reception.
- Clean up resources (close files, delete temporary directories) on exit.

## Testing

- Aim for >85% test coverage in all packages.
- Use `internal/testutil` for common testing helpers.
- Include property-based tests for complex logic (using `rapid` or similar).
- Run `checkpoint` before submitting PRs.
