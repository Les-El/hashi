# Contributing to chexum

Welcome contributions to `chexum`. Contributions, criticism, and security testing are very appreciated!

## CLI Theory

[Command_Line_Interface_Guidelines.md](docs/dev/Command_Line_Interface_Guidelines.md) was foundational in designing chexum. It's a really good read if you like that sort of thing, and the UX concepts can extend past just CLI. If you are vibe coding then it is imperative to highlight and reference this guide to your AI tools so they understand the human-first design philosophy of chexum.

## How to Contribute

The general workflow for contributing is as follows:

1.  **Fork** the `chexum` repository on GitHub.
2.  **Clone** your forked repository to your local machine.
3.  **Create a new branch** for your feature or bug fix: `git checkout -b feature/your-feature-name` or `git checkout -b bugfix/issue-description`.
4.  **Make your changes**, adhering to the coding standards and guidelines below.
5.  **Write tests** for your changes.
6.  **Run all tests** to ensure nothing is broken.
7.  **Commit your changes** with a clear and descriptive commit message.
8.  **Push your branch** to your forked repository.
9.  **Open a Pull Request** to the `main` branch of the upstream `chexum` repository.

## Setting Up Your Development Environment

`chexum` is built with Go. You'll need:

*   **Go (1.24.0+):** Install the latest stable version of Go.
*   **Git:** For version control.

**Steps:**

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/Les-El/chexum.git
    cd chexum
    ```
2.  **Download dependencies:**
    ```bash
    go mod download
    ```
3.  **Editor Setup:** We recommend setting up your editor to automatically run `gofmt` on save. This ensures consistent code formatting across the project. Consider integrating Go linters (e.g., `golangci-lint`) into your workflow.

## Building the Project

### Standard Build

To build the main `chexum` executable for your current system:

```bash
go build -o chexum ./cmd/chexum
```

### Cross-Compilation (Testing different binary versions)

To help us test `chexum`'s platform agnosticism, we encourage contributors to build and test on various operating systems and architectures.

You can cross-compile `chexum` using `GOOS` and `GOARCH` environment variables. Save the binaries to the `build/` directory.

**Example: Building for Linux, Windows, and macOS (amd64 and ARM64)**

```bash
mkdir -p build

# Linux
GOOS=linux GOARCH=amd64 go build -o build/chexum-linux-amd64 ./cmd/chexum
GOOS=linux GOARCH=arm64 go build -o build/chexum-linux-arm64 ./cmd/chexum

# Windows
GOOS=windows GOARCH=amd64 go build -o build/chexum-windows-amd64.exe ./cmd/chexum
GOOS=windows GOARCH=arm64 go build -o build/chexum-windows-arm64.exe ./cmd/chexum

# macOS
GOOS=darwin GOARCH=amd64 go build -o build/chexum-darwin-amd64 ./cmd/chexum
GOOS=darwin GOARCH=arm64 go build -o build/chexum-darwin-arm64 ./cmd/chexum
```

## Running Tests

It's crucial that all tests pass before submitting a Pull Request.

*   **Run all tests:**
    ```bash
    go test ./...
    ```
*   **Run with verbose output:**
    ```bash
    go test -v ./...
    ```
*   **Run specific tests:**
    ```bash
    go test ./path/to/package -run TestFunctionName
    ```
*   **Run with race detector:**
    ```bash
    go test -race ./...
    ```

## Coding Standards and Guidelines

Adhering to these standards helps maintain code quality and consistency.

*   **Formatting**: Always run `gofmt` on your code.
*   **Modularity**: Keep CLI entry points in `cmd/` and core logic in `internal/`.
*   **Documentation**:
    *   All exported functions, types, and variables should have clear Godoc comments.
    *   Update relevant user documentation (`docs/user/`) and developer documentation (`docs/dev/`) for any new features or changes.
    *   Ensure help text (`chexum --help`) is updated for new flags or commands.
    *   Focus on *why* something is done, not just *what*.
*   **Human-First Design**: `chexum` prioritizes intuitive user experience. Ensure:
    *   Clear and helpful error messages.
    *   Sensible defaults and flag behaviors.
    *   Appropriate use of colorized output (where a TTY is detected).
*   **Platform Agnosticism**: Contributions must work correctly and consistently across Linux, Windows, and macOS. Avoid platform-specific assumptions unless absolutely necessary and properly abstracted.
*   **Security First**:
    *   Be mindful of potential security implications of your changes.
    *   Avoid introducing known vulnerabilities (e.g., command injection).
    *   Ensure proper input validation and error handling.
    *   `chexum` is read-only for source files; maintain this principle.
*   **Testing**: New features and bug fixes *must* be accompanied by unit tests. Integration tests are also highly encouraged.

## Licensing and Dependencies

`chexum` is licensed under the MIT License. To maintain compliance and security, please follow these rules when adding new dependencies:

1.  **Research the License:** Before adding any new package or import, research and confirm its license.
2.  **Permitted Licenses:** We generally accept dependencies under permissive licenses such as MIT, Apache 2.0, BSD, or similar. Avoid GPL or other restrictive licenses without explicit maintainer approval.
3.  **Update Documentation:** If a new dependency is approved and added, you MUST:
    *   Add the dependency and its license type to `THIRD_PARTY_LICENSES.md`.
    *   Ensure any required attribution is included as per the dependency's license.
4.  **AI Code Attribution:** If you use AI tools (like Gemini, ChatGPT, etc.) to generate code, ensure that the code does not include snippets from other projects without proper acknowledgment or permission. If you use a significant snippet that requires attribution, you must provide it.

When submitting a Pull Request, please ensure:

*   **One logical change per PR**: If your PR addresses multiple unrelated issues, please split them into separate PRs.
*   **Descriptive title and body**:
    *   The title should briefly summarize the change.
    *   The body should explain *why* the change was made, *what* problem it solves, and *how* it was implemented.
    *   Reference any related GitHub issues (e.g., `Fixes #123`).
*   **Automated checks pass**: Your PR should pass all CI/CD checks (linting, testing, etc.).
*   **Self-review your code**: Before submitting, review your own changes. Does it meet the coding standards? Is it clear? Is it efficient?
*   **User Experience (UX) Impact**: Clearly describe the impact of your changes on the user experience, both before and after.
*   **Breaking Changes**: If your PR introduces breaking changes, clearly document them and explain the migration path.

## Reporting Bugs

If you find a bug, please open an issue on our [GitHub Issue Tracker](https://github.com/Les-El/chexum/issues). When reporting a bug, please include:

*   **Steps to reproduce**: Clear, concise instructions.
*   **Expected behavior**.
*   **Actual behavior**.
*   **Environment details**: Your OS, `chexum` version, Go version.
*   **Any relevant error messages or logs**.

## Suggesting Enhancements

If you have an idea for a new feature or an improvement, please open an issue on our [GitHub Issue Tracker](https://github.com/Les-El/chexum/issues) as well. Describe:

*   The problem you're trying to solve.
*   Your proposed solution.
*   Any alternative solutions you've considered.
*   Use cases where this enhancement would be beneficial.

## Reporting Security Vulnerabilities

If you discover a security vulnerability, please do **not** open a public issue. Instead, please refer to our security policy (if one exists, otherwise contact project maintainers directly via email if an address is provided, or through GitHub's private vulnerability reporting feature).

Thank you for contributing to `chexum`!