# Contributing to hashi

We welcome contributions! To ensure `hashi` remains a high-quality, maintainable project, please adhere to the following standards.

## 🎯 Core Principles

These principles guide all design and implementation decisions:

### 1. Developer Continuity
This project must be documented so any competent developer can pick it up and continue work. This principle itself must be explicit in the project.

- Every function explains *why* it exists, not just *what* it does
- Architecture decisions are documented with rationale
- The codebase serves as a teaching tool, not just a product

### 2. User-First Design
Everything should be designed with the question: "What functionality does the user need, and what behavior does the user expect?"

- Features solve real user problems
- Defaults match user expectations
- Error messages guide users to solutions

### 3. No Lock-Out
A user must never be locked out of functionality due to our design choices.

- Every default behavior has an escape hatch
- Example: `--preserve-order` bypasses default grouping when users need input order
- When adding new "smart" defaults, always provide a flag to override them

## 📖 Documentation Standards

This project treats documentation as a first-class citizen. Code without context is technical debt.

1.  **Intent Over Implementation:**
    *   Do not just describe *what* the code does (the code shows that). Describe *why* it exists and what user problem it solves.
    *   **Bad:** "The `-r` flag sets `recursive=true`."
    *   **Good:** "The `-r` flag allows users to process entire directory trees, useful for verifying backups or bulk downloads."

2.  **Living Documentation:**
    *   If you change a feature, you **must** update the corresponding section in `hashi_features.md` and `hashi_help_screen.md` in the same commit.
    *   `README.md` is the source of truth for the project's vision.

3.  **Examples First:**
    *   When explaining a new feature, start with a CLI example. Users learn by copying commands.

## 🛠️ Development Guidelines

### 1. The "CLI Guidelines" Standard
Before proposing a UI change or a new flag, consult **`CLI_guidelines.md`** (sourced from [cli-guidelines/cli-guidelines](https://github.com/cli-guidelines/cli-guidelines)).
*   **Flag Naming:** Use standard flags (e.g., `-r` for recursive, not `-a`).
*   **Output:** Respect `NO_COLOR` and non-TTY environments.
*   **Errors:** Error messages must be helpful suggestions, not just stack traces.

### 2. Go Coding Standards
*   **Formatting:** All code must run through `gofmt`.
*   **Modularity:** Keep package separation clean:
    *   `cmd/`: CLI entry points and flag parsing.
    *   `internal/hash/`: Hashing logic.
    *   `internal/ui/`: Output formatting, colors, and progress bars.
*   **Testing:** New features require unit tests.

## 🚀 Getting Started

1.  **Clone the repo.**
2.  **Read the Specs:** Review `hashi_features.md` to understand the roadmap.
3.  **Pick a Task:** Check for `TODO` items or open issues.
4.  **Build:**
    ```bash
    go build -o hashi
    ```
