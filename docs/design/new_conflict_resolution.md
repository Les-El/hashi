1# New Conflict Resolution Strategy: The Pipeline of Intent

**Status:** Finalized Design (Jan 18, 2026)
**Context:** Replacing the brittle "Pairwise Precedence" list with a "Human-First" State Machine and Global Split Streams.

## 1. Core Philosophy: The Pipeline of Intent

Instead of defining conflicts between every possible pair of flags (N! complexity), we model flag handling as a **State Machine** (or Pipeline) that processes User Intent in three distinct phases. This approach prioritizes **Utility** and **Composition** over rigid hierarchy.

### The Pipeline Phases

1.  **Intent Collection:** Collect all flags provided by the user without immediate judgment.
2.  **State Construction:** Apply intents to a `RunState` in a specific semantic order:
    *   **Step 1: Mode** (What are we doing? e.g., `Standard`, `Boolean`)
    *   **Step 2: Format** (How does it look? e.g., `Default`, `JSON`, `Plain`)
    *   **Step 3: Verbosity** (How loud is it? e.g., `Normal`, `Quiet`, `Verbose`)
3.  **Validation:** Ensure the final state is sane.

---

## 2. Foundational Architecture: Global Split Streams

To resolve the "Split Personality" of mixed output types and ensure compatibility with shell redirection/piping, `hashi` enforces a strict **Global Split Stream** discipline.

### The Rules
1.  **stdout (Standard Output) = DATA**
    *   Contains *only* the requested result (The "Payload").
    *   Examples: Clean JSON, the list of hashes, the boolean `true`/`false`.
    *   **Guarantee:** This stream is always safe to pipe (`|`) or redirect (`>`) to a file. It is never polluted by logs.
2.  **stderr (Standard Error) = CONTEXT**
    *   Contains *everything else*.
    *   Examples: Progress bars, "Scanning..." messages, Warnings, Errors, Verbose debug logs.
    *   **Guarantee:** This stream is for the human user. It allows progress monitoring even when data is being redirected.

### The "Tee" Model for File Output
Internal file writing flags (`--output` and `--log-file`) act as "Tees" (duplicators) of these streams. They do not change *what* is generated, only *where else* it is saved.

*   **`stdout`** -> Terminal Screen (Data) **AND** `cfg.OutputFile` (if set).
*   **`stderr`** -> Terminal Screen (Context) **AND** `cfg.LogFile` (if set).

**Why this matters:**
*   It decouples "Generation" from "Destination."
*   It prevents "garbage" (logs) from corrupting data files.
*   It allows `hashi` to offer "Safe File Writing" (checked against blacklists/extensions) as a user-friendly alternative to dangerous shell redirection.

---

## 3. Conflict Resolution Decisions

### A. Quiet vs. Verbose
**Decision:** `Quiet` overrides `Verbose`.
**Resolution:** If both `-q` and `-v` are present, the application runs in **Quiet** mode.

*   **Rationale ("The Alias Defense"):** Users often alias commands to be verbose by default (e.g., `alias hashi='hashi -v'`). If a user explicitly types `hashi -q`, they are signaling a specific intent to suppress that default. Throwing an error would break scripts and aliased environments. The specific request for silence trumps the general configuration for verbosity.

### B. JSON + Verbose
**Decision:** **Split Streams** (Composition, not Conflict).
**Resolution:** The two flags are allowed to coexist.
*   **stdout:** Receives pure, syntax-valid JSON output.
*   **stderr:** Receives verbose logs.

*   **Rationale:** Enabled by the "Global Split Streams" rule. Preserves the utility of both flags. A user can pipe data to a processor while monitoring progress on screen:
    ```bash
    hashi --json --verbose > data.json
    ```

### C. Bool vs. JSON
**Decision:** `Bool` overrides `JSON`.
**Resolution:** If `--bool` is present, `--json` is ignored. The output is a simple boolean string/code.

*   **Rationale (Mode > Format):**
    *   `--bool` is a **Mode** (Predicate Check). It changes *what* the tool does (returns a status).
    *   `--json` is a **Format**. It changes *how* a report is displayed.
    *   **Principle:** "A hashi JSON always acts like a hashi JSON." We must not fragment the JSON schema by outputting a single `{"status": true}` object just to satisfy a flag combination. If the mode is Boolean, the standard JSON report schema is not applicable.

### D. JSON vs. Plain (Format vs. Format)
**Decision:** **Last One Wins**.
**Resolution:** If both `--json` and `--plain` are present, the one specified last on the command line takes effect.

*   **Rationale:** Similar to the "Alias Defense," this allows users to override a default output format set in a shell alias or configuration file. If a user has `alias hashi='hashi --json'`, they can still request plain output via `hashi --plain`. The explicit, most recent intent is honored.

---

## 4. Implementation Strategy

To implement this system, we will move away from the pairwise comparison logic in `internal/conflict`.

1.  **Refactor `internal/conflict`:** Implement a `ResolveState(flags) -> RunState` function.
    *   This function will apply the "Pipeline" logic steps defined in Section 1.
    *   It will return a finalized configuration object that strictly defines the Output Stream and Log Stream behaviors.

2.  **Update `internal/config`:**
    *   Ensure the parser tracks the *order* of flags (crucial for "Last One Wins" logic).
    *   Pass the raw flag list to the new Resolver.

3.  **Refactor Output Calls:**
    *   Audit `cmd/hashi/main.go` and `internal/output`.
    *   Ensure ALL "Result" data (hashes, json) uses `fmt.Fprintf(stdout, ...)` or `fmt.Printf(...)`.
    *   Ensure ALL "Context" data (logs, progress) uses `fmt.Fprintf(stderr, ...)` or a dedicated Logger.
    *   *Critically:* Ensure the `--output` and `--log-file` writing logic is hooked into these streams (The "Tee" implementation).