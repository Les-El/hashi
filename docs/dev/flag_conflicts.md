# Flag Conflict Documentation

This document catalogs known flag interactions, conflicts, and resolution strategies for the `chexum` project.

## Conflict Resolution Philosophy: Pipeline of Intent

Chexum uses a three-phase resolution system to avoid "Matrix Hell":
1. **Mode Detection**: (Standard, Boolean, Validation, Comparison)
2. **Format Resolution**: (Default, JSON, JSONL, Plain, Verbose)
3. **Verbosity Filtering**: (Quiet, Normal, Verbose)

### Resolution Rules
- **Mode > Format**: If a mode implies a specific format (e.g. `--bool` forces boolean output), it overrides any explicit `--format` flags.
- **Last One Wins**: For flags of the same category (e.g. `--json` then `--plain`), the last flag provided by the user is the one honored.
- **Quiet Overrides Verbose**: If both `-q` and `-v` are present, the user's request for silence is considered more restrictive and takes precedence.

## Known Conflict Matrix

| Flag A | Flag B | Type | Resolution |
|--------|--------|------|------------|
| `--bool` | `--json/--plain` | Override | `--bool` wins; warning emitted. |
| `--bool` | `--quiet` | Synergy | `--bool` implies silence; no conflict. |
| `--json` | `--jsonl` | Sibling | Last one wins. |
| `--quiet` | `--verbose` | Logical | `--quiet` wins; warning emitted. |
| `--manifest` | `--recursive` | Synergy | Manifest filters the recursive discovery results. |
| `--only-changed`| (No Manifest) | Dependency | Warning: `--only-changed` has no effect without `--manifest`. |

## Conflict Review Checklist

When adding a new flag, answer the following:
1. **Category**: Does this flag define a new Mode, a new Format, or a level of Verbosity?
2. **Exclusivity**: Does this flag make any other flag physically impossible or logically meaningless?
3. **Override**: If this flag is provided last, should it override previous settings?
4. **Silence**: How does this flag behave under `--quiet`? (Stdout should always be empty in quiet mode).

## Automated Combination Testing

The test `internal/conflict/property_test.go` uses Property-Based Testing to ensure these rules hold true across hundreds of random combinations.

### Fuzzing Strategy (Moonshot)
We use a custom fuzzer to:
1. Generate random command strings.
2. Execute the `chexum` binary.
3. Verify that the exit code and output format match the "Last One Wins" expectation.
4. Ensure no combination of flags results in a panic or crash.
