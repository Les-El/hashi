# Flag Conflict Matrix and Resolution Rules

This document defines how `hashi` resolves conflicting or overlapping flags. It follows the "Pipeline of Intent" model where intents are applied in phases: Mode -> Format -> Verbosity.

## 1. Mode Conflicts (Behavioral)

Modes define *what* the tool is doing. They are generally mutually exclusive because they represent fundamentally different processing logic.

| Mode A | Mode B | Resolution | Rationale |
|--------|--------|------------|-----------|
| `--verify` | `--raw` | **Hard Error** | Incompatible ways of treating a file. |
| `--verify` | Hash String | **Hard Error** | Verification checks internal integrity; hash strings check external match. |
| `--bool` | Standard | **Override** | `--bool` is a specialized mode that changes output and exit logic. |

## 2. Format Conflicts (Visual)

Formats define *how* the results are displayed on `stdout`.

| Format A | Format B | Resolution | Rationale |
|----------|----------|------------|-----------|
| `--json` | `--plain` | **Last One Wins** | Allows overriding aliased defaults. |
| `--json` | `--format=X`| **Last One Wins** | Consistent with "Last One Wins" for same-level flags. |
| `--bool` | `--json` | **Override (Warning)** | `--bool` is a Mode that dictates a specific, non-JSON output. |

## 3. Verbosity Conflicts (Logging)

Verbosity defines *how much* context is displayed on `stderr`.

| Flag A | Flag B | Resolution | Rationale |
|--------|--------|------------|-----------|
| `--quiet` | `--verbose` | **Quiet Wins (Warning)** | Specific request for silence trumps general request for loudness. |
| `--bool` | `--verbose` | **Composition** | `--bool` outputs to `stdout`, `--verbose` logs to `stderr`. They coexist via Split Streams. |
| `--bool` | `--quiet` | **Composition** | `--bool` already implies quiet on `stdout`. `--quiet` ensures silence on `stderr` too. |

## 4. Other Interactions

| Flag A | Flag B | Resolution | Rationale |
|--------|--------|------------|-----------|
| `--match-required` | `--bool` | **Composition** | `--match-required` changes the predicate logic that `--bool` reports. |
| `--match-required` | `--verify` | **Composition** | Exit 0 only if verification passes AND (if no files) ... wait. |
| `--output` | shell `>` | **Composition (Tee)** | `--output` writes to file AND terminal. Shell `>` redirects terminal output. |

## 5. Summary of Hard Errors

The following combinations are strictly prohibited and will cause `hashi` to exit with code 3 (Invalid Arguments):

1.  `--raw` AND `--verify`
2.  `--verify` AND any positional hash string (e.g., `hashi --verify file.zip 1a2b3c4d`)
3.  `--output` with an invalid file extension (must be `.txt`, `.log`, or `.json`)
4.  `--config` combined with a positional `config` argument (to prevent recursive confusion)
